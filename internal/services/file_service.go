package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"backend/internal/db"
	"backend/internal/storage"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

// FileService orchestrates dedup logic, DB updates and calls to storage.
type FileService struct {
	db      *gorm.DB
	storage *storage.MinioClient
}

// NewFileService
func NewFileService(dbConn *gorm.DB, st *storage.MinioClient) *FileService {
	return &FileService{db: dbConn, storage: st}
}

// ProcessUpload:
//   - userID: uploader's ID
//   - filename: original filename (for metadata); storage uses hash objectKey
//   - tmpFilePath: path to temporary file on disk (handler should remove when finished)
//   - size: file size in bytes
//   - mimeType: detected mime
//   - hash: sha256 hex string
//
// Returns created/existing file ID (db.File.ID) on success.
func (s *FileService) ProcessUpload(ctx context.Context, userID uuid.UUID, filename, tmpFilePath string, size int64, mimeType, hash string) (string, error) {
	// open temp file for upload
	f, err := os.Open(tmpFilePath)
	if err != nil {
		return "", fmt.Errorf("open tmp: %w", err)
	}
	defer f.Close()

	// object key strategy: use hash so uploads are idempotent
	objectKey := hash

	// 1) Try to find existing file by hash
	var existing db.File
	err = s.db.Where("hash = ?", hash).First(&existing).Error
	if err == nil {
		// file exists, check if user already linked to it
		var userFile db.UserFile
		errUF := s.db.Where("user_id = ? AND file_id = ?", userID, existing.ID).First(&userFile).Error
		if errUF == nil {
			// user already has this file linked — nothing to do
			return existing.ID.String(), nil
		}

		// Create a link and increment ref_count and user's used_storage in a transaction
		tx := s.db.Begin()
		if err := tx.Model(&existing).Update("ref_count", gorm.Expr("ref_count + ?", 1)).Error; err != nil {
			tx.Rollback()
			return "", fmt.Errorf("update ref_count: %w", err)
		}
		uf := db.UserFile{
			UserID:  userID,
			FileID:  existing.ID,
			IsOwner: false,
		}
		if err := tx.Create(&uf).Error; err != nil {
			tx.Rollback()
			return "", fmt.Errorf("create user_file: %w", err)
		}
		// increase user's used storage
		if err := tx.Model(&db.User{}).Where("id = ?", userID).Update("used_storage", gorm.Expr("used_storage + ?", existing.Size)).Error; err != nil {
			tx.Rollback()
			return "", fmt.Errorf("update user storage: %w", err)
		}
		if err := tx.Commit().Error; err != nil {
			return "", err
		}
		return existing.ID.String(), nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// some other DB error
		return "", fmt.Errorf("db find file: %w", err)
	}

	// 2) File not found in DB -> upload to MinIO
	// Use the temp file reader (seek to beginning)
	if _, err := f.Seek(0, 0); err != nil {
		return "", fmt.Errorf("seek tmp: %w", err)
	}

	_, err = s.storage.Upload(ctx, objectKey, mimeType, f, size)
	if err != nil {
		// upload failed
		return "", fmt.Errorf("minio upload: %w", err)
	}

	// 3) Create DB records in transaction. Handle potential race (unique hash) gracefully.
	tx := s.db.Begin()

	newFile := db.File{
		Hash:       hash,
		ObjectName: objectKey,
		Size:       size,
		MimeType:   mimeType,
		RefCount:   1,
	}

	// Attempt to create new file row
	if err := tx.Create(&newFile).Error; err != nil {
		// If unique constraint violation happened (race), then another process created the file concurrently.
		// Fall back: find existing and create user_file + increment ref_count.
		// To detect unique constraint we try to look for existing file by hash.
		if isUniqueConstraintErr(err) {
			tx.Rollback()
			var existing2 db.File
			if err2 := s.db.Where("hash = ?", hash).First(&existing2).Error; err2 != nil {
				return "", fmt.Errorf("concurrent create: find existing failed: %w", err2)
			}
			// same flow as earlier for existing file
			tx2 := s.db.Begin()
			if err := tx2.Model(&existing2).Update("ref_count", gorm.Expr("ref_count + ?", 1)).Error; err != nil {
				tx2.Rollback()
				return "", fmt.Errorf("update ref_count (concurrent): %w", err)
			}
			uf := db.UserFile{UserID: userID, FileID: existing2.ID, IsOwner: false}
			if err := tx2.Create(&uf).Error; err != nil {
				tx2.Rollback()
				return "", fmt.Errorf("create user_file (concurrent): %w", err)
			}
			if err := tx2.Model(&db.User{}).Where("id = ?", userID).Update("used_storage", gorm.Expr("used_storage + ?", existing2.Size)).Error; err != nil {
				tx2.Rollback()
				return "", fmt.Errorf("update user storage (concurrent): %w", err)
			}
			if err := tx2.Commit().Error; err != nil {
				return "", err
			}
			return existing2.ID.String(), nil
		}
		tx.Rollback()
		return "", fmt.Errorf("create file record: %w", err)
	}

	// create user_file (owner=true)
	uf := db.UserFile{
		UserID:  userID,
		FileID:  newFile.ID,
		IsOwner: true,
	}
	if err := tx.Create(&uf).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("create user_file: %w", err)
	}

	// update user's used_storage
	if err := tx.Model(&db.User{}).Where("id = ?", userID).Update("used_storage", gorm.Expr("used_storage + ?", newFile.Size)).Error; err != nil {
		tx.Rollback()
		return "", fmt.Errorf("update user used_storage: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return "", err
	}

	return newFile.ID.String(), nil
}

// isUniqueConstraintErr tries to detect a unique-violation during insertion.
// Implemented conservatively: checks common Postgres error signatures.
func isUniqueConstraintErr(err error) bool {
	if err == nil {
		return false
	}
	// GORM wraps driver errors; check for classic Postgres duplicate key text
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "duplicate key") || strings.Contains(lower, "unique constraint") || strings.Contains(lower, "pq: duplicate key") {
		return true
	}
	return false
}

// List all files for user
func ListUserFiles(userID uuid.UUID) ([]db.UserFile, error) {
	var userFiles []db.UserFile
	err := db.DB.Preload("File").Where("user_id = ?", userID).Find(&userFiles).Error
	return userFiles, err
}

// Delete a user's file reference
func DeleteUserFile(userID uuid.UUID, fileID uuid.UUID) error {
	var userFile db.UserFile
	err := db.DB.Where("user_id = ? AND file_id = ?", userID, fileID).First(&userFile).Error
	if err != nil {
		return errors.New("file not found or not owned")
	}

	//only owner can delete
	if !userFile.IsOwner {
		return errors.New("not allowed only owner can delete")
	}

	//Delete the reference
	err = db.DB.Delete(&userFile).Error
	if err != nil {
		return err
	}

	//Decrement the ref count
	var file db.File
	if err := db.DB.First(&file, "id = ?", fileID).Error; err != nil {
		return err
	}

	file.RefCount -= 1
	if file.RefCount <= 0 {
		//Delete the file record and miniIO object
		if err := db.DB.Delete(&file).Error; err != nil {
			return err
		}
		// Later --- todo --- minioDeletion storage.DeleteObject on  file.ObjectName
	} else {
		db.DB.Save(&file)
	}
	return nil
}

package services

import (
	"backend/internal/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FolderService struct {
	db *gorm.DB
}

func NewFolderService(dbConn *gorm.DB) *FolderService {
	return &FolderService{db: dbConn}
}

func (s *FolderService) CreateFolder(ownerID uuid.UUID, name string, parentID *uuid.UUID) (*db.Folder, error) {
	f := &db.Folder{
		Name:     name,
		ParentID: parentID,
		OwnerID:  ownerID,
	}
	if err := s.db.Create(f).Error; err != nil {
		return nil, err
	}
	return f, nil
}

func (s *FolderService) ListUserFolders(ownerID uuid.UUID, parentID *uuid.UUID) ([]db.Folder, error) {
	var folders []db.Folder
	q := s.db.Where("owner_id = ?", ownerID)
	if parentID != nil {
		q = q.Where("parent_id = ?", *parentID)
	} else {
		q = q.Where("parent_id IS NULL")
	}
	err := q.Find(&folders).Error
	return folders, err
}

func (s *FolderService) DeleteFolder(ownerID, folderID uuid.UUID) error {
	// Only owner can delete
	var folder db.Folder
	if err := s.db.First(&folder, "id = ? AND owner_id = ?", folderID, ownerID).Error; err != nil {
		return err
	}
	return s.db.Delete(&folder).Error
}

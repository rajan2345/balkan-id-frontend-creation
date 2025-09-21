package services

import (
	"errors"

	"backend/internal/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShareService struct {
	db *gorm.DB
}

func NewShareService(dbConn *gorm.DB) *ShareService {
	return &ShareService{db: dbConn}
}

// Create a new Share (public or specified user)
func (s *ShareService) CreateShare(userID, fileID uuid.UUID, isPublic bool, sharedWith *string) (*db.Share, error) {
	//Verify ownership
	var uf db.UserFile
	if err := s.db.Where("user_id = ? AND file_id = ? AND is_owner = true", userID, fileID).First(&uf).Error; err != nil {
		return nil, errors.New("not allowed only owner have right to share")
	}

	share := db.Share{
		FileID:     fileID,
		UserID:     userID,
		IsPublic:   isPublic,
		SharedWith: sharedWith,
	}
	if err := s.db.Create(&share).Error; err != nil {
		return nil, err
	}
	return &share, nil
}

// Get a share by ID
func (s *ShareService) GetShare(shareID uuid.UUID) (*db.Share, error) {
	var share db.Share
	if err := s.db.Preload("File").First(&share, "id = ?", shareID).Error; err != nil {
		return nil, err
	}
	return &share, nil
}

// Increment download counter
func (s *ShareService) IncrementDownloads(shareID uuid.UUID) error {
	return s.db.Model(&db.Share{}).Where("id = ?", shareID).
		Update("downloads", gorm.Expr("downloads + 1")).Error
}

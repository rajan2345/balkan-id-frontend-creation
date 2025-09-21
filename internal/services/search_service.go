package services

import (
	"backend/internal/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SearchService struct {
	db *gorm.DB
}

func NewSearchService(dbConn *gorm.DB) *SearchService {
	return &SearchService{db: dbConn}
}

// Search files by name substring (case-insensitive) for a specific user
func (s *SearchService) SearchFiles(userID uuid.UUID, query string) ([]db.File, error) {
	var files []db.File
	err := s.db.Joins("JOIN user_files uf ON uf.file_id = files.id").
		Where("uf.user_id = ? AND files.filename ILIKE ?", userID, "%"+query+"%").Find(&files).Error

	return files, err
}

// Filter files by mime-type or Size range
func (s *SearchService) FilterFiles(userID uuid.UUID, mime *string, minSize, maxSize *int64) ([]db.File, error) {
	q := s.db.Joins("JOIN user_files uf ON uf.file_id = files.id").Where("uf.user_id = ?", userID)

	if mime != nil {
		q = q.Where("files.mime_type = ?", *mime)
	}
	if minSize != nil {
		q = q.Where("files.size >= ?", *minSize)
	}
	if maxSize != nil {
		q = q.Where("files.size <= ?", *maxSize)
	}

	var files []db.File
	err := q.Find(&files).Error
	return files, err
}

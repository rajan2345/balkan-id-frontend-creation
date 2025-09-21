package services

import (
	"backend/internal/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StatsService struct {
	db *gorm.DB
}

func NewStatsService(dbConn *gorm.DB) *StatsService {
	return &StatsService{db: dbConn}
}

// Per user usage stats
type UserStats struct {
	TotalFiles int64 `json:"total_files"`
	TotalSize  int64 `json:"total_size"`
	Quota      int64 `json:"quota"`
	Used       int64 `json:"used"`
}

func (s *StatsService) GetUserStats(userID uuid.UUID) (*UserStats, error) {
	var totalFiles int64
	var totalSize int64

	err := s.db.Model(&db.UserFile{}).Where("user_id = ?", userID).Count(&totalFiles).Error
	if err != nil {
		return nil, err
	}

	err = s.db.Joins("JOIN files f ON f.id = user_files.file_id").
		Model(&db.UserFile{}).
		Where("user_files.user_id = ?", userID).
		Select("COALESCE(SUM(f.size), 0)").Scan(&totalSize).Error
	if err != nil {
		return nil, err
	}

	var user db.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &UserStats{
		TotalFiles: totalFiles,
		TotalSize:  totalSize,
		Quota:      user.Quota,
		Used:       user.UsedStorage,
	}, nil
}

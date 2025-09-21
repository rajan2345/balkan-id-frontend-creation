package services

import (
	"backend/internal/db"

	"gorm.io/gorm"
)

type AdminService struct {
	db *gorm.DB
}

func NewAdminService(dbConn *gorm.DB) *AdminService {
	return &AdminService{db: dbConn}
}

// List all users
func (s *AdminService) ListUsers() ([]db.User, error) {
	var users []db.User
	err := s.db.Find(&users).Error
	return users, err
}

// Delete a user and Cascade their file and folder (folder logic not implemented till now)
func (s *AdminService) DeleteUser(userID string) error {
	return s.db.Delete(&db.User{}, "id = ?", userID).Error
}

// Get total storage stats accross all users
func (s *AdminService) GetSystemStats() (map[string]interface{}, error) {
	var totalUsed int64
	var totalQuota int64

	err := s.db.Model(&db.User{}).Select("sum(used_storage)").Scan(&totalUsed).Error
	if err != nil {
		return nil, err
	}

	err = s.db.Model(&db.User{}).Select("sum(quota)").Scan(&totalQuota).Error
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_used":  totalUsed,
		"total_quota": totalQuota,
	}, nil
}

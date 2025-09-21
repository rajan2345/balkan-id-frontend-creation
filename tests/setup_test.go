package tests

import (
	// "fmt"
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/db"

	// "backend/internal/middleware"
	"backend/internal/services"
	"backend/internal/storage"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupTest creates DB conn, runs migrations and returns FileService plus created test user.
func SetupTest(t *testing.T) (*services.FileService, *db.User, *gorm.DB) {
	t.Helper()

	cfg := config.Load()

	// Use DATABASE_URL from config (default configured in config.Load())
	dsn := cfg.DatabaseURL

	// Retry connecting briefly (CI may take time for services to come up)
	var dbConn *gorm.DB
	var err error
	retries := 15
	for i := 0; i < retries; i++ {
		dbConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		t.Logf("waiting for db (%d/%d) ... %v", i+1, retries, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		t.Fatalf("failed to connect to DB: %v", err)
	}

	// AutoMigrate required models used in tests
	if err := dbConn.AutoMigrate(&db.User{}, &db.File{}, &db.UserFile{}, &db.Share{}, &db.Folder{}); err != nil {
		t.Fatalf("migrate error: %v", err)
	}

	// Save global DB reference used elsewhere
	db.DB = dbConn

	// Ensure MinIO is available (retry)
	st, err := storage.NewMinioClient(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, "files", cfg.MinioUseSSL)
	if err != nil {
		// retry a few times
		for i := 0; i < retries; i++ {
			time.Sleep(2 * time.Second)
			st, err = storage.NewMinioClient(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, "files", cfg.MinioUseSSL)
			if err == nil {
				break
			}
		}
		if err != nil {
			t.Fatalf("failed init minio: %v", err)
		}
	}

	// Create test user (fixed UUID for predictability)
	testUserID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	user := &db.User{}
	if err := dbConn.First(user, "id = ?", testUserID).Error; err != nil {
		// not found -> create
		user = &db.User{
			ID:           testUserID,
			Username:     "testuser",
			PasswordHash: "test-hash",
			Email:        "test@example.com",
			Quota:        10485760, // 10 MB
			UsedStorage:  0,
		}
		if err := dbConn.Create(user).Error; err != nil {
			t.Fatalf("create test user: %v", err)
		}
	}

	fs := services.NewFileService(dbConn, st)
	return fs, user, dbConn
}

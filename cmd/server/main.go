package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"backend/internal/api"
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/middleware"
	"backend/internal/services"
	"backend/internal/storage"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// === Load Config ===
	cfg := config.Load()

	// === Setup Database ===
	dbConn, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// AutoMigrate models (alternative: run raw migrations)
	if err := dbConn.AutoMigrate(&db.User{}, &db.File{}, &db.UserFile{}, &db.Folder{}, &db.Share{}); err != nil {
		log.Fatalf("failed to migrate DB: %v", err)
	}
	db.DB = dbConn // make global ref available

	// === Setup MinIO Storage ===
	minioClient, err := storage.NewMinioClient(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		"files",
		cfg.MinioUseSSL,
	)
	if err != nil {
		log.Fatalf("failed to init MinIO: %v", err)
	}

	// === Setup Services ===
	fileService := services.NewFileService(dbConn, minioClient)
	adminService := services.NewAdminService(dbConn)
	shareService := services.NewShareService(dbConn)
	searchService := services.NewSearchService(dbConn)
	statsService := services.NewStatsService(dbConn)

	// === Setup Router ===
	r := mux.NewRouter()

	// Middlewares
	authMw := middleware.AuthMiddleware(middleware.AuthOptions{
		JWTSecret: os.Getenv("JWT_SECRET"), // set in docker-compose
		DB:        dbConn,
	})
	quotaMw := middleware.QuotaMiddleware(dbConn)

	//Rate limiter instance
	rl := middleware.NewRateLimiter(2, 2) // new ratelimiter instance
	rateLimitMw := middleware.RateLimitMiddleware(rl)

	mwChain := func(h http.Handler) http.Handler {
		return rateLimitMw(quotaMw(authMw(h)))
	}

	// === API Routes ===
	// Upload
	r.Handle("/upload", mwChain(api.NewUploadHandler(fileService))).Methods("POST")

	// Files
	r.Handle("/files", mwChain(http.HandlerFunc(api.ListUserFiles))).Methods("GET")
	r.Handle("/files", mwChain(http.HandlerFunc(api.DeleteFileHandler))).Methods("DELETE")

	// Shares
	r.Handle("/shares", mwChain(api.NewShareHandler(shareService))).Methods("POST", "GET", "DELETE")

	// Search
	r.Handle("/search", mwChain(api.NewSearchHandler(searchService))).Methods("GET")

	// Stats
	r.Handle("/stats", mwChain(api.NewStatsHandler(statsService))).Methods("GET")

	// Admin
	r.PathPrefix("/admin/").Handler(mwChain(api.NewAdminHandler(adminService)))

	// Health check
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}).Methods("GET")

	// === Start Server ===
	addr := ":" + cfg.Port
	fmt.Printf("Server running on %s\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func initDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Example fallback update if needed
		dsn = "host=localhost user=postgres password=postgres dbname=balkanID_db port=5432 sslmode=disable"
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database %v", err)
	}

	fmt.Println("Database connected and migrated successfully")
}

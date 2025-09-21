package config

import "os"

type Config struct {
	DatabaseURL    string
	Port           string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioUseSSL    bool
}

func Load() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/balkanID_db?sslmode=disable"),
		Port:           getEnv("PORT", "8080"),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioUseSSL:    false,
	}
}
func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

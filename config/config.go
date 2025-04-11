package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort     string
	DBConnection   string
	UploadTempDir  string
	UploadFinalDir string
	MaxFileSize    int64
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioUseSSL    bool
	EnabledMinio   bool
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	return &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		DBConnection:   getEnv("DB_CONNECTION", "host=localhost user=postgres password=postgres dbname=fileuploader port=5432 sslmode=disable"),
		UploadTempDir:  getEnv("UPLOAD_TEMP_DIR", "./uploads/temp"),
		UploadFinalDir: getEnv("UPLOAD_FINAL_DIR", "./uploads/files"),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "Q3AM3TQ867SPQQA43P2F"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "zuf+tfteSls5A6y2sxDzsv8+M+3w=="),
		MinioUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
		MaxFileSize:    100 * 1024 * 1024, // 100MB default
		EnabledMinio:   getEnv("ENABLE_MINIO", "false") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

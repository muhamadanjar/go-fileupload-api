package main

import (
	"context"
	"fileupload/config"
	"fileupload/internal/delivery/http/route"
	"fileupload/internal/repository"
	"fileupload/internal/usecase"
	"fileupload/pkg/logger"
	"fileupload/pkg/minio"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	logger.Log.Info("Aplikasi dimulai")
	minio.Init(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioUseSSL)

	db, err := gorm.Open(postgres.Open(cfg.DBConnection), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal("Failed to connect to database: ", err)
	}

	db.AutoMigrate(&repository.UploadModel{}, &repository.FileModel{})

	if err := os.MkdirAll(cfg.UploadTempDir, os.ModePerm); err != nil {
		logger.Log.Fatalf("Failed to create temporary upload directory: %v", err)
	}
	if err := os.MkdirAll(cfg.UploadFinalDir, os.ModePerm); err != nil {
		logger.Log.Fatalf("Failed to create final upload directory: %v", err)
	}

	// Initialize repositories
	fileRepo := repository.NewFileRepository(db)

	// Initialize use cases
	fileUseCase := usecase.NewFileUseCase(fileRepo, cfg)

	// Setup Gin
	r := gin.Default()

	// Register routes
	route.SetupRoutes(r, fileUseCase)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server is running on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

package route

import (
	"fileupload/internal/delivery/http/handler"
	"fileupload/internal/delivery/http/middleware"
	"fileupload/internal/usecase"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, fileUseCase usecase.FileUseCase) {
	// Apply global middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.CheckContentTypeMiddleware())

	// Create handlers
	fileHandler := handler.NewFileHandler(fileUseCase)

	// API routes
	api := r.Group("/api")
	{
		// Upload routes
		uploads := api.Group("/uploads")
		{
			uploads.POST("", fileHandler.InitiateUpload)
			uploads.GET("/:upload_id", fileHandler.GetUploadStatus)
			uploads.POST("/:upload_id/chunks", fileHandler.UploadChunk)
			uploads.POST("/:upload_id/finalize", fileHandler.FinalizeUpload)
		}

		api.POST("/files", fileHandler.UploadFile)
	}
}

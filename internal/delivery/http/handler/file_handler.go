package handler

import (
	"fileupload/internal/usecase"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileHandler struct {
	fileUseCase usecase.FileUseCase
}

func NewFileHandler(fileUseCase usecase.FileUseCase) *FileHandler {
	return &FileHandler{
		fileUseCase: fileUseCase,
	}
}

// InitiateUpload godoc
// @Summary Initiate a new file upload
// @Description Start the process of uploading a file in chunks
// @Tags files
// @Accept json
// @Produce json
// @Param request body InitiateUploadRequest true "Upload information"
// @Success 201 {object} InitiateUploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /uploads [post]
func (h *FileHandler) InitiateUpload(c *gin.Context) {
	var req struct {
		FileName string `json:"file_name" binding:"required"`
		FileSize int64  `json:"file_size" binding:"required,min=1"`
		MimeType string `json:"mime_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	upload, err := h.fileUseCase.InitiateUpload(req.FileName, req.FileSize, req.MimeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"upload_id":  upload.ID,
		"file_name":  upload.OriginalName,
		"total_size": upload.TotalSize,
		"status":     upload.Status,
		"created_at": upload.CreatedAt,
	})
}

// UploadChunk godoc
// @Summary Upload a chunk of a file
// @Description Upload a chunk of a file using Content-Range header
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param upload_id path string true "Upload ID"
// @Param file formData file true "File chunk"
// @Param Content-Range header string true "Content range (e.g., bytes 0-1023/10240)"
// @Success 200 {object} UploadChunkResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /uploads/{upload_id}/chunks [post]
func (h *FileHandler) UploadChunk(c *gin.Context) {
	uploadIDStr := c.Param("upload_id")
	uploadID, err := uuid.Parse(uploadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid upload ID"})
		return
	}

	contentRange := c.GetHeader("Content-Range")
	if contentRange == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content-Range header is required"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file upload error: %v", err)})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to open uploaded file: %v", err)})
		return
	}
	defer src.Close()

	upload, err := h.fileUseCase.ProcessChunk(uploadID, src, contentRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_id":      upload.ID,
		"status":         upload.Status,
		"uploaded_size":  upload.UploadedSize,
		"total_size":     upload.TotalSize,
		"upload_percent": float64(upload.UploadedSize) / float64(upload.TotalSize) * 100,
	})
}

// FinalizeUpload godoc
// @Summary Finalize an upload
// @Description Complete the upload process and move the file to its final location
// @Tags files
// @Produce json
// @Param upload_id path string true "Upload ID"
// @Success 200 {object} FinalizeUploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /uploads/{upload_id}/finalize [post]
func (h *FileHandler) FinalizeUpload(c *gin.Context) {
	uploadIDStr := c.Param("upload_id")
	uploadID, err := uuid.Parse(uploadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid upload ID"})
		return
	}

	file, err := h.fileUseCase.FinalizeUpload(uploadID)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "upload incomplete") ||
			strings.Contains(err.Error(), "already completed") ||
			strings.Contains(err.Error(), "has failed") {
			status = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file_id":    file.ID,
		"file_name":  file.OriginalName,
		"size":       file.Size,
		"mime_type":  file.MimeType,
		"created_at": file.CreatedAt,
	})
}

// GetUploadStatus godoc
// @Summary Get upload status
// @Description Get the status of an ongoing or completed upload
// @Tags files
// @Produce json
// @Param upload_id path string true "Upload ID"
// @Success 200 {object} UploadStatusResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /uploads/{upload_id} [get]
func (h *FileHandler) GetUploadStatus(c *gin.Context) {
	uploadIDStr := c.Param("upload_id")
	uploadID, err := uuid.Parse(uploadIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid upload ID"})
		return
	}

	upload, err := h.fileUseCase.GetUploadStatus(uploadID)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"upload_id":      upload.ID,
		"file_name":      upload.OriginalName,
		"status":         upload.Status,
		"uploaded_size":  upload.UploadedSize,
		"total_size":     upload.TotalSize,
		"upload_percent": float64(upload.UploadedSize) / float64(upload.TotalSize) * 100,
		"created_at":     upload.CreatedAt,
		"updated_at":     upload.UpdatedAt,
	}

	if upload.CompletedAt != nil {
		response["completed_at"] = upload.CompletedAt
	}

	c.JSON(http.StatusOK, response)
}

// UploadFile godoc
// @Summary Upload a file (regular, non-chunked method)
// @Description Upload a file using the standard multipart/form-data method
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 200 {object} FileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files [post]
func (h *FileHandler) UploadFile(c *gin.Context) {
	// Get the file from form data
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to get file: %v", err)})
		return
	}
	defer file.Close()

	fileEntity, err := h.fileUseCase.DirectUpload(file, fileHeader)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create file record: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file_id":    fileEntity.ID,
		"file_name":  fileEntity.OriginalName,
		"size":       fileEntity.Size,
		"mime_type":  fileEntity.MimeType,
		"created_at": fileEntity.CreatedAt,
	})
}

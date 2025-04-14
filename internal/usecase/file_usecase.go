package usecase

import (
	"context"
	"errors"
	"fileupload/config"
	"fileupload/internal/domain/entity"
	"fileupload/internal/repository"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"fileupload/pkg/logger"
	minioClient "fileupload/pkg/minio"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type FileUseCase interface {
	InitiateUpload(originalName string, totalSize int64, mimeType string) (*entity.Upload, error)
	ProcessChunk(uploadID uuid.UUID, chunkReader io.Reader, contentRange string) (*entity.Upload, error)
	FinalizeUpload(uploadID uuid.UUID) (*entity.File, error)
	GetUploadStatus(uploadID uuid.UUID) (*entity.Upload, error)
	DirectUpload(file multipart.File, fileHeader *multipart.FileHeader) (*entity.File, error)
}

type fileUseCase struct {
	fileRepo repository.FileRepository
	config   *config.Config
}

func NewFileUseCase(fileRepo repository.FileRepository, config *config.Config) FileUseCase {
	return &fileUseCase{
		fileRepo: fileRepo,
		config:   config,
	}
}

func (u *fileUseCase) InitiateUpload(originalName string, totalSize int64, mimeType string) (*entity.Upload, error) {
	// Check file size limit
	if totalSize > u.config.MaxFileSize {
		return nil, errors.New("file size exceeds maximum allowed size")
	}

	// Generate a unique ID for the upload
	uploadID := uuid.New()

	// Generate a unique file name
	ext := filepath.Ext(originalName)
	fileName := uuid.New().String() + ext
	tempPath := filepath.Join(u.config.UploadTempDir, fileName)

	// Create an empty file for writing chunks
	file, err := os.Create(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer file.Close()

	now := time.Now()
	upload := &entity.Upload{
		ID:           uploadID,
		FileName:     fileName,
		OriginalName: originalName,
		TotalSize:    totalSize,
		UploadedSize: 0,
		MimeType:     mimeType,
		Status:       "pending",
		TempPath:     tempPath,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = u.fileRepo.CreateUpload(upload)
	if err != nil {
		// Clean up the temporary file
		os.Remove(tempPath)
		return nil, fmt.Errorf("failed to create upload record: %w", err)
	}

	return upload, nil
}

func (u *fileUseCase) ProcessChunk(uploadID uuid.UUID, chunkReader io.Reader, contentRange string) (*entity.Upload, error) {
	upload, err := u.fileRepo.GetUploadByID(uploadID)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %w", err)
	}

	if upload.Status == "completed" {
		return nil, errors.New("upload already completed")
	}

	if upload.Status == "failed" {
		return nil, errors.New("upload has failed")
	}

	// Parse content range header (format: bytes start-end/total)
	var start, end, total int64
	_, err = fmt.Sscanf(contentRange, "bytes %d-%d/%d", &start, &end, &total)
	if err != nil {
		return nil, fmt.Errorf("invalid content range format: %w", err)
	}

	// Validate range
	if total != upload.TotalSize {
		return nil, errors.New("total size mismatch")
	}

	if start > upload.UploadedSize {
		return nil, errors.New("chunk out of order")
	}

	// Open the temporary file for writing
	file, err := os.OpenFile(upload.TempPath, os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open temporary file: %w", err)
	}
	defer file.Close()

	// Seek to the correct position
	_, err = file.Seek(start, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to seek in file: %w", err)
	}

	// Write the chunk
	written, err := io.Copy(file, chunkReader)
	if err != nil {
		upload.Status = "failed"
		u.fileRepo.UpdateUpload(upload)
		return nil, fmt.Errorf("failed to write chunk: %w", err)
	}

	// Update upload status
	chunkSize := end - start + 1
	if written != chunkSize {
		upload.Status = "failed"
		u.fileRepo.UpdateUpload(upload)
		return nil, fmt.Errorf("chunk size mismatch: expected %d, got %d", chunkSize, written)
	}

	// Update upload progress
	newUploadedSize := start + written
	if newUploadedSize > upload.UploadedSize {
		upload.UploadedSize = newUploadedSize
	}
	upload.Status = "uploading"
	upload.UpdatedAt = time.Now()

	err = u.fileRepo.UpdateUpload(upload)
	if err != nil {
		return nil, fmt.Errorf("failed to update upload record: %w", err)
	}

	return upload, nil
}

func (u *fileUseCase) FinalizeUpload(uploadID uuid.UUID) (*entity.File, error) {
	upload, err := u.fileRepo.GetUploadByID(uploadID)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %w", err)
	}

	if upload.Status == "completed" {
		return nil, errors.New("upload already completed")
	}

	if upload.Status == "failed" {
		return nil, errors.New("upload has failed")
	}

	// Check if all chunks have been uploaded
	if upload.UploadedSize != upload.TotalSize {
		return nil, fmt.Errorf("upload incomplete: expected %d bytes, got %d bytes", upload.TotalSize, upload.UploadedSize)
	}

	// Move the file from temp directory to final directory
	finalPath := filepath.Join(u.config.UploadFinalDir, upload.FileName)
	err = os.Rename(upload.TempPath, finalPath)
	if err != nil {
		upload.Status = "failed"
		u.fileRepo.UpdateUpload(upload)
		return nil, fmt.Errorf("failed to move file to final location: %w", err)
	}

	// Update upload status
	now := time.Now()
	upload.Status = "completed"
	upload.UpdatedAt = now
	upload.CompletedAt = &now

	err = u.fileRepo.UpdateUpload(upload)
	if err != nil {
		return nil, fmt.Errorf("failed to update upload record: %w", err)
	}

	// Create file record
	file := &entity.File{
		ID:           uuid.New(),
		FileName:     upload.FileName,
		OriginalName: upload.OriginalName,
		Size:         upload.TotalSize,
		MimeType:     upload.MimeType,
		Path:         finalPath,
		UploadID:     upload.ID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = u.fileRepo.CreateFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}
	if u.config.EnabledMinio {
		f, err := os.Open(finalPath)
		if err != nil {
			upload.Status = "failed"
			u.fileRepo.UpdateUpload(upload)
			return nil, fmt.Errorf("failed to open file for minio upload: %w", err)
		}
		defer f.Close()

		fileStat, err := f.Stat()
		if err != nil {
			upload.Status = "failed"
			u.fileRepo.UpdateUpload(upload)
			return nil, fmt.Errorf("failed to stat file: %w", err)
		}

		_, err = minioClient.Client.PutObject(context.Background(), "uploads",
			upload.FileName, f, fileStat.Size(), minio.PutObjectOptions{
				ContentType: upload.MimeType,
				UserMetadata: map[string]string{
					"originalName": upload.OriginalName,
					"uploadID":     upload.ID.String(),
				},
			})
		if err != nil {
			upload.Status = "failed"
			u.fileRepo.UpdateUpload(upload)
			logger.UploadLog.Errorf("failed to upload file to MinIO: %v", err)
			return nil, fmt.Errorf("failed to upload file to MinIO: %w", err)
		}

	}

	return file, nil
}

func (u *fileUseCase) GetUploadStatus(uploadID uuid.UUID) (*entity.Upload, error) {
	return u.fileRepo.GetUploadByID(uploadID)
}

func (u *fileUseCase) DirectUpload(file multipart.File, fileHeader *multipart.FileHeader) (*entity.File, error) {
	if fileHeader.Size > u.config.MaxFileSize {
		return nil, errors.New("file size exceeds maximum allowed size")
	}

	// Generate a unique ID for the upload
	uploadID := uuid.New()

	ext := filepath.Ext(fileHeader.Filename)
	fileName := uuid.New().String() + ext
	finalDir := u.config.UploadFinalDir

	if err := os.MkdirAll(finalDir, os.ModePerm); err != nil {
		return nil, errors.New("failed to create directory")
	}

	finalPath := filepath.Join(finalDir, fileName)
	out, err := os.Create(finalPath)
	if err != nil {
		return nil, errors.New("failed to create file")
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return nil, errors.New("failed to write file")
	}

	now := time.Now()

	fileEntity := &entity.File{
		ID:           uuid.New(),
		FileName:     fileName,
		OriginalName: fileHeader.Filename,
		Size:         fileHeader.Size,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		Path:         finalPath,
		UploadID:     uploadID, // We still create a reference to a "virtual" upload
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = u.fileRepo.CreateFile(fileEntity)
	if err != nil {
		return nil, errors.New("failed to create file record")
	}

	return fileEntity, nil
}

package repository

import (
	"fileupload/internal/domain/entity"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	FileName     string
	OriginalName string
	TotalSize    int64
	UploadedSize int64
	MimeType     string
	Status       string
	TempPath     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
}

type FileModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	FileName     string
	OriginalName string
	Size         int64
	MimeType     string
	Path         string
	UploadID     uuid.UUID `gorm:"type:uuid"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type FileRepository interface {
	CreateUpload(upload *entity.Upload) error
	GetUploadByID(id uuid.UUID) (*entity.Upload, error)
	UpdateUpload(upload *entity.Upload) error
	CreateFile(file *entity.File) error
	GetFileByID(id uuid.UUID) (*entity.File, error)
}

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{
		db: db,
	}
}

func (r *fileRepository) CreateUpload(upload *entity.Upload) error {
	model := &UploadModel{
		ID:           upload.ID,
		FileName:     upload.FileName,
		OriginalName: upload.OriginalName,
		TotalSize:    upload.TotalSize,
		UploadedSize: upload.UploadedSize,
		MimeType:     upload.MimeType,
		Status:       upload.Status,
		TempPath:     upload.TempPath,
		CreatedAt:    upload.CreatedAt,
		UpdatedAt:    upload.UpdatedAt,
		CompletedAt:  upload.CompletedAt,
	}
	return r.db.Create(model).Error
}

func (r *fileRepository) GetUploadByID(id uuid.UUID) (*entity.Upload, error) {
	var model UploadModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if err != nil {
		return nil, err
	}

	return &entity.Upload{
		ID:           model.ID,
		FileName:     model.FileName,
		OriginalName: model.OriginalName,
		TotalSize:    model.TotalSize,
		UploadedSize: model.UploadedSize,
		MimeType:     model.MimeType,
		Status:       model.Status,
		TempPath:     model.TempPath,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		CompletedAt:  model.CompletedAt,
	}, nil
}

func (r *fileRepository) UpdateUpload(upload *entity.Upload) error {
	model := &UploadModel{
		ID:           upload.ID,
		FileName:     upload.FileName,
		OriginalName: upload.OriginalName,
		TotalSize:    upload.TotalSize,
		UploadedSize: upload.UploadedSize,
		MimeType:     upload.MimeType,
		Status:       upload.Status,
		TempPath:     upload.TempPath,
		CreatedAt:    upload.CreatedAt,
		UpdatedAt:    upload.UpdatedAt,
		CompletedAt:  upload.CompletedAt,
	}
	return r.db.Save(model).Error
}

func (r *fileRepository) CreateFile(file *entity.File) error {
	model := &FileModel{
		ID:           file.ID,
		FileName:     file.FileName,
		OriginalName: file.OriginalName,
		Size:         file.Size,
		MimeType:     file.MimeType,
		Path:         file.Path,
		UploadID:     file.UploadID,
		CreatedAt:    file.CreatedAt,
		UpdatedAt:    file.UpdatedAt,
	}
	return r.db.Create(model).Error
}

func (r *fileRepository) GetFileByID(id uuid.UUID) (*entity.File, error) {
	var model FileModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if err != nil {
		return nil, err
	}

	return &entity.File{
		ID:           model.ID,
		FileName:     model.FileName,
		OriginalName: model.OriginalName,
		Size:         model.Size,
		MimeType:     model.MimeType,
		Path:         model.Path,
		UploadID:     model.UploadID,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}, nil
}

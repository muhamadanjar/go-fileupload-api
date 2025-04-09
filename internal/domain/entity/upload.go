package entity

import (
	"time"

	"github.com/google/uuid"
)

type Upload struct {
	ID           uuid.UUID
	FileName     string
	OriginalName string
	TotalSize    int64
	UploadedSize int64
	MimeType     string
	Status       string // "pending", "uploading", "completed", "failed"
	TempPath     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
}

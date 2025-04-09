package entity

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID           uuid.UUID
	FileName     string
	OriginalName string
	Size         int64
	MimeType     string
	Path         string
	UploadID     uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

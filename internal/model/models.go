package model

import "time"

// TaskStatus represents the current state of an image processing task.
type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusProcessing TaskStatus = "processing"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
)

type ImageProcessingTask struct {
	ID               string     `db:"id"`
	OriginalFilename string     `db:"original_filename"`
	StorageKey       string     `db:"storage_key"`
	Status           TaskStatus `db:"status"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
	ErrorMessage     string     `db:"error_message"`
}

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
	ID               int64      `db:"id"`
	OriginalFilename string     `db:"original_filename"`
	StorageKey       string     `db:"storage_key"`
	Status           TaskStatus `db:"status"`
	ErrorMessage     string     `db:"error_message"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`

	// Relation
	processedImage *ProcessedImage
}

type ProcessedImage struct {
	ID         int64     `db:"id"`
	TaskID     int64     `db:"task_id"`
	Format     string    `db:"format"`
	Size       string    `db:"size"`
	StorageKey string    `db:"storage_key"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

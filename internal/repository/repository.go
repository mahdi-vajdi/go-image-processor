package repository

import (
	"context"
	"errors"

	"github.com/mahdi-vajdi/go-image-processor/internal/model"
)

type Repository interface {
	CreateTask(ctx context.Context, task *model.ImageProcessingTask) (*model.ImageProcessingTask, error)

	GetTaskByID(ctx context.Context, id int64) (*model.ImageProcessingTask, error)

	UpdateTaskStatus(ctx context.Context, id int64, status model.TaskStatus, errorMessage string) error

	GetPendingTasks(ctx context.Context, limit int) ([]model.ImageProcessingTask, error)

	CreateProcessedImageDetail(ctx context.Context, detail *model.ProcessedImage) (*model.ProcessedImage, error)
}

var ErrTaskNotFound = errors.New("repository: task not found")

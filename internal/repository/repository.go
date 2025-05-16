package repository

import (
	"context"
	"errors"

	"github.com/mahdi-vajdi/go-image-processor/internal/model"
)

type Repository interface {
	CreateTask(ctx context.Context, task *model.ImageProcessingTask) (*model.ImageProcessingTask, error)

	GetTaskByID(ctx context.Context, id string) (*model.ImageProcessingTask, error)

	UpdateTaskStatus(ctx context.Context, id string, status model.TaskStatus, errorMessage string) error

	GetPendingTasks(ctx context.Context, limit int) ([]model.ImageProcessingTask, error)
}

var ErrTaskNotFound = errors.New("repository: task not found")

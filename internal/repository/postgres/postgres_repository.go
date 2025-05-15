package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mahdi-vajdi/go-image-processor/internal/model"
	"github.com/mahdi-vajdi/go-image-processor/internal/repository"
)

type Repository struct {
	db *sqlx.DB
}

var _ repository.Repository = (*Repository)(nil)

func NewTaskRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTask(ctx context.Context, task *model.ImageProcessingTask) (*model.ImageProcessingTask, error) {
	task.ID = uuid.New().String()
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	task.Status = model.StatusPending

	query := `INSERT INTO image_processing_tasks (id, original_filename, storage_key, status, created_at, updated_at, error_message)
			  VALUES (id, original_filename, storage_key, status, created_at, updated_at, error_message)
			  `

	_, err := r.db.NamedExecContext(ctx, query, task)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

func (r *Repository) GetTaskByID(ctx context.Context, id string) (*model.ImageProcessingTask, error) {
	var task model.ImageProcessingTask
	query := `SELECT id, original_filename, storage_key, status, created_at, updated_at, error_message 
			  FROM image_processing_tasks 
			  WHERE id = $1
			  `

	err := r.db.GetContext(ctx, task, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task with ID %s was not found", id)
		}
		return nil, fmt.Errorf("failed to get task by ID %s: %w", id, err)
	}

	return &task, nil
}

func (r *Repository) UpdateTaskStatus(ctx context.Context, id string, status model.TaskStatus, errorMessage string) error {
	query := `UPDATE image_processing_tasks SET status = $1, error_message = $2, updated_at = DEFAULT WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, status, errorMessage, id)
	if err != nil {
		return fmt.Errorf("failed to udpate task status with ID %s: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after updating task status for ID %s: %w", id, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no task status with ID %s was found to update", id)
	}

	return nil
}

func (r *Repository) GetPendingTasks(ctx context.Context, limit int) ([]model.ImageProcessingTask, error) {
	var tasks []model.ImageProcessingTask
	query := `SELECT id, original_filename, storage_key, status, created_at, updated_at, error_message 
			  FROM image_processing_tasks 
			  WHERE status = $1
			  ORDER BY created_at 
			  LIMIT $2
	`

	err := r.db.SelectContext(ctx, tasks, query, model.StatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending tasks: %w", err)
	}

	return tasks, nil
}

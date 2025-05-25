package processing

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/mahdi-vajdi/go-image-processor/internal/model"
	"github.com/mahdi-vajdi/go-image-processor/internal/repository"
	"github.com/mahdi-vajdi/go-image-processor/internal/storage"
)

type ServiceConfig struct {
	WorkerPoolSize  int
	PollingInterval time.Duration
	TaskBatchSize   int
}

type Service struct {
	repo    repository.Repository
	storage storage.Storage
	config  ServiceConfig

	taskChan chan model.ImageProcessingTask
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewService(repo repository.Repository, storage storage.Storage, config ServiceConfig) *Service {
	if config.WorkerPoolSize <= 0 {
		config.WorkerPoolSize = 5
		log.Printf("Warning: WorkerPoolSize not set or invalid, defaulting to %d", config.WorkerPoolSize)
	}
	if config.PollingInterval <= 0 {
		config.PollingInterval = 5 * time.Second
		log.Printf("Warning: PollingInterval not set or invalid, defaulting to %s", config.PollingInterval)
	}
	if config.TaskBatchSize <= 0 {
		config.TaskBatchSize = 10
		log.Printf("Warning: TaskBatchSize not set or invalid, defaulting to %d", config.TaskBatchSize)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		repo:     repo,
		storage:  storage,
		config:   config,
		taskChan: make(chan model.ImageProcessingTask, config.TaskBatchSize*2),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (s *Service) Start() {
	log.Println("Starting image processing service...")

	for i := 0; i < s.config.WorkerPoolSize; i++ {
		s.wg.Add(1)
		go s.worker(i + 1) // Pass the ID
	}

	log.Printf("Image processing service started with %d workers and polling every %s", s.config.WorkerPoolSize, s.config.PollingInterval)
}

func (s *Service) Stop(ctx context.Context) {
	log.Println("stopping image processing service...")

	// Signal workers to stop
	close(s.taskChan)

	done := make(chan struct{})
	go func() {
		s.wg.Wait() // Wait for all goroutines to finish
		close(done) // Signal that all goroutines are finished
	}()

	// Wait for either all goroutines to finish or the provided shutdown context to expire
	select {
	case <-done:
		log.Println("All image processing tasks finished")
	case <-ctx.Done():
		log.Println("Image processing service shutdown context timed out. Some tasks may not have finished.")
	}

	log.Println("Image processing service stopped.")
}

func (s *Service) SubmitTask(task model.ImageProcessingTask) {
	select {
	case s.taskChan <- task:
		log.Printf("task %d submistted to processing channel.", task.ID)
	case <-s.ctx.Done():
		log.Printf("Could not submit task %d. Service is shutting down.", task.ID)
	}
}

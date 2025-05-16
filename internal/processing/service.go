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

	s.wg.Add(1)
	go s.dispatcher()

	for i := 0; i < s.config.WorkerPoolSize; i++ {
		s.wg.Add(1)
		go s.worker(i + 1) // Pass the ID
	}

	log.Printf("Image processing service started with %d workers and polling every %s", s.config.WorkerPoolSize, s.config.PollingInterval)
}

func (s *Service) Stop(ctx context.Context) {
	log.Println("stopping image processing service...")

	// Signal the dispatcher to stop
	s.cancel()

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

func (s *Service) dispatcher() {
	defer s.wg.Done()

	log.Println("Dispatcher started")

	// Loop indefinitely until the service context is cancelled.
	for {
		select {
		case <-s.ctx.Done():
			log.Println("dispatcher received shutdown signal. Closing the task channel...")
			close(s.taskChan)
			return

		default:
			// Continue
		}

		tasks, err := s.repo.GetPendingTasks(s.ctx, s.config.TaskBatchSize)
		if err != nil {
			log.Printf("Error fetching pending task: %v. Retrying in %s", err, s.config.PollingInterval)
			time.Sleep(s.config.PollingInterval)
			continue // Start the next iteration
		}

		if len(tasks) == 0 {
			log.Println("No pending tasks were found. Waiting...")
			time.Sleep(s.config.PollingInterval)
			continue
		}

		log.Printf("Dispatcher fetched %d tasks", len(tasks))

		// Send the fetched tasks to the task channel
		for _, task := range tasks {
			select {
			case s.taskChan <- task:
			// Task successfully was sent to the channel
			case <-s.ctx.Done():
				log.Printf("Dispatcher shutting down while sending tasks. Task %s might not be processed from this batch", task.ID)
				close(s.taskChan)
				return
			}
		}

		// Wait for the polling interval before fetching the next batch
		time.Sleep(s.config.PollingInterval)
	}
}

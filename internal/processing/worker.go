package processing

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/mahdi-vajdi/go-image-processor/internal/model"
)

func (s *Service) worker(id int) {
	defer s.wg.Done()

	log.Printf("Worker #%d started", id)

	for task := range s.taskChan {
		log.Printf("Worker #%d processing task %s (original: %s)...", id, task.ID, task.OriginalFilename)

		processingCtx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
		err := s.repo.UpdateTaskStatus(processingCtx, task.ID, model.StatusProcessing, "")
		cancel()

		if err != nil {
			log.Printf("Worker #%d failed to set the task %s to 'procssing': %v", id, task.ID, err)
			continue
		}

		// Initialize the processing
		processErr := s.processTask(s.ctx, &task)

		// Update the task with the processing results
		status := model.StatusCompleted
		errorMessage := ""
		if processErr != nil {
			status = model.StatusFailed
			errorMessage = processErr.Error()

			log.Printf("Worker #%d: Task %s failed: %v", id, task.ID, processErr)
		} else {
			log.Printf("Worker #%d: Task %s completed successfully", id, task.ID)
		}

		updateContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err = s.repo.UpdateTaskStatus(updateContext, task.ID, status, errorMessage)
		cancel()

		if err != nil {
			// TODO: I need to create a cleanup process for this (maybe retry)
			log.Printf("Worker #%d FATAL: failed to update task %s with final status '%s': %v", id, task.ID, status, err)
		}
	}

	log.Printf("Worker #%d exiting.", id)
}

func (s *Service) processTask(ctx context.Context, task *model.ImageProcessingTask) error {
	// Check if the context has been cancelled before starting or during long operations.
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Context is not done
	}

	log.Printf("Processing image for task %s (Storage key: %s)", task.ID, task.StorageKey)

	// Download the file
	originalImageReader, err := s.storage.Get(ctx, task.StorageKey)
	if err != nil {
		return fmt.Errorf("failed to download original image %s: %w", task.StorageKey, err)
	}
	defer originalImageReader.Close()

	// Decode the image
	// Determine the image format based on the original file extension
	img, format, err := image.Decode(originalImageReader)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	log.Printf("Image decoded successfully (Format: %s, Size: %dx%d)", format, img.Bounds().Dx(), img.Bounds().Dy())

	// Resize the image
	targetWidth := 800
	resizedImage := imaging.Resize(img, targetWidth, 0, imaging.Lanczos)

	log.Printf("Image resized to %dx%d", resizedImage.Bounds().Dx(), resizedImage.Bounds().Dy())

	// Encode the processed image
	var buf bytes.Buffer
	outputFormat := imaging.JPEG // TODO: this can be configurable by params
	originalExt := filepath.Ext(task.OriginalFilename)

	// New filename
	// FIXME: use a better naming structure
	processedFilename := fmt.Sprintf("%s_%dx%d.%s",
		strings.TrimSuffix(task.OriginalFilename, originalExt),
		resizedImage.Bounds().Dx(),
		resizedImage.Bounds().Dy(),
		strings.ToLower(outputFormat.String()),
	)

	err = imaging.Encode(&buf, resizedImage, outputFormat)
	if err != nil {
		return fmt.Errorf("failed to encode processed image: %w", err)
	}

	// Upload the processed image
	processedStorageKey, err := s.storage.Save(ctx, processedFilename, &buf)
	if err != nil {
		return fmt.Errorf("failed to upload processed image %s: %w", processedFilename, err)
	}

	log.Printf("Processed image uploaded successfully with key: %s", processedStorageKey)

	// TODO: save the details for the uploaded image in the database

	return nil
}

package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/mahdi-vajdi/go-image-processor/internal/model"
	"github.com/mahdi-vajdi/go-image-processor/internal/processing"
	"github.com/mahdi-vajdi/go-image-processor/internal/repository"
	"github.com/mahdi-vajdi/go-image-processor/internal/storage"
)

type Handler interface {
	Ping(w http.ResponseWriter, r *http.Request)
	UploadImage(w http.ResponseWriter, r *http.Request)
	GetImageStatus(w http.ResponseWriter, r *http.Request)
	GetImage(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	repo       repository.Repository
	imageStore storage.Storage
	processor  *processing.Service
}

func NewHandler(repo repository.Repository, imageStore storage.Storage, processor *processing.Service) Handler {
	return &handler{
		repo:       repo,
		imageStore: imageStore,
		processor:  processor,
	}
}

func (h *handler) Ping(w http.ResponseWriter, _ *http.Request) {
	ResponseJSON(w, http.StatusOK, map[string]string{"message": "pong"})
}

func (h *handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		ErrorJSON(w, http.StatusBadGateway, fmt.Sprintf("failed to parse multipart form: %v", err))
		return
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			ErrorJSON(w, http.StatusBadRequest, "missing file 'image' in the form data")
		} else {
			ErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("failed to get file from form: %v", err))
		}
		return
	}
	defer file.Close()

	originalFilename := fileHeader.Filename
	if originalFilename == "" {
		ErrorJSON(w, http.StatusBadRequest, "missing file name")
		return
	}

	storageKey, err := h.imageStore.Save(ctx, originalFilename, file)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("failed to save image: %v", err))
		return
	}

	task := &model.ImageProcessingTask{OriginalFilename: originalFilename, StorageKey: storageKey}

	createdTask, err := h.repo.CreateTask(ctx, task)
	if err != nil {
		log.Printf("Warning: failed to create task for storage key %s: %v", storageKey, err)
		cleanUpErr := h.imageStore.Delete(context.Background(), storageKey)
		if cleanUpErr != nil {
			log.Printf("Warning: failed to clean up storage key %s after task creation failed: %v", storageKey, cleanUpErr)
		}
		ErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("failed to create processing task: %v", err))
		return
	}

	h.processor.SubmitTask(*createdTask)

	ResponseJSON(w, http.StatusAccepted, map[string]string{"id": strconv.FormatInt(createdTask.ID, 10), "status": string(createdTask.Status), "createdAt": createdTask.CreatedAt.String()})
}

func (h *handler) GetImageStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	taskIDString := vars["taskId"]
	if taskIDString == "" {
		ErrorJSON(w, http.StatusBadRequest, "missing task ID")
		return
	}

	taskID, err := strconv.ParseInt(taskIDString, 10, 64)
	if err != nil {
		ErrorJSON(w, http.StatusBadRequest, "invalid task ID")
		return
	}

	task, err := h.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			ErrorJSON(w, http.StatusNotFound, "task not found")
		} else {
			ErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("failed to get task: %v", err))
		}
		return
	}

	ResponseJSON(w, http.StatusOK, task)
}

func (h *handler) GetImage(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	imageKey := vars["imageKey"]
	if imageKey == "" {
		ErrorJSON(w, http.StatusBadRequest, "missing image key")
		return
	}

	imageReader, err := h.imageStore.Get(ctx, imageKey)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			ErrorJSON(w, http.StatusNotFound, "image not found")
		} else {
			ErrorJSON(w, http.StatusInternalServerError, fmt.Sprintf("failed to get image from storage: %v", err))
		}
		return
	}
	defer imageReader.Close()

	// TODO: store and retrieve the content type using the storage metadata
	imageExt := strings.ToLower(filepath.Ext(imageKey))
	contentType := "application/octet-stream"
	switch imageExt {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	}

	w.Header().Set("Content-Type", contentType)
	// TODO: set the size too

	_, err = io.Copy(w, imageReader)
	if err != nil {
		// Headers are already sent so cannot return error
		log.Printf("Error streaming image data for key %s: %v", imageKey, err)
	}
}

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mahdi-vajdi/go-image-processor/internal/processing"
	"github.com/mahdi-vajdi/go-image-processor/internal/repository/postgres"
	"github.com/mahdi-vajdi/go-image-processor/internal/router"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mahdi-vajdi/go-image-processor/internal/config"
	"github.com/mahdi-vajdi/go-image-processor/internal/handler"
	"github.com/mahdi-vajdi/go-image-processor/internal/storage"
	"github.com/mahdi-vajdi/go-image-processor/internal/storage/local"
	"github.com/mahdi-vajdi/go-image-processor/internal/storage/s3"
)

func main() {
	appCtx := context.Background()

	if err := config.LoadEnv(""); err != nil {
		log.Printf("Warning: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configurtation: %v", err)
	}

	// Database
	db, err := sqlx.Connect("postgres", cfg.Database.PostgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(appCtx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database successfully")

	// Repository
	repo := postgres.NewRepository(db)

	// Storage
	var imageStore storage.Storage
	if cfg.Storage.Type == "local" {
		localStore, err := localStorage.NewLocalStore(cfg.Storage.Local.BaseDir)
		if err != nil {
			log.Fatalf("Failed to create local storage: %v", err)
		}
		imageStore = localStore
	} else if cfg.Storage.Type == "s3" {
		s3Store, err := s3Storage.NewS3Store(
			cfg.Storage.S3.EndpointURL,
			cfg.Storage.S3.AccessKeyID,
			cfg.Storage.S3.SecretAccessKey,
			cfg.Storage.S3.Bucket,
			cfg.Storage.S3.Prefix,
		)
		if err != nil {
			log.Fatalf("Failed to create S3 storage: %v", err)
		}
		imageStore = s3Store
	} else {
		log.Fatalf("Unknown storage type: %s", cfg.Storage.Type)
	}

	// Processing service
	processingService := processing.NewService(repo, imageStore, processing.ServiceConfig{
		WorkerPoolSize:  cfg.ProcessingService.WorkerPoolSize,
		PollingInterval: cfg.ProcessingService.PollingInterval,
		TaskBatchSize:   cfg.ProcessingService.TaskBatchSize,
	})
	processingService.Start()

	// Handler
	apiHandler := handler.NewHandler(repo, imageStore)

	r := router.New(apiHandler)

	serverAddr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		fmt.Printf("Starting http server on %s\n", serverAddr)
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start the HTTP server: %v", err)
		}

	}()

	// Setup graceful shutdown
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)
	<-quitChan

	log.Println("Shutting down server...")

	// Create a context with timeout for shutdown
	appCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.IdleTimeout)
	defer cancel()

	if err = server.Shutdown(appCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")

}

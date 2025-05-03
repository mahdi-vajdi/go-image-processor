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

	"github.com/mahdi-vajdi/go-image-processor/internal/config"
	"github.com/mahdi-vajdi/go-image-processor/internal/handler"
	"github.com/mahdi-vajdi/go-image-processor/internal/router"
)

func main() {
	if err := config.LoadEnv(""); err != nil {
		log.Printf("Warning: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configurtation: %v", err)
	}

	// Initialize handlers
	publicHandler := handler.NewPublicHandler()

	r := router.New(publicHandler)

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
	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.IdleTimeout)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")

}

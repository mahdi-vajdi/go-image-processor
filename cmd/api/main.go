package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mahdi-vajdi/go-image-processor/internal/config"
)

func main() {
	if err := config.LoadEnv(""); err != nil {
		log.Printf("Warning: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configurtation: %v", err)
	}

	serverAddr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	server := &http.Server{
		Addr:         serverAddr,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	fmt.Printf("starting http server on %s\n", serverAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start the HTTP server: %v", err)
	}
}

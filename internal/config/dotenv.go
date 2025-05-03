package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnv(envFile string) error {
	if envFile == "" {
		// Check for APP_ENV variable to determine which .env file to load
		env := os.Getenv("APP_ENV")
		if env == "" {
			env = "development"
		}

		// Try environment-specific file first
		envFile = fmt.Sprintf(".env.%s", env)
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			envFile = ".env"
		}
	}
	// Check if file exists
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		log.Printf("Warning: Environment file %s not found, using system environment variables only", envFile)
		return nil
	}

	// Look for the file in current directory and parent directories
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		path := filepath.Join(currentDir, envFile)
		if _, err = os.Stat(path); err == nil {
			err = godotenv.Load(path)
			if err != nil {
				return fmt.Errorf("error loading env file %s: %w", path, err)
			}
			log.Printf("Loaded enviroment from %s", path)
			return nil
		}

		// Go up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached the root
			break
		}
		currentDir = parentDir
	}

	return fmt.Errorf("couldn't find %s in current or parent directories", envFile)
}

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTP ServerConfig
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func LoadConfig() (*Config, error) {
	config := &Config{
		HTTP: ServerConfig{
			Host:         getEnv("HTTP_HOST", "0.0.0.0"),
			Port:         getEnvAsInt("HTTP_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: getEnvAsDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvAsDuration("HTTP_IDLE_TIMEOUT", 120*time.Second),
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
		fmt.Printf("Warning: Enviroment %s has invalid integer value, using defuault\n", key)
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
		fmt.Printf("Warning: Enviroment %s has invalid boolean value, using defuault\n", key)
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		fmt.Printf("Warning: Environment %s has invalid duration value, using defuault\n", key)
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string, separator string) []string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.Split(value, separator)
	}
	return defaultValue
}

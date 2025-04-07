package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration settings
type Config struct {
	Port     int
	NodePort int
	Mode     string
	Storage  StorageConfig
}

// StorageConfig holds storage-related configuration
type StorageConfig struct {
	Type      string
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file only in local environment
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Printf("⚠️ Warning: Could not load .env file: %v", err)
		} else {
			log.Println("✅ .env file loaded for local development")
		}
	} else {
		log.Println("ℹ️ .env not found — using system environment variables")
	}

	cfg := &Config{
		Port:     getEnvInt("PORT", 8080),
		NodePort: getEnvInt("NODE_PORT", 8081),
		Mode:     getEnv("MODE", "web"),
		Storage: StorageConfig{
			Type:      getEnv("STORAGE_TYPE", "s3"),
			Endpoint:  getEnv("STORAGE_ENDPOINT", ""),
			AccessKey: getEnv("STORAGE_ACCESS_KEY", ""),
			SecretKey: getEnv("STORAGE_SECRET_KEY", ""),
			Bucket:    getEnv("STORAGE_BUCKET", ""),
			UseSSL:    getEnvBool("STORAGE_USE_SSL", true),
		},
	}

	// Validate required storage configuration
	if cfg.Storage.Type == "s3" {
		if cfg.Storage.Endpoint == "" {
			return nil, fmt.Errorf("STORAGE_ENDPOINT is required for s3 storage")
		}
		if cfg.Storage.AccessKey == "" {
			return nil, fmt.Errorf("STORAGE_ACCESS_KEY is required for s3 storage")
		}
		if cfg.Storage.SecretKey == "" {
			return nil, fmt.Errorf("STORAGE_SECRET_KEY is required for s3 storage")
		}
		if cfg.Storage.Bucket == "" {
			return nil, fmt.Errorf("STORAGE_BUCKET is required for s3 storage")
		}
	}

	return cfg, nil
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper function to get integer environment variable with default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Helper function to get boolean environment variable with default
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// Optional utility method
func (c *Config) GetStorageEndpoint() string {
	return c.Storage.Endpoint
}

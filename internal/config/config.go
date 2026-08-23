package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	Storage  StorageConfig
	Logging  LoggingConfig
}
type HTTPConfig struct {
	Address                                    string
	ReadTimeout, WriteTimeout, ShutdownTimeout time.Duration
	AllowedOrigins                             []string
}
type DatabaseConfig struct {
	URL      string
	Required bool
	Timeout  time.Duration
}
type StorageConfig struct {
	Root           string
	MaxUploadBytes int64
	AllowedTypes   []string
}
type LoggingConfig struct {
	Level       string
	Development bool
}

func Load() (Config, error) {
	cfg := Config{HTTP: HTTPConfig{Address: get("HTTP_ADDRESS", ":8080"), ReadTimeout: duration("HTTP_READ_TIMEOUT", 10*time.Second), WriteTimeout: duration("HTTP_WRITE_TIMEOUT", 30*time.Second), ShutdownTimeout: duration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second), AllowedOrigins: split(get("CORS_ALLOWED_ORIGINS", "http://localhost:5173"))}, Database: DatabaseConfig{URL: get("DATABASE_URL", "postgres://cry076:cry076@localhost:5432/cry076?sslmode=disable"), Required: boolean("DATABASE_REQUIRED", true), Timeout: duration("DATABASE_TIMEOUT", 5*time.Second)}, Storage: StorageConfig{Root: get("STORAGE_ROOT", "./var/data"), MaxUploadBytes: int64Value("MAX_UPLOAD_BYTES", 8<<20), AllowedTypes: split(get("ALLOWED_UPLOAD_TYPES", "application/pdf,image/png,image/jpeg"))}, Logging: LoggingConfig{Level: get("LOG_LEVEL", "info"), Development: boolean("LOG_DEVELOPMENT", false)}}
	if cfg.Database.Required && cfg.Database.URL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.Storage.MaxUploadBytes <= 0 {
		return Config{}, fmt.Errorf("MAX_UPLOAD_BYTES must be positive")
	}
	return cfg, nil
}
func get(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
func split(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := strings.TrimSpace(part); item != "" {
			result = append(result, item)
		}
	}
	return result
}
func boolean(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
func duration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
func int64Value(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

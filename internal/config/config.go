// Package config provides environment-based configuration loading.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	// Server settings
	Port            int
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration

	// Database settings
	DatabaseURL     string
	MaxDBConns      int
	MinDBConns      int
	DBConnMaxLife   time.Duration

	// Application settings
	BaseURL         string
	MaxContentSize  int64
	DefaultExpiry   time.Duration
	MinExpiry       time.Duration
	MaxExpiry       time.Duration
	CleanupInterval time.Duration

	// Rate limiting
	PostRateLimit int
	GetRateLimit  int
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		// Server defaults
		Port:            getEnvInt("PORT", 8080),
		Host:            getEnvString("HOST", "0.0.0.0"),
		ReadTimeout:     getEnvDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getEnvDuration("WRITE_TIMEOUT", 30*time.Second),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),

		// Database defaults
		DatabaseURL:   getEnvString("DATABASE_URL", ""),
		MaxDBConns:    getEnvInt("MAX_DB_CONNS", 25),
		MinDBConns:    getEnvInt("MIN_DB_CONNS", 5),
		DBConnMaxLife: getEnvDuration("DB_CONN_MAX_LIFE", 5*time.Minute),

		// Application defaults
		BaseURL:         getEnvString("BASE_URL", "http://localhost:8080"),
		MaxContentSize:  getEnvInt64("MAX_CONTENT_SIZE", 1<<20), // 1 MiB
		DefaultExpiry:   getEnvDuration("DEFAULT_EXPIRY", 72*time.Hour),
		MinExpiry:       getEnvDuration("MIN_EXPIRY", 10*time.Minute),
		MaxExpiry:       getEnvDuration("MAX_EXPIRY", 30*24*time.Hour),
		CleanupInterval: getEnvDuration("CLEANUP_INTERVAL", 5*time.Minute),

		// Rate limiting defaults
		PostRateLimit: getEnvInt("POST_RATE_LIMIT", 30),
		GetRateLimit:  getEnvInt("GET_RATE_LIMIT", 300),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535")
	}
	if c.MaxContentSize < 1 {
		return fmt.Errorf("MAX_CONTENT_SIZE must be positive")
	}
	if c.MinExpiry > c.MaxExpiry {
		return fmt.Errorf("MIN_EXPIRY cannot be greater than MAX_EXPIRY")
	}
	if c.DefaultExpiry < c.MinExpiry || c.DefaultExpiry > c.MaxExpiry {
		return fmt.Errorf("DEFAULT_EXPIRY must be between MIN_EXPIRY and MAX_EXPIRY")
	}
	return nil
}

// Addr returns the server address in host:port format.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func getEnvString(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvInt64(key string, defaultVal int64) int64 {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}

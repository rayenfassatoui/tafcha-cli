package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// Set required DATABASE_URL
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	defer os.Unsetenv("DATABASE_URL")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, 10*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.WriteTimeout)
	assert.Equal(t, int64(1<<20), cfg.MaxContentSize)
	assert.Equal(t, 72*time.Hour, cfg.DefaultExpiry)
	assert.Equal(t, 10*time.Minute, cfg.MinExpiry)
	assert.Equal(t, 30*24*time.Hour, cfg.MaxExpiry)
	assert.Equal(t, 30, cfg.PostRateLimit)
	assert.Equal(t, 300, cfg.GetRateLimit)
}

func TestLoad_CustomValues(t *testing.T) {
	envVars := map[string]string{
		"DATABASE_URL":     "postgres://custom/db",
		"PORT":             "3000",
		"HOST":             "127.0.0.1",
		"MAX_CONTENT_SIZE": "2097152",
		"DEFAULT_EXPIRY":   "24h",
		"POST_RATE_LIMIT":  "60",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "postgres://custom/db", cfg.DatabaseURL)
	assert.Equal(t, 3000, cfg.Port)
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, int64(2097152), cfg.MaxContentSize)
	assert.Equal(t, 24*time.Hour, cfg.DefaultExpiry)
	assert.Equal(t, 60, cfg.PostRateLimit)
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	os.Unsetenv("DATABASE_URL")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL is required")
}

func TestValidate_InvalidPort(t *testing.T) {
	cfg := &Config{
		DatabaseURL:   "postgres://localhost/test",
		Port:          70000,
		MaxContentSize: 1024,
		MinExpiry:     time.Minute,
		MaxExpiry:     time.Hour,
		DefaultExpiry: 30 * time.Minute,
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PORT must be between")
}

func TestValidate_InvalidExpiryRange(t *testing.T) {
	cfg := &Config{
		DatabaseURL:    "postgres://localhost/test",
		Port:           8080,
		MaxContentSize: 1024,
		MinExpiry:      time.Hour,
		MaxExpiry:      time.Minute, // Less than MinExpiry
		DefaultExpiry:  30 * time.Minute,
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MIN_EXPIRY cannot be greater than MAX_EXPIRY")
}

func TestAddr(t *testing.T) {
	cfg := &Config{Host: "localhost", Port: 3000}
	assert.Equal(t, "localhost:3000", cfg.Addr())
}

package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	RedisURL       string
	SessionTTL     time.Duration
	MigrateOnStart bool

	WorkerPollInterval time.Duration
	WorkerBatchSize    int
	WorkerMaxIntentos  int
	WorkerBackoffBase  time.Duration
	SunatFakeMode      string
}

func Load() (Config, error) {
	ttlStr := getenv("SESSION_TTL", "8h")
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		return Config{}, fmt.Errorf("SESSION_TTL: %w", err)
	}
	migrateOnStart, err := strconv.ParseBool(getenv("MIGRATE_ON_START", "true"))
	if err != nil {
		return Config{}, fmt.Errorf("MIGRATE_ON_START: %w", err)
	}

	pollStr := getenv("WORKER_POLL_INTERVAL", "1s")
	pollInterval, err := time.ParseDuration(pollStr)
	if err != nil {
		return Config{}, fmt.Errorf("WORKER_POLL_INTERVAL: %w", err)
	}
	batchSize, err := strconv.Atoi(getenv("WORKER_BATCH_SIZE", "10"))
	if err != nil {
		return Config{}, fmt.Errorf("WORKER_BATCH_SIZE: %w", err)
	}
	maxIntentos, err := strconv.Atoi(getenv("WORKER_MAX_INTENTOS", "5"))
	if err != nil {
		return Config{}, fmt.Errorf("WORKER_MAX_INTENTOS: %w", err)
	}
	backoffStr := getenv("WORKER_BACKOFF_BASE", "1s")
	backoffBase, err := time.ParseDuration(backoffStr)
	if err != nil {
		return Config{}, fmt.Errorf("WORKER_BACKOFF_BASE: %w", err)
	}

	cfg := Config{
		HTTPAddr:           getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		RedisURL:             os.Getenv("REDIS_URL"),
		SessionTTL:           ttl,
		MigrateOnStart:       migrateOnStart,
		WorkerPollInterval:   pollInterval,
		WorkerBatchSize:      batchSize,
		WorkerMaxIntentos:    maxIntentos,
		WorkerBackoffBase:    backoffBase,
		SunatFakeMode:        getenv("SUNAT_FAKE_MODE", "accept"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

// ValidateAPI exige Redis; el worker no lo usa.
func (c Config) ValidateAPI() error {
	if c.RedisURL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}
	return nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

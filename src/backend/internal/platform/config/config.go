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

	cfg := Config{
		HTTPAddr:       getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		RedisURL:       os.Getenv("REDIS_URL"),
		SessionTTL:     ttl,
		MigrateOnStart: migrateOnStart,
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.RedisURL == "" {
		return Config{}, fmt.Errorf("REDIS_URL is required")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

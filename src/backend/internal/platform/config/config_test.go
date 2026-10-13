package config_test

import (
	"testing"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/config"
)

func TestLoad_DefaultsAndEnv(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("DATABASE_URL", "postgres://ff:ff@localhost:5432/flowfacture?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://127.0.0.1:6379/0")
	t.Setenv("SESSION_TTL", "30m")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr=%q", cfg.HTTPAddr)
	}
	if cfg.DatabaseURL == "" {
		t.Fatal("DatabaseURL empty")
	}
	if cfg.RedisURL != "redis://127.0.0.1:6379/0" {
		t.Fatalf("RedisURL=%q", cfg.RedisURL)
	}
	if cfg.SessionTTL != 30*time.Minute {
		t.Fatalf("SessionTTL=%v want 30m", cfg.SessionTTL)
	}
}

func TestLoad_DefaultHTTPAddr(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("SESSION_TTL", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("default HTTPAddr=%q want :8080", cfg.HTTPAddr)
	}
}

func TestLoad_DefaultSessionTTL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("SESSION_TTL", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.SessionTTL != 8*time.Hour {
		t.Fatalf("default SessionTTL=%v want 8h", cfg.SessionTTL)
	}
}

func TestLoad_RequiresRedisURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("REDIS_URL", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when REDIS_URL missing")
	}
}

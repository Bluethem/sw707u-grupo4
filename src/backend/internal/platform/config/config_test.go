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

func TestLoad_AllowsMissingRedisURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("REDIS_URL", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.RedisURL != "" {
		t.Fatalf("RedisURL=%q want empty", cfg.RedisURL)
	}
}

func TestConfig_ValidateAPI_RequiresRedisURL(t *testing.T) {
	cfg := config.Config{DatabaseURL: "postgres://x", RedisURL: ""}
	if err := cfg.ValidateAPI(); err == nil {
		t.Fatal("expected error when REDIS_URL missing")
	}
}

func TestLoad_WorkerDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.WorkerPollInterval != time.Second {
		t.Fatalf("WorkerPollInterval=%v want 1s", cfg.WorkerPollInterval)
	}
	if cfg.WorkerBatchSize != 10 {
		t.Fatalf("WorkerBatchSize=%d want 10", cfg.WorkerBatchSize)
	}
	if cfg.WorkerMaxIntentos != 5 {
		t.Fatalf("WorkerMaxIntentos=%d want 5", cfg.WorkerMaxIntentos)
	}
	if cfg.WorkerBackoffBase != time.Second {
		t.Fatalf("WorkerBackoffBase=%v want 1s", cfg.WorkerBackoffBase)
	}
	if cfg.SunatFakeMode != "accept" {
		t.Fatalf("SunatFakeMode=%q want accept", cfg.SunatFakeMode)
	}
}

func TestLoad_WorkerEnvOverrides(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("WORKER_POLL_INTERVAL", "2s")
	t.Setenv("WORKER_BATCH_SIZE", "25")
	t.Setenv("WORKER_MAX_INTENTOS", "3")
	t.Setenv("WORKER_BACKOFF_BASE", "500ms")
	t.Setenv("SUNAT_FAKE_MODE", "reject")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.WorkerPollInterval != 2*time.Second {
		t.Fatalf("WorkerPollInterval=%v", cfg.WorkerPollInterval)
	}
	if cfg.WorkerBatchSize != 25 || cfg.WorkerMaxIntentos != 3 {
		t.Fatalf("batch=%d max=%d", cfg.WorkerBatchSize, cfg.WorkerMaxIntentos)
	}
	if cfg.WorkerBackoffBase != 500*time.Millisecond {
		t.Fatalf("WorkerBackoffBase=%v", cfg.WorkerBackoffBase)
	}
	if cfg.SunatFakeMode != "reject" {
		t.Fatalf("SunatFakeMode=%q", cfg.SunatFakeMode)
	}
}

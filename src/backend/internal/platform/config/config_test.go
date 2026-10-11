package config_test

import (
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/config"
)

func TestLoad_DefaultsAndEnv(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("DATABASE_URL", "postgres://ff:ff@localhost:5432/flowfacture?sslmode=disable")

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
}

func TestLoad_DefaultHTTPAddr(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("DATABASE_URL", "postgres://x")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("default HTTPAddr=%q want :8080", cfg.HTTPAddr)
	}
}

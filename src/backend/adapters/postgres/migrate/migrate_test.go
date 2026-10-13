package migrate

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestUp(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}
	if err := Up(ctx, pool); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	tables := []string{
		"productos_stock",
		"series_correlativos",
		"comprobantes",
		"comprobante_lineas",
		"outbox_envios",
	}
	for _, table := range tables {
		var exists bool
		err := pool.QueryRow(
			ctx,
			`SELECT to_regclass('public.' || $1) IS NOT NULL`,
			table,
		).Scan(&exists)
		if err != nil {
			t.Fatalf("check table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("table %s was not created", table)
		}
	}
}

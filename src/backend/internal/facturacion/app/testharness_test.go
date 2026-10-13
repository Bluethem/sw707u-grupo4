package app_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		os.Exit(m.Run())
	}

	ctx := context.Background()
	var err error
	testPool, err = pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create postgres pool:", err)
		os.Exit(1)
	}
	if err := testPool.Ping(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "ping postgres:", err)
		os.Exit(1)
	}
	if err := migrate.Up(ctx, testPool); err != nil {
		fmt.Fprintln(os.Stderr, "migrate postgres:", err)
		os.Exit(1)
	}

	code := m.Run()
	testPool.Close()
	os.Exit(code)
}

func seed(t *testing.T, stock int64) {
	t.Helper()
	if testPool == nil {
		t.Skip("DATABASE_URL no configurado")
	}

	ctx := context.Background()
	tx, err := testPool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin seed: %v", err)
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `
		TRUNCATE outbox_envios, comprobante_lineas, comprobantes,
		         series_correlativos, productos_stock RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate seed: %v", err)
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO productos_stock(empresa_id, producto_id, cantidad)
		VALUES ('e1', 'p1', $1)`,
		stock,
	); err != nil {
		t.Fatalf("insert stock seed: %v", err)
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO series_correlativos(empresa_id, tipo, serie, siguiente)
		VALUES ('e1', 'boleta', 'B001', 1)`); err != nil {
		t.Fatalf("insert serie seed: %v", err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatalf("commit seed: %v", err)
	}
}

func queryInt64(t *testing.T, query string, args ...any) int64 {
	t.Helper()
	var got int64
	if err := testPool.QueryRow(context.Background(), query, args...).Scan(&got); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return got
}

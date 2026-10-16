package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres/migrate"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/storage"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/sunat"
	factapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/app"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		log.Printf("worker: %v", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("postgres ping: %w", err)
	}
	if cfg.MigrateOnStart {
		if err := migrate.Up(ctx, pool); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	procesador := &factapp.ProcesadorEnvio{
		Outbox:       postgres.NewOutboxConsumer(pool),
		Comprobantes: postgres.NewComprobanteQuery(pool),
		Resumenes:    postgres.NewResumenRepository(pool),
		Storage:      storage.NewFake(),
		Sunat:        sunat.NewFake(sunat.Mode(cfg.SunatFakeMode)),
		BatchSize:    cfg.WorkerBatchSize,
		MaxIntentos:  cfg.WorkerMaxIntentos,
		BackoffBase:  cfg.WorkerBackoffBase,
	}

	runCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("worker started (poll=%s batch=%d max_intentos=%d sunat_fake=%s)",
		cfg.WorkerPollInterval, cfg.WorkerBatchSize, cfg.WorkerMaxIntentos, cfg.SunatFakeMode)

	for {
		if runCtx.Err() != nil {
			break
		}
		if err := procesador.RunOnce(runCtx); err != nil {
			log.Printf("worker run once: %v", err)
		}
		select {
		case <-runCtx.Done():
			log.Println("worker shutting down")
			return nil
		case <-time.After(cfg.WorkerPollInterval):
		}
	}
	return nil
}

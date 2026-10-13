package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/firma"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/pdf"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres/migrate"
	redisadapter "github.com/Bluethem/sw707u-grupo4/src/backend/adapters/redis"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/storage"
	factapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/app"
	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/config"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/httpapi"
	"github.com/jackc/pgx/v5/pgxpool"
	redislib "github.com/redis/go-redis/v9"
)

func main() {
	if err := run(); err != nil {
		log.Printf("api: %v", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	if err := cfg.ValidateAPI(); err != nil {
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

	redisOptions, err := redislib.ParseURL(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("redis url: %w", err)
	}
	redisClient := redislib.NewClient(redisOptions)
	defer redisClient.Close()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}

	empresas := postgres.NewEmpresaRepository(pool)
	usuarios := postgres.NewUsuarioRepository(pool)
	series := postgres.NewSeriesGestionRepository(pool)
	clientes := postgres.NewClienteRepository(pool)
	productos := postgres.NewProductoRepository(pool)
	gestion := gestionapp.NewService(empresas, usuarios, series, clientes, productos, productos)
	sessions := redisadapter.NewSessionStore(redisClient, cfg.SessionTTL)

	comprobantes := postgres.NewComprobante()
	emisor := &factapp.Emisor{
		UoW:       postgres.NewUnitOfWork(pool),
		Stock:     postgres.NewStock(),
		Productos: postgres.NewProductoEmisionConsulta(pool),
		Series:    postgres.NewSeries(),
		Comps:   comprobantes,
		Outbox:  postgres.NewOutbox(),
		Firma:   firma.Fake{},
		PDF:     pdf.Fake{},
		Storage: storage.NewFake(),
	}
	query := postgres.NewComprobanteQuery(pool)
	handler := httpapi.NewAPIHandler(httpapi.Dependencies{
		Gestion: gestion, Usuarios: usuarios, Sessions: sessions,
		Emisor: emisor, Query: query,
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	shutdownDone := make(chan error, 1)
	go func() {
		<-shutdownCtx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdownDone <- server.Shutdown(ctx)
	}()

	log.Printf("api listening on %s", cfg.HTTPAddr)
	serveErr := server.ListenAndServe()
	stop()
	shutdownErr := <-shutdownDone
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		return fmt.Errorf("listen: %w", serveErr)
	}
	if shutdownErr != nil {
		return fmt.Errorf("shutdown: %w", shutdownErr)
	}
	return nil
}

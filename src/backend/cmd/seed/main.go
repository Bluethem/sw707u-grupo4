package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres/migrate"
	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultRUC           = "20100070970"
	defaultRazonSocial   = "FlowFacture Demo SAC"
	defaultAdminEmail    = "admin@flowfacture.local"
	defaultAdminPassword = "Admin123!"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("seed: %v", err)
	}
}

func run() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL es requerida")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("conectar a Postgres: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping a Postgres: %w", err)
	}
	if err := migrate.Up(ctx, pool); err != nil {
		return fmt.Errorf("ejecutar migraciones: %w", err)
	}

	service := gestionapp.NewService(
		postgres.NewEmpresaRepository(pool),
		postgres.NewUsuarioRepository(pool),
		postgres.NewSeriesGestionRepository(pool),
		postgres.NewClienteRepository(pool),
	)
	ruc := getenv("SEED_RUC", defaultRUC)
	razonSocial := getenv("SEED_RAZON_SOCIAL", defaultRazonSocial)
	adminEmail := getenv("SEED_ADMIN_EMAIL", defaultAdminEmail)
	adminPassword := getenv("SEED_ADMIN_PASSWORD", defaultAdminPassword)

	empresa, err := service.CrearEmpresa(ctx, gestionapp.CrearEmpresaInput{
		RUC: ruc, RazonSocial: razonSocial, DomicilioFiscal: "Lima", Regimen: "general",
	})
	if err != nil {
		return fmt.Errorf("crear empresa: %w", err)
	}
	if _, err := service.CrearUsuario(ctx, empresa.ID, gestionapp.CrearUsuarioInput{
		Email: adminEmail, Password: adminPassword, Nombre: "Administrador",
		Rol: gestiondomain.RolAdministrador,
	}); err != nil {
		return fmt.Errorf("crear administrador: %w", err)
	}
	for _, serie := range []gestionapp.CrearSerieInput{
		{Tipo: gestiondomain.TipoBoleta, Serie: "B001", Siguiente: 1},
		{Tipo: gestiondomain.TipoFactura, Serie: "F001", Siguiente: 1},
	} {
		if _, err := service.CrearSerie(ctx, empresa.ID, serie); err != nil {
			return fmt.Errorf("crear serie %s: %w", serie.Serie, err)
		}
	}

	fmt.Printf("Seed completado\nEmpresa: %s (%s)\nEmail: %s\nPassword: %s\nSeries: B001, F001\n",
		razonSocial, ruc, adminEmail, adminPassword)
	return nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

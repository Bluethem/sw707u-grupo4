package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres/migrate"
	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

const (
	defaultRUC            = "20100070970"
	defaultRazonSocial    = "FlowFacture Demo SAC"
	defaultAdminEmail     = "admin@flowfacture.local"
	defaultAdminPassword  = "Admin123!"
	defaultProductoCodigo = "SKU-1"
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

	empresas := postgres.NewEmpresaRepository(pool)
	usuarios := postgres.NewUsuarioRepository(pool)
	series := postgres.NewSeriesGestionRepository(pool)
	clientes := postgres.NewClienteRepository(pool)
	productos := postgres.NewProductoRepository(pool)
	service := gestionapp.NewService(empresas, usuarios, series, clientes, productos, productos)

	ruc := getenv("SEED_RUC", defaultRUC)
	razonSocial := getenv("SEED_RAZON_SOCIAL", defaultRazonSocial)
	adminEmail := getenv("SEED_ADMIN_EMAIL", defaultAdminEmail)
	adminPassword := getenv("SEED_ADMIN_PASSWORD", defaultAdminPassword)

	empresaID, createdFresh, err := ensureEmpresa(ctx, pool, service, ruc, razonSocial)
	if err != nil {
		return err
	}

	if err := ensureAdmin(ctx, service, usuarios, empresaID, adminEmail, adminPassword); err != nil {
		return err
	}
	if err := ensureSeries(ctx, service, empresaID); err != nil {
		return err
	}
	if err := ensureDemoCatalog(ctx, service, empresaID); err != nil {
		return err
	}

	if createdFresh {
		fmt.Printf("Seed completado\nEmpresa: %s (%s)\nEmail: %s\nPassword: %s\nSeries: B001, F001\nProducto: %s\n",
			razonSocial, ruc, adminEmail, adminPassword, defaultProductoCodigo)
	} else {
		fmt.Printf("Seed ya aplicado / completado\nEmpresaID: %s\nEmail: %s\nPassword: %s\nProducto: %s\n",
			empresaID, adminEmail, adminPassword, defaultProductoCodigo)
	}
	return nil
}

func ensureEmpresa(ctx context.Context, pool *pgxpool.Pool, service *gestionapp.Service, ruc, razonSocial string) (string, bool, error) {
	var id string
	err := pool.QueryRow(ctx, `SELECT id FROM empresas WHERE ruc = $1`, ruc).Scan(&id)
	if err == nil {
		return id, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", false, fmt.Errorf("buscar empresa: %w", err)
	}

	empresa, err := service.CrearEmpresa(ctx, gestionapp.CrearEmpresaInput{
		RUC: ruc, RazonSocial: razonSocial, DomicilioFiscal: "Lima", Regimen: "general",
	})
	if err == nil {
		return empresa.ID, true, nil
	}
	if !errors.Is(err, gestiondomain.ErrConflicto) {
		return "", false, fmt.Errorf("crear empresa: %w", err)
	}
	if err := pool.QueryRow(ctx, `SELECT id FROM empresas WHERE ruc = $1`, ruc).Scan(&id); err != nil {
		return "", false, fmt.Errorf("empresa en conflicto pero no legible: %w", err)
	}
	return id, false, nil
}

func ensureAdmin(ctx context.Context, service *gestionapp.Service, usuarios gestionapp.UsuarioRepository, empresaID, email, password string) error {
	_, err := usuarios.ObtenerPorEmail(ctx, gestiondomain.NormalizarEmail(email))
	if err == nil {
		return nil
	}
	if !errors.Is(err, gestiondomain.ErrNoEncontrado) {
		return fmt.Errorf("buscar admin: %w", err)
	}
	if _, err := service.CrearUsuario(ctx, empresaID, gestionapp.CrearUsuarioInput{
		Email: email, Password: password, Nombre: "Administrador",
		Rol: gestiondomain.RolAdministrador,
	}); err != nil && !errors.Is(err, gestiondomain.ErrConflicto) {
		return fmt.Errorf("crear administrador: %w", err)
	}
	return nil
}

func ensureSeries(ctx context.Context, service *gestionapp.Service, empresaID string) error {
	for _, serie := range []gestionapp.CrearSerieInput{
		{Tipo: gestiondomain.TipoBoleta, Serie: "B001", Siguiente: 1},
		{Tipo: gestiondomain.TipoFactura, Serie: "F001", Siguiente: 1},
	} {
		if _, err := service.CrearSerie(ctx, empresaID, serie); err != nil && !errors.Is(err, gestiondomain.ErrConflicto) {
			return fmt.Errorf("crear serie %s: %w", serie.Serie, err)
		}
	}
	return nil
}

func ensureDemoCatalog(ctx context.Context, service *gestionapp.Service, empresaID string) error {
	items, _, err := service.ListarProductos(ctx, empresaID, gestionapp.ListarProductosFilter{
		Q: defaultProductoCodigo, Limit: 5,
	})
	if err != nil {
		return fmt.Errorf("listar productos: %w", err)
	}
	for _, p := range items {
		if p.Producto.Codigo == defaultProductoCodigo {
			return nil
		}
	}
	stock := int64(100)
	if _, err := service.CrearProducto(ctx, empresaID, gestionapp.CrearProductoInput{
		Codigo:        defaultProductoCodigo,
		Descripcion:   "Producto demo",
		Unidad:        "NIU",
		Precio:        decimal.RequireFromString("10.00"),
		Afectacion:    "gravado",
		ControlaStock: true,
		StockInicial:  &stock,
	}); err != nil && !errors.Is(err, gestiondomain.ErrConflicto) {
		return fmt.Errorf("crear producto demo: %w", err)
	}
	return nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

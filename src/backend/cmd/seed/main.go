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
)

const (
	defaultRUC            = "20100070970"
	defaultRazonSocial    = "FlowFacture Demo SAC"
	defaultAdminEmail     = "admin@flowfacture.local"
	defaultAdminPassword  = "Admin123!"
	defaultCajeroEmail    = "cajero@flowfacture.local"
	defaultCajeroPassword = "Cajero123!"
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
	configCatalogo := postgres.NewConfigCatalogoRepository(pool)
	service := gestionapp.NewService(empresas, usuarios, series, clientes, productos, productos, configCatalogo)

	ruc := getenv("SEED_RUC", defaultRUC)
	razonSocial := getenv("SEED_RAZON_SOCIAL", defaultRazonSocial)
	adminEmail := getenv("SEED_ADMIN_EMAIL", defaultAdminEmail)
	adminPassword := getenv("SEED_ADMIN_PASSWORD", defaultAdminPassword)
	cajeroEmail := getenv("SEED_CAJERO_EMAIL", defaultCajeroEmail)
	cajeroPassword := getenv("SEED_CAJERO_PASSWORD", defaultCajeroPassword)

	empresaID, createdFresh, err := ensureEmpresa(ctx, pool, service, ruc, razonSocial)
	if err != nil {
		return err
	}

	if err := ensureUsuario(ctx, service, usuarios, empresaID, adminEmail, adminPassword, "Administrador", gestiondomain.RolAdministrador); err != nil {
		return err
	}
	if err := ensureUsuario(ctx, service, usuarios, empresaID, cajeroEmail, cajeroPassword, "Cajero demo", gestiondomain.RolCajero); err != nil {
		return err
	}
	if err := ensureSeries(ctx, service, empresaID); err != nil {
		return err
	}
	if err := ensureConfigCatalogo(ctx, pool, empresaID); err != nil {
		return err
	}
	if err := ensureDemoCatalog(ctx, service, empresaID); err != nil {
		return err
	}
	if err := ensureDemoClientes(ctx, service, empresaID); err != nil {
		return err
	}

	ventas, err := ensureDemoVentas(ctx, pool, empresaID)
	if err != nil {
		return err
	}

	fmt.Printf("Seed completado\nEmpresa: %s (%s)\nAdmin: %s / %s\nCajero: %s / %s\nSeries: B001, F001\nProductos: %d (incluye %s)\nClientes: %d\n",
		razonSocial, ruc, adminEmail, adminPassword, cajeroEmail, cajeroPassword,
		len(demoProductos), defaultProductoCodigo, len(demoClientes))
	if createdFresh {
		fmt.Println("Empresa creada en esta ejecución.")
	} else {
		fmt.Printf("EmpresaID existente: %s\n", empresaID)
	}
	if ventas.Aplicado {
		fmt.Printf("Comprobantes demo: %d boletas + %d facturas (hoy: %d vigentes / %d anulados)\n",
			ventas.Boletas, ventas.Facturas, ventas.HoyVigentes, ventas.HoyAnulados)
		fmt.Printf("Resúmenes diarios: %d aceptados (días -%v). Hoy y ayer quedan libres para Generar.\n",
			len(ventas.ResumenDias), ventas.ResumenDias)
		fmt.Println("En /admin/reportes usa el rango de los últimos 30 días para ver el histórico.")
	} else {
		fmt.Println("Ventas demo ya aplicadas (no se duplicaron comprobantes).")
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
		RUC: ruc, RazonSocial: razonSocial, DomicilioFiscal: "Av. Javier Prado 1234, Lima", Regimen: "general",
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

func ensureUsuario(
	ctx context.Context,
	service *gestionapp.Service,
	usuarios gestionapp.UsuarioRepository,
	empresaID, email, password, nombre string,
	rol gestiondomain.Rol,
) error {
	_, err := usuarios.ObtenerPorEmail(ctx, gestiondomain.NormalizarEmail(email))
	if err == nil {
		return nil
	}
	if !errors.Is(err, gestiondomain.ErrNoEncontrado) {
		return fmt.Errorf("buscar usuario %s: %w", email, err)
	}
	if _, err := service.CrearUsuario(ctx, empresaID, gestionapp.CrearUsuarioInput{
		Email: email, Password: password, Nombre: nombre, Rol: rol,
	}); err != nil && !errors.Is(err, gestiondomain.ErrConflicto) {
		return fmt.Errorf("crear usuario %s: %w", email, err)
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

func ensureConfigCatalogo(ctx context.Context, pool *pgxpool.Pool, empresaID string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO config_catalogo (empresa_id)
		VALUES ($1)
		ON CONFLICT (empresa_id) DO NOTHING`,
		empresaID,
	)
	if err != nil {
		return fmt.Errorf("seed config_catalogo: %w", err)
	}
	return nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

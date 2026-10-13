package postgres

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres/migrate"
	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestClienteRepository_CrearObtenerListar(t *testing.T) {
	ctx := context.Background()
	pool := newClienteTestPool(t, ctx)
	repo := NewClienteRepository(pool)
	empresaID := insertEmpresaClienteTest(t, ctx, pool)

	c1, err := repo.Crear(ctx, gestiondomain.Cliente{
		ID: uuid.NewString(), EmpresaID: empresaID, TipoDoc: gestiondomain.TipoDocDNI,
		Numero: "12345678", RazonSocial: "Ana López", Direccion: "Lima 1", Activo: true,
	})
	if err != nil {
		t.Fatalf("crear c1: %v", err)
	}
	if c1.CreadoEn.IsZero() || c1.ActualizadoEn.IsZero() {
		t.Fatalf("timestamps esperados: %#v", c1)
	}

	c2, err := repo.Crear(ctx, gestiondomain.Cliente{
		ID: uuid.NewString(), EmpresaID: empresaID, TipoDoc: gestiondomain.TipoDocDNI,
		Numero: "87654321", RazonSocial: "Bruno Vega", Direccion: "Lima 2", Activo: true,
	})
	if err != nil {
		t.Fatalf("crear c2: %v", err)
	}

	got, err := repo.Obtener(ctx, empresaID, c1.ID)
	if err != nil || got.Numero != "12345678" {
		t.Fatalf("obtener: %#v, err=%v", got, err)
	}
	if _, err := repo.Obtener(ctx, uuid.NewString(), c1.ID); !errors.Is(err, gestiondomain.ErrNoEncontrado) {
		t.Fatalf("cross-tenant debio ser ErrNoEncontrado: %v", err)
	}

	items, total, err := repo.Listar(ctx, empresaID, gestionapp.ListarClientesFilter{Limit: 10})
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("listar todos: total=%d len=%d err=%v", total, len(items), err)
	}

	items, total, err = repo.Listar(ctx, empresaID, gestionapp.ListarClientesFilter{Q: "Bruno", Limit: 50})
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != c2.ID {
		t.Fatalf("listar Q: total=%d items=%#v err=%v", total, items, err)
	}

	activo := true
	items, total, err = repo.Listar(ctx, empresaID, gestionapp.ListarClientesFilter{Activo: &activo, Limit: 50})
	if err != nil || total != 2 {
		t.Fatalf("filtro activo: total=%d err=%v", total, err)
	}

	if _, err := repo.Crear(ctx, gestiondomain.Cliente{
		ID: uuid.NewString(), EmpresaID: empresaID, TipoDoc: gestiondomain.TipoDocDNI,
		Numero: "12345678", RazonSocial: "Duplicado", Direccion: "X", Activo: true,
	}); !errors.Is(err, gestiondomain.ErrConflicto) {
		t.Fatalf("duplicado debio ser ErrConflicto: %v", err)
	}

	updated, err := repo.Actualizar(ctx, gestiondomain.Cliente{
		ID: c1.ID, EmpresaID: empresaID, TipoDoc: c1.TipoDoc, Numero: c1.Numero,
		RazonSocial: "Ana Actualizada", Direccion: c1.Direccion, Correo: c1.Correo, Activo: false,
	})
	if err != nil || !updated.ActualizadoEn.After(c1.ActualizadoEn) || updated.Activo {
		t.Fatalf("actualizar: %#v err=%v", updated, err)
	}
}

func newClienteTestPool(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL no configurado")
	}
	adminPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("crear pool administrativo: %v", err)
	}
	t.Cleanup(adminPool.Close)
	if err := adminPool.Ping(ctx); err != nil {
		t.Fatalf("conectar postgres: %v", err)
	}

	esquema := "cliente_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+esquema); err != nil {
		t.Fatalf("crear esquema aislado: %v", err)
	}
	t.Cleanup(func() {
		_, _ = adminPool.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+esquema+" CASCADE")
	})

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("leer DATABASE_URL: %v", err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = esquema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("crear pool aislado: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := migrate.Up(ctx, pool); err != nil {
		t.Fatalf("migrar: %v", err)
	}
	return pool
}

func insertEmpresaClienteTest(t *testing.T, ctx context.Context, pool *pgxpool.Pool) string {
	t.Helper()
	id := uuid.NewString()
	_, err := pool.Exec(ctx, `
		INSERT INTO empresas (id, ruc, razon_social, domicilio_fiscal, regimen, activa)
		VALUES ($1, $2, $3, '', 'general', true)`,
		id, "20100070970", "Empresa Test",
	)
	if err != nil {
		t.Fatalf("insertar empresa: %v", err)
	}
	return id
}

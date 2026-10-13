package postgres

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres/migrate"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

func TestComprobanteQuery_GetByIDAndList(t *testing.T) {
	ctx := context.Background()
	pool := newComprobanteTestPool(t, ctx)
	repo := NewComprobante()
	query := NewComprobanteQuery(pool)
	uow := NewUnitOfWork(pool)

	insertar := func(c *domain.Comprobante, creadoEn time.Time) {
		t.Helper()
		if err := uow.WithinTransaction(ctx, func(ctx context.Context, tx ports.Tx) error {
			return repo.Insert(ctx, tx, c)
		}); err != nil {
			t.Fatalf("insertar comprobante %s: %v", c.ID, err)
		}
		if _, err := pool.Exec(ctx, `UPDATE comprobantes SET creado_en = $2 WHERE id = $1`, c.ID, creadoEn); err != nil {
			t.Fatalf("fijar fecha de %s: %v", c.ID, err)
		}
	}

	base := time.Date(2026, time.October, 10, 12, 0, 0, 0, time.UTC)
	antiguo := comprobanteDePrueba("comp-antiguo", "empresa-a", domain.TipoBoleta, domain.EstadoPendiente)
	reciente := comprobanteDePrueba("comp-reciente", "empresa-a", domain.TipoFactura, domain.EstadoAceptado)
	otroTenant := comprobanteDePrueba("comp-otro", "empresa-b", domain.TipoBoleta, domain.EstadoPendiente)
	insertar(antiguo, base)
	insertar(reciente, base.AddDate(0, 0, 1))
	insertar(otroTenant, base.AddDate(0, 0, 2))

	obtenido, err := query.GetByID(ctx, "empresa-a", reciente.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if obtenido == nil || obtenido.ID != reciente.ID || len(obtenido.Lineas) != 1 {
		t.Fatalf("detalle inesperado: %#v", obtenido)
	}
	noVisible, err := query.GetByID(ctx, "empresa-b", reciente.ID)
	if err != nil || noVisible != nil {
		t.Fatalf("GetByID cross-tenant = %#v, err = %v", noVisible, err)
	}

	items, total, err := query.List(ctx, "empresa-a", ports.ListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 2 || len(items) != 2 || items[0].ID != reciente.ID || items[1].ID != antiguo.ID {
		t.Fatalf("lista inesperada: total=%d items=%#v", total, items)
	}
	if items[0].Lineas != nil || items[1].Lineas != nil {
		t.Fatalf("List debe omitir lineas: %#v", items)
	}

	diaReciente := base.AddDate(0, 0, 1).Truncate(24 * time.Hour)
	items, total, err = query.List(ctx, "empresa-a", ports.ListFilter{
		FechaDesde: &diaReciente,
		FechaHasta: &diaReciente,
		Tipo:       string(domain.TipoFactura),
		Estado:     string(domain.EstadoAceptado),
		Limit:      1,
	})
	if err != nil {
		t.Fatalf("List con filtros: %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].ID != reciente.ID {
		t.Fatalf("lista filtrada inesperada: total=%d items=%#v", total, items)
	}
}

func newComprobanteTestPool(t *testing.T, ctx context.Context) *pgxpool.Pool {
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

	esquema := "comprobante_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
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

func comprobanteDePrueba(id, empresaID string, tipo domain.TipoComprobante, estado domain.EstadoComprobante) *domain.Comprobante {
	return &domain.Comprobante{
		ID:             id,
		EmpresaID:      empresaID,
		Tipo:           tipo,
		Serie:          "B001",
		Numero:         1,
		Estado:         estado,
		ValorVenta:     decimal.RequireFromString("100.00"),
		IGV:            decimal.RequireFromString("18.00"),
		Total:          decimal.RequireFromString("118.00"),
		IdempotencyKey: "key-" + id,
		Lineas: []domain.LineaComprobante{{
			ProductoID: "producto-1",
			Cantidad:   1,
			PrecioUnit: decimal.RequireFromString("118.00"),
			Afectacion: domain.AfectacionGravado,
			ValorVenta: decimal.RequireFromString("100.00"),
			IGV:        decimal.RequireFromString("18.00"),
			Total:      decimal.RequireFromString("118.00"),
		}},
	}
}

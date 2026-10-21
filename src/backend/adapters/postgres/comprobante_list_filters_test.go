package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
)

func TestComprobanteQuery_ListSerieNumeroQ(t *testing.T) {
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

	base := time.Date(2026, time.October, 15, 10, 0, 0, 0, time.UTC)
	b001 := comprobanteDePrueba("comp-b001", "empresa-a", domain.TipoBoleta, domain.EstadoPendiente)
	b001.Serie = "B001"
	b001.Numero = 42
	b001.Cliente = &domain.DatosCliente{
		TipoDoc: "6", Numero: "20123456789", RazonSocial: "ACME COMERCIAL SAC",
	}
	f001 := comprobanteDePrueba("comp-f001", "empresa-a", domain.TipoFactura, domain.EstadoAceptado)
	f001.Serie = "F001"
	f001.Numero = 7
	f001.Cliente = &domain.DatosCliente{
		TipoDoc: "1", Numero: "44556677", RazonSocial: "JUAN PEREZ",
	}
	insertar(b001, base)
	insertar(f001, base.Add(time.Hour))

	numero42 := int64(42)
	items, total, err := query.List(ctx, "empresa-a", ports.ListFilter{Serie: "B001"})
	if err != nil {
		t.Fatalf("List serie: %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].ID != b001.ID {
		t.Fatalf("filtro serie: total=%d items=%#v", total, items)
	}

	items, total, err = query.List(ctx, "empresa-a", ports.ListFilter{Serie: "b001"})
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != b001.ID {
		t.Fatalf("serie case-insensitive en adapter (upper en handler): total=%d err=%v", total, err)
	}

	items, total, err = query.List(ctx, "empresa-a", ports.ListFilter{Numero: &numero42})
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != b001.ID {
		t.Fatalf("filtro numero: total=%d items=%#v err=%v", total, items, err)
	}

	items, total, err = query.List(ctx, "empresa-a", ports.ListFilter{Q: "acme"})
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != b001.ID {
		t.Fatalf("filtro q razon: total=%d items=%#v err=%v", total, items, err)
	}

	items, total, err = query.List(ctx, "empresa-a", ports.ListFilter{Q: "4455"})
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != f001.ID {
		t.Fatalf("filtro q numero cliente: total=%d items=%#v err=%v", total, items, err)
	}
}

package postgres

import (
	"context"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
)

func TestComprobanteQuery_UpdateEstado(t *testing.T) {
	ctx := context.Background()
	pool := newComprobanteTestPool(t, ctx)
	repo := NewComprobante()
	query := NewComprobanteQuery(pool)
	uow := NewUnitOfWork(pool)

	c := comprobanteDePrueba("c-upd-est", "empresa-a", domain.TipoBoleta, domain.EstadoPendiente)
	if err := uow.WithinTransaction(ctx, func(ctx context.Context, tx ports.Tx) error {
		return repo.Insert(ctx, tx, c)
	}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	if err := query.UpdateEstado(ctx, c.ID, domain.EstadoAceptado); err != nil {
		t.Fatalf("UpdateEstado: %v", err)
	}

	got, err := query.GetByComprobanteID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByComprobanteID: %v", err)
	}
	if got.Estado != domain.EstadoAceptado {
		t.Fatalf("estado = %q; want ACEPTADO", got.Estado)
	}
}

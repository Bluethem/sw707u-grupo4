package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	"github.com/google/uuid"
)

type Resumenes struct {
	UoW    ports.UnitOfWork
	Store  ports.ResumenStore
	Outbox ports.OutboxPublisher
	NewID  func() string
}

func (s *Resumenes) GenerarResumen(ctx context.Context, empresaID string, fecha time.Time) (*domain.ResumenDiario, error) {
	if s.UoW == nil || s.Store == nil || s.Outbox == nil {
		return nil, fmt.Errorf("servicio resumenes no configurado")
	}
	if fecha.IsZero() {
		return nil, domain.ErrFechaResumenInvalida
	}
	fecha = time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 0, 0, 0, 0, time.UTC)

	var createdID string
	err := s.UoW.WithinTransaction(ctx, func(ctx context.Context, tx ports.Tx) error {
		boletas, err := s.Store.BoletasInformables(ctx, tx, empresaID, fecha)
		if err != nil {
			return err
		}
		if len(boletas) == 0 {
			return domain.ErrResumenSinBoletas
		}
		correlativo, err := s.Store.NextCorrelativo(ctx, tx, empresaID, fecha)
		if err != nil {
			return err
		}
		id := s.newID()
		resumen := &domain.ResumenDiario{
			ID:          id,
			EmpresaID:   empresaID,
			Fecha:       fecha,
			Correlativo: correlativo,
			Estado:      domain.ResumenPendiente,
			Comprobantes: make([]domain.ResumenComprobanteItem, 0, len(boletas)),
		}
		for _, b := range boletas {
			resumen.Comprobantes = append(resumen.Comprobantes, domain.ResumenComprobanteItem{
				ComprobanteID: b.ID,
				Anulado:       b.Estado == domain.EstadoPendienteBaja,
			})
		}
		if err := s.Store.Insert(ctx, tx, resumen); err != nil {
			return err
		}
		if err := s.Outbox.EnqueueResumen(ctx, tx, id); err != nil {
			return err
		}
		createdID = id
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Store.GetByID(ctx, empresaID, createdID)
}

func (s *Resumenes) Listar(ctx context.Context, empresaID string) ([]domain.ResumenDiario, error) {
	if s.Store == nil {
		return nil, fmt.Errorf("servicio resumenes no configurado")
	}
	return s.Store.List(ctx, empresaID)
}

func (s *Resumenes) Obtener(ctx context.Context, empresaID, id string) (*domain.ResumenDiario, error) {
	if s.Store == nil {
		return nil, fmt.Errorf("servicio resumenes no configurado")
	}
	r, err := s.Store.GetByID(ctx, empresaID, id)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, domain.ErrResumenNoEncontrado
	}
	return r, nil
}

func (s *Resumenes) newID() string {
	if s.NewID != nil {
		return s.NewID()
	}
	return uuid.NewString()
}

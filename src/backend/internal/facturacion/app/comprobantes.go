package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
)

type Comprobantes struct {
	UoW         ports.UnitOfWork
	Query       ports.ComprobanteQuery
	Lifecycle   ports.ComprobanteLifecycle
	Outbox      ports.OutboxPublisher
	OutboxRead  ports.EnvioOutboxReader
}

func (s *Comprobantes) ReenviarComprobante(
	ctx context.Context,
	empresaID, id string,
) (*domain.Comprobante, error) {
	if s.Query == nil || s.Outbox == nil || s.UoW == nil {
		return nil, errServicioIncompleto()
	}
	c, err := s.Query.GetByID(ctx, empresaID, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, domain.ErrComprobanteNoEncontrado
	}
	if !puedeReenviar(c, s.outboxEnvio(ctx, id)) {
		return nil, domain.ErrOperacionNoPermitida
	}
	err = s.UoW.WithinTransaction(ctx, func(ctx context.Context, tx ports.Tx) error {
		if err := s.Lifecycle.PonerEstadoPendiente(ctx, tx, id); err != nil {
			return err
		}
		return s.Outbox.RequeueEnvio(ctx, tx, id)
	})
	if err != nil {
		return nil, err
	}
	return s.Query.GetByID(ctx, empresaID, id)
}

func (s *Comprobantes) SolicitarBaja(
	ctx context.Context,
	empresaID, id, motivo string,
) (*domain.Comprobante, error) {
	motivo = strings.TrimSpace(motivo)
	if motivo == "" {
		return nil, domain.ErrMotivoBajaInvalido
	}
	if s.Query == nil || s.Outbox == nil || s.UoW == nil || s.Lifecycle == nil {
		return nil, errServicioIncompleto()
	}
	c, err := s.Query.GetByID(ctx, empresaID, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, domain.ErrComprobanteNoEncontrado
	}
	if !puedeSolicitarBaja(c) {
		return nil, domain.ErrOperacionNoPermitida
	}
	err = s.UoW.WithinTransaction(ctx, func(ctx context.Context, tx ports.Tx) error {
		if err := s.Lifecycle.MarcarPendienteBaja(ctx, tx, empresaID, id, motivo); err != nil {
			return err
		}
		if c.Tipo == domain.TipoFactura {
			return s.Outbox.EnqueueBaja(ctx, tx, id)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Query.GetByID(ctx, empresaID, id)
}

type outboxEnvioInfo struct {
	estado      string
	ultimoError string
	ok          bool
}

func (s *Comprobantes) outboxEnvio(ctx context.Context, comprobanteID string) outboxEnvioInfo {
	if s.OutboxRead == nil {
		return outboxEnvioInfo{}
	}
	estado, ultimo, ok, err := s.OutboxRead.LatestEnvioJob(ctx, comprobanteID)
	if err != nil || !ok {
		return outboxEnvioInfo{}
	}
	return outboxEnvioInfo{estado: estado, ultimoError: ultimo, ok: true}
}

func puedeReenviar(c *domain.Comprobante, outbox outboxEnvioInfo) bool {
	switch c.Estado {
	case domain.EstadoAceptado, domain.EstadoRechazado, domain.EstadoAnulada, domain.EstadoPendienteBaja:
		return false
	case domain.EstadoPendiente, domain.EstadoEnviado:
		return true
	}
	if outbox.ok {
		if outbox.estado == "failed" {
			return true
		}
		if outbox.estado == "pending" && strings.TrimSpace(outbox.ultimoError) != "" {
			return true
		}
	}
	return false
}

func puedeSolicitarBaja(c *domain.Comprobante) bool {
	switch c.Tipo {
	case domain.TipoFactura:
		return c.Estado == domain.EstadoAceptado
	case domain.TipoBoleta:
		return c.Estado == domain.EstadoAceptado ||
			c.Estado == domain.EstadoPendiente ||
			c.Estado == domain.EstadoEnviado
	default:
		return false
	}
}

func errServicioIncompleto() error {
	return fmt.Errorf("servicio comprobantes no configurado")
}

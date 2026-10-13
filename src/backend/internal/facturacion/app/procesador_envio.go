package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
)

const defaultClaimBatch = 10

type ProcesadorEnvio struct {
	Outbox       ports.OutboxConsumer
	Comprobantes ports.ComprobanteEnvioStore
	Storage      ports.StoragePort
	Sunat        ports.SunatGateway
	BatchSize    int
	MaxIntentos  int
	BackoffBase  time.Duration
	Now          func() time.Time
}

func CalcularBackoff(intentos int, base time.Duration) time.Duration {
	if base <= 0 {
		base = time.Second
	}
	if intentos < 0 {
		intentos = 0
	}
	exp := 1 << intentos
	if exp > 60 {
		exp = 60
	}
	return time.Duration(exp) * base
}

func (p *ProcesadorEnvio) RunOnce(ctx context.Context) error {
	limit := p.BatchSize
	if limit <= 0 {
		limit = defaultClaimBatch
	}
	jobs, err := p.Outbox.Claim(ctx, limit)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if err := p.procesarJob(ctx, job); err != nil {
			return err
		}
	}
	return nil
}

func (p *ProcesadorEnvio) procesarJob(ctx context.Context, job ports.OutboxJob) error {
	c, err := p.Comprobantes.GetByComprobanteID(ctx, job.ComprobanteID)
	if err != nil {
		return err
	}
	if c == nil {
		return p.Outbox.Fail(ctx, job.ID, "comprobante no encontrado")
	}

	xml, err := p.obtenerXML(ctx, c.ID)
	if err != nil {
		return err
	}

	if err := p.Comprobantes.UpdateEstado(ctx, c.ID, domain.EstadoEnviado); err != nil {
		return err
	}

	res, err := p.Sunat.Enviar(ctx, c.ID, xml)
	if err != nil {
		_ = p.Comprobantes.UpdateEstado(ctx, c.ID, domain.EstadoPendiente)
		return p.reintentarOFallar(ctx, job, err.Error())
	}

	if res.Aceptado {
		if len(res.CDR) > 0 && p.Storage != nil {
			_ = p.Storage.Guardar(ctx, c.ID+".cdr", res.CDR)
		}
		if err := p.Comprobantes.UpdateEstado(ctx, c.ID, domain.EstadoAceptado); err != nil {
			return err
		}
		return p.Outbox.Complete(ctx, job.ID)
	}

	if err := p.Comprobantes.UpdateEstado(ctx, c.ID, domain.EstadoRechazado); err != nil {
		return err
	}
	return p.Outbox.Complete(ctx, job.ID)
}

func (p *ProcesadorEnvio) obtenerXML(ctx context.Context, comprobanteID string) ([]byte, error) {
	key := comprobanteID + ".xml"
	if p.Storage != nil {
		if data, err := p.Storage.Leer(ctx, key); err == nil {
			return data, nil
		}
	}
	return []byte(fmt.Sprintf("<comprobante id=%q/>", comprobanteID)), nil
}

func (p *ProcesadorEnvio) reintentarOFallar(ctx context.Context, job ports.OutboxJob, errMsg string) error {
	intentos := job.Intentos + 1
	max := p.MaxIntentos
	if max <= 0 {
		max = 5
	}
	if intentos >= max {
		return p.Outbox.Fail(ctx, job.ID, errMsg)
	}
	backoff := CalcularBackoff(intentos, p.BackoffBase)
	disponible := p.now().Add(backoff)
	return p.Outbox.Retry(ctx, job.ID, intentos, disponible, errMsg)
}

func (p *ProcesadorEnvio) now() time.Time {
	if p.Now != nil {
		return p.Now()
	}
	return time.Now()
}

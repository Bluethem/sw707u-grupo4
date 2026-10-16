package postgres

import (
	"context"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
)

type Outbox struct{}

func NewOutbox() *Outbox {
	return &Outbox{}
}

func (o *Outbox) EnqueueEnvio(ctx context.Context, transaction ports.Tx, comprobanteID string) error {
	tx, err := unwrapTx(transaction)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_envios (comprobante_id, tipo_trabajo)
		VALUES ($1, $2)`,
		comprobanteID, ports.TipoTrabajoEnvioComprobante,
	)
	return err
}

func (o *Outbox) EnqueueBaja(ctx context.Context, transaction ports.Tx, comprobanteID string) error {
	tx, err := unwrapTx(transaction)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_envios (comprobante_id, tipo_trabajo)
		VALUES ($1, $2)`,
		comprobanteID, ports.TipoTrabajoBajaFactura,
	)
	return err
}

func (o *Outbox) EnqueueResumen(ctx context.Context, transaction ports.Tx, resumenID string) error {
	tx, err := unwrapTx(transaction)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_envios (referencia_id, tipo_trabajo)
		VALUES ($1, $2)`,
		resumenID, ports.TipoTrabajoResumenDiario,
	)
	return err
}

func (o *Outbox) RequeueEnvio(ctx context.Context, transaction ports.Tx, comprobanteID string) error {
	tx, err := unwrapTx(transaction)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `
		UPDATE outbox_envios
		SET estado = 'pending',
		    intentos = 0,
		    ultimo_error = '',
		    disponible_en = now(),
		    actualizado_en = now()
		WHERE id = (
			SELECT id FROM outbox_envios
			WHERE comprobante_id = $1 AND tipo_trabajo = $2
			ORDER BY id DESC
			LIMIT 1
		)`,
		comprobanteID, ports.TipoTrabajoEnvioComprobante,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_envios (comprobante_id, tipo_trabajo)
		VALUES ($1, $2)`,
		comprobanteID, ports.TipoTrabajoEnvioComprobante,
	)
	return err
}

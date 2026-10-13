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
		INSERT INTO outbox_envios (comprobante_id)
		VALUES ($1)`,
		comprobanteID,
	)
	return err
}

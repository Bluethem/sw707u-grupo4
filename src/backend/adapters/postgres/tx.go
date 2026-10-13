package postgres

import (
	"fmt"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	"github.com/jackc/pgx/v5"
)

type pgTx struct {
	tx pgx.Tx
}

func unwrapTx(tx ports.Tx) (pgx.Tx, error) {
	switch value := tx.(type) {
	case pgTx:
		return value.tx, nil
	case *pgTx:
		if value != nil {
			return value.tx, nil
		}
	}
	return nil, fmt.Errorf("postgres: transaccion invalida")
}

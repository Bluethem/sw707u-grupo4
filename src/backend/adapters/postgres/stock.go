package postgres

import (
	"context"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
)

type Stock struct{}

func NewStock() *Stock {
	return &Stock{}
}

func (s *Stock) Descontar(ctx context.Context, transaction ports.Tx, empresaID string, lineas []ports.StockLinea) error {
	tx, err := unwrapTx(transaction)
	if err != nil {
		return err
	}
	for _, linea := range lineas {
		tag, err := tx.Exec(ctx, `
			UPDATE productos_stock
			SET cantidad = cantidad - $1
			WHERE empresa_id = $2 AND producto_id = $3 AND cantidad >= $1`,
			linea.Cantidad, empresaID, linea.ProductoID,
		)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return domain.ErrStockInsuficiente
		}
	}
	return nil
}

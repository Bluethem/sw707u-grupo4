package postgres

import (
	"context"
	"errors"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	"github.com/jackc/pgx/v5"
)

type Series struct{}

func NewSeries() *Series {
	return &Series{}
}

func (s *Series) SiguienteNumero(
	ctx context.Context,
	transaction ports.Tx,
	empresaID string,
	tipo domain.TipoComprobante,
	serie string,
) (int64, error) {
	tx, err := unwrapTx(transaction)
	if err != nil {
		return 0, err
	}
	var numero int64
	err = tx.QueryRow(ctx, `
		UPDATE series_correlativos
		SET siguiente = siguiente + 1
		WHERE empresa_id = $1 AND tipo = $2 AND serie = $3
		RETURNING siguiente - 1`,
		empresaID, tipo, serie,
	).Scan(&numero)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domain.ErrSerieNoEncontrada
	}
	return numero, err
}

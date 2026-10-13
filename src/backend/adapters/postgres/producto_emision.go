package postgres

import (
	"context"
	"errors"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductoEmisionConsulta struct {
	pool *pgxpool.Pool
}

func NewProductoEmisionConsulta(pool *pgxpool.Pool) *ProductoEmisionConsulta {
	return &ProductoEmisionConsulta{pool: pool}
}

func (c *ProductoEmisionConsulta) ObtenerParaEmision(
	ctx context.Context,
	empresaID, productoID string,
) (controlaStock bool, activo bool, err error) {
	err = c.pool.QueryRow(ctx, `
		SELECT controla_stock, activo
		FROM productos
		WHERE empresa_id = $1 AND id = $2`,
		empresaID, productoID,
	).Scan(&controlaStock, &activo)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, false, domain.ErrProductoNoEncontrado
	}
	if err != nil {
		return false, false, err
	}
	return controlaStock, activo, nil
}

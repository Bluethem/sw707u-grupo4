package postgres

import (
	"context"
	"errors"
	"fmt"

	facturaciondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
)

func (r *ProductoRepository) AjustarStock(
	ctx context.Context,
	empresaID, productoID, usuarioID, tipo string,
	cantidad int64,
	motivo string,
) (gestionapp.AjustarStockResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return gestionapp.AjustarStockResult{}, fmt.Errorf("ajustar stock: %w", err)
	}
	defer tx.Rollback(ctx)

	var controla bool
	err = tx.QueryRow(ctx, `
		SELECT controla_stock FROM productos
		WHERE empresa_id = $1 AND id = $2`,
		empresaID, productoID,
	).Scan(&controla)
	if err != nil {
		return gestionapp.AjustarStockResult{}, mapGestionError("ajustar stock producto", err)
	}
	if !controla {
		return gestionapp.AjustarStockResult{}, fmt.Errorf(
			"ajustar stock: %w: producto no controla stock",
			gestiondomain.ErrInvalido,
		)
	}

	var delta int64
	switch tipo {
	case gestiondomain.TipoMovimientoEntrada:
		delta = cantidad
	case gestiondomain.TipoMovimientoSalida:
		delta = -cantidad
	default:
		return gestionapp.AjustarStockResult{}, fmt.Errorf(
			"ajustar stock: %w: tipo invalido",
			gestiondomain.ErrInvalido,
		)
	}

	var saldo int64
	if delta < 0 {
		err = tx.QueryRow(ctx, `
			UPDATE productos_stock
			SET cantidad = cantidad + $3
			WHERE empresa_id = $1 AND producto_id = $2 AND cantidad >= $4
			RETURNING cantidad`,
			empresaID, productoID, delta, cantidad,
		).Scan(&saldo)
		if err != nil {
			if errors.Is(mapGestionError("ajustar stock salida", err), gestiondomain.ErrNoEncontrado) {
				return gestionapp.AjustarStockResult{}, facturaciondomain.ErrStockInsuficiente
			}
			return gestionapp.AjustarStockResult{}, mapGestionError("ajustar stock salida", err)
		}
	} else {
		err = tx.QueryRow(ctx, `
			INSERT INTO productos_stock (empresa_id, producto_id, cantidad)
			VALUES ($1, $2, $3)
			ON CONFLICT (empresa_id, producto_id) DO UPDATE
			SET cantidad = productos_stock.cantidad + EXCLUDED.cantidad
			RETURNING cantidad`,
			empresaID, productoID, delta,
		).Scan(&saldo)
		if err != nil {
			return gestionapp.AjustarStockResult{}, mapGestionError("ajustar stock entrada", err)
		}
	}

	var movID int64
	var uid *string
	if usuarioID != "" {
		uid = &usuarioID
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO stock_movimientos (empresa_id, producto_id, tipo, cantidad, motivo, usuario_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		empresaID, productoID, tipo, cantidad, motivo, uid,
	).Scan(&movID)
	if err != nil {
		return gestionapp.AjustarStockResult{}, mapGestionError("registrar movimiento stock", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return gestionapp.AjustarStockResult{}, fmt.Errorf("ajustar stock: %w", err)
	}
	return gestionapp.AjustarStockResult{
		ProductoID: productoID, Cantidad: saldo, MovimientoID: movID,
	}, nil
}

func (r *ProductoRepository) ListarMovimientos(
	ctx context.Context,
	empresaID, productoID string,
	limit, offset int,
) ([]gestiondomain.StockMovimiento, int, error) {
	const filtros = `
		FROM stock_movimientos
		WHERE empresa_id = $1 AND producto_id = $2`

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*)`+filtros, empresaID, productoID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar movimientos: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, empresa_id, producto_id, tipo, cantidad, motivo, usuario_id, creado_en`+
		filtros+`
		ORDER BY creado_en DESC, id DESC
		LIMIT $3 OFFSET $4`,
		empresaID, productoID, limit, offset,
	)
	if err != nil {
		return nil, 0, mapGestionError("listar movimientos", err)
	}
	defer rows.Close()

	items := make([]gestiondomain.StockMovimiento, 0)
	for rows.Next() {
		var m gestiondomain.StockMovimiento
		if err := rows.Scan(
			&m.ID, &m.EmpresaID, &m.ProductoID, &m.Tipo, &m.Cantidad, &m.Motivo, &m.UsuarioID, &m.CreadoEn,
		); err != nil {
			return nil, 0, mapGestionError("leer movimiento", err)
		}
		items = append(items, m)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, mapGestionError("listar movimientos", err)
	}
	return items, total, nil
}

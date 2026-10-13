package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Comprobante struct{}

func NewComprobante() *Comprobante {
	return &Comprobante{}
}

type ComprobanteQuery struct {
	pool *pgxpool.Pool
}

func NewComprobanteQuery(pool *pgxpool.Pool) *ComprobanteQuery {
	return &ComprobanteQuery{pool: pool}
}

func (r *ComprobanteQuery) GetByID(
	ctx context.Context,
	empresaID string,
	id string,
) (*domain.Comprobante, error) {
	c, err := scanComprobante(r.pool.QueryRow(ctx, `
		SELECT id, empresa_id, tipo, serie, numero, estado,
		       cliente_tipo, cliente_numero, cliente_razon,
		       valor_venta, igv, total, idempotency_key
		FROM comprobantes
		WHERE id = $1 AND empresa_id = $2`,
		id, empresaID,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("obtener comprobante: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT producto_id, cantidad, precio_unit, afectacion, valor_venta, igv, total
		FROM comprobante_lineas
		WHERE comprobante_id = $1
		ORDER BY id`,
		c.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("obtener lineas de comprobante: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var linea domain.LineaComprobante
		if err := rows.Scan(
			&linea.ProductoID, &linea.Cantidad, &linea.PrecioUnit, &linea.Afectacion,
			&linea.ValorVenta, &linea.IGV, &linea.Total,
		); err != nil {
			return nil, fmt.Errorf("leer linea de comprobante: %w", err)
		}
		c.Lineas = append(c.Lineas, linea)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recorrer lineas de comprobante: %w", err)
	}
	return c, nil
}

func (r *ComprobanteQuery) List(
	ctx context.Context,
	empresaID string,
	f ports.ListFilter,
) ([]domain.Comprobante, int, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	} else if limit > 100 {
		limit = 100
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	var fechaDesde, fechaHastaExclusiva any
	if f.FechaDesde != nil {
		fechaDesde = *f.FechaDesde
	}
	if f.FechaHasta != nil {
		fechaHastaExclusiva = f.FechaHasta.AddDate(0, 0, 1)
	}
	args := []any{empresaID, fechaDesde, fechaHastaExclusiva, f.Tipo, f.Estado}
	const filtros = `
		FROM comprobantes
		WHERE empresa_id = $1
		  AND ($2::timestamptz IS NULL OR creado_en >= $2)
		  AND ($3::timestamptz IS NULL OR creado_en < $3)
		  AND ($4 = '' OR tipo = $4)
		  AND ($5 = '' OR estado = $5)`

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*)`+filtros, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar comprobantes: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, empresa_id, tipo, serie, numero, estado,
		       cliente_tipo, cliente_numero, cliente_razon,
		       valor_venta, igv, total, idempotency_key`+
		filtros+`
		ORDER BY creado_en DESC
		LIMIT $6 OFFSET $7`,
		append(args, limit, offset)...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listar comprobantes: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Comprobante, 0)
	for rows.Next() {
		c, err := scanComprobante(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("leer comprobante: %w", err)
		}
		items = append(items, *c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("recorrer comprobantes: %w", err)
	}
	return items, total, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanComprobante(row rowScanner) (*domain.Comprobante, error) {
	c := &domain.Comprobante{}
	var clienteTipo, clienteNumero, clienteRazon *string
	if err := row.Scan(
		&c.ID, &c.EmpresaID, &c.Tipo, &c.Serie, &c.Numero, &c.Estado,
		&clienteTipo, &clienteNumero, &clienteRazon,
		&c.ValorVenta, &c.IGV, &c.Total, &c.IdempotencyKey,
	); err != nil {
		return nil, err
	}
	if clienteTipo != nil || clienteNumero != nil || clienteRazon != nil {
		c.Cliente = &domain.DatosCliente{}
		if clienteTipo != nil {
			c.Cliente.TipoDoc = *clienteTipo
		}
		if clienteNumero != nil {
			c.Cliente.Numero = *clienteNumero
		}
		if clienteRazon != nil {
			c.Cliente.RazonSocial = *clienteRazon
		}
	}
	return c, nil
}

func (r *Comprobante) GetByIdempotencyKey(
	ctx context.Context,
	transaction ports.Tx,
	empresaID string,
	key string,
) (*domain.Comprobante, error) {
	tx, err := unwrapTx(transaction)
	if err != nil {
		return nil, err
	}

	c := &domain.Comprobante{}
	var clienteTipo, clienteNumero, clienteRazon *string
	err = tx.QueryRow(ctx, `
		SELECT id, empresa_id, tipo, serie, numero, estado,
		       cliente_tipo, cliente_numero, cliente_razon,
		       valor_venta, igv, total, idempotency_key
		FROM comprobantes
		WHERE empresa_id = $1 AND idempotency_key = $2`,
		empresaID, key,
	).Scan(
		&c.ID, &c.EmpresaID, &c.Tipo, &c.Serie, &c.Numero, &c.Estado,
		&clienteTipo, &clienteNumero, &clienteRazon,
		&c.ValorVenta, &c.IGV, &c.Total, &c.IdempotencyKey,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if clienteTipo != nil || clienteNumero != nil || clienteRazon != nil {
		c.Cliente = &domain.DatosCliente{}
		if clienteTipo != nil {
			c.Cliente.TipoDoc = *clienteTipo
		}
		if clienteNumero != nil {
			c.Cliente.Numero = *clienteNumero
		}
		if clienteRazon != nil {
			c.Cliente.RazonSocial = *clienteRazon
		}
	}

	rows, err := tx.Query(ctx, `
		SELECT producto_id, cantidad, precio_unit, afectacion, valor_venta, igv, total
		FROM comprobante_lineas
		WHERE comprobante_id = $1
		ORDER BY id`,
		c.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var linea domain.LineaComprobante
		if err := rows.Scan(
			&linea.ProductoID, &linea.Cantidad, &linea.PrecioUnit, &linea.Afectacion,
			&linea.ValorVenta, &linea.IGV, &linea.Total,
		); err != nil {
			return nil, err
		}
		c.Lineas = append(c.Lineas, linea)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *Comprobante) Insert(ctx context.Context, transaction ports.Tx, c *domain.Comprobante) error {
	tx, err := unwrapTx(transaction)
	if err != nil {
		return err
	}

	var clienteTipo, clienteNumero, clienteRazon any
	if c.Cliente != nil {
		clienteTipo = c.Cliente.TipoDoc
		clienteNumero = c.Cliente.Numero
		clienteRazon = c.Cliente.RazonSocial
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO comprobantes (
			id, empresa_id, tipo, serie, numero, estado,
			cliente_tipo, cliente_numero, cliente_razon,
			valor_venta, igv, total, idempotency_key
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		c.ID, c.EmpresaID, c.Tipo, c.Serie, c.Numero, c.Estado,
		clienteTipo, clienteNumero, clienteRazon,
		c.ValorVenta, c.IGV, c.Total, c.IdempotencyKey,
	)
	if err != nil {
		return err
	}

	for _, linea := range c.Lineas {
		_, err = tx.Exec(ctx, `
			INSERT INTO comprobante_lineas (
				comprobante_id, producto_id, cantidad, precio_unit,
				afectacion, valor_venta, igv, total
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			c.ID, linea.ProductoID, linea.Cantidad, linea.PrecioUnit,
			linea.Afectacion, linea.ValorVenta, linea.IGV, linea.Total,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

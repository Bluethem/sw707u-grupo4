package postgres

import (
	"context"
	"errors"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	"github.com/jackc/pgx/v5"
)

type Comprobante struct{}

func NewComprobante() *Comprobante {
	return &Comprobante{}
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

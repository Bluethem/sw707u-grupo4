package postgres

import (
	"context"
	"fmt"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	"github.com/jackc/pgx/v5"
)

func (r *ComprobanteQuery) Match(
	ctx context.Context,
	c ports.ConsultaPublicaCriteria,
) (*ports.ConsultaPublicaMatch, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT c.id, c.empresa_id, c.tipo, c.serie, c.numero, c.estado,
		       e.razon_social, c.total, c.creado_en
		FROM comprobantes c
		INNER JOIN empresas e ON e.id = c.empresa_id
		WHERE e.ruc = $1
		  AND UPPER(c.serie) = UPPER($2)
		  AND c.numero = $3
		  AND c.cliente_tipo = $4
		  AND c.cliente_numero = $5
		  AND c.cliente_tipo IS NOT NULL
		  AND c.cliente_numero IS NOT NULL
		  AND c.creado_en::date = $6::date
		  AND ROUND(c.total, 2) = ROUND($7::numeric, 2)
		LIMIT 2`,
		c.RUC, c.Serie, c.NumeroComprobante, c.TipoDoc, c.ClienteNumero,
		c.Fecha, c.Monto,
	)
	var m ports.ConsultaPublicaMatch
	if err := row.Scan(
		&m.ID, &m.EmpresaID, &m.Tipo, &m.Serie, &m.Numero, &m.Estado,
		&m.EmisorRazon, &m.Total, &m.CreadoEn,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("match consulta publica: %w", err)
	}
	// Segunda fila: ambigüedad tratada como no coincidencia.
	var extra string
	if err := r.pool.QueryRow(ctx, `
		SELECT c.id
		FROM comprobantes c
		INNER JOIN empresas e ON e.id = c.empresa_id
		WHERE e.ruc = $1
		  AND UPPER(c.serie) = UPPER($2)
		  AND c.numero = $3
		  AND c.cliente_tipo = $4
		  AND c.cliente_numero = $5
		  AND c.cliente_tipo IS NOT NULL
		  AND c.cliente_numero IS NOT NULL
		  AND c.creado_en::date = $6::date
		  AND ROUND(c.total, 2) = ROUND($7::numeric, 2)
		  AND c.id <> $8
		LIMIT 1`,
		c.RUC, c.Serie, c.NumeroComprobante, c.TipoDoc, c.ClienteNumero,
		c.Fecha, c.Monto, m.ID,
	).Scan(&extra); err == nil {
		return nil, nil
	} else if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("verificar unicidad consulta publica: %w", err)
	}
	return &m, nil
}

var _ ports.ConsultaPublicaRepository = (*ComprobanteQuery)(nil)

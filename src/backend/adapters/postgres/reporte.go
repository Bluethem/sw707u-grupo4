package postgres

import (
	"context"
	"fmt"
	"time"

	factapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/app"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type ReporteRepository struct {
	pool *pgxpool.Pool
}

func NewReporteRepository(pool *pgxpool.Pool) *ReporteRepository {
	return &ReporteRepository{pool: pool}
}

func (r *ReporteRepository) Ventas(
	ctx context.Context,
	empresaID string,
	desde, hastaExclusivo time.Time,
) (factapp.VentasReporte, error) {
	var rep factapp.VentasReporte
	rows, err := r.pool.Query(ctx, `
		SELECT tipo, count(*)::int,
		       coalesce(sum(valor_venta), 0),
		       coalesce(sum(igv), 0),
		       coalesce(sum(total), 0)
		FROM comprobantes
		WHERE empresa_id = $1
		  AND creado_en >= $2
		  AND creado_en < $3
		  AND estado NOT IN ('ANULADA', 'PENDIENTE_BAJA')
		GROUP BY tipo
		ORDER BY tipo`,
		empresaID, desde, hastaExclusivo,
	)
	if err != nil {
		return rep, fmt.Errorf("reporte ventas por tipo: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var row factapp.VentasTipo
		if err := rows.Scan(&row.Tipo, &row.Cantidad, &row.ValorVenta, &row.IGV, &row.Total); err != nil {
			return rep, fmt.Errorf("scan por tipo: %w", err)
		}
		rep.PorTipo = append(rep.PorTipo, row)
		rep.Totales.Cantidad += row.Cantidad
		rep.Totales.ValorVenta = rep.Totales.ValorVenta.Add(row.ValorVenta)
		rep.Totales.IGV = rep.Totales.IGV.Add(row.IGV)
		rep.Totales.Total = rep.Totales.Total.Add(row.Total)
	}
	if err := rows.Err(); err != nil {
		return rep, fmt.Errorf("iter por tipo: %w", err)
	}

	var anulTotal decimal.Decimal
	if err := r.pool.QueryRow(ctx, `
		SELECT count(*)::int, coalesce(sum(total), 0)
		FROM comprobantes
		WHERE empresa_id = $1
		  AND creado_en >= $2
		  AND creado_en < $3
		  AND estado IN ('ANULADA', 'PENDIENTE_BAJA')`,
		empresaID, desde, hastaExclusivo,
	).Scan(&rep.Anulados.Cantidad, &anulTotal); err != nil {
		return rep, fmt.Errorf("reporte anulados: %w", err)
	}
	rep.Anulados.Total = anulTotal
	return rep, nil
}

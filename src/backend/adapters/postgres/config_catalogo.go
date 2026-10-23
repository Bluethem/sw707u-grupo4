package postgres

import (
	"context"
	"errors"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type ConfigCatalogoRepository struct {
	pool *pgxpool.Pool
}

func NewConfigCatalogoRepository(pool *pgxpool.Pool) *ConfigCatalogoRepository {
	return &ConfigCatalogoRepository{pool: pool}
}

// GetByEmpresaID retorna la configuración de la empresa.
// Si no existe fila en BD devuelve el default del dominio (nunca error por ausencia).
func (r *ConfigCatalogoRepository) GetByEmpresaID(ctx context.Context, empresaID string) (domain.ConfigCatalogo, error) {
	const q = `
		SELECT empresa_id,
		       precio_minimo, precio_maximo,
		       carga_max_filas, carga_max_mb,
		       unidades_permitidas, afectaciones_permitidas,
		       actualizado_en
		FROM config_catalogo
		WHERE empresa_id = $1`

	row := r.pool.QueryRow(ctx, q, empresaID)
	c, err := scanConfigCatalogo(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DefaultConfigCatalogo(empresaID), nil
	}
	return c, err
}

// Upsert crea o actualiza la configuración de la empresa en una sola operación.
func (r *ConfigCatalogoRepository) Upsert(ctx context.Context, c domain.ConfigCatalogo) (domain.ConfigCatalogo, error) {
	const q = `
		INSERT INTO config_catalogo (
		    empresa_id,
		    precio_minimo, precio_maximo,
		    carga_max_filas, carga_max_mb,
		    unidades_permitidas, afectaciones_permitidas,
		    actualizado_en
		) VALUES ($1,$2,$3,$4,$5,$6,$7, now())
		ON CONFLICT (empresa_id) DO UPDATE SET
		    precio_minimo           = EXCLUDED.precio_minimo,
		    precio_maximo           = EXCLUDED.precio_maximo,
		    carga_max_filas         = EXCLUDED.carga_max_filas,
		    carga_max_mb            = EXCLUDED.carga_max_mb,
		    unidades_permitidas     = EXCLUDED.unidades_permitidas,
		    afectaciones_permitidas = EXCLUDED.afectaciones_permitidas,
		    actualizado_en          = now()
		RETURNING
		    empresa_id,
		    precio_minimo, precio_maximo,
		    carga_max_filas, carga_max_mb,
		    unidades_permitidas, afectaciones_permitidas,
		    actualizado_en`

	var precioMax *string
	if c.PrecioMaximo != nil {
		s := c.PrecioMaximo.String()
		precioMax = &s
	}

	row := r.pool.QueryRow(ctx, q,
		c.EmpresaID,
		c.PrecioMinimo.String(),
		precioMax,
		c.CargaMaxFilas,
		c.CargaMaxMB,
		c.UnidadesPermitidas,
		c.AfectacionesPermitidas,
	)
	return scanConfigCatalogo(row)
}

func scanConfigCatalogo(row pgx.Row) (domain.ConfigCatalogo, error) {
	var (
		c            domain.ConfigCatalogo
		precioMin    string
		precioMaxStr *string
	)
	err := row.Scan(
		&c.EmpresaID,
		&precioMin,
		&precioMaxStr,
		&c.CargaMaxFilas,
		&c.CargaMaxMB,
		&c.UnidadesPermitidas,
		&c.AfectacionesPermitidas,
		&c.ActualizadoEn,
	)
	if err != nil {
		return domain.ConfigCatalogo{}, err
	}
	c.PrecioMinimo, err = decimal.NewFromString(precioMin)
	if err != nil {
		return domain.ConfigCatalogo{}, err
	}
	if precioMaxStr != nil {
		d, err := decimal.NewFromString(*precioMaxStr)
		if err != nil {
			return domain.ConfigCatalogo{}, err
		}
		c.PrecioMaximo = &d
	}
	return c, nil
}

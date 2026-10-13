package postgres

import (
	"context"

	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SeriesGestionRepository struct {
	pool *pgxpool.Pool
}

func NewSeriesGestionRepository(pool *pgxpool.Pool) *SeriesGestionRepository {
	return &SeriesGestionRepository{pool: pool}
}

func (r *SeriesGestionRepository) Crear(ctx context.Context, serie gestiondomain.Serie) (gestiondomain.Serie, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO series_correlativos (empresa_id, tipo, serie, siguiente, activa)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING empresa_id, tipo, serie, siguiente, activa`,
		serie.EmpresaID, serie.Tipo, serie.Serie, serie.Siguiente, serie.Activa,
	).Scan(&serie.EmpresaID, &serie.Tipo, &serie.Serie, &serie.Siguiente, &serie.Activa)
	if err != nil {
		return gestiondomain.Serie{}, mapGestionError("crear serie", err)
	}
	return serie, nil
}

func (r *SeriesGestionRepository) Obtener(ctx context.Context, empresaID string, tipo gestiondomain.TipoSerie, codigo string) (gestiondomain.Serie, error) {
	var serie gestiondomain.Serie
	err := r.pool.QueryRow(ctx, `
		SELECT empresa_id, tipo, serie, siguiente, activa
		FROM series_correlativos
		WHERE empresa_id = $1 AND tipo = $2 AND serie = $3`,
		empresaID, tipo, codigo,
	).Scan(&serie.EmpresaID, &serie.Tipo, &serie.Serie, &serie.Siguiente, &serie.Activa)
	if err != nil {
		return gestiondomain.Serie{}, mapGestionError("obtener serie", err)
	}
	return serie, nil
}

func (r *SeriesGestionRepository) Listar(ctx context.Context, empresaID string) ([]gestiondomain.Serie, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT empresa_id, tipo, serie, siguiente, activa
		FROM series_correlativos
		WHERE empresa_id = $1
		ORDER BY tipo, serie`,
		empresaID,
	)
	if err != nil {
		return nil, mapGestionError("listar series", err)
	}
	defer rows.Close()

	series := make([]gestiondomain.Serie, 0)
	for rows.Next() {
		var serie gestiondomain.Serie
		if err := rows.Scan(&serie.EmpresaID, &serie.Tipo, &serie.Serie, &serie.Siguiente, &serie.Activa); err != nil {
			return nil, mapGestionError("leer serie", err)
		}
		series = append(series, serie)
	}
	if err := rows.Err(); err != nil {
		return nil, mapGestionError("listar series", err)
	}
	return series, nil
}

func (r *SeriesGestionRepository) Actualizar(ctx context.Context, serie gestiondomain.Serie) (gestiondomain.Serie, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE series_correlativos
		SET siguiente = $4, activa = $5
		WHERE empresa_id = $1 AND tipo = $2 AND serie = $3
		RETURNING empresa_id, tipo, serie, siguiente, activa`,
		serie.EmpresaID, serie.Tipo, serie.Serie, serie.Siguiente, serie.Activa,
	).Scan(&serie.EmpresaID, &serie.Tipo, &serie.Serie, &serie.Siguiente, &serie.Activa)
	if err != nil {
		return gestiondomain.Serie{}, mapGestionError("actualizar serie", err)
	}
	return serie, nil
}

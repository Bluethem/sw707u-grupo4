package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ResumenRepository struct {
	pool *pgxpool.Pool
}

func NewResumenRepository(pool *pgxpool.Pool) *ResumenRepository {
	return &ResumenRepository{pool: pool}
}

func (r *ResumenRepository) List(ctx context.Context, empresaID string) ([]domain.ResumenDiario, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, empresa_id, fecha, correlativo, estado, ticket, creado_en
		FROM resumenes_diarios
		WHERE empresa_id = $1
		ORDER BY fecha DESC, correlativo DESC`,
		empresaID,
	)
	if err != nil {
		return nil, fmt.Errorf("listar resumenes: %w", err)
	}
	defer rows.Close()

	var items []domain.ResumenDiario
	for rows.Next() {
		item, err := scanResumenHeader(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recorrer resumenes: %w", err)
	}
	return items, nil
}

func (r *ResumenRepository) GetByID(ctx context.Context, empresaID, id string) (*domain.ResumenDiario, error) {
	resumen, err := scanResumenHeader(r.pool.QueryRow(ctx, `
		SELECT id, empresa_id, fecha, correlativo, estado, ticket, creado_en
		FROM resumenes_diarios
		WHERE id = $1 AND empresa_id = $2`,
		id, empresaID,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("obtener resumen: %w", err)
	}
	rows, err := r.pool.Query(ctx, `
		SELECT comprobante_id, anulado
		FROM resumen_comprobantes
		WHERE resumen_id = $1
		ORDER BY comprobante_id`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("listar comprobantes del resumen: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.ResumenComprobanteItem
		if err := rows.Scan(&item.ComprobanteID, &item.Anulado); err != nil {
			return nil, fmt.Errorf("leer comprobante resumen: %w", err)
		}
		resumen.Comprobantes = append(resumen.Comprobantes, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recorrer comprobantes resumen: %w", err)
	}
	return &resumen, nil
}

func (r *ResumenRepository) BoletasInformables(
	ctx context.Context,
	transaction ports.Tx,
	empresaID string,
	fecha time.Time,
) ([]ports.BoletaInformable, error) {
	tx, err := unwrapTx(transaction)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `
		SELECT c.id, c.estado
		FROM comprobantes c
		WHERE c.empresa_id = $1
		  AND c.tipo = 'boleta'
		  AND c.creado_en::date = $2::date
		  AND c.estado IN ('ACEPTADO', 'PENDIENTE', 'ENVIADO', 'PENDIENTE_BAJA')
		  AND NOT EXISTS (
		    SELECT 1
		    FROM resumen_comprobantes rc
		    JOIN resumenes_diarios rd ON rd.id = rc.resumen_id
		    WHERE rc.comprobante_id = c.id AND rd.estado = 'ACEPTADO'
		  )
		ORDER BY c.creado_en, c.id`,
		empresaID, fecha,
	)
	if err != nil {
		return nil, fmt.Errorf("boletas informables: %w", err)
	}
	defer rows.Close()

	var items []ports.BoletaInformable
	for rows.Next() {
		var item ports.BoletaInformable
		var estado string
		if err := rows.Scan(&item.ID, &estado); err != nil {
			return nil, fmt.Errorf("leer boleta informable: %w", err)
		}
		item.Estado = domain.EstadoComprobante(estado)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recorrer boletas informables: %w", err)
	}
	return items, nil
}

func (r *ResumenRepository) NextCorrelativo(
	ctx context.Context,
	transaction ports.Tx,
	empresaID string,
	fecha time.Time,
) (int, error) {
	tx, err := unwrapTx(transaction)
	if err != nil {
		return 0, err
	}
	var next int
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(correlativo), 0) + 1
		FROM resumenes_diarios
		WHERE empresa_id = $1 AND fecha = $2::date`,
		empresaID, fecha,
	).Scan(&next)
	if err != nil {
		return 0, fmt.Errorf("correlativo resumen: %w", err)
	}
	return next, nil
}

func (r *ResumenRepository) Insert(ctx context.Context, transaction ports.Tx, resumen *domain.ResumenDiario) error {
	tx, err := unwrapTx(transaction)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO resumenes_diarios (id, empresa_id, fecha, correlativo, estado, ticket)
		VALUES ($1, $2, $3::date, $4, $5, $6)`,
		resumen.ID, resumen.EmpresaID, resumen.Fecha, resumen.Correlativo, resumen.Estado, resumen.Ticket,
	)
	if err != nil {
		return fmt.Errorf("insert resumen: %w", err)
	}
	for _, item := range resumen.Comprobantes {
		_, err = tx.Exec(ctx, `
			INSERT INTO resumen_comprobantes (resumen_id, comprobante_id, anulado)
			VALUES ($1, $2, $3)`,
			resumen.ID, item.ComprobanteID, item.Anulado,
		)
		if err != nil {
			return fmt.Errorf("insert resumen comprobante: %w", err)
		}
	}
	return nil
}

func (r *ResumenRepository) GetWorkerData(
	ctx context.Context,
	resumenID string,
) (domain.EstadoResumen, []ports.ResumenWorkerItem, error) {
	var estado string
	err := r.pool.QueryRow(ctx, `
		SELECT estado FROM resumenes_diarios WHERE id = $1`, resumenID,
	).Scan(&estado)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, fmt.Errorf("resumen %q no encontrado", resumenID)
	}
	if err != nil {
		return "", nil, fmt.Errorf("estado resumen worker: %w", err)
	}
	rows, err := r.pool.Query(ctx, `
		SELECT comprobante_id, anulado
		FROM resumen_comprobantes
		WHERE resumen_id = $1`,
		resumenID,
	)
	if err != nil {
		return "", nil, fmt.Errorf("items resumen worker: %w", err)
	}
	defer rows.Close()
	var items []ports.ResumenWorkerItem
	for rows.Next() {
		var item ports.ResumenWorkerItem
		if err := rows.Scan(&item.ComprobanteID, &item.Anulado); err != nil {
			return "", nil, fmt.Errorf("leer item resumen worker: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return "", nil, fmt.Errorf("recorrer items resumen worker: %w", err)
	}
	return domain.EstadoResumen(estado), items, nil
}

func (r *ResumenRepository) UpdateEstado(
	ctx context.Context,
	resumenID string,
	estado domain.EstadoResumen,
	ticket string,
) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE resumenes_diarios SET estado = $2, ticket = $3 WHERE id = $1`,
		resumenID, estado, ticket,
	)
	if err != nil {
		return fmt.Errorf("actualizar estado resumen: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("resumen %q no encontrado", resumenID)
	}
	return nil
}

type resumenHeaderScanner interface {
	Scan(dest ...any) error
}

func scanResumenHeader(s resumenHeaderScanner) (domain.ResumenDiario, error) {
	var resumen domain.ResumenDiario
	var estado string
	if err := s.Scan(
		&resumen.ID, &resumen.EmpresaID, &resumen.Fecha, &resumen.Correlativo,
		&estado, &resumen.Ticket, &resumen.CreadoEn,
	); err != nil {
		return domain.ResumenDiario{}, err
	}
	resumen.Estado = domain.EstadoResumen(estado)
	return resumen, nil
}

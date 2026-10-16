package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxConsumer struct {
	pool *pgxpool.Pool
}

func NewOutboxConsumer(pool *pgxpool.Pool) *OutboxConsumer {
	return &OutboxConsumer{pool: pool}
}

func (c *OutboxConsumer) Claim(ctx context.Context, limit int) ([]ports.OutboxJob, error) {
	if limit <= 0 {
		return nil, nil
	}
	rows, err := c.pool.Query(ctx, `
		WITH cte AS (
			SELECT id FROM outbox_envios
			WHERE estado = 'pending' AND disponible_en <= now()
			ORDER BY id
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		UPDATE outbox_envios o
		SET estado = 'processing', actualizado_en = now()
		FROM cte WHERE o.id = cte.id
		RETURNING o.id, o.comprobante_id, o.referencia_id, o.tipo_trabajo, o.estado, o.intentos,
		          o.disponible_en, o.creado_en, o.ultimo_error, o.actualizado_en`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("claim outbox: %w", err)
	}
	defer rows.Close()

	var jobs []ports.OutboxJob
	for rows.Next() {
		job, err := scanOutboxJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recorrer claim outbox: %w", err)
	}
	return jobs, nil
}

func (c *OutboxConsumer) Complete(ctx context.Context, jobID int64) error {
	tag, err := c.pool.Exec(ctx, `
		UPDATE outbox_envios
		SET estado = 'done', actualizado_en = now()
		WHERE id = $1`,
		jobID,
	)
	if err != nil {
		return fmt.Errorf("complete outbox: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("complete outbox: job %d no encontrado", jobID)
	}
	return nil
}

func (c *OutboxConsumer) CompleteWithError(ctx context.Context, jobID int64, errMsg string) error {
	tag, err := c.pool.Exec(ctx, `
		UPDATE outbox_envios
		SET estado = 'done', ultimo_error = $2, actualizado_en = now()
		WHERE id = $1`,
		jobID, errMsg,
	)
	if err != nil {
		return fmt.Errorf("complete outbox con error: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("complete outbox con error: job %d no encontrado", jobID)
	}
	return nil
}

func (c *OutboxConsumer) Retry(
	ctx context.Context,
	jobID int64,
	intentos int,
	disponibleEn time.Time,
	errMsg string,
) error {
	tag, err := c.pool.Exec(ctx, `
		UPDATE outbox_envios
		SET estado = 'pending',
		    intentos = $2,
		    disponible_en = $3,
		    ultimo_error = $4,
		    actualizado_en = now()
		WHERE id = $1`,
		jobID, intentos, disponibleEn, errMsg,
	)
	if err != nil {
		return fmt.Errorf("retry outbox: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("retry outbox: job %d no encontrado", jobID)
	}
	return nil
}

func (c *OutboxConsumer) Fail(ctx context.Context, jobID int64, errMsg string) error {
	tag, err := c.pool.Exec(ctx, `
		UPDATE outbox_envios
		SET estado = 'failed',
		    ultimo_error = $2,
		    actualizado_en = now()
		WHERE id = $1`,
		jobID, errMsg,
	)
	if err != nil {
		return fmt.Errorf("fail outbox: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("fail outbox: job %d no encontrado", jobID)
	}
	return nil
}

type outboxRowScanner interface {
	Scan(dest ...any) error
}

func scanOutboxJob(s outboxRowScanner) (ports.OutboxJob, error) {
	var job ports.OutboxJob
	var comprobanteID, referenciaID sql.NullString
	if err := s.Scan(
		&job.ID,
		&comprobanteID,
		&referenciaID,
		&job.TipoTrabajo,
		&job.Estado,
		&job.Intentos,
		&job.DisponibleEn,
		&job.CreadoEn,
		&job.UltimoError,
		&job.ActualizadoEn,
	); err != nil {
		return ports.OutboxJob{}, fmt.Errorf("leer outbox job: %w", err)
	}
	if comprobanteID.Valid {
		job.ComprobanteID = comprobanteID.String
	}
	if referenciaID.Valid {
		job.ReferenciaID = referenciaID.String
	}
	return job, nil
}

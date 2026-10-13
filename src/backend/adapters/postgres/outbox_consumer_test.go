package postgres

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestOutboxConsumer_ClaimCompleteRetryFail(t *testing.T) {
	ctx := context.Background()
	pool := newOutboxTestPool(t, ctx)
	consumer := NewOutboxConsumer(pool)

	id1 := insertOutboxJob(t, ctx, pool, "pending", time.Now())
	id2 := insertOutboxJob(t, ctx, pool, "pending", time.Now())

	jobs, err := consumer.Claim(ctx, 10)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("claim len=%d, want 2", len(jobs))
	}
	for _, j := range jobs {
		if j.Estado != "processing" {
			t.Fatalf("job %d estado=%q want processing", j.ID, j.Estado)
		}
	}

	if err := consumer.Complete(ctx, id1); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if estado := outboxEstado(t, ctx, pool, id1); estado != "done" {
		t.Fatalf("id1 estado=%q want done", estado)
	}

	retryAt := time.Now().UTC().Add(2 * time.Minute).Truncate(time.Microsecond)
	if err := consumer.Retry(ctx, id2, 3, retryAt, "timeout"); err != nil {
		t.Fatalf("retry: %v", err)
	}
	var intentos int
	var disp time.Time
	var ultimo string
	var estado string
	err = pool.QueryRow(ctx, `
		SELECT intentos, disponible_en, ultimo_error, estado
		FROM outbox_envios WHERE id = $1`, id2,
	).Scan(&intentos, &disp, &ultimo, &estado)
	if err != nil {
		t.Fatalf("leer retry: %v", err)
	}
	if intentos != 3 || ultimo != "timeout" || estado != "pending" {
		t.Fatalf("retry fields: intentos=%d ultimo=%q estado=%q", intentos, ultimo, estado)
	}
	if !disp.Equal(retryAt) {
		t.Fatalf("disponible_en=%v want %v", disp, retryAt)
	}

	id3 := insertOutboxJob(t, ctx, pool, "processing", time.Now())
	if err := consumer.Fail(ctx, id3, "agotado"); err != nil {
		t.Fatalf("fail: %v", err)
	}
	if estado := outboxEstado(t, ctx, pool, id3); estado != "failed" {
		t.Fatalf("id3 estado=%q want failed", estado)
	}
}

func TestOutboxConsumer_ClaimSkipsNotReady(t *testing.T) {
	ctx := context.Background()
	pool := newOutboxTestPool(t, ctx)
	consumer := NewOutboxConsumer(pool)

	insertOutboxJob(t, ctx, pool, "pending", time.Now().Add(time.Hour))
	insertOutboxJob(t, ctx, pool, "done", time.Now())
	ready := insertOutboxJob(t, ctx, pool, "pending", time.Now())

	jobs, err := consumer.Claim(ctx, 5)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	if len(jobs) != 1 || jobs[0].ID != ready {
		t.Fatalf("claim=%#v want solo job %d", jobs, ready)
	}
}

func TestOutboxConsumer_ClaimSkipLocked(t *testing.T) {
	ctx := context.Background()
	pool := newOutboxTestPool(t, ctx)
	consumer := NewOutboxConsumer(pool)

	insertOutboxJob(t, ctx, pool, "pending", time.Now())

	var wg sync.WaitGroup
	var claimed atomic.Int32
	errCh := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			jobs, err := consumer.Claim(ctx, 1)
			if err != nil {
				errCh <- err
				return
			}
			claimed.Add(int32(len(jobs)))
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("claim concurrente: %v", err)
	}
	if claimed.Load() != 1 {
		t.Fatalf("claimed=%d want 1 (sin duplicados)", claimed.Load())
	}
}

func newOutboxTestPool(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()
	return newClienteTestPool(t, ctx)
}

func insertOutboxComprobante(t *testing.T, ctx context.Context, pool *pgxpool.Pool) string {
	t.Helper()
	id := uuid.NewString()
	empresaID := uuid.NewString()
	ruc := rucUnico(t, id)
	_, err := pool.Exec(ctx, `
		INSERT INTO empresas (id, ruc, razon_social, domicilio_fiscal, regimen, activa)
		VALUES ($1, $2, 'Outbox Test', '', 'general', true)`,
		empresaID, ruc,
	)
	if err != nil {
		t.Fatalf("insert empresa: %v", err)
	}
	key := uuid.NewString()
	numero := int64(uuid.New().ID()%900_000) + 1
	_, err = pool.Exec(ctx, `
		INSERT INTO comprobantes (
			id, empresa_id, tipo, serie, numero, estado,
			valor_venta, igv, total, idempotency_key
		) VALUES ($1, $2, 'boleta', 'B001', $3, 'PENDIENTE', 10, 1.8, 11.8, $4)`,
		id, empresaID, numero, key,
	)
	if err != nil {
		t.Fatalf("insert comprobante: %v", err)
	}
	return id
}

func insertOutboxJob(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	estado string,
	disponibleEn time.Time,
) int64 {
	t.Helper()
	comprobanteID := insertOutboxComprobante(t, ctx, pool)
	var id int64
	err := pool.QueryRow(ctx, `
		INSERT INTO outbox_envios (comprobante_id, estado, disponible_en)
		VALUES ($1, $2, $3)
		RETURNING id`,
		comprobanteID, estado, disponibleEn,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert outbox: %v", err)
	}
	return id
}

func outboxEstado(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id int64) string {
	t.Helper()
	var estado string
	if err := pool.QueryRow(ctx, `SELECT estado FROM outbox_envios WHERE id = $1`, id).Scan(&estado); err != nil {
		t.Fatalf("leer estado: %v", err)
	}
	return estado
}

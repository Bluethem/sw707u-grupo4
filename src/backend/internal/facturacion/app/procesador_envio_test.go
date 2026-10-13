package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/pdf"
	postgresadapter "github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/storage"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/sunat"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/app"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
)

func TestCalcularBackoff(t *testing.T) {
	base := time.Second
	if got := app.CalcularBackoff(0, base); got != time.Second {
		t.Fatalf("intentos 0: %v; want 1s", got)
	}
	if got := app.CalcularBackoff(3, base); got != 8*time.Second {
		t.Fatalf("intentos 3: %v; want 8s", got)
	}
	if got := app.CalcularBackoff(10, base); got != 60*time.Second {
		t.Fatalf("intentos 10: %v; want cap 60s", got)
	}
}

func nuevoProcesador(gateway *sunat.Fake) *app.ProcesadorEnvio {
	store := storage.NewFake()
	return &app.ProcesadorEnvio{
		Outbox:       postgresadapter.NewOutboxConsumer(testPool),
		Comprobantes: postgresadapter.NewComprobanteQuery(testPool),
		Storage:      store,
		Sunat:        gateway,
		MaxIntentos:  5,
		BackoffBase:  time.Second,
	}
}

func TestWorker_ProcesarAceptado(t *testing.T) {
	seed(t, 10)
	emisor := nuevoEmisor(pdf.Fake{})
	got, _, err := emisor.Emitir(context.Background(), comando("proc-acept", 1))
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}

	proc := nuevoProcesador(sunat.NewFake(sunat.ModeAccept))
	if err := proc.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	estado := queryString(t, `SELECT estado FROM comprobantes WHERE id=$1`, got.ID)
	if estado != string(domain.EstadoAceptado) {
		t.Fatalf("estado = %q; want ACEPTADO", estado)
	}
	if outbox := queryString(t, `SELECT estado FROM outbox_envios WHERE comprobante_id=$1`, got.ID); outbox != "done" {
		t.Fatalf("outbox = %q; want done", outbox)
	}
}

func TestWorker_Rechazo(t *testing.T) {
	seed(t, 10)
	got, _, err := nuevoEmisor(pdf.Fake{}).Emitir(context.Background(), comando("proc-rech", 1))
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}

	if err := nuevoProcesador(sunat.NewFake(sunat.ModeReject)).RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	estado := queryString(t, `SELECT estado FROM comprobantes WHERE id=$1`, got.ID)
	if estado != string(domain.EstadoRechazado) {
		t.Fatalf("estado = %q; want RECHAZADO", estado)
	}
}

func TestWorker_RetryError(t *testing.T) {
	seed(t, 10)
	got, _, err := nuevoEmisor(pdf.Fake{}).Emitir(context.Background(), comando("proc-err", 1))
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}

	now := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	proc := nuevoProcesador(sunat.NewFake(sunat.ModeError))
	proc.Now = func() time.Time { return now }

	if err := proc.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	estado := queryString(t, `SELECT estado FROM comprobantes WHERE id=$1`, got.ID)
	if estado != string(domain.EstadoPendiente) {
		t.Fatalf("estado = %q; want PENDIENTE", estado)
	}

	var jobEstado string
	var intentos int
	var disponible time.Time
	err = testPool.QueryRow(context.Background(), `
		SELECT estado, intentos, disponible_en FROM outbox_envios WHERE comprobante_id=$1`,
		got.ID,
	).Scan(&jobEstado, &intentos, &disponible)
	if err != nil {
		t.Fatalf("query outbox: %v", err)
	}
	if jobEstado != "pending" {
		t.Fatalf("job estado = %q; want pending", jobEstado)
	}
	if intentos != 1 {
		t.Fatalf("intentos = %d; want 1", intentos)
	}
	wantDisp := now.Add(2 * time.Second)
	if !disponible.Equal(wantDisp) {
		t.Fatalf("disponible_en = %v; want %v", disponible, wantDisp)
	}
}

func queryString(t *testing.T, query string, args ...any) string {
	t.Helper()
	var got string
	if err := testPool.QueryRow(context.Background(), query, args...).Scan(&got); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return got
}

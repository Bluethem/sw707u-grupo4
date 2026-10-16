package app_test

import (
	"context"
	"testing"
	"time"

	postgresadapter "github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/sunat"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/app"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
)

func generarResumenWorker(t *testing.T) (resumenID, boletaID string) {
	t.Helper()
	seed(t, 10)
	got, _, err := nuevoEmisor(nil).Emitir(context.Background(), comando("res-worker", 1))
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}
	if err := nuevoProcesador(sunat.NewFake(sunat.ModeAccept)).RunOnce(context.Background()); err != nil {
		t.Fatalf("aceptar boleta: %v", err)
	}

	uow := postgresadapter.NewUnitOfWork(testPool)
	store := postgresadapter.NewResumenRepository(testPool)
	outbox := postgresadapter.NewOutbox()
	svc := &app.Resumenes{UoW: uow, Store: store, Outbox: outbox}

	var fecha time.Time
	if err := testPool.QueryRow(context.Background(), `
		SELECT creado_en::date FROM comprobantes WHERE id=$1`, got.ID,
	).Scan(&fecha); err != nil {
		t.Fatalf("fecha: %v", err)
	}
	resumen, err := svc.GenerarResumen(context.Background(), "e1", fecha)
	if err != nil {
		t.Fatalf("GenerarResumen: %v", err)
	}
	return resumen.ID, got.ID
}

func TestWorker_ResumenAceptado(t *testing.T) {
	resumenID, boletaID := generarResumenWorker(t)
	if err := nuevoProcesador(sunat.NewFake(sunat.ModeAccept)).RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if estado := queryString(t, `SELECT estado FROM resumenes_diarios WHERE id=$1`, resumenID); estado != "ACEPTADO" {
		t.Fatalf("resumen = %q; want ACEPTADO", estado)
	}
	if estado := queryString(t, `SELECT estado FROM comprobantes WHERE id=$1`, boletaID); estado != "ACEPTADO" {
		t.Fatalf("boleta = %q; want ACEPTADO", estado)
	}
}

func TestWorker_ResumenRechazado(t *testing.T) {
	resumenID, boletaID := generarResumenWorker(t)
	if err := nuevoProcesador(sunat.NewFake(sunat.ModeReject)).RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if estado := queryString(t, `SELECT estado FROM resumenes_diarios WHERE id=$1`, resumenID); estado != "RECHAZADO" {
		t.Fatalf("resumen = %q; want RECHAZADO", estado)
	}
	if estado := queryString(t, `SELECT estado FROM comprobantes WHERE id=$1`, boletaID); estado != "ACEPTADO" {
		t.Fatalf("boleta debe quedar ACEPTADA tras rechazo resumen, got %q", estado)
	}
}

func TestWorker_ResumenAnulaPendienteBaja(t *testing.T) {
	seed(t, 10)
	got, _, err := nuevoEmisor(nil).Emitir(context.Background(), comando("res-anul", 1))
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}
	proc := nuevoProcesador(sunat.NewFake(sunat.ModeAccept))
	if err := proc.RunOnce(context.Background()); err != nil {
		t.Fatalf("aceptar: %v", err)
	}
	if _, err := testPool.Exec(context.Background(), `
		UPDATE comprobantes SET estado='PENDIENTE_BAJA', baja_motivo='test' WHERE id=$1`, got.ID); err != nil {
		t.Fatalf("marcar baja: %v", err)
	}

	uow := postgresadapter.NewUnitOfWork(testPool)
	svc := &app.Resumenes{
		UoW: uow, Store: postgresadapter.NewResumenRepository(testPool), Outbox: postgresadapter.NewOutbox(),
	}
	var fecha time.Time
	if err := testPool.QueryRow(context.Background(), `
		SELECT creado_en::date FROM comprobantes WHERE id=$1`, got.ID,
	).Scan(&fecha); err != nil {
		t.Fatalf("fecha: %v", err)
	}
	resumen, err := svc.GenerarResumen(context.Background(), "e1", fecha)
	if err != nil {
		t.Fatalf("GenerarResumen: %v", err)
	}
	if err := proc.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce resumen: %v", err)
	}
	if estado := queryString(t, `SELECT estado FROM comprobantes WHERE id=$1`, got.ID); estado != string(domain.EstadoAnulada) {
		t.Fatalf("boleta = %q; want ANULADA", estado)
	}
	if anulado := queryString(t, `
		SELECT anulado::text FROM resumen_comprobantes WHERE resumen_id=$1 AND comprobante_id=$2`,
		resumen.ID, got.ID); anulado != "true" {
		t.Fatalf("anulado flag = %q; want true", anulado)
	}
}

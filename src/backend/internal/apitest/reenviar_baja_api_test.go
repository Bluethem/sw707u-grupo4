package apitest

import (
	"context"
	"net/http"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/sunat"
)

func seedFacturaSeries(t *testing.T, h *harness, empresaID string) {
	t.Helper()
	if _, err := h.pool.Exec(context.Background(), `
		INSERT INTO series_correlativos (empresa_id, tipo, serie, siguiente, activa)
		VALUES ($1, 'factura', 'F001', 1, true)`,
		empresaID,
	); err != nil {
		t.Fatalf("seed serie factura: %v", err)
	}
}

func emitirFactura(t *testing.T, h *harness, token, key string) comprobanteAPI {
	t.Helper()
	body := map[string]any{
		"tipo": "factura", "serie": "F001",
		"cliente": map[string]string{
			"tipo_doc": "RUC", "numero": "20100070970", "razon_social": "Cliente SAC",
		},
		"lineas": []map[string]any{{
			"producto_id": "p1", "cantidad": int64(1),
			"precio_unitario": "118.00", "afectacion": "gravado",
		}},
	}
	res := h.requestWithHeaders(t, http.MethodPost, "/api/v1/comprobantes", token,
		body, map[string]string{"Idempotency-Key": key})
	requireStatus(t, res, http.StatusCreated)
	var c comprobanteAPI
	decodeResponse(t, res, &c)
	return c
}

func aceptarComprobante(t *testing.T, h *harness) {
	t.Helper()
	proc := h.newProcesadorEnvio(sunat.ModeAccept)
	if err := proc.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce aceptar: %v", err)
	}
}

func TestAPI_Reenviar_OK(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, emitido := h.emitir(t, seed.Token, "reenviar-ok", 1)
	requireStatus(t, emitirRes, http.StatusCreated)

	if _, err := h.pool.Exec(context.Background(), `
		UPDATE outbox_envios SET estado='failed', ultimo_error='timeout fake'
		WHERE comprobante_id=$1`, emitido.ID); err != nil {
		t.Fatalf("marcar outbox failed: %v", err)
	}

	res := h.request(t, http.MethodPost, "/api/v1/comprobantes/"+emitido.ID+"/reenviar", seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var got comprobanteAPI
	decodeResponse(t, res, &got)
	if got.Estado != "PENDIENTE" {
		t.Fatalf("estado = %q; want PENDIENTE", got.Estado)
	}

	var outboxEstado string
	var intentos int
	if err := h.pool.QueryRow(context.Background(), `
		SELECT estado, intentos FROM outbox_envios WHERE comprobante_id=$1`, emitido.ID,
	).Scan(&outboxEstado, &intentos); err != nil {
		t.Fatalf("query outbox: %v", err)
	}
	if outboxEstado != "pending" || intentos != 0 {
		t.Fatalf("outbox = %q intentos=%d; want pending/0", outboxEstado, intentos)
	}

	aceptarComprobante(t, h)
	getRes := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+emitido.ID, seed.Token, nil)
	requireStatus(t, getRes, http.StatusOK)
	decodeResponse(t, getRes, &got)
	if got.Estado != "ACEPTADO" {
		t.Fatalf("estado final = %q; want ACEPTADO", got.Estado)
	}
}

func TestAPI_Reenviar_ForbiddenAceptado(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, emitido := h.emitir(t, seed.Token, "reenviar-forbidden", 1)
	requireStatus(t, emitirRes, http.StatusCreated)
	aceptarComprobante(t, h)

	res := h.request(t, http.MethodPost, "/api/v1/comprobantes/"+emitido.ID+"/reenviar", seed.Token, nil)
	requireStatus(t, res, http.StatusConflict)
	res.Body.Close()
}

func TestAPI_Reenviar_CajeroForbidden(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, emitido := h.emitir(t, seed.Token, "reenviar-rbac", 1)
	requireStatus(t, emitirRes, http.StatusCreated)
	cajeroToken := seedCajero(t, h, seed.EmpresaID)

	res := h.request(t, http.MethodPost, "/api/v1/comprobantes/"+emitido.ID+"/reenviar", cajeroToken, nil)
	requireStatus(t, res, http.StatusForbidden)
	res.Body.Close()
}

func TestAPI_BajaFactura_OK(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	seedFacturaSeries(t, h, seed.EmpresaID)
	emitido := emitirFactura(t, h, seed.Token, "baja-factura-ok")
	aceptarComprobante(t, h)

	res := h.request(t, http.MethodPost, "/api/v1/comprobantes/"+emitido.ID+"/baja", seed.Token,
		map[string]string{"motivo": "  Error de emision  "})
	requireStatus(t, res, http.StatusOK)
	var got struct {
		Estado     string `json:"estado"`
		BajaMotivo string `json:"baja_motivo"`
	}
	decodeResponse(t, res, &got)
	if got.Estado != "PENDIENTE_BAJA" || got.BajaMotivo != "Error de emision" {
		t.Fatalf("respuesta baja inesperada: %+v", got)
	}

	var bajaJobs int
	if err := h.pool.QueryRow(context.Background(), `
		SELECT count(*) FROM outbox_envios
		WHERE comprobante_id=$1 AND tipo_trabajo='baja_factura'`, emitido.ID,
	).Scan(&bajaJobs); err != nil {
		t.Fatalf("contar baja outbox: %v", err)
	}
	if bajaJobs != 1 {
		t.Fatalf("jobs baja = %d; want 1", bajaJobs)
	}

	aceptarComprobante(t, h)
	getRes := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+emitido.ID, seed.Token, nil)
	requireStatus(t, getRes, http.StatusOK)
	decodeResponse(t, getRes, &got)
	if got.Estado != "ANULADA" {
		t.Fatalf("estado final = %q; want ANULADA", got.Estado)
	}
}

func TestAPI_BajaFactura_ForbiddenPendiente(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	seedFacturaSeries(t, h, seed.EmpresaID)
	emitido := emitirFactura(t, h, seed.Token, "baja-factura-forbidden")

	res := h.request(t, http.MethodPost, "/api/v1/comprobantes/"+emitido.ID+"/baja", seed.Token,
		map[string]string{"motivo": "motivo"})
	requireStatus(t, res, http.StatusConflict)
	res.Body.Close()
}

func TestAPI_BajaFactura_BoletaSinOutbox(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, emitido := h.emitir(t, seed.Token, "baja-boleta", 1)
	requireStatus(t, emitirRes, http.StatusCreated)

	res := h.request(t, http.MethodPost, "/api/v1/comprobantes/"+emitido.ID+"/baja", seed.Token,
		map[string]string{"motivo": "Cliente desistió"})
	requireStatus(t, res, http.StatusOK)
	var got struct {
		Estado     string `json:"estado"`
		BajaMotivo string `json:"baja_motivo"`
	}
	decodeResponse(t, res, &got)
	if got.Estado != "PENDIENTE_BAJA" || got.BajaMotivo != "Cliente desistió" {
		t.Fatalf("baja boleta inesperada: %+v", got)
	}
	var bajaJobs int
	if err := h.pool.QueryRow(context.Background(), `
		SELECT count(*) FROM outbox_envios
		WHERE comprobante_id=$1 AND tipo_trabajo='baja_factura'`, emitido.ID,
	).Scan(&bajaJobs); err != nil {
		t.Fatalf("contar baja outbox: %v", err)
	}
	if bajaJobs != 0 {
		t.Fatalf("boleta no debe encolar baja; jobs=%d", bajaJobs)
	}
	var motivoDB string
	if err := h.pool.QueryRow(context.Background(), `
		SELECT baja_motivo FROM comprobantes WHERE id=$1`, emitido.ID,
	).Scan(&motivoDB); err != nil {
		t.Fatalf("leer baja_motivo: %v", err)
	}
	if motivoDB != "Cliente desistió" {
		t.Fatalf("baja_motivo DB = %q", motivoDB)
	}
}

func TestAPI_BajaFactura_MotivoVacio(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, emitido := h.emitir(t, seed.Token, "baja-motivo-vacio", 1)
	requireStatus(t, emitirRes, http.StatusCreated)

	res := h.request(t, http.MethodPost, "/api/v1/comprobantes/"+emitido.ID+"/baja", seed.Token,
		map[string]string{"motivo": "   "})
	requireStatus(t, res, http.StatusBadRequest)
	res.Body.Close()
}

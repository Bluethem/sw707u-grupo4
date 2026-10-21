package apitest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

const consultaSeedRUC = "20100070970"

func emitirBodyConCliente(cantidad int64) map[string]any {
	body := emitirBody(cantidad)
	body["cliente"] = map[string]string{
		"tipo_doc": "DNI", "numero": "12345678", "razon_social": "Cliente Consulta",
	}
	return body
}

func (h *harness) emitirConCliente(
	t *testing.T,
	token, key string,
	cantidad int64,
) (*http.Response, comprobanteAPI) {
	t.Helper()
	res := h.requestWithHeaders(t, http.MethodPost, "/api/v1/comprobantes", token,
		emitirBodyConCliente(cantidad), map[string]string{"Idempotency-Key": key})
	var comprobante comprobanteAPI
	if res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusOK {
		decodeResponse(t, res, &comprobante)
	}
	return res, comprobante
}

func (h *harness) consultaPublica(
	t *testing.T,
	body map[string]any,
	headers map[string]string,
) *http.Response {
	t.Helper()
	return h.requestWithHeaders(t, http.MethodPost, "/api/v1/publico/consulta", "", body, headers)
}

func comprobanteFechaTotal(t *testing.T, h *harness, id string) (string, string) {
	t.Helper()
	var fecha time.Time
	var total decimal.Decimal
	err := h.pool.QueryRow(context.Background(), `
		SELECT creado_en::date, total FROM comprobantes WHERE id = $1`, id,
	).Scan(&fecha, &total)
	if err != nil {
		t.Fatalf("leer comprobante para consulta: %v", err)
	}
	return fecha.Format(time.DateOnly), total.StringFixed(2)
}

func TestAPI_ConsultaPublica_OK(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, emitido := h.emitirConCliente(t, seed.Token, "consulta-ok", 2)
	requireStatus(t, emitirRes, http.StatusCreated)
	fecha, monto := comprobanteFechaTotal(t, h, emitido.ID)

	res := h.consultaPublica(t, map[string]any{
		"ruc": consultaSeedRUC, "tipo_doc": "DNI", "numero": "12345678",
		"serie": "B001", "numero_comprobante": emitido.Numero,
		"fecha": fecha, "monto": monto,
	}, nil)
	requireStatus(t, res, http.StatusOK)
	var body struct {
		ID                string `json:"id"`
		Token             string `json:"token"`
		RazonSocialEmisor string `json:"razon_social_emisor"`
	}
	decodeResponse(t, res, &body)
	if body.ID != emitido.ID || body.Token == "" || body.RazonSocialEmisor == "" {
		t.Fatalf("respuesta consulta inesperada: %+v", body)
	}

	pdfRes := h.request(t, http.MethodGet,
		fmt.Sprintf("/api/v1/publico/comprobantes/%s/pdf?token=%s", emitido.ID, body.Token), "", nil)
	requireStatus(t, pdfRes, http.StatusOK)
	pdfRes.Body.Close()
}

func TestAPI_ConsultaPublica_DatosNoCoinciden404(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, emitido := h.emitirConCliente(t, seed.Token, "consulta-404", 1)
	requireStatus(t, emitirRes, http.StatusCreated)
	fecha, _ := comprobanteFechaTotal(t, h, emitido.ID)

	res := h.consultaPublica(t, map[string]any{
		"ruc": consultaSeedRUC, "tipo_doc": "DNI", "numero": "12345678",
		"serie": "B001", "numero_comprobante": emitido.Numero,
		"fecha": fecha, "monto": "0.01",
	}, nil)
	requireStatus(t, res, http.StatusNotFound)
	var errBody struct {
		Error string `json:"error"`
	}
	decodeResponse(t, res, &errBody)
	if errBody.Error != "datos no coinciden" {
		t.Fatalf("error = %q, esperado datos no coinciden", errBody.Error)
	}
}

func TestAPI_ConsultaPublica_Expirada410(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, emitido := h.emitirConCliente(t, seed.Token, "consulta-410", 1)
	requireStatus(t, emitirRes, http.StatusCreated)
	fecha, monto := comprobanteFechaTotal(t, h, emitido.ID)
	if _, err := h.pool.Exec(context.Background(), `
		UPDATE comprobantes SET creado_en = NOW() - INTERVAL '366 days' WHERE id = $1`, emitido.ID); err != nil {
		t.Fatalf("retroceder creado_en: %v", err)
	}
	fecha, _ = comprobanteFechaTotal(t, h, emitido.ID)

	res := h.consultaPublica(t, map[string]any{
		"ruc": consultaSeedRUC, "tipo_doc": "DNI", "numero": "12345678",
		"serie": "B001", "numero_comprobante": emitido.Numero,
		"fecha": fecha, "monto": monto,
	}, nil)
	requireStatus(t, res, http.StatusGone)
}

func TestAPI_ConsultaPublica_RateLimit429(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, emitido := h.emitirConCliente(t, seed.Token, "consulta-429", 1)
	requireStatus(t, emitirRes, http.StatusCreated)
	fecha, _ := comprobanteFechaTotal(t, h, emitido.ID)
	headers := map[string]string{"X-Forwarded-For": "203.0.113.50"}
	body := map[string]any{
		"ruc": consultaSeedRUC, "tipo_doc": "DNI", "numero": "12345678",
		"serie": "B001", "numero_comprobante": emitido.Numero,
		"fecha": fecha, "monto": "0.01",
	}
	for i := 0; i < 10; i++ {
		res := h.consultaPublica(t, body, headers)
		requireStatus(t, res, http.StatusNotFound)
		res.Body.Close()
	}
	res := h.consultaPublica(t, body, headers)
	requireStatus(t, res, http.StatusTooManyRequests)
	res.Body.Close()
}

func TestAPI_ConsultaPublica_TokenInvalido401(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, emitido := h.emitirConCliente(t, seed.Token, "consulta-token", 1)
	requireStatus(t, emitirRes, http.StatusCreated)

	badTokenRes := h.request(t, http.MethodGet,
		fmt.Sprintf("/api/v1/publico/comprobantes/%s/pdf?token=invalido", emitido.ID), "", nil)
	requireStatus(t, badTokenRes, http.StatusUnauthorized)
	badTokenRes.Body.Close()

	fecha, monto := comprobanteFechaTotal(t, h, emitido.ID)
	res := h.consultaPublica(t, map[string]any{
		"ruc": consultaSeedRUC, "tipo_doc": "DNI", "numero": "12345678",
		"serie": "B001", "numero_comprobante": emitido.Numero,
		"fecha": fecha, "monto": monto,
	}, nil)
	requireStatus(t, res, http.StatusOK)
	var okBody struct {
		Token string `json:"token"`
	}
	decodeResponse(t, res, &okBody)

	mismatchRes := h.request(t, http.MethodGet,
		fmt.Sprintf("/api/v1/publico/comprobantes/otro-id/pdf?token=%s", okBody.Token), "", nil)
	requireStatus(t, mismatchRes, http.StatusUnauthorized)
	mismatchRes.Body.Close()

	xmlRes := h.request(t, http.MethodGet,
		fmt.Sprintf("/api/v1/publico/comprobantes/%s/xml?token=%s", emitido.ID, okBody.Token), "", nil)
	requireStatus(t, xmlRes, http.StatusOK)
	if _, err := io.ReadAll(xmlRes.Body); err != nil {
		t.Fatalf("leer xml publico: %v", err)
	}
	xmlRes.Body.Close()
}

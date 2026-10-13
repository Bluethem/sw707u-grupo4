package apitest

import (
	"net/http"
	"testing"
)

func TestE2E_EmitirComprobante(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)

	res := h.request(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email": seed.AdminEmail, "password": seed.AdminPassword,
	})
	requireStatus(t, res, http.StatusOK)
	var loginBody struct {
		Token string `json:"token"`
	}
	decodeResponse(t, res, &loginBody)
	if loginBody.Token == "" {
		t.Fatal("login no devolvio token")
	}
	token := loginBody.Token

	const idempotencyKey = "e2e-emitir-key"
	emitirRes, emitido := h.emitir(t, token, idempotencyKey, 2)
	requireStatus(t, emitirRes, http.StatusCreated)
	if emitido.ID == "" {
		t.Fatal("emitir no devolvio id")
	}

	listRes := h.request(t, http.MethodGet, "/api/v1/comprobantes?limit=50", token, nil)
	requireStatus(t, listRes, http.StatusOK)
	var listado struct {
		Items []comprobanteAPI `json:"items"`
		Total int              `json:"total"`
	}
	decodeResponse(t, listRes, &listado)
	found := false
	for _, item := range listado.Items {
		if item.ID == emitido.ID {
			found = true
			break
		}
	}
	if !found || listado.Total < 1 {
		t.Fatalf("listado no contiene comprobante emitido: id=%q listado=%+v", emitido.ID, listado)
	}

	reemitRes, reemitido := h.emitir(t, token, idempotencyKey, 2)
	requireStatus(t, reemitRes, http.StatusOK)
	if reemitido.ID != emitido.ID {
		t.Fatalf("re-POST idempotente devolvio otro id: %q vs %q", reemitido.ID, emitido.ID)
	}
}

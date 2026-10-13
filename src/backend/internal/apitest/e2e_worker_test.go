package apitest

import (
	"context"
	"net/http"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/sunat"
)

func TestE2E_EmitirYAceptarSUNATFake(t *testing.T) {
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
	token := loginBody.Token
	if token == "" {
		t.Fatal("login no devolvio token")
	}

	emitirRes, emitido := h.emitir(t, token, "e2e-worker-aceptar", 1)
	requireStatus(t, emitirRes, http.StatusCreated)
	if emitido.Estado != "PENDIENTE" {
		t.Fatalf("estado inicial = %q; want PENDIENTE", emitido.Estado)
	}

	proc := h.newProcesadorEnvio(sunat.ModeAccept)
	if err := proc.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	getRes := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+emitido.ID, token, nil)
	requireStatus(t, getRes, http.StatusOK)
	var got comprobanteAPI
	decodeResponse(t, getRes, &got)
	if got.Estado != "ACEPTADO" {
		t.Fatalf("estado tras worker = %q; want ACEPTADO", got.Estado)
	}
}

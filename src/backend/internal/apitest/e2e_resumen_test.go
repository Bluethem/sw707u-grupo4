package apitest

import (
	"context"
	"net/http"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/sunat"
)

func TestE2E_ResumenDiario(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)

	emitirRes, emitido := h.emitir(t, seed.Token, "e2e-resumen", 1)
	requireStatus(t, emitirRes, http.StatusCreated)
	if emitido.Estado != "PENDIENTE" {
		t.Fatalf("estado inicial = %q; want PENDIENTE", emitido.Estado)
	}

	proc := h.newProcesadorEnvio(sunat.ModeAccept)
	if err := proc.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce envio: %v", err)
	}

	fecha := comprobanteFecha(t, h, emitido.ID)
	genRes := h.request(t, http.MethodPost, "/api/v1/resumenes", seed.Token, map[string]string{"fecha": fecha})
	requireStatus(t, genRes, http.StatusCreated)
	var generado resumenAPI
	decodeResponse(t, genRes, &generado)

	if err := proc.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce resumen: %v", err)
	}

	getRes := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+emitido.ID, seed.Token, nil)
	requireStatus(t, getRes, http.StatusOK)
	var got comprobanteAPI
	decodeResponse(t, getRes, &got)
	if got.Estado != "ACEPTADO" {
		t.Fatalf("boleta tras resumen = %q; want ACEPTADO", got.Estado)
	}

	detRes := h.request(t, http.MethodGet, "/api/v1/resumenes/"+generado.ID, seed.Token, nil)
	requireStatus(t, detRes, http.StatusOK)
	var det resumenAPI
	decodeResponse(t, detRes, &det)
	if det.Estado != "ACEPTADO" {
		t.Fatalf("resumen = %q; want ACEPTADO", det.Estado)
	}
}

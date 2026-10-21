package apitest

import (
	"context"
	"net/http"
	"testing"
)

func TestAPI_Descarga_PDFXML_OK(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	res, c := h.emitir(t, seed.Token, "desc-ok", 1)
	requireStatus(t, res, http.StatusCreated)
	pdfRes := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+c.ID+"/pdf", seed.Token, nil)
	requireStatus(t, pdfRes, http.StatusOK)
	xmlRes := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+c.ID+"/xml", seed.Token, nil)
	requireStatus(t, xmlRes, http.StatusOK)
}

func TestAPI_Descarga_CDR_Pendiente409(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	_, c := h.emitir(t, seed.Token, "desc-cdr", 1)
	res := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+c.ID+"/cdr", seed.Token, nil)
	requireStatus(t, res, http.StatusConflict)
}

func TestAPI_Descarga_CDR_Final404(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	_, c := h.emitir(t, seed.Token, "desc-cdr-final404", 1)
	if _, err := h.pool.Exec(context.Background(), `
		UPDATE comprobantes SET estado = 'ACEPTADO' WHERE id = $1`, c.ID); err != nil {
		t.Fatalf("actualizar estado ACEPTADO: %v", err)
	}
	if _, ok := h.storage.Get(c.ID + ".cdr"); ok {
		t.Fatalf("storage no debe tener %s.cdr", c.ID)
	}
	res := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+c.ID+"/cdr", seed.Token, nil)
	requireStatus(t, res, http.StatusNotFound)
}

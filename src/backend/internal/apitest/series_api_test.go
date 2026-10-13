package apitest

import (
	"net/http"
	"testing"
)

func TestAPI_SeriesRBAC(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	res := h.request(t, http.MethodPost, "/api/v1/series", seed.Token, map[string]any{
		"tipo": "factura", "serie": "F001", "siguiente": 7,
	})
	requireStatus(t, res, http.StatusCreated)
	var serie struct {
		EmpresaID string `json:"empresa_id"`
		Tipo      string `json:"tipo"`
		Serie     string `json:"serie"`
		Siguiente int64  `json:"siguiente"`
	}
	decodeResponse(t, res, &serie)
	if serie.EmpresaID != seed.EmpresaID || serie.Tipo != "factura" || serie.Serie != "F001" || serie.Siguiente != 7 {
		t.Fatalf("serie creada inesperada: %+v", serie)
	}

	res = h.request(t, http.MethodGet, "/api/v1/series", seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var series []struct {
		Serie string `json:"serie"`
	}
	decodeResponse(t, res, &series)
	if len(series) != 1 || series[0].Serie != "F001" {
		t.Fatalf("series listadas inesperadas: %+v", series)
	}

	res = h.request(t, http.MethodPost, "/api/v1/series", seed.Token, map[string]any{
		"tipo": "boleta", "serie": "F002",
	})
	requireStatus(t, res, http.StatusBadRequest)
	res.Body.Close()

	cajeroToken := seedCajero(t, h, seed.EmpresaID)
	res = h.request(t, http.MethodPost, "/api/v1/series", cajeroToken, map[string]any{
		"tipo": "boleta", "serie": "B001",
	})
	requireStatus(t, res, http.StatusForbidden)
	res.Body.Close()
}

func TestAPI_EmpresasMe(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	res := h.request(t, http.MethodGet, "/api/v1/empresas/me", seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var empresa struct {
		ID          string `json:"id"`
		RazonSocial string `json:"razon_social"`
	}
	decodeResponse(t, res, &empresa)
	if empresa.ID != seed.EmpresaID || empresa.RazonSocial != "Empresa API SAC" {
		t.Fatalf("empresa inesperada: %+v", empresa)
	}

	res = h.request(t, http.MethodPatch, "/api/v1/empresas/me", seed.Token, map[string]any{
		"razon_social":     "Empresa Actualizada SAC",
		"domicilio_fiscal": "Arequipa",
	})
	requireStatus(t, res, http.StatusOK)
	var actualizada struct {
		RazonSocial     string `json:"razon_social"`
		DomicilioFiscal string `json:"domicilio_fiscal"`
	}
	decodeResponse(t, res, &actualizada)
	if actualizada.RazonSocial != "Empresa Actualizada SAC" || actualizada.DomicilioFiscal != "Arequipa" {
		t.Fatalf("empresa actualizada inesperada: %+v", actualizada)
	}
}

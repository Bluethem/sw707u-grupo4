package apitest

import (
	"net/http"
	"testing"
)

type configCatalogoAPI struct {
	EmpresaID              string   `json:"empresa_id"`
	PrecioMinimo           string   `json:"precio_minimo"`
	PrecioMaximo           *string  `json:"precio_maximo"`
	CargaMaxFilas          int      `json:"carga_max_filas"`
	CargaMaxMB             int      `json:"carga_max_mb"`
	UnidadesPermitidas     []string `json:"unidades_permitidas"`
	AfectacionesPermitidas []string `json:"afectaciones_permitidas"`
}

func TestAPI_ConfigCatalogo_GetDefaultYPatch(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	res := h.request(t, http.MethodGet, "/api/v1/configuracion/catalogo", seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var cfg configCatalogoAPI
	decodeResponse(t, res, &cfg)
	if cfg.EmpresaID != seed.EmpresaID {
		t.Fatalf("empresa_id=%q want %q", cfg.EmpresaID, seed.EmpresaID)
	}
	if cfg.PrecioMinimo == "" || cfg.CargaMaxFilas < 1 || len(cfg.UnidadesPermitidas) == 0 {
		t.Fatalf("defaults incompletos: %+v", cfg)
	}

	res = h.request(t, http.MethodPatch, "/api/v1/configuracion/catalogo", seed.Token, map[string]any{
		"precio_minimo":           "1.50",
		"precio_maximo":           "50.00",
		"carga_max_filas":         25,
		"carga_max_mb":            1,
		"unidades_permitidas":     []string{"NIU", "ZZ"},
		"afectaciones_permitidas": []string{"gravado"},
	})
	requireStatus(t, res, http.StatusOK)
	decodeResponse(t, res, &cfg)
	if cfg.PrecioMinimo != "1.5" && cfg.PrecioMinimo != "1.50" {
		t.Fatalf("precio_minimo=%q", cfg.PrecioMinimo)
	}
	if cfg.PrecioMaximo == nil || (*cfg.PrecioMaximo != "50" && *cfg.PrecioMaximo != "50.00") {
		t.Fatalf("precio_maximo=%v", cfg.PrecioMaximo)
	}
	if cfg.CargaMaxFilas != 25 || cfg.CargaMaxMB != 1 {
		t.Fatalf("limites=%d/%d", cfg.CargaMaxFilas, cfg.CargaMaxMB)
	}

	// Producto fuera de rango debe rechazarse con la config activa.
	res, _ = h.crearProducto(t, seed.Token, map[string]any{
		"codigo": "CFG-HI", "descripcion": "caro", "unidad": "NIU",
		"precio": "99.00", "afectacion": "gravado", "controla_stock": false,
	})
	requireStatus(t, res, http.StatusBadRequest)
	res.Body.Close()

	res, _ = h.crearProducto(t, seed.Token, map[string]any{
		"codigo": "CFG-OK", "descripcion": "ok", "unidad": "NIU",
		"precio": "10.00", "afectacion": "gravado", "controla_stock": false,
	})
	requireStatus(t, res, http.StatusCreated)

	// Limpiar límite superior.
	res = h.request(t, http.MethodPatch, "/api/v1/configuracion/catalogo", seed.Token, map[string]any{
		"precio_maximo": "",
	})
	requireStatus(t, res, http.StatusOK)
	decodeResponse(t, res, &cfg)
	if cfg.PrecioMaximo != nil {
		t.Fatalf("precio_maximo debería ser null, got %v", *cfg.PrecioMaximo)
	}
}

func TestAPI_ConfigCatalogo_PatchInvalido(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	res := h.request(t, http.MethodPatch, "/api/v1/configuracion/catalogo", seed.Token, map[string]any{
		"precio_minimo": "10",
		"precio_maximo": "1",
	})
	requireStatus(t, res, http.StatusBadRequest)
	res.Body.Close()

	res = h.request(t, http.MethodPatch, "/api/v1/configuracion/catalogo", seed.Token, map[string]any{
		"unidades_permitidas": []string{},
	})
	requireStatus(t, res, http.StatusBadRequest)
	res.Body.Close()
}

func TestAPI_ConfigCatalogo_CajeroForbidden(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)
	cajero := seedCajero(t, h, seed.EmpresaID)

	res := h.request(t, http.MethodGet, "/api/v1/configuracion/catalogo", cajero, nil)
	requireStatus(t, res, http.StatusForbidden)
	res.Body.Close()
}

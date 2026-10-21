package apitest

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestAPI_ListarComprobantes_RangoFechasExcede90Dias(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)

	desde := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	hasta := desde.AddDate(0, 0, 91)
	path := "/api/v1/comprobantes?fecha_desde=" + desde.Format(time.DateOnly) +
		"&fecha_hasta=" + hasta.Format(time.DateOnly)

	res := h.request(t, http.MethodGet, path, seed.Token, nil)
	requireStatus(t, res, http.StatusBadRequest)
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatalf("leer body: %v", err)
	}
	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		t.Fatalf("decodificar error: %v body=%s", err, string(body))
	}
	if errResp.Error != "rango de fechas excede 90 dias" {
		t.Fatalf("mensaje inesperado: %q", errResp.Error)
	}
}

func TestAPI_ListarComprobantes_SerieYQ(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, emitido := h.emitir(t, seed.Token, "listado-serie-q", 1)
	requireStatus(t, emitirRes, http.StatusCreated)

	res := h.request(t, http.MethodGet,
		"/api/v1/comprobantes?serie=b001&numero=1",
		seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var listado struct {
		Items []comprobanteAPI `json:"items"`
		Total int              `json:"total"`
	}
	decodeResponse(t, res, &listado)
	if listado.Total != 1 || len(listado.Items) != 1 || listado.Items[0].ID != emitido.ID {
		t.Fatalf("filtro serie/numero: %+v", listado)
	}

	res = h.request(t, http.MethodGet, "/api/v1/comprobantes?serie=F001", seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	decodeResponse(t, res, &listado)
	if listado.Total != 0 {
		t.Fatalf("serie inexistente deberia estar vacio: %+v", listado)
	}
}

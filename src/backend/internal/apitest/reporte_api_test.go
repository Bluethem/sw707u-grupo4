package apitest

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

type ventasTipoAPI struct {
	Tipo       string `json:"tipo"`
	Cantidad   int    `json:"cantidad"`
	ValorVenta string `json:"valor_venta"`
	IGV        string `json:"igv"`
	Total      string `json:"total"`
}

type reporteVentasAPI struct {
	FechaDesde string `json:"fecha_desde"`
	FechaHasta string `json:"fecha_hasta"`
	PorTipo    []ventasTipoAPI `json:"por_tipo"`
	Anulados   struct {
		Cantidad int    `json:"cantidad"`
		Total    string `json:"total"`
	} `json:"anulados"`
	Totales ventasTipoAPI `json:"totales"`
}

func todayReportQuery() string {
	return "?fecha_desde=" + time.Now().UTC().Format(time.DateOnly) +
		"&fecha_hasta=" + time.Now().UTC().Format(time.DateOnly)
}

func (h *harness) reporteVentas(t *testing.T, token, query string) (*http.Response, reporteVentasAPI) {
	t.Helper()
	res := h.request(t, http.MethodGet, "/api/v1/reportes/ventas"+query, token, nil)
	var body reporteVentasAPI
	if res.StatusCode == http.StatusOK {
		decodeResponse(t, res, &body)
	}
	return res, body
}

func TestAPI_ReporteVentas_DosBoletasUnaAnulada(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)

	res1, b1 := h.emitir(t, seed.Token, "rep-b1", 1)
	requireStatus(t, res1, http.StatusCreated)
	res2, b2 := h.emitir(t, seed.Token, "rep-b2", 1)
	requireStatus(t, res2, http.StatusCreated)

	if _, err := h.pool.Exec(context.Background(), `
		UPDATE comprobantes SET estado = 'ANULADA' WHERE id = $1`, b2.ID); err != nil {
		t.Fatalf("anular boleta: %v", err)
	}
	_ = b1

	res, rep := h.reporteVentas(t, seed.Token, todayReportQuery())
	requireStatus(t, res, http.StatusOK)

	if rep.Totales.Cantidad != 1 {
		t.Fatalf("totales.cantidad = %d; want 1", rep.Totales.Cantidad)
	}
	if rep.Anulados.Cantidad != 1 {
		t.Fatalf("anulados.cantidad = %d; want 1", rep.Anulados.Cantidad)
	}
	wantTotal := decimal.RequireFromString("118.00")
	if got, err := decimal.NewFromString(rep.Totales.Total); err != nil || !got.Equal(wantTotal) {
		t.Fatalf("totales.total = %q; want 118.00", rep.Totales.Total)
	}
	if got, err := decimal.NewFromString(rep.Anulados.Total); err != nil || !got.Equal(wantTotal) {
		t.Fatalf("anulados.total = %q; want 118.00", rep.Anulados.Total)
	}
	if len(rep.PorTipo) != 1 || rep.PorTipo[0].Tipo != "boleta" || rep.PorTipo[0].Cantidad != 1 {
		t.Fatalf("por_tipo inesperado: %+v", rep.PorTipo)
	}
}

func TestAPI_ReporteVentas_RangoExcede90Dias(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)

	res := h.request(t, http.MethodGet,
		"/api/v1/reportes/ventas?fecha_desde=2025-01-01&fecha_hasta=2025-05-01",
		seed.Token, nil)
	requireStatus(t, res, http.StatusBadRequest)
	res.Body.Close()
}

func TestAPI_ReporteVentas_CSV(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, _ := h.emitir(t, seed.Token, "rep-csv", 1)
	requireStatus(t, emitirRes, http.StatusCreated)

	res := h.request(t, http.MethodGet, "/api/v1/reportes/ventas.csv"+todayReportQuery(), seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	defer res.Body.Close()
	if ct := res.Header.Get("Content-Type"); !strings.Contains(ct, "text/csv") {
		t.Fatalf("content-type = %q", ct)
	}
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("leer csv: %v", err)
	}
	text := string(raw)
	if !strings.HasPrefix(text, "tipo,cantidad,valor_venta,igv,total\n") {
		t.Fatalf("cabecera csv inesperada: %q", text)
	}
	if !strings.Contains(text, "boleta,") || !strings.Contains(text, "totales,") {
		t.Fatalf("filas csv inesperadas: %q", text)
	}
}

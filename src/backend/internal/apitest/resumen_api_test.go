package apitest

import (
	"context"
	"net/http"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/sunat"
)

type resumenAPI struct {
	ID           string `json:"id"`
	Fecha        string `json:"fecha"`
	Estado       string `json:"estado"`
	Correlativo  int    `json:"correlativo"`
	Comprobantes []struct {
		ComprobanteID string `json:"comprobante_id"`
		Anulado       bool   `json:"anulado"`
		Tipo          string `json:"tipo"`
		Serie         string `json:"serie"`
		Numero        int64  `json:"numero"`
		Estado        string `json:"estado"`
	} `json:"comprobantes"`
}

func comprobanteFecha(t *testing.T, h *harness, id string) string {
	t.Helper()
	var fecha string
	if err := h.pool.QueryRow(context.Background(), `
		SELECT to_char(creado_en, 'YYYY-MM-DD') FROM comprobantes WHERE id=$1`, id,
	).Scan(&fecha); err != nil {
		t.Fatalf("fecha comprobante: %v", err)
	}
	return fecha
}

func TestAPI_Resumen_GenerarListarObtener(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, emitido := h.emitir(t, seed.Token, "resumen-api-1", 1)
	requireStatus(t, emitirRes, http.StatusCreated)
	aceptarComprobante(t, h)

	fecha := comprobanteFecha(t, h, emitido.ID)
	genRes := h.request(t, http.MethodPost, "/api/v1/resumenes", seed.Token, map[string]string{"fecha": fecha})
	requireStatus(t, genRes, http.StatusCreated)
	var generado resumenAPI
	decodeResponse(t, genRes, &generado)
	if generado.Estado != "PENDIENTE" {
		t.Fatalf("resumen generado inesperado: %+v", generado)
	}
	assertResumenReferenciaBoleta(t, generado, emitido)

	listRes := h.request(t, http.MethodGet, "/api/v1/resumenes", seed.Token, nil)
	requireStatus(t, listRes, http.StatusOK)
	var lista []resumenAPI
	decodeResponse(t, listRes, &lista)
	if len(lista) != 1 || lista[0].ID != generado.ID {
		t.Fatalf("lista resumenes: %+v", lista)
	}
	assertResumenReferenciaBoleta(t, lista[0], emitido)

	getRes := h.request(t, http.MethodGet, "/api/v1/resumenes/"+generado.ID, seed.Token, nil)
	requireStatus(t, getRes, http.StatusOK)
	var detalle resumenAPI
	decodeResponse(t, getRes, &detalle)
	if detalle.ID != generado.ID {
		t.Fatalf("detalle resumen: %+v", detalle)
	}
	assertResumenReferenciaBoleta(t, detalle, emitido)
}

func assertResumenReferenciaBoleta(t *testing.T, resumen resumenAPI, boleta comprobanteAPI) {
	t.Helper()
	if len(resumen.Comprobantes) != 1 {
		t.Fatalf("resumen debe referenciar las boletas del día, got %+v", resumen.Comprobantes)
	}
	item := resumen.Comprobantes[0]
	if item.ComprobanteID != boleta.ID || item.Anulado {
		t.Fatalf("referencia inesperada: %+v", item)
	}
	if item.Tipo != boleta.Tipo || item.Serie != boleta.Serie || item.Numero != boleta.Numero {
		t.Fatalf("serie/numero ausentes en referencia: %+v (boleta %+v)", item, boleta)
	}
}

func TestAPI_Resumen_SinBoletas(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	res := h.request(t, http.MethodPost, "/api/v1/resumenes", seed.Token, map[string]string{"fecha": "2020-01-01"})
	requireStatus(t, res, http.StatusConflict)
}

func TestAPI_Resumen_WorkerAceptaBoletaYBaja(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	_, b1 := h.emitir(t, seed.Token, "res-w1", 1)
	_, b2 := h.emitir(t, seed.Token, "res-w2", 1)
	aceptarComprobante(t, h)
	aceptarComprobante(t, h)

	bajaRes := h.request(t, http.MethodPost, "/api/v1/comprobantes/"+b2.ID+"/baja", seed.Token,
		map[string]string{"motivo": "Devolucion"})
	requireStatus(t, bajaRes, http.StatusOK)

	fecha := comprobanteFecha(t, h, b1.ID)
	genRes := h.request(t, http.MethodPost, "/api/v1/resumenes", seed.Token, map[string]string{"fecha": fecha})
	requireStatus(t, genRes, http.StatusCreated)
	var generado resumenAPI
	decodeResponse(t, genRes, &generado)

	proc := h.newProcesadorEnvio(sunat.ModeAccept)
	if err := proc.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce resumen: %v", err)
	}

	getRes := h.request(t, http.MethodGet, "/api/v1/resumenes/"+generado.ID, seed.Token, nil)
	requireStatus(t, getRes, http.StatusOK)
	var detalle resumenAPI
	decodeResponse(t, getRes, &detalle)
	if detalle.Estado != "ACEPTADO" {
		t.Fatalf("resumen estado = %q; want ACEPTADO", detalle.Estado)
	}

	getB1 := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+b1.ID, seed.Token, nil)
	requireStatus(t, getB1, http.StatusOK)
	var got1 comprobanteAPI
	decodeResponse(t, getB1, &got1)
	if got1.Estado != "ACEPTADO" {
		t.Fatalf("b1 estado = %q; want ACEPTADO", got1.Estado)
	}

	getB2 := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+b2.ID, seed.Token, nil)
	requireStatus(t, getB2, http.StatusOK)
	var got2 comprobanteAPI
	decodeResponse(t, getB2, &got2)
	if got2.Estado != "ANULADA" {
		t.Fatalf("b2 estado = %q; want ANULADA", got2.Estado)
	}
}

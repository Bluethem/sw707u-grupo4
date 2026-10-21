package apitest

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/sunat"
)

func TestE2E_Documentos_EmitirDescargarYConsultar(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)

	emitirRes, emitido := h.emitirConCliente(t, seed.Token, "e2e-docs-flujo", 2)
	requireStatus(t, emitirRes, http.StatusCreated)
	if emitido.Estado != "PENDIENTE" {
		t.Fatalf("estado inicial = %q; want PENDIENTE", emitido.Estado)
	}

	proc := h.newProcesadorEnvio(sunat.ModeAccept)
	if err := proc.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	getRes := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+emitido.ID, seed.Token, nil)
	requireStatus(t, getRes, http.StatusOK)
	var got comprobanteAPI
	decodeResponse(t, getRes, &got)
	if got.Estado != "ACEPTADO" {
		t.Fatalf("estado tras worker = %q; want ACEPTADO", got.Estado)
	}

	pdfRes := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+emitido.ID+"/pdf", seed.Token, nil)
	requireStatus(t, pdfRes, http.StatusOK)
	pdfRes.Body.Close()

	xmlRes := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+emitido.ID+"/xml", seed.Token, nil)
	requireStatus(t, xmlRes, http.StatusOK)
	xmlRes.Body.Close()

	cdrRes := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+emitido.ID+"/cdr", seed.Token, nil)
	requireStatus(t, cdrRes, http.StatusOK)
	cdrRes.Body.Close()

	fecha, monto := comprobanteFechaTotal(t, h, emitido.ID)
	consultaRes := h.consultaPublica(t, map[string]any{
		"ruc": consultaSeedRUC, "tipo_doc": "DNI", "numero": "12345678",
		"serie": "B001", "numero_comprobante": emitido.Numero,
		"fecha": fecha, "monto": monto,
	}, nil)
	requireStatus(t, consultaRes, http.StatusOK)
	var consultaBody struct {
		Token string `json:"token"`
	}
	decodeResponse(t, consultaRes, &consultaBody)
	if consultaBody.Token == "" {
		t.Fatal("consulta sin token")
	}

	pubPDF := h.request(t, http.MethodGet,
		fmt.Sprintf("/api/v1/publico/comprobantes/%s/pdf?token=%s", emitido.ID, consultaBody.Token), "", nil)
	requireStatus(t, pubPDF, http.StatusOK)
	pubPDF.Body.Close()
}

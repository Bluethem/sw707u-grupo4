package apitest

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	redisadapter "github.com/Bluethem/sw707u-grupo4/src/backend/adapters/redis"
)

type comprobanteAPI struct {
	ID             string `json:"id"`
	Tipo           string `json:"tipo"`
	Serie          string `json:"serie"`
	Numero         int64  `json:"numero"`
	Estado         string `json:"estado"`
	IdempotencyKey string `json:"idempotency_key"`
	Lineas         []struct {
		ProductoID string `json:"producto_id"`
		Cantidad   int64  `json:"cantidad"`
	} `json:"lineas"`
}

func emitirBody(cantidad int64) map[string]any {
	return map[string]any{
		"tipo": "boleta", "serie": "B001",
		"lineas": []map[string]any{{
			"producto_id": "p1", "cantidad": cantidad,
			"precio_unitario": "118.00", "afectacion": "gravado",
		}},
	}
}

func (h *harness) emitir(
	t *testing.T,
	token, key string,
	cantidad int64,
) (*http.Response, comprobanteAPI) {
	t.Helper()
	res := h.requestWithHeaders(t, http.MethodPost, "/api/v1/comprobantes", token,
		emitirBody(cantidad), map[string]string{"Idempotency-Key": key})
	var comprobante comprobanteAPI
	if res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusOK {
		decodeResponse(t, res, &comprobante)
	}
	return res, comprobante
}

func TestAPI_Emitir_OK(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)

	res, got := h.emitir(t, seed.Token, "emitir-ok", 2)
	requireStatus(t, res, http.StatusCreated)
	if got.ID == "" || got.Tipo != "boleta" || got.Serie != "B001" ||
		got.Numero != 1 || got.Estado != "PENDIENTE" || len(got.Lineas) != 1 {
		t.Fatalf("comprobante inesperado: %+v", got)
	}
	if got.Lineas[0].ProductoID != "p1" || got.Lineas[0].Cantidad != 2 {
		t.Fatalf("linea inesperada: %+v", got.Lineas[0])
	}
	var stock int64
	if err := h.pool.QueryRow(context.Background(), `
		SELECT cantidad FROM productos_stock WHERE empresa_id=$1 AND producto_id='p1'`,
		seed.EmpresaID,
	).Scan(&stock); err != nil {
		t.Fatalf("consultar stock: %v", err)
	}
	if stock != 8 {
		t.Fatalf("stock = %d; esperado 8", stock)
	}
}

func TestAPI_Emitir_Idempotencia(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)

	primeraRes, primera := h.emitir(t, seed.Token, "misma-key-api", 2)
	requireStatus(t, primeraRes, http.StatusCreated)
	segundaRes, segunda := h.emitir(t, seed.Token, "misma-key-api", 2)
	requireStatus(t, segundaRes, http.StatusOK)
	if primera.ID == "" || primera.ID != segunda.ID {
		t.Fatalf("IDs idempotentes distintos: %q y %q", primera.ID, segunda.ID)
	}
	var comprobantes int
	var stock int64
	if err := h.pool.QueryRow(context.Background(), `
		SELECT count(*) FROM comprobantes WHERE empresa_id=$1`, seed.EmpresaID,
	).Scan(&comprobantes); err != nil {
		t.Fatalf("contar comprobantes: %v", err)
	}
	if err := h.pool.QueryRow(context.Background(), `
		SELECT cantidad FROM productos_stock WHERE empresa_id=$1 AND producto_id='p1'`,
		seed.EmpresaID,
	).Scan(&stock); err != nil {
		t.Fatalf("consultar stock: %v", err)
	}
	if comprobantes != 1 || stock != 8 {
		t.Fatalf("idempotencia altero datos: comprobantes=%d stock=%d", comprobantes, stock)
	}
}

func TestAPI_Emitir_StockInsuficiente(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)

	// Precio bajo para que el total quede bajo el umbral boleta sin documento (700).
	body := emitirBody(11)
	body["lineas"] = []map[string]any{{
		"producto_id": "p1", "cantidad": int64(11),
		"precio_unitario": "10.00", "afectacion": "gravado",
	}}
	res := h.requestWithHeaders(t, http.MethodPost, "/api/v1/comprobantes", seed.Token,
		body, map[string]string{"Idempotency-Key": "sin-stock-api"})
	requireStatus(t, res, http.StatusConflict)
	res.Body.Close()

	var comprobantes int
	if err := h.pool.QueryRow(context.Background(), `SELECT count(*) FROM comprobantes`).
		Scan(&comprobantes); err != nil {
		t.Fatalf("contar comprobantes: %v", err)
	}
	if comprobantes != 0 {
		t.Fatalf("se persistieron %d comprobantes sin stock", comprobantes)
	}
}

func TestAPI_Emitir_SinIdempotencyKey(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)

	res := h.request(t, http.MethodPost, "/api/v1/comprobantes", seed.Token, emitirBody(1))
	requireStatus(t, res, http.StatusBadRequest)
	res.Body.Close()
}

func TestAPI_Emitir_SinToken(t *testing.T) {
	h := newHarness(t)
	h.SeedEmisionReady(t)

	res := h.requestWithHeaders(t, http.MethodPost, "/api/v1/comprobantes", "",
		emitirBody(1), map[string]string{"Idempotency-Key": "sin-token"})
	requireStatus(t, res, http.StatusUnauthorized)
	res.Body.Close()
}

func TestAPI_ObtenerComprobante(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, emitido := h.emitir(t, seed.Token, "obtener-api", 1)
	requireStatus(t, emitirRes, http.StatusCreated)

	res := h.request(t, http.MethodGet, "/api/v1/comprobantes/"+emitido.ID, seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var got comprobanteAPI
	decodeResponse(t, res, &got)
	if got.ID != emitido.ID || got.IdempotencyKey != "obtener-api" ||
		len(got.Lineas) != 1 || got.Lineas[0].ProductoID != "p1" {
		t.Fatalf("detalle inesperado: %+v", got)
	}
}

func TestAPI_ListarComprobantes(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmisionReady(t)
	emitirRes, emitido := h.emitir(t, seed.Token, "listar-api", 1)
	requireStatus(t, emitirRes, http.StatusCreated)

	res := h.request(t, http.MethodGet,
		"/api/v1/comprobantes?tipo=boleta&estado=PENDIENTE&limit=10&offset=0",
		seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var listado struct {
		Items []comprobanteAPI `json:"items"`
		Total int              `json:"total"`
	}
	decodeResponse(t, res, &listado)
	if listado.Total != 1 || len(listado.Items) != 1 || listado.Items[0].ID != emitido.ID {
		t.Fatalf("listado filtrado inesperado: %+v", listado)
	}

	res = h.request(t, http.MethodGet, "/api/v1/comprobantes?tipo=factura", seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	decodeResponse(t, res, &listado)
	if listado.Total != 0 || len(listado.Items) != 0 {
		t.Fatalf("filtro por tipo inesperado: %+v", listado)
	}

	const tenantBToken = "token-tenant-b"
	if err := h.sessions.Create(context.Background(), tenantBToken, redisadapter.SessionData{
		UserID: "usuario-b", EmpresaID: "empresa-b", Rol: "administrador",
	}); err != nil {
		t.Fatalf("crear sesion tenant B: %v", err)
	}
	res = h.request(t, http.MethodGet, "/api/v1/comprobantes", tenantBToken, nil)
	requireStatus(t, res, http.StatusOK)
	decodeResponse(t, res, &listado)
	if listado.Total != 0 || len(listado.Items) != 0 {
		t.Fatalf("tenant B vio comprobantes de A: %+v", listado)
	}
	res = h.request(t, http.MethodGet,
		fmt.Sprintf("/api/v1/comprobantes/%s", emitido.ID), tenantBToken, nil)
	requireStatus(t, res, http.StatusNotFound)
	res.Body.Close()
}

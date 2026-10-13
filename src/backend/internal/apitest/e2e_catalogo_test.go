package apitest

import (
	"context"
	"net/http"
	"testing"
)

func TestE2E_CatalogoProducto(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	res := h.request(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email": seed.AdminEmail, "password": seed.AdminPassword,
	})
	requireStatus(t, res, http.StatusOK)
	var loginBody struct {
		Token string `json:"token"`
	}
	decodeResponse(t, res, &loginBody)
	if loginBody.Token == "" {
		t.Fatal("login no devolvio token")
	}
	token := loginBody.Token

	body := productoBody("E2E-CAT")
	body["stock_inicial"] = int64(10)
	crearRes, creado := h.crearProducto(t, token, body)
	requireStatus(t, crearRes, http.StatusCreated)
	if creado.ID == "" || creado.Stock == nil || *creado.Stock != 10 {
		t.Fatalf("producto creado inesperado: %+v", creado)
	}

	listRes := h.request(t, http.MethodGet, "/api/v1/productos?limit=50", token, nil)
	requireStatus(t, listRes, http.StatusOK)
	var listado struct {
		Items []productoAPI `json:"items"`
		Total int           `json:"total"`
	}
	decodeResponse(t, listRes, &listado)
	found := false
	for _, item := range listado.Items {
		if item.ID == creado.ID {
			found = true
			break
		}
	}
	if !found || listado.Total < 1 {
		t.Fatalf("listado no contiene producto: id=%q listado=%+v", creado.ID, listado)
	}

	ajusteRes := h.request(t, http.MethodPost, "/api/v1/productos/"+creado.ID+"/ajustes-stock", token, map[string]any{
		"tipo": "salida", "cantidad": 3, "motivo": "venta mostrador",
	})
	requireStatus(t, ajusteRes, http.StatusOK)
	var ajuste struct {
		Cantidad int64 `json:"cantidad"`
	}
	decodeResponse(t, ajusteRes, &ajuste)
	if ajuste.Cantidad != 7 {
		t.Fatalf("saldo tras salida = %d, esperado 7", ajuste.Cantidad)
	}

	movRes := h.request(t, http.MethodGet, "/api/v1/productos/"+creado.ID+"/movimientos?limit=50", token, nil)
	requireStatus(t, movRes, http.StatusOK)
	var movs struct {
		Items []struct {
			Tipo     string `json:"tipo"`
			Cantidad int64  `json:"cantidad"`
		} `json:"items"`
		Total int `json:"total"`
	}
	decodeResponse(t, movRes, &movs)
	if movs.Total < 2 {
		t.Fatalf("movimientos insuficientes: %+v", movs)
	}
	hasInicial, hasSalida := false, false
	for _, m := range movs.Items {
		switch m.Tipo {
		case "inicial":
			hasInicial = true
		case "salida":
			hasSalida = true
			if m.Cantidad != 3 {
				t.Fatalf("cantidad salida inesperada: %+v", m)
			}
		}
	}
	if !hasInicial || !hasSalida {
		t.Fatalf("faltan tipos de movimiento: inicial=%v salida=%v items=%+v", hasInicial, hasSalida, movs.Items)
	}

	if _, err := h.pool.Exec(context.Background(), `
		INSERT INTO series_correlativos (empresa_id, tipo, serie, siguiente, activa)
		VALUES ($1, 'boleta', 'B001', 1, true)`,
		seed.EmpresaID,
	); err != nil {
		t.Fatalf("seed serie emision: %v", err)
	}

	emitirRes := h.requestWithHeaders(t, http.MethodPost, "/api/v1/comprobantes", token, map[string]any{
		"tipo": "boleta", "serie": "B001",
		"lineas": []map[string]any{{
			"producto_id": creado.ID, "cantidad": 1,
			"precio_unitario": "118.00", "afectacion": "gravado",
		}},
	}, map[string]string{"Idempotency-Key": "e2e-catalogo-emitir"})
	requireStatus(t, emitirRes, http.StatusCreated)
	var comprobante comprobanteAPI
	decodeResponse(t, emitirRes, &comprobante)
	if comprobante.ID == "" || len(comprobante.Lineas) != 1 || comprobante.Lineas[0].ProductoID != creado.ID {
		t.Fatalf("comprobante emitido inesperado: %+v", comprobante)
	}

	getRes := h.request(t, http.MethodGet, "/api/v1/productos/"+creado.ID, token, nil)
	requireStatus(t, getRes, http.StatusOK)
	var detalle productoAPI
	decodeResponse(t, getRes, &detalle)
	if detalle.Stock == nil || *detalle.Stock != 6 {
		t.Fatalf("stock tras emision = %v, esperado 6", detalle.Stock)
	}
}

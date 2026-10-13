package apitest

import (
	"fmt"
	"net/http"
	"testing"
)

type productoAPI struct {
	ID            string `json:"id"`
	Codigo        string `json:"codigo"`
	Descripcion   string `json:"descripcion"`
	Unidad        string `json:"unidad"`
	Precio        string `json:"precio"`
	Afectacion    string `json:"afectacion"`
	ControlaStock bool   `json:"controla_stock"`
	Activo        bool   `json:"activo"`
	Stock         *int64 `json:"stock"`
}

func productoBody(codigo string) map[string]any {
	stock := int64(10)
	return map[string]any{
		"codigo": codigo, "descripcion": "Producto API " + codigo, "unidad": "NIU",
		"precio": "118.00", "afectacion": "gravado", "controla_stock": true, "stock_inicial": stock,
	}
}

func (h *harness) crearProducto(t *testing.T, token string, body map[string]any) (*http.Response, productoAPI) {
	t.Helper()
	res := h.request(t, http.MethodPost, "/api/v1/productos", token, body)
	var p productoAPI
	if res.StatusCode == http.StatusCreated {
		decodeResponse(t, res, &p)
	}
	return res, p
}

func TestAPI_Producto_CrearOK(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	res, got := h.crearProducto(t, seed.Token, productoBody("SKU-001"))
	requireStatus(t, res, http.StatusCreated)
	if got.ID == "" || got.Codigo != "SKU-001" || got.Descripcion != "Producto API SKU-001" ||
		got.Unidad != "NIU" || got.Precio != "118.00" || got.Afectacion != "gravado" ||
		!got.ControlaStock || !got.Activo || got.Stock == nil || *got.Stock != 10 {
		t.Fatalf("producto creado inesperado: %+v", got)
	}
}

func TestAPI_Producto_Duplicado(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	res, _ := h.crearProducto(t, seed.Token, productoBody("SKU-DUP"))
	requireStatus(t, res, http.StatusCreated)

	res, _ = h.crearProducto(t, seed.Token, productoBody("SKU-DUP"))
	requireStatus(t, res, http.StatusConflict)
	res.Body.Close()
}

func TestAPI_Producto_CajeroForbiddenCreate(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)
	cajeroToken := seedCajero(t, h, seed.EmpresaID)

	res, _ := h.crearProducto(t, cajeroToken, productoBody("SKU-CAJERO"))
	requireStatus(t, res, http.StatusForbidden)
	res.Body.Close()

	res = h.request(t, http.MethodPost, fmt.Sprintf("/api/v1/productos/%s/ajustes-stock", "cualquier-id"), cajeroToken, map[string]any{
		"tipo": "entrada", "cantidad": 1, "motivo": "test",
	})
	requireStatus(t, res, http.StatusForbidden)
	res.Body.Close()
}

func TestAPI_Producto_ListarYObtener(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	crearRes, creado := h.crearProducto(t, seed.Token, productoBody("SKU-LIST"))
	requireStatus(t, crearRes, http.StatusCreated)

	res := h.request(t, http.MethodGet, "/api/v1/productos?limit=10&offset=0", seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var listado struct {
		Items []productoAPI `json:"items"`
		Total int           `json:"total"`
	}
	decodeResponse(t, res, &listado)
	if listado.Total != 1 || len(listado.Items) != 1 || listado.Items[0].ID != creado.ID {
		t.Fatalf("listado inesperado: %+v", listado)
	}

	res = h.request(t, http.MethodGet, "/api/v1/productos?q=SKU-LIST&limit=10", seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	decodeResponse(t, res, &listado)
	if listado.Total != 1 || listado.Items[0].Codigo != "SKU-LIST" {
		t.Fatalf("filtro q inesperado: %+v", listado)
	}

	res = h.request(t, http.MethodGet, "/api/v1/productos/"+creado.ID, seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var detalle productoAPI
	decodeResponse(t, res, &detalle)
	if detalle.ID != creado.ID || detalle.Stock == nil || *detalle.Stock != 10 {
		t.Fatalf("detalle inesperado: %+v", detalle)
	}

	cajeroToken := seedCajero(t, h, seed.EmpresaID)
	res = h.request(t, http.MethodGet, "/api/v1/productos/"+creado.ID, cajeroToken, nil)
	requireStatus(t, res, http.StatusOK)
	res.Body.Close()
}

func TestAPI_Producto_Desactivar(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	crearRes, creado := h.crearProducto(t, seed.Token, productoBody("SKU-OFF"))
	requireStatus(t, crearRes, http.StatusCreated)

	res := h.request(t, http.MethodPatch, "/api/v1/productos/"+creado.ID, seed.Token, map[string]any{
		"activo": false,
	})
	requireStatus(t, res, http.StatusOK)
	var actualizado productoAPI
	decodeResponse(t, res, &actualizado)
	if actualizado.Activo {
		t.Fatalf("producto sigue activo: %+v", actualizado)
	}

	res = h.request(t, http.MethodGet, "/api/v1/productos?activo=true&limit=10", seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var listado struct {
		Items []productoAPI `json:"items"`
		Total int           `json:"total"`
	}
	decodeResponse(t, res, &listado)
	if listado.Total != 0 {
		t.Fatalf("listado activo=true incluye desactivado: %+v", listado)
	}
}

func TestAPI_AjusteStock_EntradaSalida(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	body := productoBody("SKU-AJ")
	body["stock_inicial"] = int64(5)
	crearRes, creado := h.crearProducto(t, seed.Token, body)
	requireStatus(t, crearRes, http.StatusCreated)

	res := h.request(t, http.MethodPost, "/api/v1/productos/"+creado.ID+"/ajustes-stock", seed.Token, map[string]any{
		"tipo": "entrada", "cantidad": 3, "motivo": "compra",
	})
	requireStatus(t, res, http.StatusOK)
	var ajuste struct {
		ProductoID   string `json:"producto_id"`
		Cantidad     int64  `json:"cantidad"`
		MovimientoID int64  `json:"movimiento_id"`
	}
	decodeResponse(t, res, &ajuste)
	if ajuste.ProductoID != creado.ID || ajuste.Cantidad != 8 || ajuste.MovimientoID == 0 {
		t.Fatalf("ajuste entrada inesperado: %+v", ajuste)
	}

	res = h.request(t, http.MethodPost, "/api/v1/productos/"+creado.ID+"/ajustes-stock", seed.Token, map[string]any{
		"tipo": "salida", "cantidad": 2, "motivo": "merma",
	})
	requireStatus(t, res, http.StatusOK)
	decodeResponse(t, res, &ajuste)
	if ajuste.Cantidad != 6 {
		t.Fatalf("saldo tras salida = %d, esperado 6", ajuste.Cantidad)
	}

	res = h.request(t, http.MethodGet, "/api/v1/productos/"+creado.ID, seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var detalle productoAPI
	decodeResponse(t, res, &detalle)
	if detalle.Stock == nil || *detalle.Stock != 6 {
		t.Fatalf("stock en detalle inesperado: %+v", detalle)
	}
}

func TestAPI_AjusteStock_SalidaExcesiva(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	body := productoBody("SKU-NEG")
	body["stock_inicial"] = int64(2)
	crearRes, creado := h.crearProducto(t, seed.Token, body)
	requireStatus(t, crearRes, http.StatusCreated)

	res := h.request(t, http.MethodPost, "/api/v1/productos/"+creado.ID+"/ajustes-stock", seed.Token, map[string]any{
		"tipo": "salida", "cantidad": 5, "motivo": "demasiado",
	})
	requireStatus(t, res, http.StatusConflict)
	res.Body.Close()
}

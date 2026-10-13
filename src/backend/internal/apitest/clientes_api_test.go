package apitest

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	redisadapter "github.com/Bluethem/sw707u-grupo4/src/backend/adapters/redis"
)

type clienteAPI struct {
	ID            string `json:"id"`
	TipoDoc       string `json:"tipo_doc"`
	Numero        string `json:"numero"`
	RazonSocial   string `json:"razon_social"`
	Direccion     string `json:"direccion"`
	Correo        string `json:"correo"`
	Activo        bool   `json:"activo"`
}

func clienteBodyDN() map[string]any {
	return map[string]any{
		"tipo_doc": "DNI", "numero": "12345678",
		"razon_social": "Juan Pérez", "direccion": "Av. Test 100", "correo": "",
	}
}

func (h *harness) crearCliente(t *testing.T, token string, body map[string]any) (*http.Response, clienteAPI) {
	t.Helper()
	res := h.request(t, http.MethodPost, "/api/v1/clientes", token, body)
	var c clienteAPI
	if res.StatusCode == http.StatusCreated {
		decodeResponse(t, res, &c)
	}
	return res, c
}

func TestAPI_Cliente_CrearOK(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	res, got := h.crearCliente(t, seed.Token, clienteBodyDN())
	requireStatus(t, res, http.StatusCreated)
	if got.ID == "" || got.TipoDoc != "DNI" || got.Numero != "12345678" ||
		got.RazonSocial != "Juan Pérez" || got.Direccion != "Av. Test 100" || !got.Activo {
		t.Fatalf("cliente creado inesperado: %+v", got)
	}

	res, got = h.crearCliente(t, seed.Token, map[string]any{
		"tipo_doc": "RUC", "numero": "20100070970",
		"razon_social": "Cliente RUC SAC", "direccion": "Lima 1", "correo": "ruc@test.pe",
	})
	requireStatus(t, res, http.StatusCreated)
	if got.TipoDoc != "RUC" || got.Numero != "20100070970" || got.Correo != "ruc@test.pe" {
		t.Fatalf("cliente RUC inesperado: %+v", got)
	}
}

func TestAPI_Cliente_DocumentoInvalido(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	body := clienteBodyDN()
	body["numero"] = "123"
	res, _ := h.crearCliente(t, seed.Token, body)
	requireStatus(t, res, http.StatusBadRequest)
	res.Body.Close()

	body = clienteBodyDN()
	body["tipo_doc"] = "PASAPORTE"
	res, _ = h.crearCliente(t, seed.Token, body)
	requireStatus(t, res, http.StatusBadRequest)
	res.Body.Close()
}

func TestAPI_Cliente_Duplicado(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	res, _ := h.crearCliente(t, seed.Token, clienteBodyDN())
	requireStatus(t, res, http.StatusCreated)

	res, _ = h.crearCliente(t, seed.Token, clienteBodyDN())
	requireStatus(t, res, http.StatusConflict)
	res.Body.Close()
}

func TestAPI_Cliente_Obtener(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	crearRes, creado := h.crearCliente(t, seed.Token, clienteBodyDN())
	requireStatus(t, crearRes, http.StatusCreated)

	res := h.request(t, http.MethodGet, "/api/v1/clientes/"+creado.ID, seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var got clienteAPI
	decodeResponse(t, res, &got)
	if got.ID != creado.ID || got.Numero != creado.Numero {
		t.Fatalf("detalle inesperado: %+v", got)
	}

	const tenantBToken = "token-cliente-tenant-b"
	if err := h.sessions.Create(context.Background(), tenantBToken, redisadapter.SessionData{
		UserID: "usuario-b", EmpresaID: "empresa-b", Rol: "administrador",
	}); err != nil {
		t.Fatalf("crear sesion tenant B: %v", err)
	}
	res = h.request(t, http.MethodGet, fmt.Sprintf("/api/v1/clientes/%s", creado.ID), tenantBToken, nil)
	requireStatus(t, res, http.StatusNotFound)
	res.Body.Close()

	res = h.request(t, http.MethodGet, "/api/v1/clientes/id-inexistente", seed.Token, nil)
	requireStatus(t, res, http.StatusNotFound)
	res.Body.Close()
}

func TestAPI_Cliente_Listar(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	crearRes, creado := h.crearCliente(t, seed.Token, clienteBodyDN())
	requireStatus(t, crearRes, http.StatusCreated)
	rucRes, _ := h.crearCliente(t, seed.Token, map[string]any{
		"tipo_doc": "RUC", "numero": "20100070970",
		"razon_social": "Otro Cliente SAC", "direccion": "Lima", "correo": "",
	})
	requireStatus(t, rucRes, http.StatusCreated)

	res := h.request(t, http.MethodGet, "/api/v1/clientes?limit=10&offset=0", seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var listado struct {
		Items []clienteAPI `json:"items"`
		Total int          `json:"total"`
	}
	decodeResponse(t, res, &listado)
	if listado.Total != 2 || len(listado.Items) != 2 {
		t.Fatalf("listado total inesperado: %+v", listado)
	}

	res = h.request(t, http.MethodGet, "/api/v1/clientes?q=Juan&limit=10", seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	decodeResponse(t, res, &listado)
	if listado.Total != 1 || len(listado.Items) != 1 || listado.Items[0].ID != creado.ID {
		t.Fatalf("filtro q inesperado: %+v", listado)
	}

	res = h.request(t, http.MethodGet, "/api/v1/clientes?activo=true&limit=10", seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	decodeResponse(t, res, &listado)
	if listado.Total != 2 {
		t.Fatalf("filtro activo=true inesperado: %+v", listado)
	}
}

func TestAPI_Cliente_ActualizarYDesactivar(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	crearRes, creado := h.crearCliente(t, seed.Token, clienteBodyDN())
	requireStatus(t, crearRes, http.StatusCreated)

	nuevaRS := "Juan Actualizado"
	res := h.request(t, http.MethodPatch, "/api/v1/clientes/"+creado.ID, seed.Token, map[string]any{
		"razon_social": nuevaRS,
	})
	requireStatus(t, res, http.StatusOK)
	var actualizado clienteAPI
	decodeResponse(t, res, &actualizado)
	if actualizado.RazonSocial != nuevaRS || actualizado.Numero != creado.Numero {
		t.Fatalf("PATCH parcial inesperado: %+v", actualizado)
	}

	desactivo := false
	res = h.request(t, http.MethodPatch, "/api/v1/clientes/"+creado.ID, seed.Token, map[string]any{
		"activo": desactivo,
	})
	requireStatus(t, res, http.StatusOK)
	decodeResponse(t, res, &actualizado)
	if actualizado.Activo {
		t.Fatalf("cliente sigue activo: %+v", actualizado)
	}

	res = h.request(t, http.MethodGet, "/api/v1/clientes?activo=true&limit=10", seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var listado struct {
		Items []clienteAPI `json:"items"`
		Total int          `json:"total"`
	}
	decodeResponse(t, res, &listado)
	if listado.Total != 0 {
		t.Fatalf("listado activo=true incluye desactivado: %+v", listado)
	}

	res = h.request(t, http.MethodGet, "/api/v1/clientes/"+creado.ID, seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	decodeResponse(t, res, &actualizado)
	if actualizado.Activo {
		t.Fatalf("GET id muestra activo=true: %+v", actualizado)
	}
}

func TestAPI_Cliente_SinToken(t *testing.T) {
	h := newHarness(t)
	h.SeedEmpresaAdmin(t)

	res := h.request(t, http.MethodPost, "/api/v1/clientes", "", clienteBodyDN())
	requireStatus(t, res, http.StatusUnauthorized)
	res.Body.Close()

	res = h.request(t, http.MethodGet, "/api/v1/clientes", "", nil)
	requireStatus(t, res, http.StatusUnauthorized)
	res.Body.Close()
}

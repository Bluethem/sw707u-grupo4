package apitest

import (
	"net/http"
	"testing"
)

func TestE2E_GestionarCliente(t *testing.T) {
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

	body := map[string]any{
		"tipo_doc": "DNI", "numero": "87654321",
		"razon_social": "Cliente E2E", "direccion": "Av. E2E 200", "correo": "e2e@test.pe",
	}
	crearRes, creado := h.crearCliente(t, token, body)
	requireStatus(t, crearRes, http.StatusCreated)
	if creado.ID == "" || !creado.Activo {
		t.Fatalf("cliente creado inesperado: %+v", creado)
	}

	listRes := h.request(t, http.MethodGet, "/api/v1/clientes?limit=50", token, nil)
	requireStatus(t, listRes, http.StatusOK)
	var listado struct {
		Items []clienteAPI `json:"items"`
		Total int          `json:"total"`
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
		t.Fatalf("listado no contiene cliente creado: id=%q listado=%+v", creado.ID, listado)
	}

	patchRes := h.request(t, http.MethodPatch, "/api/v1/clientes/"+creado.ID, token, map[string]any{
		"activo": false,
	})
	requireStatus(t, patchRes, http.StatusOK)
	var desactivado clienteAPI
	decodeResponse(t, patchRes, &desactivado)
	if desactivado.Activo {
		t.Fatalf("PATCH no desactivo cliente: %+v", desactivado)
	}

	listActivoRes := h.request(t, http.MethodGet, "/api/v1/clientes?activo=true&limit=50", token, nil)
	requireStatus(t, listActivoRes, http.StatusOK)
	decodeResponse(t, listActivoRes, &listado)
	for _, item := range listado.Items {
		if item.ID == creado.ID {
			t.Fatalf("listado activo=true incluye cliente desactivado: %+v", listado)
		}
	}

	getRes := h.request(t, http.MethodGet, "/api/v1/clientes/"+creado.ID, token, nil)
	requireStatus(t, getRes, http.StatusOK)
	var detalle clienteAPI
	decodeResponse(t, getRes, &detalle)
	if detalle.Activo || detalle.ID != creado.ID {
		t.Fatalf("GET id inesperado tras desactivar: %+v", detalle)
	}
}

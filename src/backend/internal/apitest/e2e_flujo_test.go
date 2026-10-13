package apitest

import (
	"context"
	"net/http"
	"testing"

	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
)

func TestE2E_FlujoSeguridadM1(t *testing.T) {
	h := newHarness(t)
	tenantA := h.SeedEmpresaAdmin(t)

	login := func(email, password string) string {
		t.Helper()
		res := h.request(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
			"email": email, "password": password,
		})
		requireStatus(t, res, http.StatusOK)
		var body struct {
			Token string `json:"token"`
		}
		decodeResponse(t, res, &body)
		if body.Token == "" {
			t.Fatal("login no devolvio token")
		}
		return body.Token
	}

	adminAToken := login(tenantA.AdminEmail, tenantA.AdminPassword)
	const (
		cajeroEmail    = "cajero-e2e@empresa-a.test"
		cajeroPassword = "Cajero123!"
	)
	res := h.request(t, http.MethodPost, "/api/v1/usuarios", adminAToken, map[string]any{
		"email": cajeroEmail, "password": cajeroPassword,
		"nombre": "Cajero E2E", "rol": "cajero",
	})
	requireStatus(t, res, http.StatusCreated)
	var cajero struct {
		EmpresaID string `json:"empresa_id"`
		Rol       string `json:"rol"`
	}
	decodeResponse(t, res, &cajero)
	if cajero.EmpresaID != tenantA.EmpresaID || cajero.Rol != "cajero" {
		t.Fatalf("cajero creado inesperado: %+v", cajero)
	}

	cajeroToken := login(cajeroEmail, cajeroPassword)
	res = h.request(t, http.MethodGet, "/api/v1/usuarios", cajeroToken, nil)
	requireStatus(t, res, http.StatusForbidden)
	res.Body.Close()

	res = h.request(t, http.MethodPost, "/api/v1/series", adminAToken, map[string]any{
		"tipo": "factura", "serie": "F001", "siguiente": 1,
	})
	requireStatus(t, res, http.StatusCreated)
	res.Body.Close()

	empresaB, err := h.gestion.CrearEmpresa(context.Background(), gestionapp.CrearEmpresaInput{
		RUC: "20100070988", RazonSocial: "Empresa B E2E SAC",
		DomicilioFiscal: "Cusco", Regimen: "general",
	})
	if err != nil {
		t.Fatalf("seed empresa B: %v", err)
	}
	const (
		adminBEmail    = "admin@empresa-b.test"
		adminBPassword = "AdminB123!"
	)
	if _, err := h.gestion.CrearUsuario(context.Background(), empresaB.ID, gestionapp.CrearUsuarioInput{
		Email: adminBEmail, Password: adminBPassword,
		Nombre: "Admin B E2E", Rol: gestiondomain.RolAdministrador,
	}); err != nil {
		t.Fatalf("seed admin B: %v", err)
	}

	adminBToken := login(adminBEmail, adminBPassword)
	res = h.request(t, http.MethodGet, "/api/v1/usuarios", adminBToken, nil)
	requireStatus(t, res, http.StatusOK)
	var usuariosB []struct {
		EmpresaID string `json:"empresa_id"`
		Email     string `json:"email"`
	}
	decodeResponse(t, res, &usuariosB)
	if len(usuariosB) != 1 || usuariosB[0].EmpresaID != empresaB.ID || usuariosB[0].Email != adminBEmail {
		t.Fatalf("tenant B vio usuarios inesperados: %+v", usuariosB)
	}

	res = h.request(t, http.MethodGet, "/api/v1/series", adminBToken, nil)
	requireStatus(t, res, http.StatusOK)
	var seriesB []struct {
		EmpresaID string `json:"empresa_id"`
		Serie     string `json:"serie"`
	}
	decodeResponse(t, res, &seriesB)
	if len(seriesB) != 0 {
		t.Fatalf("tenant B vio series de tenant A: %+v", seriesB)
	}
}

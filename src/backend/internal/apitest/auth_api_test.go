package apitest

import (
	"context"
	"net/http"
	"testing"
)

func TestAPI_Login(t *testing.T) {
	t.Run("credenciales validas", func(t *testing.T) {
		h := newHarness(t)
		seed := h.SeedEmpresaAdmin(t)
		res := h.request(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
			"email": seed.AdminEmail, "password": seed.AdminPassword,
		})
		requireStatus(t, res, http.StatusOK)
		var body struct {
			Token string `json:"token"`
		}
		decodeResponse(t, res, &body)
		if body.Token == "" {
			t.Fatal("login no devolvio token")
		}
	})

	t.Run("credenciales invalidas", func(t *testing.T) {
		h := newHarness(t)
		seed := h.SeedEmpresaAdmin(t)
		res := h.request(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
			"email": seed.AdminEmail, "password": "password-incorrecto",
		})
		requireStatus(t, res, http.StatusUnauthorized)
		res.Body.Close()
	})

	t.Run("usuario inactivo", func(t *testing.T) {
		h := newHarness(t)
		seed := h.SeedEmpresaAdmin(t)
		if _, err := h.pool.Exec(context.Background(),
			`UPDATE usuarios SET activo = false WHERE id = $1`, seed.AdminID); err != nil {
			t.Fatalf("inactivar usuario: %v", err)
		}
		res := h.request(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
			"email": seed.AdminEmail, "password": seed.AdminPassword,
		})
		requireStatus(t, res, http.StatusUnauthorized)
		res.Body.Close()
	})
}

func TestAPI_Me(t *testing.T) {
	t.Run("token valido", func(t *testing.T) {
		h := newHarness(t)
		seed := h.SeedEmpresaAdmin(t)
		res := h.request(t, http.MethodGet, "/api/v1/auth/me", seed.Token, nil)
		requireStatus(t, res, http.StatusOK)
		var body struct {
			Rol     string `json:"rol"`
			Usuario struct {
				Email string `json:"email"`
			} `json:"usuario"`
			Empresa struct {
				ID string `json:"id"`
			} `json:"empresa"`
		}
		decodeResponse(t, res, &body)
		if body.Usuario.Email != seed.AdminEmail || body.Empresa.ID != seed.EmpresaID || body.Rol != "administrador" {
			t.Fatalf("respuesta me inesperada: %+v", body)
		}
	})

	t.Run("sin token", func(t *testing.T) {
		h := newHarness(t)
		h.SeedEmpresaAdmin(t)
		res := h.request(t, http.MethodGet, "/api/v1/auth/me", "", nil)
		requireStatus(t, res, http.StatusUnauthorized)
		res.Body.Close()
	})

	t.Run("token basura", func(t *testing.T) {
		h := newHarness(t)
		h.SeedEmpresaAdmin(t)
		res := h.request(t, http.MethodGet, "/api/v1/auth/me", "token-inexistente", nil)
		requireStatus(t, res, http.StatusUnauthorized)
		res.Body.Close()
	})
}

func TestAPI_Logout(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)
	res := h.request(t, http.MethodPost, "/api/v1/auth/logout", seed.Token, nil)
	requireStatus(t, res, http.StatusNoContent)
	res.Body.Close()

	res = h.request(t, http.MethodGet, "/api/v1/auth/me", seed.Token, nil)
	requireStatus(t, res, http.StatusUnauthorized)
	res.Body.Close()
}

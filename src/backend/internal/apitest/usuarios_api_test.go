package apitest

import (
	"context"
	"net/http"
	"testing"

	redisadapter "github.com/Bluethem/sw707u-grupo4/src/backend/adapters/redis"
	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
)

func TestAPI_UsuariosRBAC(t *testing.T) {
	h := newHarness(t)
	seed := h.SeedEmpresaAdmin(t)

	res := h.request(t, http.MethodPost, "/api/v1/usuarios", seed.Token, map[string]any{
		"email": "cajero@empresa.test", "password": "Cajero123!",
		"nombre": "Cajero API", "rol": "cajero",
	})
	requireStatus(t, res, http.StatusCreated)
	var cajero struct {
		ID        string `json:"id"`
		EmpresaID string `json:"empresa_id"`
		Email     string `json:"email"`
		Rol       string `json:"rol"`
	}
	decodeResponse(t, res, &cajero)
	if cajero.EmpresaID != seed.EmpresaID || cajero.Email != "cajero@empresa.test" || cajero.Rol != "cajero" {
		t.Fatalf("cajero creado inesperado: %+v", cajero)
	}

	res = h.request(t, http.MethodGet, "/api/v1/usuarios", seed.Token, nil)
	requireStatus(t, res, http.StatusOK)
	var usuarios []struct {
		ID string `json:"id"`
	}
	decodeResponse(t, res, &usuarios)
	if len(usuarios) != 2 {
		t.Fatalf("usuarios listados = %d, esperado 2", len(usuarios))
	}

	cajeroToken, err := redisadapter.GenerateToken()
	if err != nil {
		t.Fatalf("generar token cajero: %v", err)
	}
	if err := h.sessions.Create(context.Background(), cajeroToken, redisadapter.SessionData{
		UserID: cajero.ID, EmpresaID: seed.EmpresaID, Rol: "cajero",
	}); err != nil {
		t.Fatalf("crear sesion cajero: %v", err)
	}
	res = h.request(t, http.MethodGet, "/api/v1/usuarios", cajeroToken, nil)
	requireStatus(t, res, http.StatusForbidden)
	res.Body.Close()
}

func seedCajero(t *testing.T, h *harness, empresaID string) string {
	t.Helper()
	cajero, err := h.gestion.CrearUsuario(context.Background(), empresaID, gestionapp.CrearUsuarioInput{
		Email: "otro-cajero@empresa.test", Password: "Cajero123!",
		Nombre: "Otro Cajero", Rol: gestiondomain.RolCajero,
	})
	if err != nil {
		t.Fatalf("seed cajero: %v", err)
	}
	token, err := redisadapter.GenerateToken()
	if err != nil {
		t.Fatalf("generar token cajero: %v", err)
	}
	if err := h.sessions.Create(context.Background(), token, redisadapter.SessionData{
		UserID: cajero.ID, EmpresaID: empresaID, Rol: string(cajero.Rol),
	}); err != nil {
		t.Fatalf("crear sesion cajero: %v", err)
	}
	return token
}

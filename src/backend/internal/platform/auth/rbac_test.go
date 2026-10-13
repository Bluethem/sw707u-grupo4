package auth_test

import (
	"errors"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/auth"
)

func TestAllow_Cajero_NoGestionarUsuarios(t *testing.T) {
	if auth.Allow(auth.RolCajero, auth.ActionGestionarUsuarios) {
		t.Fatal("cajero no debe poder gestionar usuarios")
	}
}

func TestAllow_Admin_GestionarUsuarios(t *testing.T) {
	if !auth.Allow(auth.RolAdministrador, auth.ActionGestionarUsuarios) {
		t.Fatal("administrador debe poder gestionar usuarios")
	}
}

func TestRequireAdmin(t *testing.T) {
	if err := auth.RequireAdmin(auth.RolAdministrador); err != nil {
		t.Fatalf("admin: %v", err)
	}
	err := auth.RequireAdmin(auth.RolCajero)
	if err == nil {
		t.Fatal("cajero debe fallar RequireAdmin")
	}
	if !errors.Is(err, auth.ErrRequireAdmin) {
		t.Fatalf("error=%v want ErrRequireAdmin", err)
	}
}

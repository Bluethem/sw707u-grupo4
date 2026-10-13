package auth_test

import (
	"strings"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/auth"
)

func TestHashPassword_CheckPassword_Roundtrip(t *testing.T) {
	hash, err := auth.HashPassword("secreto-123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" || hash == "secreto-123" {
		t.Fatalf("hash inesperado: %q", hash)
	}
	if !strings.HasPrefix(hash, "$2a$10$") && !strings.HasPrefix(hash, "$2b$10$") {
		t.Fatalf("esperaba bcrypt cost 10, got prefix %q", hash[:7])
	}
	if !auth.CheckPassword(hash, "secreto-123") {
		t.Fatal("CheckPassword debe ser true para la contraseña correcta")
	}
	if auth.CheckPassword(hash, "otra") {
		t.Fatal("CheckPassword debe ser false para contraseña incorrecta")
	}
}

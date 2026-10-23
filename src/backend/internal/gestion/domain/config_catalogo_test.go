package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestPrecioEnRango_NoRedondeaPorDebajo(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfigCatalogo("emp-1")
	err := cfg.PrecioEnRango(decimal.RequireFromString("0.001"))
	if !errors.Is(err, ErrInvalido) {
		t.Fatalf("se esperaba ErrInvalido, se obtuvo %v", err)
	}
	if !strings.Contains(err.Error(), "0.00") || !strings.Contains(err.Error(), "0.01") {
		t.Fatalf("mensaje inesperado: %v", err)
	}
	if strings.Contains(err.Error(), "0.00 es menor") && !strings.Contains(err.Error(), "precio 0.00") {
		t.Fatalf("mensaje inesperado: %v", err)
	}
}

func TestValidarContraCatalogo_UnidadYAfectacion(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfigCatalogo("emp-1")
	cfg.UnidadesPermitidas = []string{"NIU"}
	cfg.AfectacionesPermitidas = []string{AfectacionGravado}

	p := productoValido()
	if err := cfg.ValidarContraCatalogo(p); err != nil {
		t.Fatalf("producto default debe pasar: %v", err)
	}

	p.Unidad = "KGM"
	if err := cfg.ValidarContraCatalogo(p); !errors.Is(err, ErrInvalido) {
		t.Fatalf("unidad fuera de lista debe fallar: %v", err)
	}

	p = productoValido()
	p.Afectacion = AfectacionInafecto
	if err := cfg.ValidarContraCatalogo(p); !errors.Is(err, ErrInvalido) {
		t.Fatalf("afectacion fuera de lista debe fallar: %v", err)
	}
}

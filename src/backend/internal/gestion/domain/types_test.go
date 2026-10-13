package domain

import (
	"errors"
	"testing"
)

func TestValidarEmpresa(t *testing.T) {
	t.Parallel()
	if err := ValidarEmpresa(Empresa{RUC: "20100070970", RazonSocial: "Empresa SAC"}); err != nil {
		t.Fatalf("empresa valida rechazada: %v", err)
	}
	if err := ValidarEmpresa(Empresa{RUC: "20100070971", RazonSocial: "Empresa SAC"}); !errors.Is(err, ErrInvalido) {
		t.Fatalf("se esperaba ErrInvalido para RUC, se obtuvo %v", err)
	}
	if err := ValidarEmpresa(Empresa{RUC: "20100070970"}); !errors.Is(err, ErrInvalido) {
		t.Fatalf("se esperaba ErrInvalido para razon social, se obtuvo %v", err)
	}
}

func TestValidarSerie(t *testing.T) {
	t.Parallel()
	casos := []struct {
		nombre string
		tipo   TipoSerie
		serie  string
		valida bool
	}{
		{"boleta", TipoBoleta, "B001", true},
		{"factura", TipoFactura, "FABC", true},
		{"prefijo incorrecto", TipoFactura, "B001", false},
		{"minusculas normalizables", TipoBoleta, "b001", true},
		{"longitud", TipoBoleta, "B01", false},
		{"tipo desconocido", "nota", "B001", false},
	}
	for _, caso := range casos {
		caso := caso
		t.Run(caso.nombre, func(t *testing.T) {
			t.Parallel()
			err := ValidarSerie(caso.tipo, caso.serie)
			if caso.valida && err != nil {
				t.Fatalf("serie valida rechazada: %v", err)
			}
			if !caso.valida && !errors.Is(err, ErrInvalido) {
				t.Fatalf("se esperaba ErrInvalido, se obtuvo %v", err)
			}
		})
	}
}

func TestNormalizarYValidarEmail(t *testing.T) {
	t.Parallel()
	if got := NormalizarEmail(" Admin@Example.COM "); got != "admin@example.com" {
		t.Fatalf("email normalizado = %q", got)
	}
	if err := ValidarEmail("sin-arroba"); !errors.Is(err, ErrInvalido) {
		t.Fatalf("se esperaba ErrInvalido, se obtuvo %v", err)
	}
}

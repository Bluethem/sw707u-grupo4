package domain

import (
	"errors"
	"testing"
)

func clienteValidoDNI() Cliente {
	return Cliente{
		TipoDoc:     TipoDocDNI,
		Numero:      "12345678",
		RazonSocial: "Juan Pérez",
		Direccion:   "Av. Ejemplo 123",
	}
}

func TestValidarCliente_DNIValido(t *testing.T) {
	t.Parallel()
	if err := ValidarCliente(clienteValidoDNI()); err != nil {
		t.Fatalf("cliente DNI valido rechazado: %v", err)
	}
}

func TestValidarCliente_RUCValido(t *testing.T) {
	t.Parallel()
	c := clienteValidoDNI()
	c.TipoDoc = TipoDocRUC
	c.Numero = "20100070970"
	if err := ValidarCliente(c); err != nil {
		t.Fatalf("cliente RUC valido rechazado: %v", err)
	}
}

func TestValidarCliente_TipoDocInvalido(t *testing.T) {
	t.Parallel()
	c := clienteValidoDNI()
	c.TipoDoc = "PASAPORTE"
	if err := ValidarCliente(c); !errors.Is(err, ErrInvalido) {
		t.Fatalf("se esperaba ErrInvalido, se obtuvo %v", err)
	}
}

func TestValidarCliente_DNIInvalido(t *testing.T) {
	t.Parallel()
	casos := []string{"1234567", "123456789", "1234567a"}
	for _, numero := range casos {
		c := clienteValidoDNI()
		c.Numero = numero
		if err := ValidarCliente(c); !errors.Is(err, ErrInvalido) {
			t.Fatalf("DNI %q: se esperaba ErrInvalido, se obtuvo %v", numero, err)
		}
	}
}

func TestValidarCliente_RUCInvalido(t *testing.T) {
	t.Parallel()
	c := clienteValidoDNI()
	c.TipoDoc = TipoDocRUC
	c.Numero = "20100070971"
	if err := ValidarCliente(c); !errors.Is(err, ErrInvalido) {
		t.Fatalf("se esperaba ErrInvalido, se obtuvo %v", err)
	}
}

func TestValidarCliente_RazonSocialVacia(t *testing.T) {
	t.Parallel()
	c := clienteValidoDNI()
	c.RazonSocial = "   "
	if err := ValidarCliente(c); !errors.Is(err, ErrInvalido) {
		t.Fatalf("se esperaba ErrInvalido, se obtuvo %v", err)
	}
}

func TestValidarCliente_DireccionVacia(t *testing.T) {
	t.Parallel()
	c := clienteValidoDNI()
	c.Direccion = ""
	if err := ValidarCliente(c); !errors.Is(err, ErrInvalido) {
		t.Fatalf("se esperaba ErrInvalido, se obtuvo %v", err)
	}
}

func TestValidarCliente_CorreoOpcional(t *testing.T) {
	t.Parallel()
	c := clienteValidoDNI()
	c.Correo = ""
	if err := ValidarCliente(c); err != nil {
		t.Fatalf("correo vacio debe ser valido: %v", err)
	}
}

func TestValidarCliente_CorreoInvalido(t *testing.T) {
	t.Parallel()
	c := clienteValidoDNI()
	c.Correo = "no-es-email"
	if err := ValidarCliente(c); !errors.Is(err, ErrInvalido) {
		t.Fatalf("se esperaba ErrInvalido, se obtuvo %v", err)
	}
}

func TestValidarCliente_CorreoValido(t *testing.T) {
	t.Parallel()
	c := clienteValidoDNI()
	c.Correo = "cliente@example.com"
	if err := ValidarCliente(c); err != nil {
		t.Fatalf("correo valido rechazado: %v", err)
	}
}

package domain

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func productoValido() Producto {
	return Producto{
		Codigo:        "SKU-001",
		Descripcion:   "Producto ejemplo",
		Unidad:        UnidadNIU,
		Precio:        decimal.RequireFromString("10.50"),
		Afectacion:    AfectacionGravado,
		ControlaStock: true,
	}
}

func TestValidarProducto_OK(t *testing.T) {
	t.Parallel()
	if err := ValidarProducto(productoValido(), nil); err != nil {
		t.Fatalf("producto valido rechazado: %v", err)
	}
}

func TestValidarProducto_CodigoVacio(t *testing.T) {
	t.Parallel()
	p := productoValido()
	p.Codigo = "  "
	if err := ValidarProducto(p, nil); !errors.Is(err, ErrInvalido) {
		t.Fatalf("se esperaba ErrInvalido, se obtuvo %v", err)
	}
}

func TestValidarProducto_DescripcionVacia(t *testing.T) {
	t.Parallel()
	p := productoValido()
	p.Descripcion = ""
	if err := ValidarProducto(p, nil); !errors.Is(err, ErrInvalido) {
		t.Fatalf("se esperaba ErrInvalido, se obtuvo %v", err)
	}
}

func TestValidarProducto_PrecioInvalido(t *testing.T) {
	t.Parallel()
	casos := []string{"0", "-1", "0.00"}
	for _, precio := range casos {
		p := productoValido()
		p.Precio = decimal.RequireFromString(precio)
		if err := ValidarProducto(p, nil); !errors.Is(err, ErrInvalido) {
			t.Fatalf("precio %q: se esperaba ErrInvalido, se obtuvo %v", precio, err)
		}
	}
}

func TestValidarProducto_AfectacionInvalida(t *testing.T) {
	t.Parallel()
	p := productoValido()
	p.Afectacion = "exportacion"
	if err := ValidarProducto(p, nil); !errors.Is(err, ErrInvalido) {
		t.Fatalf("se esperaba ErrInvalido, se obtuvo %v", err)
	}
}

func TestValidarProducto_AfectacionesValidas(t *testing.T) {
	t.Parallel()
	for _, af := range []string{AfectacionGravado, AfectacionExonerado, AfectacionInafecto} {
		p := productoValido()
		p.Afectacion = af
		if err := ValidarProducto(p, nil); err != nil {
			t.Fatalf("afectacion %q rechazada: %v", af, err)
		}
	}
}

func TestValidarProducto_StockInicialConControl(t *testing.T) {
	t.Parallel()
	p := productoValido()
	p.ControlaStock = true
	stock := int64(0)
	if err := ValidarProducto(p, &stock); err != nil {
		t.Fatalf("stock 0 debe ser valido: %v", err)
	}
	stock = 10
	if err := ValidarProducto(p, &stock); err != nil {
		t.Fatalf("stock 10 debe ser valido: %v", err)
	}
	stock = -1
	if err := ValidarProducto(p, &stock); !errors.Is(err, ErrInvalido) {
		t.Fatalf("se esperaba ErrInvalido, se obtuvo %v", err)
	}
}

func TestValidarProducto_StockInicialSinControl(t *testing.T) {
	t.Parallel()
	p := productoValido()
	p.ControlaStock = false
	stock := int64(0)
	if err := ValidarProducto(p, &stock); !errors.Is(err, ErrInvalido) {
		t.Fatalf("stock informado sin control debe fallar: %v", err)
	}
}

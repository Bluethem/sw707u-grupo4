package domain_test

import (
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/shopspring/decimal"
)

func TestCalcularLinea_GravadoIGVIncluido(t *testing.T) {
	// 1 × 118.00 → valor 100.00, igv 18.00
	l := domain.LineaCalculo{
		Cantidad:       1,
		PrecioUnitario: decimal.RequireFromString("118.00"),
		Afectacion:     domain.AfectacionGravado,
	}
	tot, err := domain.CalcularLinea(l)
	if err != nil {
		t.Fatal(err)
	}
	if !tot.ValorVenta.Equal(decimal.RequireFromString("100.00")) {
		t.Fatalf("valor=%s", tot.ValorVenta)
	}
	if !tot.IGV.Equal(decimal.RequireFromString("18.00")) {
		t.Fatalf("igv=%s", tot.IGV)
	}
	if !tot.Total.Equal(decimal.RequireFromString("118.00")) {
		t.Fatalf("total=%s", tot.Total)
	}
}

func TestCalcularLinea_Exonerado(t *testing.T) {
	l := domain.LineaCalculo{
		Cantidad:       2,
		PrecioUnitario: decimal.RequireFromString("10.50"),
		Afectacion:     domain.AfectacionExonerado,
	}
	tot, err := domain.CalcularLinea(l)
	if err != nil {
		t.Fatal(err)
	}
	if !tot.IGV.IsZero() || !tot.ValorVenta.Equal(decimal.RequireFromString("21.00")) {
		t.Fatalf("%+v", tot)
	}
}

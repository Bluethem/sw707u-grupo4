package domain_test

import (
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/shopspring/decimal"
)

func TestValidarIdentificacion_BoletaBajoUmbralSinCliente(t *testing.T) {
	err := domain.ValidarIdentificacion(domain.TipoBoleta, decimal.RequireFromString("100.00"), nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestValidarIdentificacion_BoletaSobreUmbralSinDoc(t *testing.T) {
	err := domain.ValidarIdentificacion(domain.TipoBoleta, decimal.RequireFromString("700.01"), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidarIdentificacion_FacturaRequiereRUC(t *testing.T) {
	err := domain.ValidarIdentificacion(domain.TipoFactura, decimal.RequireFromString("50.00"), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	c := &domain.DatosCliente{TipoDoc: "RUC", Numero: "20131312955", RazonSocial: "ACME"}
	if err := domain.ValidarIdentificacion(domain.TipoFactura, decimal.RequireFromString("50.00"), c); err != nil {
		t.Fatal(err)
	}
}

func TestValidarRUC_Formato(t *testing.T) {
	if err := domain.ValidarRUC("123"); err == nil {
		t.Fatal("expected error")
	}
}

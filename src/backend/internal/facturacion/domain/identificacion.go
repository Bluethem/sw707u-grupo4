package domain

import (
	"fmt"
	"unicode"

	"github.com/shopspring/decimal"
)

var UmbralBoletaSinDoc = decimal.RequireFromString("700.00")

type DatosCliente struct {
	TipoDoc     string
	Numero      string
	RazonSocial string
}

func ValidarIdentificacion(tipo TipoComprobante, total decimal.Decimal, c *DatosCliente) error {
	switch tipo {
	case TipoFactura:
		if c == nil || c.TipoDoc != "RUC" {
			return ErrClienteRequerido
		}
		return ValidarRUC(c.Numero)
	case TipoBoleta:
		if total.GreaterThan(UmbralBoletaSinDoc) {
			if c == nil || c.Numero == "" || (c.TipoDoc != "DNI" && c.TipoDoc != "RUC") {
				return ErrClienteRequerido
			}
			if c.TipoDoc == "RUC" {
				return ValidarRUC(c.Numero)
			}
			if c.TipoDoc == "DNI" && (len(c.Numero) != 8 || !soloDigitos(c.Numero)) {
				return fmt.Errorf("DNI invalido")
			}
		}
		return nil
	default:
		return fmt.Errorf("tipo comprobante invalido")
	}
}

func ValidarRUC(ruc string) error {
	if len(ruc) != 11 || !soloDigitos(ruc) {
		return ErrRUCInvalido
	}
	factors := []int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 10; i++ {
		sum += int(ruc[i]-'0') * factors[i]
	}
	mod := sum % 11
	dig := 11 - mod
	if dig == 10 {
		dig = 0
	} else if dig == 11 {
		dig = 1
	}
	if dig != int(ruc[10]-'0') {
		return ErrRUCInvalido
	}
	return nil
}

func soloDigitos(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

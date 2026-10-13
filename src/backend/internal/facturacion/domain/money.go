package domain

import (
	"fmt"

	"github.com/shopspring/decimal"
)

var (
	TasaIGV     = decimal.RequireFromString("0.18")
	FactorIGV   = decimal.RequireFromString("1.18")
	errCantidad = fmt.Errorf("cantidad debe ser > 0")
	errPrecio   = fmt.Errorf("precio debe ser > 0")
)

type LineaCalculo struct {
	Cantidad       int64
	PrecioUnitario decimal.Decimal
	Afectacion     Afectacion
}

type TotalesLinea struct {
	ValorVenta decimal.Decimal
	IGV        decimal.Decimal
	Total      decimal.Decimal
}

func CalcularLinea(l LineaCalculo) (TotalesLinea, error) {
	if l.Cantidad <= 0 {
		return TotalesLinea{}, errCantidad
	}
	if !l.PrecioUnitario.GreaterThan(decimal.Zero) {
		return TotalesLinea{}, errPrecio
	}
	total := l.PrecioUnitario.Mul(decimal.NewFromInt(l.Cantidad)).Round(2)
	switch l.Afectacion {
	case AfectacionGravado:
		valor := total.Div(FactorIGV).Round(2)
		igv := total.Sub(valor)
		return TotalesLinea{ValorVenta: valor, IGV: igv, Total: total}, nil
	case AfectacionExonerado, AfectacionInafecto:
		return TotalesLinea{ValorVenta: total, IGV: decimal.Zero, Total: total}, nil
	default:
		return TotalesLinea{}, fmt.Errorf("afectacion invalida: %s", l.Afectacion)
	}
}

func SumarTotales(lineas []TotalesLinea) TotalesLinea {
	var v, i, t decimal.Decimal
	for _, l := range lineas {
		v = v.Add(l.ValorVenta)
		i = i.Add(l.IGV)
		t = t.Add(l.Total)
	}
	return TotalesLinea{ValorVenta: v.Round(2), IGV: i.Round(2), Total: t.Round(2)}
}

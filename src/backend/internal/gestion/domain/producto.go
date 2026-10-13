package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	UnidadNIU = "NIU"

	AfectacionGravado   = "gravado"
	AfectacionExonerado = "exonerado"
	AfectacionInafecto  = "inafecto"

	TipoMovimientoInicial = "inicial"
	TipoMovimientoEntrada = "entrada"
	TipoMovimientoSalida  = "salida"
	TipoMovimientoVenta   = "venta"
)

type Producto struct {
	ID            string
	EmpresaID     string
	Codigo        string
	Descripcion   string
	Unidad        string
	Precio        decimal.Decimal
	Afectacion    string
	ControlaStock bool
	Activo        bool
	CreadoEn      time.Time
	ActualizadoEn time.Time
}

// ValidarProducto valida campos del catálogo. Si stockInicial != nil, aplica reglas de alta:
// con controla_stock el saldo inicial debe ser >= 0; sin controla_stock no debe informarse stock.
func ValidarProducto(p Producto, stockInicial *int64) error {
	if strings.TrimSpace(p.Codigo) == "" {
		return fmt.Errorf("%w: codigo requerido", ErrInvalido)
	}
	if strings.TrimSpace(p.Descripcion) == "" {
		return fmt.Errorf("%w: descripcion requerida", ErrInvalido)
	}
	if !p.Precio.GreaterThan(decimal.Zero) {
		return fmt.Errorf("%w: precio debe ser > 0", ErrInvalido)
	}
	switch p.Afectacion {
	case AfectacionGravado, AfectacionExonerado, AfectacionInafecto:
	default:
		return fmt.Errorf("%w: afectacion invalida", ErrInvalido)
	}
	if stockInicial == nil {
		return nil
	}
	if p.ControlaStock {
		if *stockInicial < 0 {
			return fmt.Errorf("%w: stock inicial invalido", ErrInvalido)
		}
		return nil
	}
	return fmt.Errorf("%w: stock inicial no aplica si no controla stock", ErrInvalido)
}

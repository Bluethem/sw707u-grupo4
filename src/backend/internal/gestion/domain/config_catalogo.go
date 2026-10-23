package domain

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// Unidades de medida SUNAT habilitadas por defecto.
var UnidadesSUNAT = []string{"NIU", "ZZ", "KGM", "LTR", "MTR", "BX"}

// ConfigCatalogo contiene las reglas de validación del catálogo para una empresa.
// Si la empresa no tiene fila en BD se usan los valores de DefaultConfigCatalogo().
type ConfigCatalogo struct {
	EmpresaID              string
	PrecioMinimo           decimal.Decimal
	PrecioMaximo           *decimal.Decimal // nil = sin límite superior
	CargaMaxFilas          int
	CargaMaxMB             int
	UnidadesPermitidas     []string
	AfectacionesPermitidas []string
	ActualizadoEn          time.Time
}

// DefaultConfigCatalogo devuelve la configuración base que aplica cuando
// una empresa no tiene fila explícita en config_catalogo.
func DefaultConfigCatalogo(empresaID string) ConfigCatalogo {
	return ConfigCatalogo{
		EmpresaID:              empresaID,
		PrecioMinimo:           decimal.NewFromFloat(0.01),
		PrecioMaximo:           nil,
		CargaMaxFilas:          500,
		CargaMaxMB:             2,
		UnidadesPermitidas:     append([]string{}, UnidadesSUNAT...),
		AfectacionesPermitidas: []string{AfectacionGravado, AfectacionExonerado, AfectacionInafecto},
	}
}

// ValidarConfigCatalogo verifica que los valores propuestos sean coherentes.
func ValidarConfigCatalogo(c ConfigCatalogo) error {
	if !c.PrecioMinimo.GreaterThan(decimal.Zero) {
		return fmt.Errorf("%w: precio_minimo debe ser > 0", ErrInvalido)
	}
	if c.PrecioMaximo != nil && c.PrecioMaximo.LessThan(c.PrecioMinimo) {
		return fmt.Errorf("%w: precio_maximo debe ser >= precio_minimo", ErrInvalido)
	}
	if c.CargaMaxFilas < 1 || c.CargaMaxFilas > 2000 {
		return fmt.Errorf("%w: carga_max_filas debe estar entre 1 y 2000", ErrInvalido)
	}
	if c.CargaMaxMB < 1 || c.CargaMaxMB > 20 {
		return fmt.Errorf("%w: carga_max_mb debe estar entre 1 y 20", ErrInvalido)
	}
	if len(c.UnidadesPermitidas) == 0 {
		return fmt.Errorf("%w: debe haber al menos una unidad permitida", ErrInvalido)
	}
	for _, u := range c.UnidadesPermitidas {
		if !unidadValida(u) {
			return fmt.Errorf("%w: unidad desconocida: %q", ErrInvalido, u)
		}
	}
	if len(c.AfectacionesPermitidas) == 0 {
		return fmt.Errorf("%w: debe haber al menos una afectacion permitida", ErrInvalido)
	}
	for _, a := range c.AfectacionesPermitidas {
		switch a {
		case AfectacionGravado, AfectacionExonerado, AfectacionInafecto:
		default:
			return fmt.Errorf("%w: afectacion desconocida: %q", ErrInvalido, a)
		}
	}
	return nil
}

// UnidadPermitida verifica si una unidad está en la lista configurada.
func (c ConfigCatalogo) UnidadPermitida(u string) bool {
	for _, v := range c.UnidadesPermitidas {
		if v == u {
			return true
		}
	}
	return false
}

// AfectacionPermitida verifica si una afectación está en la lista configurada.
func (c ConfigCatalogo) AfectacionPermitida(a string) bool {
	for _, v := range c.AfectacionesPermitidas {
		if v == a {
			return true
		}
	}
	return false
}

// PrecioEnRango verifica si un precio cumple los límites configurados.
func (c ConfigCatalogo) PrecioEnRango(precio decimal.Decimal) error {
	if precio.LessThan(c.PrecioMinimo) {
		return fmt.Errorf("%w: precio %s es menor al mínimo permitido %s",
			ErrInvalido, precio.StringFixed(2), c.PrecioMinimo.StringFixed(2))
	}
	if c.PrecioMaximo != nil && precio.GreaterThan(*c.PrecioMaximo) {
		return fmt.Errorf("%w: precio %s supera el máximo permitido %s",
			ErrInvalido, precio.StringFixed(2), c.PrecioMaximo.StringFixed(2))
	}
	return nil
}

// ValidarContraCatalogo aplica las reglas paramétricas de la empresa al producto.
func (c ConfigCatalogo) ValidarContraCatalogo(p Producto) error {
	if p.Unidad != "" && !c.UnidadPermitida(p.Unidad) {
		return fmt.Errorf("%w: unidad %q no permitida", ErrInvalido, p.Unidad)
	}
	if p.Afectacion != "" && !c.AfectacionPermitida(p.Afectacion) {
		return fmt.Errorf("%w: afectacion %q no permitida", ErrInvalido, p.Afectacion)
	}
	if !p.Precio.IsZero() {
		return c.PrecioEnRango(p.Precio)
	}
	return nil
}

func unidadValida(u string) bool {
	for _, v := range UnidadesSUNAT {
		if v == u {
			return true
		}
	}
	return false
}

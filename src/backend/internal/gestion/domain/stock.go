package domain

import (
	"fmt"
	"strings"
	"time"
)

type StockMovimiento struct {
	ID         int64
	EmpresaID  string
	ProductoID string
	Tipo       string
	Cantidad   int64
	Motivo     string
	UsuarioID  *string
	CreadoEn   time.Time
}

func ValidarAjusteStock(tipo string, cantidad int64, motivo string) error {
	switch tipo {
	case TipoMovimientoEntrada, TipoMovimientoSalida:
	default:
		return fmt.Errorf("%w: tipo de ajuste invalido", ErrInvalido)
	}
	if cantidad <= 0 {
		return fmt.Errorf("%w: cantidad debe ser > 0", ErrInvalido)
	}
	if strings.TrimSpace(motivo) == "" {
		return fmt.Errorf("%w: motivo requerido", ErrInvalido)
	}
	return nil
}

package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/shopspring/decimal"
)

type VentasTipo struct {
	Tipo       string
	Cantidad   int
	ValorVenta decimal.Decimal
	IGV        decimal.Decimal
	Total      decimal.Decimal
}

type VentasReporte struct {
	FechaDesde string
	FechaHasta string
	PorTipo    []VentasTipo
	Anulados   struct {
		Cantidad int
		Total    decimal.Decimal
	}
	Totales VentasTipo
}

type VentasReporteStore interface {
	Ventas(ctx context.Context, empresaID string, desde, hastaExclusivo time.Time) (VentasReporte, error)
}

type Reportes struct {
	Store VentasReporteStore
}

func (s *Reportes) Ventas(ctx context.Context, empresaID string, desde, hasta time.Time) (VentasReporte, error) {
	if s.Store == nil {
		return VentasReporte{}, fmt.Errorf("servicio reportes no configurado")
	}
	desde = time.Date(desde.Year(), desde.Month(), desde.Day(), 0, 0, 0, 0, time.UTC)
	hasta = time.Date(hasta.Year(), hasta.Month(), hasta.Day(), 0, 0, 0, 0, time.UTC)
	if hasta.Before(desde) {
		return VentasReporte{}, domain.ErrFechaReporteInvalida
	}
	dias := hasta.Sub(desde) / (24 * time.Hour)
	if dias > 90 {
		return VentasReporte{}, domain.ErrRangoReporteExcedido
	}
	hastaExcl := hasta.AddDate(0, 0, 1)
	rep, err := s.Store.Ventas(ctx, empresaID, desde, hastaExcl)
	if err != nil {
		return VentasReporte{}, err
	}
	rep.FechaDesde = desde.Format(time.DateOnly)
	rep.FechaHasta = hasta.Format(time.DateOnly)
	return rep, nil
}

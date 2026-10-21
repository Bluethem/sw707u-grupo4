package ubl_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/ubl"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/shopspring/decimal"
)

func TestUBL_BoletaContieneIDYTotales(t *testing.T) {
	c := &domain.Comprobante{
		Tipo: domain.TipoBoleta, Serie: "B001", Numero: 1,
		CreadoEn: time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC),
		ValorVenta: decimal.RequireFromString("100.00"),
		IGV:        decimal.RequireFromString("18.00"),
		Total:      decimal.RequireFromString("118.00"),
		Lineas: []domain.LineaComprobante{{
			Cantidad: 1, PrecioUnit: decimal.RequireFromString("118.00"),
			Afectacion: domain.AfectacionGravado,
			ValorVenta: decimal.RequireFromString("100.00"),
			IGV:        decimal.RequireFromString("18.00"),
			Total:      decimal.RequireFromString("118.00"),
		}},
	}
	xml, err := ubl.Generar(c, "20100070970", "Empresa API SAC")
	if err != nil || !bytes.Contains(xml, []byte("B001-1")) ||
		!bytes.Contains(xml, []byte("20100070970")) ||
		!bytes.Contains(xml, []byte("118.00")) ||
		!bytes.Contains(xml, []byte("InvoiceTypeCode")) {
		t.Fatalf("ubl inesperado: %s err=%v", xml, err)
	}
}

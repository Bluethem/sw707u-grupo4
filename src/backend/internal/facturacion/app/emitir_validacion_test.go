package app

import (
	"context"
	"errors"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/shopspring/decimal"
)

func TestEmitir_RechazaComandoInvalido(t *testing.T) {
	emisor := &Emisor{}

	t.Run("idempotency key vacia", func(t *testing.T) {
		_, err := emisor.Emitir(context.Background(), EmitirComando{
			EmpresaID: "e1",
			Lineas: []EmitirLinea{{
				ProductoID:     "p1",
				Cantidad:       1,
				PrecioUnitario: decimal.NewFromInt(118),
				Afectacion:     domain.AfectacionGravado,
			}},
		})
		if !errors.Is(err, domain.ErrIdempotencyKeyVacia) {
			t.Fatalf("error = %v; want ErrIdempotencyKeyVacia", err)
		}
	})

	t.Run("lineas vacias", func(t *testing.T) {
		_, err := emisor.Emitir(context.Background(), EmitirComando{
			EmpresaID:      "e1",
			IdempotencyKey: "k1",
		})
		if !errors.Is(err, domain.ErrLineasVacias) {
			t.Fatalf("error = %v; want ErrLineasVacias", err)
		}
	})
}

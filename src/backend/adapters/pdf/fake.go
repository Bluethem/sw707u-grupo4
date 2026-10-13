package pdf

import (
	"context"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
)

type Fake struct{}

func (Fake) Generar(_ context.Context, _ *domain.Comprobante) ([]byte, error) {
	return []byte("%PDF-fake"), nil
}

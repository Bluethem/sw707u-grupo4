package postgres

import (
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
)

func TestAdaptersImplementPorts(t *testing.T) {
	var _ ports.UnitOfWork = (*UnitOfWork)(nil)
	var _ ports.StockPort = (*Stock)(nil)
	var _ ports.SeriesRepository = (*Series)(nil)
	var _ ports.ComprobanteRepository = (*Comprobante)(nil)
	var _ ports.OutboxPublisher = (*Outbox)(nil)
}

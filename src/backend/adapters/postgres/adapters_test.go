package postgres

import (
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
)

func TestAdaptersImplementPorts(t *testing.T) {
	var _ ports.UnitOfWork = (*UnitOfWork)(nil)
	var _ ports.StockPort = (*Stock)(nil)
	var _ ports.SeriesRepository = (*Series)(nil)
	var _ ports.ComprobanteRepository = (*Comprobante)(nil)
	var _ ports.ComprobanteQuery = (*ComprobanteQuery)(nil)
	var _ ports.ComprobanteEnvioStore = (*ComprobanteQuery)(nil)
	var _ ports.OutboxPublisher = (*Outbox)(nil)
	var _ ports.OutboxConsumer = (*OutboxConsumer)(nil)
	var _ ports.ComprobanteLifecycle = (*Comprobante)(nil)
	var _ ports.EnvioOutboxReader = (*ComprobanteQuery)(nil)
	var _ gestionapp.EmpresaRepository = (*EmpresaRepository)(nil)
	var _ gestionapp.UsuarioRepository = (*UsuarioRepository)(nil)
	var _ gestionapp.SerieRepository = (*SeriesGestionRepository)(nil)
	var _ gestionapp.ClienteRepository = (*ClienteRepository)(nil)
	var _ gestionapp.ProductoRepository = (*ProductoRepository)(nil)
}

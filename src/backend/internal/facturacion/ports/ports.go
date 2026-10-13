package ports

import (
	"context"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
)

type Tx interface{} // opaque; adapters define concrete

type UnitOfWork interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
}

type StockLinea struct {
	ProductoID string
	Cantidad   int64
}

type StockPort interface {
	Descontar(ctx context.Context, tx Tx, empresaID string, lineas []StockLinea) error
}

type SeriesRepository interface {
	SiguienteNumero(ctx context.Context, tx Tx, empresaID string, tipo domain.TipoComprobante, serie string) (int64, error)
}

type ComprobanteRepository interface {
	GetByIdempotencyKey(ctx context.Context, tx Tx, empresaID, key string) (*domain.Comprobante, error)
	Insert(ctx context.Context, tx Tx, c *domain.Comprobante) error
}

type ListFilter struct {
	FechaDesde *time.Time
	FechaHasta *time.Time
	Tipo       string
	Estado     string
	Limit      int
	Offset     int
}

type ComprobanteQuery interface {
	// GetByID retorna (nil, nil) cuando el comprobante no existe para la empresa.
	GetByID(ctx context.Context, empresaID, id string) (*domain.Comprobante, error)
	// List retorna cabeceras sin lineas. FechaDesde y FechaHasta son dias inclusivos.
	List(ctx context.Context, empresaID string, f ListFilter) (items []domain.Comprobante, total int, err error)
}

type OutboxPublisher interface {
	EnqueueEnvio(ctx context.Context, tx Tx, comprobanteID string) error
}

type FirmaPort interface {
	FirmarXML(ctx context.Context, xml []byte) ([]byte, error)
}

type PdfPort interface {
	Generar(ctx context.Context, c *domain.Comprobante) ([]byte, error)
}

type StoragePort interface {
	Guardar(ctx context.Context, key string, data []byte) error
}

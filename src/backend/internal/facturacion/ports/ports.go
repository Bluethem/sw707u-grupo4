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

// ComprobanteEnvioStore lectura/actualizacion de estado para el worker de envio.
type ComprobanteEnvioStore interface {
	GetByComprobanteID(ctx context.Context, id string) (*domain.Comprobante, error)
	UpdateEstado(ctx context.Context, id string, estado domain.EstadoComprobante) error
}

type ComprobanteLifecycle interface {
	MarcarPendienteBaja(ctx context.Context, tx Tx, empresaID, id, motivo string) error
	PonerEstadoPendiente(ctx context.Context, tx Tx, id string) error
}

type EnvioOutboxReader interface {
	LatestEnvioJob(ctx context.Context, comprobanteID string) (estado, ultimoError string, ok bool, err error)
}

const (
	TipoTrabajoEnvioComprobante = "envio_comprobante"
	TipoTrabajoBajaFactura      = "baja_factura"
	TipoTrabajoResumenDiario    = "resumen_diario"
)

type OutboxPublisher interface {
	EnqueueEnvio(ctx context.Context, tx Tx, comprobanteID string) error
	EnqueueBaja(ctx context.Context, tx Tx, comprobanteID string) error
	EnqueueResumen(ctx context.Context, tx Tx, resumenID string) error
	RequeueEnvio(ctx context.Context, tx Tx, comprobanteID string) error
}

type BoletaInformable struct {
	ID     string
	Estado domain.EstadoComprobante
}

type ResumenWorkerItem struct {
	ComprobanteID string
	Anulado       bool
}

type ResumenStore interface {
	List(ctx context.Context, empresaID string) ([]domain.ResumenDiario, error)
	GetByID(ctx context.Context, empresaID, id string) (*domain.ResumenDiario, error)
	BoletasInformables(ctx context.Context, tx Tx, empresaID string, fecha time.Time) ([]BoletaInformable, error)
	NextCorrelativo(ctx context.Context, tx Tx, empresaID string, fecha time.Time) (int, error)
	Insert(ctx context.Context, tx Tx, r *domain.ResumenDiario) error
	GetWorkerData(ctx context.Context, resumenID string) (estado domain.EstadoResumen, items []ResumenWorkerItem, err error)
	UpdateEstado(ctx context.Context, resumenID string, estado domain.EstadoResumen, ticket string) error
}

type OutboxJob struct {
	ID             int64
	ComprobanteID  string
	ReferenciaID   string
	TipoTrabajo    string
	Estado         string
	Intentos       int
	DisponibleEn   time.Time
	CreadoEn       time.Time
	UltimoError    string
	ActualizadoEn  time.Time
}

type OutboxConsumer interface {
	Claim(ctx context.Context, limit int) ([]OutboxJob, error)
	Complete(ctx context.Context, jobID int64) error
	CompleteWithError(ctx context.Context, jobID int64, errMsg string) error
	Retry(ctx context.Context, jobID int64, intentos int, disponibleEn time.Time, errMsg string) error
	Fail(ctx context.Context, jobID int64, errMsg string) error
}

type SunatResult struct {
	Aceptado bool
	Codigo   string
	Mensaje  string
	CDR      []byte
}

type SunatGateway interface {
	Enviar(ctx context.Context, comprobanteID string, xml []byte) (SunatResult, error)
	EnviarBaja(ctx context.Context, comprobanteID, motivo string) (SunatResult, error)
	EnviarResumen(ctx context.Context, resumenID string, payload []byte) (SunatResult, error)
}

type FirmaPort interface {
	FirmarXML(ctx context.Context, xml []byte) ([]byte, error)
}

type PdfPort interface {
	Generar(ctx context.Context, c *domain.Comprobante) ([]byte, error)
}

type StoragePort interface {
	Guardar(ctx context.Context, key string, data []byte) error
	Leer(ctx context.Context, key string) ([]byte, error)
}

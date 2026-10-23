package domain

import "time"

type EstadoResumen string

const (
	ResumenPendiente EstadoResumen = "PENDIENTE"
	ResumenEnviado   EstadoResumen = "ENVIADO"
	ResumenAceptado  EstadoResumen = "ACEPTADO"
	ResumenRechazado EstadoResumen = "RECHAZADO"
)

type ResumenDiario struct {
	ID          string
	EmpresaID   string
	Fecha       time.Time
	Correlativo int
	Estado      EstadoResumen
	Ticket      string
	CreadoEn    time.Time
	Comprobantes []ResumenComprobanteItem
}

type ResumenComprobanteItem struct {
	ComprobanteID string
	Anulado       bool
	Tipo          TipoComprobante
	Serie         string
	Numero        int64
	Estado        EstadoComprobante
}

package domain

import "github.com/shopspring/decimal"

type EstadoComprobante string

const (
	EstadoPendiente     EstadoComprobante = "PENDIENTE"
	EstadoEnviado       EstadoComprobante = "ENVIADO"
	EstadoAceptado      EstadoComprobante = "ACEPTADO"
	EstadoObservado     EstadoComprobante = "OBSERVADO"
	EstadoRechazado     EstadoComprobante = "RECHAZADO"
	EstadoAnulada       EstadoComprobante = "ANULADA"
	EstadoPendienteBaja EstadoComprobante = "PENDIENTE_BAJA"
)

type Comprobante struct {
	ID             string
	EmpresaID      string
	Tipo           TipoComprobante
	Serie          string
	Numero         int64
	Estado         EstadoComprobante
	Cliente        *DatosCliente
	Lineas         []LineaComprobante
	ValorVenta     decimal.Decimal
	IGV            decimal.Decimal
	Total          decimal.Decimal
	IdempotencyKey string
	BajaMotivo     string
}

type LineaComprobante struct {
	ProductoID string
	Cantidad   int64
	PrecioUnit decimal.Decimal
	Afectacion Afectacion
	ValorVenta decimal.Decimal
	IGV        decimal.Decimal
	Total      decimal.Decimal
}

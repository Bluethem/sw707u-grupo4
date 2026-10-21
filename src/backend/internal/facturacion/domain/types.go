package domain

import "fmt"

type Afectacion string

const (
	AfectacionGravado   Afectacion = "gravado"
	AfectacionExonerado Afectacion = "exonerado"
	AfectacionInafecto  Afectacion = "inafecto"
)

type TipoComprobante string

const (
	TipoBoleta  TipoComprobante = "boleta"
	TipoFactura TipoComprobante = "factura"
)

var (
	ErrClienteRequerido    = fmt.Errorf("cliente/documento requerido")
	ErrRUCInvalido         = fmt.Errorf("RUC invalido")
	ErrStockInsuficiente   = fmt.Errorf("stock insuficiente")
	ErrSerieNoEncontrada   = fmt.Errorf("serie no encontrada")
	ErrLineasVacias        = fmt.Errorf("lineas vacias")
	ErrIdempotencyKeyVacia = fmt.Errorf("idempotency key vacia")
	ErrProductoNoEncontrado = fmt.Errorf("producto no encontrado")
	ErrProductoInactivo        = fmt.Errorf("producto inactivo")
	ErrComprobanteNoEncontrado = fmt.Errorf("comprobante no encontrado")
	ErrOperacionNoPermitida    = fmt.Errorf("operacion no permitida")
	ErrMotivoBajaInvalido      = fmt.Errorf("motivo de baja requerido")
	ErrResumenNoEncontrado     = fmt.Errorf("resumen no encontrado")
	ErrResumenSinBoletas       = fmt.Errorf("no hay boletas informables para la fecha")
	ErrFechaResumenInvalida    = fmt.Errorf("fecha invalida")
	ErrFechaReporteInvalida    = fmt.Errorf("fecha invalida")
	ErrRangoReporteExcedido    = fmt.Errorf("rango de fechas excede 90 dias")
	ErrCDRNoDisponible         = fmt.Errorf("cdr no disponible")
	ErrCDRNoEncontrado         = fmt.Errorf("cdr no encontrado")
)

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
	ErrLineasVacias        = fmt.Errorf("lineas vacias")
	ErrIdempotencyKeyVacia = fmt.Errorf("idempotency key vacia")
)

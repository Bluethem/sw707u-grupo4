package app

import "context"

// ProductoConsulta resuelve productos del catálogo del tenant para emisión.
type ProductoConsulta interface {
	ObtenerParaEmision(ctx context.Context, empresaID, productoID string) (controlaStock bool, activo bool, err error)
}

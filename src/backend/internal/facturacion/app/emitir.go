package app

import (
	"context"
	"fmt"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type EmitirComando struct {
	EmpresaID      string
	Tipo           domain.TipoComprobante
	Serie          string
	Cliente        *domain.DatosCliente
	Lineas         []EmitirLinea
	IdempotencyKey string
}

type EmitirLinea struct {
	ProductoID     string
	Cantidad       int64
	PrecioUnitario decimal.Decimal
	Afectacion     domain.Afectacion
}

type Emisor struct {
	UoW       ports.UnitOfWork
	Stock     ports.StockPort
	Productos ProductoConsulta
	Series    ports.SeriesRepository
	Comps     ports.ComprobanteRepository
	Outbox    ports.OutboxPublisher
	Firma     ports.FirmaPort
	PDF       ports.PdfPort
	Storage   ports.StoragePort
	NewID     func() string
}

func (e *Emisor) Emitir(ctx context.Context, cmd EmitirComando) (c *domain.Comprobante, created bool, err error) {
	if cmd.IdempotencyKey == "" {
		return nil, false, domain.ErrIdempotencyKeyVacia
	}
	if len(cmd.Lineas) == 0 {
		return nil, false, domain.ErrLineasVacias
	}

	var comprobante *domain.Comprobante

	err = e.UoW.WithinTransaction(ctx, func(ctx context.Context, tx ports.Tx) error {
		existente, err := e.Comps.GetByIdempotencyKey(ctx, tx, cmd.EmpresaID, cmd.IdempotencyKey)
		if err != nil {
			return err
		}
		if existente != nil {
			comprobante = existente
			return nil
		}

		lineas, totales, err := calcularLineas(cmd.Lineas)
		if err != nil {
			return err
		}
		if err := domain.ValidarIdentificacion(cmd.Tipo, totales.Total, cmd.Cliente); err != nil {
			return err
		}

		var stock []ports.StockLinea
		for _, linea := range cmd.Lineas {
			if e.Productos == nil {
				return fmt.Errorf("producto consulta no configurado")
			}
			controla, activo, err := e.Productos.ObtenerParaEmision(ctx, cmd.EmpresaID, linea.ProductoID)
			if err != nil {
				return err
			}
			if !activo {
				return domain.ErrProductoInactivo
			}
			if controla {
				stock = append(stock, ports.StockLinea{ProductoID: linea.ProductoID, Cantidad: linea.Cantidad})
			}
		}
		if len(stock) > 0 {
			if err := e.Stock.Descontar(ctx, tx, cmd.EmpresaID, stock); err != nil {
				return err
			}
		}
		numero, err := e.Series.SiguienteNumero(ctx, tx, cmd.EmpresaID, cmd.Tipo, cmd.Serie)
		if err != nil {
			return err
		}

		newID := e.NewID
		if newID == nil {
			newID = func() string { return uuid.NewString() }
		}
		comprobante = &domain.Comprobante{
			ID:             newID(),
			EmpresaID:      cmd.EmpresaID,
			Tipo:           cmd.Tipo,
			Serie:          cmd.Serie,
			Numero:         numero,
			Estado:         domain.EstadoPendiente,
			Cliente:        cmd.Cliente,
			Lineas:         lineas,
			ValorVenta:     totales.ValorVenta,
			IGV:            totales.IGV,
			Total:          totales.Total,
			IdempotencyKey: cmd.IdempotencyKey,
		}
		if err := e.Comps.Insert(ctx, tx, comprobante); err != nil {
			return err
		}
		if err := e.Outbox.EnqueueEnvio(ctx, tx, comprobante.ID); err != nil {
			return err
		}
		created = true
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	if created {
		e.generarRepresentaciones(ctx, comprobante)
	}
	return comprobante, created, nil
}

func calcularLineas(lineas []EmitirLinea) ([]domain.LineaComprobante, domain.TotalesLinea, error) {
	resultado := make([]domain.LineaComprobante, 0, len(lineas))
	calculos := make([]domain.TotalesLinea, 0, len(lineas))
	for _, linea := range lineas {
		total, err := domain.CalcularLinea(domain.LineaCalculo{
			Cantidad:       linea.Cantidad,
			PrecioUnitario: linea.PrecioUnitario,
			Afectacion:     linea.Afectacion,
		})
		if err != nil {
			return nil, domain.TotalesLinea{}, err
		}
		resultado = append(resultado, domain.LineaComprobante{
			ProductoID: linea.ProductoID,
			Cantidad:   linea.Cantidad,
			PrecioUnit: linea.PrecioUnitario,
			Afectacion: linea.Afectacion,
			ValorVenta: total.ValorVenta,
			IGV:        total.IGV,
			Total:      total.Total,
		})
		calculos = append(calculos, total)
	}
	return resultado, domain.SumarTotales(calculos), nil
}

func (e *Emisor) generarRepresentaciones(ctx context.Context, c *domain.Comprobante) {
	if e.Firma != nil {
		xml := []byte(fmt.Sprintf("<comprobante id=%q/>", c.ID))
		if firmado, err := e.Firma.FirmarXML(ctx, xml); err == nil && e.Storage != nil {
			_ = e.Storage.Guardar(ctx, c.ID+".xml", firmado)
		}
	}
	if e.PDF != nil {
		if data, err := e.PDF.Generar(ctx, c); err == nil && e.Storage != nil {
			_ = e.Storage.Guardar(ctx, c.ID+".pdf", data)
		}
	}
}

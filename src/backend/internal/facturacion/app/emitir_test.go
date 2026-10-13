package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/firma"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/pdf"
	postgresadapter "github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/storage"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/app"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	"github.com/shopspring/decimal"
)

type pdfConError struct{}

func (pdfConError) Generar(context.Context, *domain.Comprobante) ([]byte, error) {
	return nil, errors.New("pdf no disponible")
}

func nuevoEmisor(pdfPort ports.PdfPort) *app.Emisor {
	return &app.Emisor{
		UoW:       postgresadapter.NewUnitOfWork(testPool),
		Stock:     postgresadapter.NewStock(),
		Productos: postgresadapter.NewProductoEmisionConsulta(testPool),
		Series:    postgresadapter.NewSeries(),
		Comps:   postgresadapter.NewComprobante(),
		Outbox:  postgresadapter.NewOutbox(),
		Firma:   firma.Fake{},
		PDF:     pdfPort,
		Storage: storage.NewFake(),
		NewID:   func() string { return "comprobante-1" },
	}
}

func comando(key string, cantidad int64) app.EmitirComando {
	return app.EmitirComando{
		EmpresaID:      "e1",
		Tipo:           domain.TipoBoleta,
		Serie:          "B001",
		IdempotencyKey: key,
		Lineas: []app.EmitirLinea{{
			ProductoID:     "p1",
			Cantidad:       cantidad,
			PrecioUnitario: decimal.RequireFromString("118.00"),
			Afectacion:     domain.AfectacionGravado,
		}},
	}
}

func TestEmitir_Feliz(t *testing.T) {
	seed(t, 10)

	got, created, err := nuevoEmisor(pdf.Fake{}).Emitir(context.Background(), comando("feliz", 2))
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}
	if !created {
		t.Fatal("created = false; want true")
	}
	if got.Estado != domain.EstadoPendiente {
		t.Fatalf("estado = %q; want %q", got.Estado, domain.EstadoPendiente)
	}
	if got.Numero != 1 {
		t.Fatalf("numero = %d; want 1", got.Numero)
	}
	if stock := queryInt64(t, `SELECT cantidad FROM productos_stock WHERE empresa_id='e1' AND producto_id='p1'`); stock != 8 {
		t.Fatalf("stock = %d; want 8", stock)
	}
	if outbox := queryInt64(t, `SELECT count(*) FROM outbox_envios`); outbox != 1 {
		t.Fatalf("outbox = %d; want 1", outbox)
	}
}

func TestEmitir_StockInsuficiente(t *testing.T) {
	seed(t, 1)

	_, _, err := nuevoEmisor(pdf.Fake{}).Emitir(context.Background(), comando("sin-stock", 5))
	if !errors.Is(err, domain.ErrStockInsuficiente) {
		t.Fatalf("error = %v; want ErrStockInsuficiente", err)
	}
	if stock := queryInt64(t, `SELECT cantidad FROM productos_stock WHERE empresa_id='e1' AND producto_id='p1'`); stock != 1 {
		t.Fatalf("stock = %d; want 1", stock)
	}
	if comprobantes := queryInt64(t, `SELECT count(*) FROM comprobantes`); comprobantes != 0 {
		t.Fatalf("comprobantes = %d; want 0", comprobantes)
	}
}

func TestEmitir_Idempotencia(t *testing.T) {
	seed(t, 10)
	emisor := nuevoEmisor(pdf.Fake{})
	cmd := comando("misma-key", 2)

	primero, created, err := emisor.Emitir(context.Background(), cmd)
	if err != nil {
		t.Fatalf("primera emision: %v", err)
	}
	if !created {
		t.Fatal("primera emision: created = false; want true")
	}
	segundo, created, err := emisor.Emitir(context.Background(), cmd)
	if err != nil {
		t.Fatalf("segunda emision: %v", err)
	}
	if created {
		t.Fatal("segunda emision: created = true; want false")
	}
	if primero.ID != segundo.ID {
		t.Fatalf("IDs distintos: %q != %q", primero.ID, segundo.ID)
	}
	if stock := queryInt64(t, `SELECT cantidad FROM productos_stock WHERE empresa_id='e1' AND producto_id='p1'`); stock != 8 {
		t.Fatalf("stock = %d; want 8", stock)
	}
	if comprobantes := queryInt64(t, `SELECT count(*) FROM comprobantes`); comprobantes != 1 {
		t.Fatalf("comprobantes = %d; want 1", comprobantes)
	}
}

func TestEmitir_PdfFallaNoRevierte(t *testing.T) {
	seed(t, 10)

	got, _, err := nuevoEmisor(pdfConError{}).Emitir(context.Background(), comando("pdf-falla", 2))
	if err != nil {
		t.Fatalf("Emitir debe ignorar error de PDF post-commit: %v", err)
	}
	if persisted := queryInt64(t, `SELECT count(*) FROM comprobantes WHERE id=$1 AND estado='PENDIENTE'`, got.ID); persisted != 1 {
		t.Fatalf("comprobante pendiente persistido = %d; want 1", persisted)
	}
}

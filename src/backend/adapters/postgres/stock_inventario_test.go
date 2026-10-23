package postgres

import (
	"context"
	"errors"
	"testing"

	facturaciondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestStockInventario_AjusteEntradaSalida(t *testing.T) {
	ctx := context.Background()
	pool := newClienteTestPool(t, ctx)
	repo := NewProductoRepository(pool)
	empresaID := insertEmpresaClienteTest(t, ctx, pool)
	svc := gestionapp.NewService(nil, nil, nil, nil, repo, repo, nil)

	stock10 := int64(10)
	p, err := repo.Crear(ctx, gestiondomain.Producto{
		ID: uuid.NewString(), EmpresaID: empresaID, Codigo: "AJ-1", Descripcion: "Ajustable",
		Unidad: gestiondomain.UnidadNIU, Precio: decimal.RequireFromString("9.99"),
		Afectacion: gestiondomain.AfectacionGravado, ControlaStock: true, Activo: true,
	}, &stock10)
	if err != nil {
		t.Fatalf("crear producto: %v", err)
	}
	uid := "admin-1"

	res, err := svc.AjustarStock(ctx, empresaID, p.ID, uid, gestionapp.AjustarStockInput{
		Tipo: gestiondomain.TipoMovimientoEntrada, Cantidad: 5, Motivo: "Compra",
	})
	if err != nil || res.Cantidad != 15 || res.MovimientoID == 0 {
		t.Fatalf("entrada: res=%#v err=%v", res, err)
	}

	res, err = svc.AjustarStock(ctx, empresaID, p.ID, uid, gestionapp.AjustarStockInput{
		Tipo: gestiondomain.TipoMovimientoSalida, Cantidad: 3, Motivo: "Merma",
	})
	if err != nil || res.Cantidad != 12 {
		t.Fatalf("salida: res=%#v err=%v", res, err)
	}

	movs, total, err := svc.ListarMovimientos(ctx, empresaID, p.ID, 20, 0)
	if err != nil || total < 3 || len(movs) < 3 {
		t.Fatalf("movimientos: total=%d len=%d err=%v", total, len(movs), err)
	}
	if movs[0].Tipo != gestiondomain.TipoMovimientoSalida || movs[0].UsuarioID == nil || *movs[0].UsuarioID != uid {
		t.Fatalf("ultimo movimiento: %#v", movs[0])
	}
}

func TestStockInventario_AjusteSalidaInsuficiente(t *testing.T) {
	ctx := context.Background()
	pool := newClienteTestPool(t, ctx)
	repo := NewProductoRepository(pool)
	empresaID := insertEmpresaClienteTest(t, ctx, pool)
	svc := gestionapp.NewService(nil, nil, nil, nil, repo, repo, nil)

	stock2 := int64(2)
	p, err := repo.Crear(ctx, gestiondomain.Producto{
		ID: uuid.NewString(), EmpresaID: empresaID, Codigo: "AJ-2", Descripcion: "Poco stock",
		Unidad: gestiondomain.UnidadNIU, Precio: decimal.RequireFromString("1.00"),
		Afectacion: gestiondomain.AfectacionGravado, ControlaStock: true, Activo: true,
	}, &stock2)
	if err != nil {
		t.Fatalf("crear producto: %v", err)
	}

	_, err = svc.AjustarStock(ctx, empresaID, p.ID, "u1", gestionapp.AjustarStockInput{
		Tipo: gestiondomain.TipoMovimientoSalida, Cantidad: 5, Motivo: "Exceso",
	})
	if !errors.Is(err, facturaciondomain.ErrStockInsuficiente) {
		t.Fatalf("error = %v; want ErrStockInsuficiente", err)
	}

	d, err := repo.Obtener(ctx, empresaID, p.ID)
	if err != nil || d.Stock == nil || *d.Stock != 2 {
		t.Fatalf("saldo debio quedar en 2: %#v err=%v", d.Stock, err)
	}
}

func TestStockInventario_AjusteSinControlStock(t *testing.T) {
	ctx := context.Background()
	pool := newClienteTestPool(t, ctx)
	repo := NewProductoRepository(pool)
	empresaID := insertEmpresaClienteTest(t, ctx, pool)
	svc := gestionapp.NewService(nil, nil, nil, nil, repo, repo, nil)

	p, err := repo.Crear(ctx, gestiondomain.Producto{
		ID: uuid.NewString(), EmpresaID: empresaID, Codigo: "SVC-AJ", Descripcion: "Servicio",
		Unidad: gestiondomain.UnidadNIU, Precio: decimal.RequireFromString("50.00"),
		Afectacion: gestiondomain.AfectacionInafecto, ControlaStock: false, Activo: true,
	}, nil)
	if err != nil {
		t.Fatalf("crear producto: %v", err)
	}

	_, err = svc.AjustarStock(ctx, empresaID, p.ID, "u1", gestionapp.AjustarStockInput{
		Tipo: gestiondomain.TipoMovimientoEntrada, Cantidad: 1, Motivo: "No aplica",
	})
	if !errors.Is(err, gestiondomain.ErrInvalido) {
		t.Fatalf("error = %v; want ErrInvalido", err)
	}
}

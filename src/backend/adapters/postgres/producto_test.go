package postgres

import (
	"context"
	"errors"
	"testing"

	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestProductoRepository_CrearObtenerListar(t *testing.T) {
	ctx := context.Background()
	pool := newClienteTestPool(t, ctx)
	repo := NewProductoRepository(pool)
	empresaID := insertEmpresaClienteTest(t, ctx, pool)

	stock10 := int64(10)
	p1, err := repo.Crear(ctx, gestiondomain.Producto{
		ID: uuid.NewString(), EmpresaID: empresaID, Codigo: "SKU-A", Descripcion: "Alpha",
		Unidad: gestiondomain.UnidadNIU, Precio: decimal.RequireFromString("12.50"),
		Afectacion: gestiondomain.AfectacionGravado, ControlaStock: true, Activo: true,
	}, &stock10)
	if err != nil {
		t.Fatalf("crear p1: %v", err)
	}

	d1, err := repo.Obtener(ctx, empresaID, p1.ID)
	if err != nil || d1.Producto.Codigo != "SKU-A" {
		t.Fatalf("obtener p1: %#v err=%v", d1, err)
	}
	if d1.Stock == nil || *d1.Stock != 10 {
		t.Fatalf("saldo esperado 10: %#v", d1.Stock)
	}

	var movCount int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM stock_movimientos
		WHERE empresa_id = $1 AND producto_id = $2 AND tipo = $3`,
		empresaID, p1.ID, gestiondomain.TipoMovimientoInicial,
	).Scan(&movCount); err != nil || movCount != 1 {
		t.Fatalf("movimiento inicial: count=%d err=%v", movCount, err)
	}

	stock0 := int64(0)
	p2, err := repo.Crear(ctx, gestiondomain.Producto{
		ID: uuid.NewString(), EmpresaID: empresaID, Codigo: "SKU-B", Descripcion: "Bravo",
		Unidad: gestiondomain.UnidadNIU, Precio: decimal.RequireFromString("5.00"),
		Afectacion: gestiondomain.AfectacionExonerado, ControlaStock: true, Activo: true,
	}, &stock0)
	if err != nil {
		t.Fatalf("crear p2 stock 0: %v", err)
	}
	d2, err := repo.Obtener(ctx, empresaID, p2.ID)
	if err != nil || d2.Stock == nil || *d2.Stock != 0 {
		t.Fatalf("obtener p2 saldo 0: %#v err=%v", d2, err)
	}

	p3, err := repo.Crear(ctx, gestiondomain.Producto{
		ID: uuid.NewString(), EmpresaID: empresaID, Codigo: "SVC-1", Descripcion: "Servicio",
		Unidad: gestiondomain.UnidadNIU, Precio: decimal.RequireFromString("20.00"),
		Afectacion: gestiondomain.AfectacionInafecto, ControlaStock: false, Activo: true,
	}, nil)
	if err != nil {
		t.Fatalf("crear p3 sin stock: %v", err)
	}
	d3, err := repo.Obtener(ctx, empresaID, p3.ID)
	if err != nil || d3.Stock != nil {
		t.Fatalf("servicio sin saldo: %#v err=%v", d3, err)
	}

	if _, err := repo.Obtener(ctx, uuid.NewString(), p1.ID); !errors.Is(err, gestiondomain.ErrNoEncontrado) {
		t.Fatalf("cross-tenant debio ser ErrNoEncontrado: %v", err)
	}

	items, total, err := repo.Listar(ctx, empresaID, gestionapp.ListarProductosFilter{Limit: 10})
	if err != nil || total != 3 || len(items) != 3 {
		t.Fatalf("listar todos: total=%d len=%d err=%v", total, len(items), err)
	}

	items, total, err = repo.Listar(ctx, empresaID, gestionapp.ListarProductosFilter{Q: "Bravo", Limit: 50})
	if err != nil || total != 1 || len(items) != 1 || items[0].ID != p2.ID {
		t.Fatalf("listar Q: total=%d items=%#v err=%v", total, items, err)
	}

	if _, err := repo.Crear(ctx, gestiondomain.Producto{
		ID: uuid.NewString(), EmpresaID: empresaID, Codigo: "SKU-A", Descripcion: "Duplicado",
		Unidad: gestiondomain.UnidadNIU, Precio: decimal.RequireFromString("1.00"),
		Afectacion: gestiondomain.AfectacionGravado, ControlaStock: true, Activo: true,
	}, &stock0); !errors.Is(err, gestiondomain.ErrConflicto) {
		t.Fatalf("codigo duplicado debio ser ErrConflicto: %v", err)
	}

	updated, err := repo.Actualizar(ctx, gestiondomain.Producto{
		ID: p1.ID, EmpresaID: empresaID, Codigo: p1.Codigo, Descripcion: "Alpha Plus",
		Unidad: p1.Unidad, Precio: decimal.RequireFromString("15.00"),
		Afectacion: gestiondomain.AfectacionGravado, ControlaStock: true, Activo: false,
	})
	if err != nil || !updated.ActualizadoEn.After(p1.ActualizadoEn) || updated.Activo {
		t.Fatalf("actualizar: %#v err=%v", updated, err)
	}
	if updated.Codigo != "SKU-A" || !updated.ControlaStock {
		t.Fatalf("codigo/controla_stock no deben cambiar: %#v", updated)
	}
}

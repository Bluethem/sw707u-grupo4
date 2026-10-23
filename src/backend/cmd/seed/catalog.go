package main

import (
	"context"
	"errors"
	"fmt"

	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	"github.com/shopspring/decimal"
)

type demoProducto struct {
	Codigo      string
	Descripcion string
	Precio      string
	Afectacion  string
}

type demoCliente struct {
	TipoDoc     string
	Numero      string
	RazonSocial string
	Direccion   string
	Correo      string
}

var demoProductos = []demoProducto{
	{Codigo: "SKU-1", Descripcion: "Producto demo", Precio: "10.00", Afectacion: "gravado"},
	{Codigo: "SKU-CAFE", Descripcion: "Café americano", Precio: "8.50", Afectacion: "gravado"},
	{Codigo: "SKU-SAND", Descripcion: "Sándwich de pollo", Precio: "14.90", Afectacion: "gravado"},
	{Codigo: "SKU-AGUA", Descripcion: "Agua mineral 600ml", Precio: "3.50", Afectacion: "exonerado"},
	{Codigo: "SKU-CUAD", Descripcion: "Cuaderno A4", Precio: "6.00", Afectacion: "inafecto"},
	{Codigo: "SKU-MENU", Descripcion: "Menú ejecutivo", Precio: "25.00", Afectacion: "gravado"},
	{Codigo: "SKU-TORTA", Descripcion: "Porción de torta", Precio: "12.00", Afectacion: "gravado"},
}

var demoClientes = []demoCliente{
	{TipoDoc: "DNI", Numero: "12345678", RazonSocial: "Ana Pérez", Direccion: "Jr. Cusco 210, Lima", Correo: "ana.perez@demo.local"},
	{TipoDoc: "DNI", Numero: "87654321", RazonSocial: "Luis Torres", Direccion: "Av. Arequipa 1450, Lima", Correo: "luis.torres@demo.local"},
	{TipoDoc: "DNI", Numero: "45678912", RazonSocial: "María Gómez", Direccion: "Calle Las Flores 88, Surco", Correo: "maria.gomez@demo.local"},
	{TipoDoc: "RUC", Numero: rucDesdeBase("2012345678"), RazonSocial: "Comercial Andina SAC", Direccion: "Av. Argentina 3200, Callao", Correo: "ventas@andina.demo"},
	{TipoDoc: "RUC", Numero: rucDesdeBase("2051234567"), RazonSocial: "Inversiones Lima Norte EIRL", Direccion: "Av. Túpac Amaru 900, Independencia", Correo: "contacto@limanorte.demo"},
}

func ensureDemoCatalog(ctx context.Context, service *gestionapp.Service, empresaID string) error {
	stock := int64(5000)
	for _, item := range demoProductos {
		if exists, err := productoExiste(ctx, service, empresaID, item.Codigo); err != nil {
			return err
		} else if exists {
			continue
		}
		if _, err := service.CrearProducto(ctx, empresaID, gestionapp.CrearProductoInput{
			Codigo:        item.Codigo,
			Descripcion:   item.Descripcion,
			Unidad:        "NIU",
			Precio:        decimal.RequireFromString(item.Precio),
			Afectacion:    item.Afectacion,
			ControlaStock: true,
			StockInicial:  &stock,
		}); err != nil && !errors.Is(err, gestiondomain.ErrConflicto) {
			return fmt.Errorf("crear producto %s: %w", item.Codigo, err)
		}
	}
	return nil
}

func ensureDemoClientes(ctx context.Context, service *gestionapp.Service, empresaID string) error {
	for _, item := range demoClientes {
		if exists, err := clienteExiste(ctx, service, empresaID, item.Numero); err != nil {
			return err
		} else if exists {
			continue
		}
		if _, err := service.CrearCliente(ctx, empresaID, gestionapp.CrearClienteInput{
			TipoDoc:     item.TipoDoc,
			Numero:      item.Numero,
			RazonSocial: item.RazonSocial,
			Direccion:   item.Direccion,
			Correo:      item.Correo,
		}); err != nil && !errors.Is(err, gestiondomain.ErrConflicto) {
			return fmt.Errorf("crear cliente %s: %w", item.Numero, err)
		}
	}
	return nil
}

func productoExiste(ctx context.Context, service *gestionapp.Service, empresaID, codigo string) (bool, error) {
	items, _, err := service.ListarProductos(ctx, empresaID, gestionapp.ListarProductosFilter{
		Q: codigo, Limit: 20,
	})
	if err != nil {
		return false, fmt.Errorf("listar productos: %w", err)
	}
	for _, p := range items {
		if p.Producto.Codigo == codigo {
			return true, nil
		}
	}
	return false, nil
}

func clienteExiste(ctx context.Context, service *gestionapp.Service, empresaID, numero string) (bool, error) {
	items, _, err := service.ListarClientes(ctx, empresaID, gestionapp.ListarClientesFilter{
		Q: numero, Limit: 20,
	})
	if err != nil {
		return false, fmt.Errorf("listar clientes: %w", err)
	}
	for _, c := range items {
		if c.Numero == numero {
			return true, nil
		}
	}
	return false, nil
}

func rucDesdeBase(base10 string) string {
	if len(base10) != 10 {
		panic("base RUC debe tener 10 dígitos")
	}
	factors := []int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 10; i++ {
		sum += int(base10[i]-'0') * factors[i]
	}
	dig := 11 - (sum % 11)
	if dig == 10 {
		dig = 0
	} else if dig == 11 {
		dig = 1
	}
	return base10 + string(rune('0'+dig))
}

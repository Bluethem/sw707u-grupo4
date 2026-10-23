package main

import (
	"context"
	"fmt"
	"time"

	factdomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

const (
	seedVentaPrefix   = "seed-venta-"
	seedResumenPrefix = "seed-resumen-"
	diasHistorico     = 29
)

type ventasSeedResult struct {
	Aplicado    bool
	Boletas     int
	Facturas    int
	HoyVigentes int
	HoyAnulados int
	ResumenDias []int
}

type productoSeed struct {
	ID         string
	Codigo     string
	Precio     decimal.Decimal
	Afectacion string
}

type clienteSeed struct {
	TipoDoc     string
	Numero      string
	RazonSocial string
}

type lineaSeed struct {
	productoCodigo string
	cantidad       int64
}

type comprobanteSeed struct {
	tipo       string
	serie      string
	numero     int64
	estado     string
	clienteNro string
	diasAtras  int
	hora       int
	lineas     []lineaSeed
	bajaMotivo string
}

func ensureDemoVentas(ctx context.Context, pool *pgxpool.Pool, empresaID string) (ventasSeedResult, error) {
	var existentes int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM comprobantes
		WHERE empresa_id = $1 AND idempotency_key LIKE $2`,
		empresaID, seedVentaPrefix+"%",
	).Scan(&existentes); err != nil {
		return ventasSeedResult{}, fmt.Errorf("contar ventas seed: %w", err)
	}
	if existentes > 0 {
		return ventasSeedResult{}, nil
	}

	productos, err := cargarProductos(ctx, pool, empresaID)
	if err != nil {
		return ventasSeedResult{}, err
	}
	clientes, err := cargarClientes(ctx, pool, empresaID)
	if err != nil {
		return ventasSeedResult{}, err
	}

	nextBoleta, err := siguienteNumero(ctx, pool, empresaID, "B001")
	if err != nil {
		return ventasSeedResult{}, err
	}
	nextFactura, err := siguienteNumero(ctx, pool, empresaID, "F001")
	if err != nil {
		return ventasSeedResult{}, err
	}

	docs, resumenDias := construirDemoComprobantes(nextBoleta, nextFactura)
	tx, err := pool.Begin(ctx)
	if err != nil {
		return ventasSeedResult{}, fmt.Errorf("iniciar tx ventas: %w", err)
	}
	defer tx.Rollback(ctx)

	result := ventasSeedResult{Aplicado: true, ResumenDias: resumenDias}
	porDia := map[int][]comprobanteInsertado{}

	for _, doc := range docs {
		id, err := insertarComprobante(ctx, tx, empresaID, doc, productos, clientes)
		if err != nil {
			return ventasSeedResult{}, err
		}
		if doc.tipo == "boleta" {
			result.Boletas++
		} else {
			result.Facturas++
		}
		if doc.diasAtras == 0 {
			if doc.estado == "ANULADA" || doc.estado == "PENDIENTE_BAJA" {
				result.HoyAnulados++
			} else {
				result.HoyVigentes++
			}
		}
		porDia[doc.diasAtras] = append(porDia[doc.diasAtras], comprobanteInsertado{
			id: id, tipo: doc.tipo, estado: doc.estado, diasAtras: doc.diasAtras,
		})
	}

	if err := actualizarSerie(ctx, tx, empresaID, "boleta", "B001", nextBoleta+int64(result.Boletas)); err != nil {
		return ventasSeedResult{}, err
	}
	if err := actualizarSerie(ctx, tx, empresaID, "factura", "F001", nextFactura+int64(result.Facturas)); err != nil {
		return ventasSeedResult{}, err
	}

	for _, dia := range resumenDias {
		if err := insertarResumenAceptado(ctx, tx, empresaID, dia, porDia[dia]); err != nil {
			return ventasSeedResult{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return ventasSeedResult{}, fmt.Errorf("confirmar ventas seed: %w", err)
	}
	return result, nil
}

type comprobanteInsertado struct {
	id        string
	tipo      string
	estado    string
	diasAtras int
}

func construirDemoComprobantes(nextBoleta, nextFactura int64) ([]comprobanteSeed, []int) {
	personas := []string{"12345678", "87654321", "45678912"}
	empresas := make([]string, 0, 2)
	for _, c := range demoClientes {
		if c.TipoDoc == "RUC" {
			empresas = append(empresas, c.Numero)
		}
	}
	codigos := make([]string, 0, len(demoProductos))
	for _, p := range demoProductos {
		codigos = append(codigos, p.Codigo)
	}

	var docs []comprobanteSeed
	boletaN, facturaN := nextBoleta, nextFactura
	resumenDias := []int{7, 14, 21}

	for daysAgo := diasHistorico; daysAgo >= 0; daysAgo-- {
		nBoletas := 2
		if daysAgo == 0 {
			nBoletas = 5
		} else if daysAgo%5 == 0 {
			nBoletas = 4
		} else if daysAgo%3 == 0 {
			nBoletas = 3
		}

		for i := 0; i < nBoletas; i++ {
			estado := "ACEPTADO"
			motivo := ""
			cliente := personas[(daysAgo+i)%len(personas)]
			if daysAgo == 0 && i == 4 {
				estado = "PENDIENTE"
			}
			if daysAgo == 0 && i == 3 {
				estado = "PENDIENTE_BAJA"
				motivo = "Cliente desistió de la compra"
			}
			if daysAgo == 1 && i == 0 {
				estado = "ENVIADO"
			}
			if daysAgo == 12 && i == 0 {
				estado = "ANULADA"
				motivo = "Error en el monto"
			}
			if daysAgo > 2 && i == 0 && daysAgo%9 == 0 {
				cliente = ""
			}
			docs = append(docs, comprobanteSeed{
				tipo: "boleta", serie: "B001", numero: boletaN, estado: estado,
				clienteNro: cliente, diasAtras: daysAgo, hora: 10 + i,
				lineas:     lineasPara(codigos, daysAgo, i),
				bajaMotivo: motivo,
			})
			boletaN++
		}

		if daysAgo == 0 || daysAgo%2 == 0 {
			nFacturas := 1
			if daysAgo == 0 {
				nFacturas = 3
			} else if daysAgo%6 == 0 {
				nFacturas = 2
			}
			for i := 0; i < nFacturas; i++ {
				estado := "ACEPTADO"
				motivo := ""
				if daysAgo == 0 && i == 2 {
					estado = "ANULADA"
					motivo = "Factura emitida por error"
				}
				if daysAgo == 10 && i == 0 {
					estado = "ANULADA"
					motivo = "Cliente solicitó anulación"
				}
				docs = append(docs, comprobanteSeed{
					tipo: "factura", serie: "F001", numero: facturaN, estado: estado,
					clienteNro: empresas[(daysAgo+i)%len(empresas)],
					diasAtras:  daysAgo, hora: 15 + i,
					lineas:     lineasPara(codigos, daysAgo+3, i+1),
					bajaMotivo: motivo,
				})
				facturaN++
			}
		}
	}
	return docs, resumenDias
}

func lineasPara(codigos []string, seed, idx int) []lineaSeed {
	primero := codigos[(seed+idx)%len(codigos)]
	lineas := []lineaSeed{{productoCodigo: primero, cantidad: int64(1 + (seed+idx)%3)}}
	if (seed+idx)%2 == 0 {
		segundo := codigos[(seed+idx+2)%len(codigos)]
		if segundo != primero {
			lineas = append(lineas, lineaSeed{productoCodigo: segundo, cantidad: 1})
		}
	}
	return lineas
}

func insertarComprobante(
	ctx context.Context,
	tx pgx.Tx,
	empresaID string,
	doc comprobanteSeed,
	productos map[string]productoSeed,
	clientes map[string]clienteSeed,
) (string, error) {
	var valor, igv, total decimal.Decimal
	type lineaCalc struct {
		producto productoSeed
		cantidad int64
		totales  factdomain.TotalesLinea
	}
	calculadas := make([]lineaCalc, 0, len(doc.lineas))
	for _, l := range doc.lineas {
		p, ok := productos[l.productoCodigo]
		if !ok {
			return "", fmt.Errorf("producto seed %s no encontrado", l.productoCodigo)
		}
		tot, err := factdomain.CalcularLinea(factdomain.LineaCalculo{
			Cantidad:       l.cantidad,
			PrecioUnitario: p.Precio,
			Afectacion:     factdomain.Afectacion(p.Afectacion),
		})
		if err != nil {
			return "", fmt.Errorf("calcular linea %s: %w", p.Codigo, err)
		}
		calculadas = append(calculadas, lineaCalc{producto: p, cantidad: l.cantidad, totales: tot})
		valor = valor.Add(tot.ValorVenta)
		igv = igv.Add(tot.IGV)
		total = total.Add(tot.Total)
	}
	valor = valor.Round(2)
	igv = igv.Round(2)
	total = total.Round(2)

	var clienteTipo, clienteNumero, clienteRazon any
	if doc.clienteNro != "" {
		c, ok := clientes[doc.clienteNro]
		if !ok {
			return "", fmt.Errorf("cliente seed %s no encontrado", doc.clienteNro)
		}
		clienteTipo = c.TipoDoc
		clienteNumero = c.Numero
		clienteRazon = c.RazonSocial
	}

	id := uuid.NewString()
	creado := time.Now().UTC().AddDate(0, 0, -doc.diasAtras)
	creado = time.Date(creado.Year(), creado.Month(), creado.Day(), doc.hora, 15, 0, 0, time.UTC)
	key := fmt.Sprintf("%s%s-%04d", seedVentaPrefix, doc.serie, doc.numero)

	if _, err := tx.Exec(ctx, `
		INSERT INTO comprobantes (
			id, empresa_id, tipo, serie, numero, estado,
			cliente_tipo, cliente_numero, cliente_razon,
			valor_venta, igv, total, idempotency_key, creado_en, baja_motivo
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		id, empresaID, doc.tipo, doc.serie, doc.numero, doc.estado,
		clienteTipo, clienteNumero, clienteRazon,
		valor, igv, total, key, creado, doc.bajaMotivo,
	); err != nil {
		return "", fmt.Errorf("insertar %s %s-%d: %w", doc.tipo, doc.serie, doc.numero, err)
	}

	for _, l := range calculadas {
		if _, err := tx.Exec(ctx, `
			INSERT INTO comprobante_lineas (
				comprobante_id, producto_id, cantidad, precio_unit,
				afectacion, valor_venta, igv, total
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			id, l.producto.ID, l.cantidad, l.producto.Precio,
			l.producto.Afectacion, l.totales.ValorVenta, l.totales.IGV, l.totales.Total,
		); err != nil {
			return "", fmt.Errorf("insertar linea %s: %w", l.producto.Codigo, err)
		}
	}
	return id, nil
}

func insertarResumenAceptado(
	ctx context.Context,
	tx pgx.Tx,
	empresaID string,
	diasAtras int,
	docs []comprobanteInsertado,
) error {
	fecha := time.Now().UTC().AddDate(0, 0, -diasAtras)
	fecha = time.Date(fecha.Year(), fecha.Month(), fecha.Day(), 0, 0, 0, 0, time.UTC)
	id := fmt.Sprintf("%s%s", seedResumenPrefix, fecha.Format(time.DateOnly))

	if _, err := tx.Exec(ctx, `
		INSERT INTO resumenes_diarios (id, empresa_id, fecha, correlativo, estado, ticket, creado_en)
		VALUES ($1, $2, $3, 1, 'ACEPTADO', $4, $5)`,
		id, empresaID, fecha, "TICKET-"+fecha.Format("20060102"), fecha.Add(20*time.Hour),
	); err != nil {
		return fmt.Errorf("insertar resumen %s: %w", fecha.Format(time.DateOnly), err)
	}

	informables := 0
	for _, d := range docs {
		if d.tipo != "boleta" {
			continue
		}
		switch d.estado {
		case "ACEPTADO", "PENDIENTE", "ENVIADO", "PENDIENTE_BAJA":
		default:
			continue
		}
		anulado := d.estado == "PENDIENTE_BAJA"
		if _, err := tx.Exec(ctx, `
			INSERT INTO resumen_comprobantes (resumen_id, comprobante_id, anulado)
			VALUES ($1, $2, $3)`,
			id, d.id, anulado,
		); err != nil {
			return fmt.Errorf("vincular boleta a resumen %s: %w", id, err)
		}
		informables++
	}
	if informables == 0 {
		return fmt.Errorf("resumen %s sin boletas informables", fecha.Format(time.DateOnly))
	}
	return nil
}

func actualizarSerie(ctx context.Context, tx pgx.Tx, empresaID, tipo, serie string, siguiente int64) error {
	_, err := tx.Exec(ctx, `
		UPDATE series_correlativos
		SET siguiente = $4
		WHERE empresa_id = $1 AND tipo = $2 AND serie = $3`,
		empresaID, tipo, serie, siguiente,
	)
	if err != nil {
		return fmt.Errorf("actualizar serie %s: %w", serie, err)
	}
	return nil
}

func siguienteNumero(ctx context.Context, pool *pgxpool.Pool, empresaID, serie string) (int64, error) {
	var n int64
	if err := pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(numero), 0) + 1
		FROM comprobantes
		WHERE empresa_id = $1 AND serie = $2`,
		empresaID, serie,
	).Scan(&n); err != nil {
		return 0, fmt.Errorf("siguiente numero %s: %w", serie, err)
	}
	return n, nil
}

func cargarProductos(ctx context.Context, pool *pgxpool.Pool, empresaID string) (map[string]productoSeed, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, codigo, precio, afectacion
		FROM productos
		WHERE empresa_id = $1`, empresaID)
	if err != nil {
		return nil, fmt.Errorf("cargar productos: %w", err)
	}
	defer rows.Close()

	out := make(map[string]productoSeed)
	for rows.Next() {
		var p productoSeed
		if err := rows.Scan(&p.ID, &p.Codigo, &p.Precio, &p.Afectacion); err != nil {
			return nil, fmt.Errorf("leer producto: %w", err)
		}
		out[p.Codigo] = p
	}
	return out, rows.Err()
}

func cargarClientes(ctx context.Context, pool *pgxpool.Pool, empresaID string) (map[string]clienteSeed, error) {
	rows, err := pool.Query(ctx, `
		SELECT tipo_doc, numero, razon_social
		FROM clientes
		WHERE empresa_id = $1`, empresaID)
	if err != nil {
		return nil, fmt.Errorf("cargar clientes: %w", err)
	}
	defer rows.Close()

	out := make(map[string]clienteSeed)
	for rows.Next() {
		var c clienteSeed
		if err := rows.Scan(&c.TipoDoc, &c.Numero, &c.RazonSocial); err != nil {
			return nil, fmt.Errorf("leer cliente: %w", err)
		}
		out[c.Numero] = c
	}
	return out, rows.Err()
}

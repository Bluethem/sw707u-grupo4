package postgres

import (
	"context"
	"database/sql"
	"fmt"

	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductoRepository struct {
	pool *pgxpool.Pool
}

func NewProductoRepository(pool *pgxpool.Pool) *ProductoRepository {
	return &ProductoRepository{pool: pool}
}

func (r *ProductoRepository) Crear(
	ctx context.Context,
	p gestiondomain.Producto,
	stockInicial *int64,
) (gestiondomain.Producto, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return gestiondomain.Producto{}, fmt.Errorf("crear producto: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO productos (
			id, empresa_id, codigo, descripcion, unidad, precio, afectacion, controla_stock, activo
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING creado_en, actualizado_en`,
		p.ID, p.EmpresaID, p.Codigo, p.Descripcion, p.Unidad, p.Precio,
		p.Afectacion, p.ControlaStock, p.Activo,
	).Scan(&p.CreadoEn, &p.ActualizadoEn)
	if err != nil {
		return gestiondomain.Producto{}, mapGestionError("crear producto", err)
	}

	if p.ControlaStock {
		saldo := *stockInicial
		if _, err := tx.Exec(ctx, `
			INSERT INTO productos_stock (empresa_id, producto_id, cantidad)
			VALUES ($1, $2, $3)
			ON CONFLICT (empresa_id, producto_id) DO UPDATE SET cantidad = EXCLUDED.cantidad`,
			p.EmpresaID, p.ID, saldo,
		); err != nil {
			return gestiondomain.Producto{}, mapGestionError("crear stock producto", err)
		}
		if saldo > 0 {
			if _, err := tx.Exec(ctx, `
				INSERT INTO stock_movimientos (empresa_id, producto_id, tipo, cantidad, motivo)
				VALUES ($1, $2, $3, $4, '')`,
				p.EmpresaID, p.ID, gestiondomain.TipoMovimientoInicial, saldo,
			); err != nil {
				return gestiondomain.Producto{}, mapGestionError("crear movimiento inicial", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return gestiondomain.Producto{}, fmt.Errorf("crear producto: %w", err)
	}
	return p, nil
}

func (r *ProductoRepository) Obtener(ctx context.Context, empresaID, id string) (gestionapp.ProductoDetalle, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT p.id, p.empresa_id, p.codigo, p.descripcion, p.unidad, p.precio, p.afectacion,
		       p.controla_stock, p.activo, p.creado_en, p.actualizado_en, ps.cantidad
		FROM productos p
		LEFT JOIN productos_stock ps
		  ON ps.empresa_id = p.empresa_id AND ps.producto_id = p.id
		WHERE p.empresa_id = $1 AND p.id = $2`,
		empresaID, id,
	)
	detalle, err := scanProductoDetalle(row)
	if err != nil {
		return gestionapp.ProductoDetalle{}, mapGestionError("obtener producto", err)
	}
	return detalle, nil
}

func (r *ProductoRepository) Listar(
	ctx context.Context,
	empresaID string,
	f gestionapp.ListarProductosFilter,
) ([]gestiondomain.Producto, int, error) {
	limit := f.Limit
	offset := f.Offset
	args := []any{empresaID, f.Q, f.Activo}
	const filtros = `
		FROM productos
		WHERE empresa_id = $1
		  AND ($2 = '' OR codigo ILIKE '%' || $2 || '%' OR descripcion ILIKE '%' || $2 || '%')
		  AND ($3::boolean IS NULL OR activo = $3)`

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*)`+filtros, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar productos: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, empresa_id, codigo, descripcion, unidad, precio, afectacion,
		       controla_stock, activo, creado_en, actualizado_en`+
		filtros+`
		ORDER BY descripcion, codigo, id
		LIMIT $4 OFFSET $5`,
		append(args, limit, offset)...,
	)
	if err != nil {
		return nil, 0, mapGestionError("listar productos", err)
	}
	defer rows.Close()

	items := make([]gestiondomain.Producto, 0)
	for rows.Next() {
		p, err := scanProducto(rows)
		if err != nil {
			return nil, 0, mapGestionError("leer producto", err)
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, mapGestionError("listar productos", err)
	}
	return items, total, nil
}

func (r *ProductoRepository) Actualizar(ctx context.Context, p gestiondomain.Producto) (gestiondomain.Producto, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE productos
		SET descripcion = $3, unidad = $4, precio = $5, afectacion = $6, activo = $7,
		    actualizado_en = now()
		WHERE empresa_id = $1 AND id = $2
		RETURNING creado_en, actualizado_en`,
		p.EmpresaID, p.ID, p.Descripcion, p.Unidad, p.Precio, p.Afectacion, p.Activo,
	).Scan(&p.CreadoEn, &p.ActualizadoEn)
	if err != nil {
		return gestiondomain.Producto{}, mapGestionError("actualizar producto", err)
	}
	return p, nil
}

type productoScanner interface {
	Scan(dest ...any) error
}

func scanProducto(s productoScanner) (gestiondomain.Producto, error) {
	var p gestiondomain.Producto
	err := s.Scan(
		&p.ID, &p.EmpresaID, &p.Codigo, &p.Descripcion, &p.Unidad, &p.Precio,
		&p.Afectacion, &p.ControlaStock, &p.Activo, &p.CreadoEn, &p.ActualizadoEn,
	)
	if err != nil {
		return gestiondomain.Producto{}, err
	}
	return p, nil
}

func scanProductoDetalle(s productoScanner) (gestionapp.ProductoDetalle, error) {
	var p gestiondomain.Producto
	var stockRaw sql.NullInt64
	err := s.Scan(
		&p.ID, &p.EmpresaID, &p.Codigo, &p.Descripcion, &p.Unidad, &p.Precio,
		&p.Afectacion, &p.ControlaStock, &p.Activo, &p.CreadoEn, &p.ActualizadoEn,
		&stockRaw,
	)
	if err != nil {
		return gestionapp.ProductoDetalle{}, err
	}
	var stock *int64
	if p.ControlaStock {
		saldo := int64(0)
		if stockRaw.Valid {
			saldo = stockRaw.Int64
		}
		stock = &saldo
	}
	return gestionapp.ProductoDetalle{Producto: p, Stock: stock}, nil
}

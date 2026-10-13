package postgres

import (
	"context"
	"fmt"

	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClienteRepository struct {
	pool *pgxpool.Pool
}

func NewClienteRepository(pool *pgxpool.Pool) *ClienteRepository {
	return &ClienteRepository{pool: pool}
}

func (r *ClienteRepository) Crear(ctx context.Context, c gestiondomain.Cliente) (gestiondomain.Cliente, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO clientes (
			id, empresa_id, tipo_doc, numero, razon_social, direccion, correo, activo
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING creado_en, actualizado_en`,
		c.ID, c.EmpresaID, c.TipoDoc, c.Numero, c.RazonSocial, c.Direccion, c.Correo, c.Activo,
	).Scan(&c.CreadoEn, &c.ActualizadoEn)
	if err != nil {
		return gestiondomain.Cliente{}, mapGestionError("crear cliente", err)
	}
	return c, nil
}

func (r *ClienteRepository) Obtener(ctx context.Context, empresaID, id string) (gestiondomain.Cliente, error) {
	return r.scanOne(ctx, `
		SELECT id, empresa_id, tipo_doc, numero, razon_social, direccion, correo, activo, creado_en, actualizado_en
		FROM clientes
		WHERE empresa_id = $1 AND id = $2`,
		empresaID, id,
	)
}

func (r *ClienteRepository) Listar(
	ctx context.Context,
	empresaID string,
	f gestionapp.ListarClientesFilter,
) ([]gestiondomain.Cliente, int, error) {
	limit := f.Limit
	offset := f.Offset
	args := []any{empresaID, f.Q, f.Activo}
	const filtros = `
		FROM clientes
		WHERE empresa_id = $1
		  AND ($2 = '' OR razon_social ILIKE '%' || $2 || '%' OR numero ILIKE '%' || $2 || '%')
		  AND ($3::boolean IS NULL OR activo = $3)`

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*)`+filtros, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar clientes: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, empresa_id, tipo_doc, numero, razon_social, direccion, correo, activo, creado_en, actualizado_en`+
		filtros+`
		ORDER BY razon_social, id
		LIMIT $4 OFFSET $5`,
		append(args, limit, offset)...,
	)
	if err != nil {
		return nil, 0, mapGestionError("listar clientes", err)
	}
	defer rows.Close()

	items := make([]gestiondomain.Cliente, 0)
	for rows.Next() {
		c, err := scanCliente(rows)
		if err != nil {
			return nil, 0, mapGestionError("leer cliente", err)
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, mapGestionError("listar clientes", err)
	}
	return items, total, nil
}

func (r *ClienteRepository) Actualizar(ctx context.Context, c gestiondomain.Cliente) (gestiondomain.Cliente, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE clientes
		SET razon_social = $3, direccion = $4, correo = $5, activo = $6, actualizado_en = now()
		WHERE empresa_id = $1 AND id = $2
		RETURNING creado_en, actualizado_en`,
		c.EmpresaID, c.ID, c.RazonSocial, c.Direccion, c.Correo, c.Activo,
	).Scan(&c.CreadoEn, &c.ActualizadoEn)
	if err != nil {
		return gestiondomain.Cliente{}, mapGestionError("actualizar cliente", err)
	}
	return c, nil
}

type clienteScanner interface {
	Scan(dest ...any) error
}

func scanCliente(s clienteScanner) (gestiondomain.Cliente, error) {
	var c gestiondomain.Cliente
	err := s.Scan(
		&c.ID, &c.EmpresaID, &c.TipoDoc, &c.Numero, &c.RazonSocial,
		&c.Direccion, &c.Correo, &c.Activo, &c.CreadoEn, &c.ActualizadoEn,
	)
	if err != nil {
		return gestiondomain.Cliente{}, err
	}
	return c, nil
}

func (r *ClienteRepository) scanOne(ctx context.Context, query string, args ...any) (gestiondomain.Cliente, error) {
	row := r.pool.QueryRow(ctx, query, args...)
	c, err := scanCliente(row)
	if err != nil {
		return gestiondomain.Cliente{}, mapGestionError("obtener cliente", err)
	}
	return c, nil
}

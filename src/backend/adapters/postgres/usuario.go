package postgres

import (
	"context"

	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsuarioRepository struct {
	pool *pgxpool.Pool
}

func NewUsuarioRepository(pool *pgxpool.Pool) *UsuarioRepository {
	return &UsuarioRepository{pool: pool}
}

func (r *UsuarioRepository) Crear(ctx context.Context, usuario gestiondomain.Usuario) (gestiondomain.Usuario, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO usuarios (id, empresa_id, email, password_hash, nombre, rol, activo)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING creado_en`,
		usuario.ID, usuario.EmpresaID, usuario.Email, usuario.PasswordHash,
		usuario.Nombre, usuario.Rol, usuario.Activo,
	).Scan(&usuario.CreadoEn)
	if err != nil {
		return gestiondomain.Usuario{}, mapGestionError("crear usuario", err)
	}
	return usuario, nil
}

func (r *UsuarioRepository) Obtener(ctx context.Context, empresaID, usuarioID string) (gestiondomain.Usuario, error) {
	return r.obtener(ctx, `
		SELECT id, empresa_id, email, password_hash, nombre, rol, activo, creado_en
		FROM usuarios
		WHERE empresa_id = $1 AND id = $2`,
		empresaID, usuarioID,
	)
}

func (r *UsuarioRepository) ObtenerPorEmail(ctx context.Context, email string) (gestiondomain.Usuario, error) {
	return r.obtener(ctx, `
		SELECT id, empresa_id, email, password_hash, nombre, rol, activo, creado_en
		FROM usuarios
		WHERE email = $1`,
		gestiondomain.NormalizarEmail(email),
	)
}

func (r *UsuarioRepository) obtener(ctx context.Context, query string, args ...any) (gestiondomain.Usuario, error) {
	var usuario gestiondomain.Usuario
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&usuario.ID, &usuario.EmpresaID, &usuario.Email, &usuario.PasswordHash,
		&usuario.Nombre, &usuario.Rol, &usuario.Activo, &usuario.CreadoEn,
	)
	if err != nil {
		return gestiondomain.Usuario{}, mapGestionError("obtener usuario", err)
	}
	return usuario, nil
}

func (r *UsuarioRepository) Listar(ctx context.Context, empresaID string) ([]gestiondomain.Usuario, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, empresa_id, email, password_hash, nombre, rol, activo, creado_en
		FROM usuarios
		WHERE empresa_id = $1
		ORDER BY creado_en, id`,
		empresaID,
	)
	if err != nil {
		return nil, mapGestionError("listar usuarios", err)
	}
	defer rows.Close()

	usuarios := make([]gestiondomain.Usuario, 0)
	for rows.Next() {
		var usuario gestiondomain.Usuario
		if err := rows.Scan(
			&usuario.ID, &usuario.EmpresaID, &usuario.Email, &usuario.PasswordHash,
			&usuario.Nombre, &usuario.Rol, &usuario.Activo, &usuario.CreadoEn,
		); err != nil {
			return nil, mapGestionError("leer usuario", err)
		}
		usuarios = append(usuarios, usuario)
	}
	if err := rows.Err(); err != nil {
		return nil, mapGestionError("listar usuarios", err)
	}
	return usuarios, nil
}

func (r *UsuarioRepository) Actualizar(ctx context.Context, usuario gestiondomain.Usuario) (gestiondomain.Usuario, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE usuarios
		SET email = $3, password_hash = $4, nombre = $5, rol = $6, activo = $7
		WHERE empresa_id = $1 AND id = $2
		RETURNING creado_en`,
		usuario.EmpresaID, usuario.ID, usuario.Email, usuario.PasswordHash,
		usuario.Nombre, usuario.Rol, usuario.Activo,
	).Scan(&usuario.CreadoEn)
	if err != nil {
		return gestiondomain.Usuario{}, mapGestionError("actualizar usuario", err)
	}
	return usuario, nil
}

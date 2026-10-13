package postgres

import (
	"context"
	"errors"
	"fmt"

	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmpresaRepository struct {
	pool *pgxpool.Pool
}

func NewEmpresaRepository(pool *pgxpool.Pool) *EmpresaRepository {
	return &EmpresaRepository{pool: pool}
}

func (r *EmpresaRepository) Crear(ctx context.Context, empresa gestiondomain.Empresa) (gestiondomain.Empresa, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO empresas (id, ruc, razon_social, domicilio_fiscal, regimen, activa)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING creado_en`,
		empresa.ID, empresa.RUC, empresa.RazonSocial, empresa.DomicilioFiscal, empresa.Regimen, empresa.Activa,
	).Scan(&empresa.CreadoEn)
	if err != nil {
		return gestiondomain.Empresa{}, mapGestionError("crear empresa", err)
	}
	return empresa, nil
}

func (r *EmpresaRepository) Obtener(ctx context.Context, empresaID string) (gestiondomain.Empresa, error) {
	var empresa gestiondomain.Empresa
	err := r.pool.QueryRow(ctx, `
		SELECT id, ruc, razon_social, domicilio_fiscal, regimen, activa, creado_en
		FROM empresas
		WHERE id = $1`,
		empresaID,
	).Scan(
		&empresa.ID, &empresa.RUC, &empresa.RazonSocial, &empresa.DomicilioFiscal,
		&empresa.Regimen, &empresa.Activa, &empresa.CreadoEn,
	)
	if err != nil {
		return gestiondomain.Empresa{}, mapGestionError("obtener empresa", err)
	}
	return empresa, nil
}

func (r *EmpresaRepository) Actualizar(ctx context.Context, empresa gestiondomain.Empresa) (gestiondomain.Empresa, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE empresas
		SET ruc = $2, razon_social = $3, domicilio_fiscal = $4, regimen = $5, activa = $6
		WHERE id = $1
		RETURNING creado_en`,
		empresa.ID, empresa.RUC, empresa.RazonSocial, empresa.DomicilioFiscal, empresa.Regimen, empresa.Activa,
	).Scan(&empresa.CreadoEn)
	if err != nil {
		return gestiondomain.Empresa{}, mapGestionError("actualizar empresa", err)
	}
	return empresa, nil
}

func mapGestionError(operacion string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", operacion, gestiondomain.ErrNoEncontrado)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%s: %w", operacion, gestiondomain.ErrConflicto)
	}
	return fmt.Errorf("%s: %w", operacion, err)
}

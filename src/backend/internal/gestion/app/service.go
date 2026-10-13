package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	platformauth "github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/auth"
	"github.com/google/uuid"
)

type EmpresaRepository interface {
	Crear(context.Context, domain.Empresa) (domain.Empresa, error)
	Obtener(context.Context, string) (domain.Empresa, error)
	Actualizar(context.Context, domain.Empresa) (domain.Empresa, error)
}

type UsuarioRepository interface {
	Crear(context.Context, domain.Usuario) (domain.Usuario, error)
	Obtener(context.Context, string, string) (domain.Usuario, error)
	ObtenerPorEmail(context.Context, string) (domain.Usuario, error)
	Listar(context.Context, string) ([]domain.Usuario, error)
	Actualizar(context.Context, domain.Usuario) (domain.Usuario, error)
}

type SerieRepository interface {
	Crear(context.Context, domain.Serie) (domain.Serie, error)
	Obtener(context.Context, string, domain.TipoSerie, string) (domain.Serie, error)
	Listar(context.Context, string) ([]domain.Serie, error)
	Actualizar(context.Context, domain.Serie) (domain.Serie, error)
}

type Service struct {
	empresas EmpresaRepository
	usuarios UsuarioRepository
	series   SerieRepository
}

func NewService(empresas EmpresaRepository, usuarios UsuarioRepository, series SerieRepository) *Service {
	return &Service{empresas: empresas, usuarios: usuarios, series: series}
}

type CrearEmpresaInput struct {
	RUC             string
	RazonSocial     string
	DomicilioFiscal string
	Regimen         string
}

func (s *Service) CrearEmpresa(ctx context.Context, input CrearEmpresaInput) (domain.Empresa, error) {
	empresa := domain.Empresa{
		ID:              uuid.NewString(),
		RUC:             strings.TrimSpace(input.RUC),
		RazonSocial:     strings.TrimSpace(input.RazonSocial),
		DomicilioFiscal: strings.TrimSpace(input.DomicilioFiscal),
		Regimen:         strings.TrimSpace(input.Regimen),
		Activa:          true,
	}
	if err := domain.ValidarEmpresa(empresa); err != nil {
		return domain.Empresa{}, err
	}
	return s.empresas.Crear(ctx, empresa)
}

func (s *Service) ObtenerEmpresa(ctx context.Context, empresaID string) (domain.Empresa, error) {
	if empresaID == "" {
		return domain.Empresa{}, fmt.Errorf("%w: empresa requerida", domain.ErrInvalido)
	}
	return s.empresas.Obtener(ctx, empresaID)
}

type ActualizarEmpresaInput struct {
	RUC             *string
	RazonSocial     *string
	DomicilioFiscal *string
	Regimen         *string
	Activa          *bool
}

func (s *Service) ActualizarEmpresaMe(ctx context.Context, empresaID string, input ActualizarEmpresaInput) (domain.Empresa, error) {
	empresa, err := s.ObtenerEmpresa(ctx, empresaID)
	if err != nil {
		return domain.Empresa{}, err
	}
	if input.RUC != nil {
		empresa.RUC = strings.TrimSpace(*input.RUC)
	}
	if input.RazonSocial != nil {
		empresa.RazonSocial = strings.TrimSpace(*input.RazonSocial)
	}
	if input.DomicilioFiscal != nil {
		empresa.DomicilioFiscal = strings.TrimSpace(*input.DomicilioFiscal)
	}
	if input.Regimen != nil {
		empresa.Regimen = strings.TrimSpace(*input.Regimen)
	}
	if input.Activa != nil {
		empresa.Activa = *input.Activa
	}
	if err := domain.ValidarEmpresa(empresa); err != nil {
		return domain.Empresa{}, err
	}
	return s.empresas.Actualizar(ctx, empresa)
}

type CrearUsuarioInput struct {
	Email    string
	Password string
	Nombre   string
	Rol      domain.Rol
}

func (s *Service) CrearUsuario(ctx context.Context, empresaID string, input CrearUsuarioInput) (domain.Usuario, error) {
	if empresaID == "" {
		return domain.Usuario{}, fmt.Errorf("%w: empresa requerida", domain.ErrInvalido)
	}
	email := domain.NormalizarEmail(input.Email)
	if err := domain.ValidarEmail(email); err != nil {
		return domain.Usuario{}, err
	}
	if err := domain.ValidarRol(input.Rol); err != nil {
		return domain.Usuario{}, err
	}
	if len(input.Password) < 8 {
		return domain.Usuario{}, fmt.Errorf("%w: password debe tener al menos 8 caracteres", domain.ErrInvalido)
	}
	hash, err := platformauth.HashPassword(input.Password)
	if err != nil {
		return domain.Usuario{}, fmt.Errorf("hash password: %w", err)
	}
	usuario := domain.Usuario{
		ID:           uuid.NewString(),
		EmpresaID:    empresaID,
		Email:        email,
		PasswordHash: hash,
		Nombre:       strings.TrimSpace(input.Nombre),
		Rol:          input.Rol,
		Activo:       true,
	}
	return s.usuarios.Crear(ctx, usuario)
}

func (s *Service) ListarUsuarios(ctx context.Context, empresaID string) ([]domain.Usuario, error) {
	if empresaID == "" {
		return nil, fmt.Errorf("%w: empresa requerida", domain.ErrInvalido)
	}
	return s.usuarios.Listar(ctx, empresaID)
}

type ActualizarUsuarioInput struct {
	Email    *string
	Password *string
	Nombre   *string
	Rol      *domain.Rol
	Activo   *bool
}

func (s *Service) ActualizarUsuario(ctx context.Context, empresaID, usuarioID string, input ActualizarUsuarioInput) (domain.Usuario, error) {
	usuario, err := s.usuarios.Obtener(ctx, empresaID, usuarioID)
	if err != nil {
		return domain.Usuario{}, err
	}
	if input.Email != nil {
		usuario.Email = domain.NormalizarEmail(*input.Email)
		if err := domain.ValidarEmail(usuario.Email); err != nil {
			return domain.Usuario{}, err
		}
	}
	if input.Password != nil {
		if len(*input.Password) < 8 {
			return domain.Usuario{}, fmt.Errorf("%w: password debe tener al menos 8 caracteres", domain.ErrInvalido)
		}
		usuario.PasswordHash, err = platformauth.HashPassword(*input.Password)
		if err != nil {
			return domain.Usuario{}, fmt.Errorf("hash password: %w", err)
		}
	}
	if input.Nombre != nil {
		usuario.Nombre = strings.TrimSpace(*input.Nombre)
	}
	if input.Rol != nil {
		if err := domain.ValidarRol(*input.Rol); err != nil {
			return domain.Usuario{}, err
		}
		usuario.Rol = *input.Rol
	}
	if input.Activo != nil {
		usuario.Activo = *input.Activo
	}
	return s.usuarios.Actualizar(ctx, usuario)
}

type CrearSerieInput struct {
	Tipo      domain.TipoSerie
	Serie     string
	Siguiente int64
}

func (s *Service) CrearSerie(ctx context.Context, empresaID string, input CrearSerieInput) (domain.Serie, error) {
	serie := domain.Serie{
		EmpresaID: empresaID,
		Tipo:      input.Tipo,
		Serie:     strings.ToUpper(strings.TrimSpace(input.Serie)),
		Siguiente: input.Siguiente,
		Activa:    true,
	}
	if serie.EmpresaID == "" {
		return domain.Serie{}, fmt.Errorf("%w: empresa requerida", domain.ErrInvalido)
	}
	if serie.Siguiente == 0 {
		serie.Siguiente = 1
	}
	if serie.Siguiente < 1 {
		return domain.Serie{}, fmt.Errorf("%w: siguiente debe ser positivo", domain.ErrInvalido)
	}
	if err := domain.ValidarSerie(serie.Tipo, serie.Serie); err != nil {
		return domain.Serie{}, err
	}
	return s.series.Crear(ctx, serie)
}

func (s *Service) ListarSeries(ctx context.Context, empresaID string) ([]domain.Serie, error) {
	if empresaID == "" {
		return nil, fmt.Errorf("%w: empresa requerida", domain.ErrInvalido)
	}
	return s.series.Listar(ctx, empresaID)
}

type ActualizarSerieInput struct {
	Siguiente *int64
	Activa    *bool
}

func (s *Service) ActualizarSerie(ctx context.Context, empresaID string, tipo domain.TipoSerie, codigo string, input ActualizarSerieInput) (domain.Serie, error) {
	codigo = strings.ToUpper(strings.TrimSpace(codigo))
	if err := domain.ValidarSerie(tipo, codigo); err != nil {
		return domain.Serie{}, err
	}
	serie, err := s.series.Obtener(ctx, empresaID, tipo, codigo)
	if err != nil {
		return domain.Serie{}, err
	}
	if input.Siguiente != nil {
		if *input.Siguiente < 1 {
			return domain.Serie{}, fmt.Errorf("%w: siguiente debe ser positivo", domain.ErrInvalido)
		}
		serie.Siguiente = *input.Siguiente
	}
	if input.Activa != nil {
		serie.Activa = *input.Activa
	}
	return s.series.Actualizar(ctx, serie)
}

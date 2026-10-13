package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	platformauth "github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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

type ClienteRepository interface {
	Crear(context.Context, domain.Cliente) (domain.Cliente, error)
	Obtener(context.Context, string, string) (domain.Cliente, error)
	Listar(context.Context, string, ListarClientesFilter) ([]domain.Cliente, int, error)
	Actualizar(context.Context, domain.Cliente) (domain.Cliente, error)
}

type ListarClientesFilter struct {
	Q      string
	Activo *bool
	Limit  int
	Offset int
}

type ProductoRepository interface {
	Crear(context.Context, domain.Producto, *int64) (domain.Producto, error)
	Obtener(context.Context, string, string) (ProductoDetalle, error)
	Listar(context.Context, string, ListarProductosFilter) ([]domain.Producto, int, error)
	Actualizar(context.Context, domain.Producto) (domain.Producto, error)
}

type StockInventarioRepository interface {
	AjustarStock(
		context.Context, string, string, string, string, int64, string,
	) (AjustarStockResult, error)
	ListarMovimientos(context.Context, string, string, int, int) ([]domain.StockMovimiento, int, error)
}

type AjustarStockResult struct {
	ProductoID   string
	Cantidad     int64
	MovimientoID int64
}

type ProductoDetalle struct {
	Producto domain.Producto
	Stock    *int64
}

type ListarProductosFilter struct {
	Q      string
	Activo *bool
	Limit  int
	Offset int
}

type Service struct {
	empresas  EmpresaRepository
	usuarios  UsuarioRepository
	series    SerieRepository
	clientes  ClienteRepository
	productos ProductoRepository
	stock     StockInventarioRepository
}

func NewService(
	empresas EmpresaRepository,
	usuarios UsuarioRepository,
	series SerieRepository,
	clientes ClienteRepository,
	productos ProductoRepository,
	stock StockInventarioRepository,
) *Service {
	return &Service{
		empresas: empresas, usuarios: usuarios, series: series,
		clientes: clientes, productos: productos, stock: stock,
	}
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

type CrearClienteInput struct {
	TipoDoc     string
	Numero      string
	RazonSocial string
	Direccion   string
	Correo      string
}

func (s *Service) CrearCliente(ctx context.Context, empresaID string, input CrearClienteInput) (domain.Cliente, error) {
	if empresaID == "" {
		return domain.Cliente{}, fmt.Errorf("%w: empresa requerida", domain.ErrInvalido)
	}
	cliente := domain.Cliente{
		ID:          uuid.NewString(),
		EmpresaID:   empresaID,
		TipoDoc:     strings.TrimSpace(input.TipoDoc),
		Numero:      strings.TrimSpace(input.Numero),
		RazonSocial: strings.TrimSpace(input.RazonSocial),
		Direccion:   strings.TrimSpace(input.Direccion),
		Correo:      strings.TrimSpace(input.Correo),
		Activo:      true,
	}
	if err := domain.ValidarCliente(cliente); err != nil {
		return domain.Cliente{}, err
	}
	return s.clientes.Crear(ctx, cliente)
}

func (s *Service) ObtenerCliente(ctx context.Context, empresaID, clienteID string) (domain.Cliente, error) {
	if empresaID == "" || clienteID == "" {
		return domain.Cliente{}, fmt.Errorf("%w: empresa y cliente requeridos", domain.ErrInvalido)
	}
	return s.clientes.Obtener(ctx, empresaID, clienteID)
}

func (s *Service) ListarClientes(ctx context.Context, empresaID string, f ListarClientesFilter) ([]domain.Cliente, int, error) {
	if empresaID == "" {
		return nil, 0, fmt.Errorf("%w: empresa requerida", domain.ErrInvalido)
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	} else if limit > 100 {
		limit = 100
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	f.Limit = limit
	f.Offset = offset
	f.Q = strings.TrimSpace(f.Q)
	return s.clientes.Listar(ctx, empresaID, f)
}

type ActualizarClienteInput struct {
	RazonSocial *string
	Direccion   *string
	Correo      *string
	Activo      *bool
}

func (s *Service) ActualizarCliente(ctx context.Context, empresaID, clienteID string, input ActualizarClienteInput) (domain.Cliente, error) {
	cliente, err := s.ObtenerCliente(ctx, empresaID, clienteID)
	if err != nil {
		return domain.Cliente{}, err
	}
	if input.RazonSocial != nil {
		cliente.RazonSocial = strings.TrimSpace(*input.RazonSocial)
	}
	if input.Direccion != nil {
		cliente.Direccion = strings.TrimSpace(*input.Direccion)
	}
	if input.Correo != nil {
		cliente.Correo = strings.TrimSpace(*input.Correo)
	}
	if input.Activo != nil {
		cliente.Activo = *input.Activo
	}
	if err := domain.ValidarCliente(cliente); err != nil {
		return domain.Cliente{}, err
	}
	return s.clientes.Actualizar(ctx, cliente)
}

type CrearProductoInput struct {
	Codigo        string
	Descripcion   string
	Unidad        string
	Precio        decimal.Decimal
	Afectacion    string
	ControlaStock bool
	StockInicial  *int64
}

func (s *Service) CrearProducto(ctx context.Context, empresaID string, input CrearProductoInput) (ProductoDetalle, error) {
	if empresaID == "" {
		return ProductoDetalle{}, fmt.Errorf("%w: empresa requerida", domain.ErrInvalido)
	}
	unidad := strings.TrimSpace(input.Unidad)
	if unidad == "" {
		unidad = domain.UnidadNIU
	}
	producto := domain.Producto{
		ID:            uuid.NewString(),
		EmpresaID:     empresaID,
		Codigo:        strings.TrimSpace(input.Codigo),
		Descripcion:   strings.TrimSpace(input.Descripcion),
		Unidad:        unidad,
		Precio:        input.Precio,
		Afectacion:    strings.TrimSpace(input.Afectacion),
		ControlaStock: input.ControlaStock,
		Activo:        true,
	}
	if producto.ControlaStock && input.StockInicial == nil {
		return ProductoDetalle{}, fmt.Errorf("%w: stock_inicial requerido", domain.ErrInvalido)
	}
	if err := domain.ValidarProducto(producto, input.StockInicial); err != nil {
		return ProductoDetalle{}, err
	}
	if _, err := s.productos.Crear(ctx, producto, input.StockInicial); err != nil {
		return ProductoDetalle{}, err
	}
	return s.productos.Obtener(ctx, empresaID, producto.ID)
}

func (s *Service) ObtenerProducto(ctx context.Context, empresaID, productoID string) (ProductoDetalle, error) {
	if empresaID == "" || productoID == "" {
		return ProductoDetalle{}, fmt.Errorf("%w: empresa y producto requeridos", domain.ErrInvalido)
	}
	return s.productos.Obtener(ctx, empresaID, productoID)
}

func (s *Service) ListarProductos(ctx context.Context, empresaID string, f ListarProductosFilter) ([]domain.Producto, int, error) {
	if empresaID == "" {
		return nil, 0, fmt.Errorf("%w: empresa requerida", domain.ErrInvalido)
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	} else if limit > 100 {
		limit = 100
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	f.Limit = limit
	f.Offset = offset
	f.Q = strings.TrimSpace(f.Q)
	return s.productos.Listar(ctx, empresaID, f)
}

type ActualizarProductoInput struct {
	Descripcion *string
	Unidad      *string
	Precio      *decimal.Decimal
	Afectacion  *string
	Activo      *bool
}

func (s *Service) ActualizarProducto(
	ctx context.Context,
	empresaID, productoID string,
	input ActualizarProductoInput,
) (ProductoDetalle, error) {
	detalle, err := s.ObtenerProducto(ctx, empresaID, productoID)
	if err != nil {
		return ProductoDetalle{}, err
	}
	producto := detalle.Producto
	if input.Descripcion != nil {
		producto.Descripcion = strings.TrimSpace(*input.Descripcion)
	}
	if input.Unidad != nil {
		unidad := strings.TrimSpace(*input.Unidad)
		if unidad == "" {
			unidad = domain.UnidadNIU
		}
		producto.Unidad = unidad
	}
	if input.Precio != nil {
		producto.Precio = *input.Precio
	}
	if input.Afectacion != nil {
		producto.Afectacion = strings.TrimSpace(*input.Afectacion)
	}
	if input.Activo != nil {
		producto.Activo = *input.Activo
	}
	if err := domain.ValidarProducto(producto, nil); err != nil {
		return ProductoDetalle{}, err
	}
	if _, err := s.productos.Actualizar(ctx, producto); err != nil {
		return ProductoDetalle{}, err
	}
	return s.productos.Obtener(ctx, empresaID, productoID)
}

type AjustarStockInput struct {
	Tipo     string
	Cantidad int64
	Motivo   string
}

func (s *Service) AjustarStock(
	ctx context.Context,
	empresaID, productoID, usuarioID string,
	input AjustarStockInput,
) (AjustarStockResult, error) {
	if empresaID == "" || productoID == "" {
		return AjustarStockResult{}, fmt.Errorf("%w: empresa y producto requeridos", domain.ErrInvalido)
	}
	if err := domain.ValidarAjusteStock(input.Tipo, input.Cantidad, input.Motivo); err != nil {
		return AjustarStockResult{}, err
	}
	detalle, err := s.ObtenerProducto(ctx, empresaID, productoID)
	if err != nil {
		return AjustarStockResult{}, err
	}
	if !detalle.Producto.ControlaStock {
		return AjustarStockResult{}, fmt.Errorf("%w: producto no controla stock", domain.ErrInvalido)
	}
	motivo := strings.TrimSpace(input.Motivo)
	return s.stock.AjustarStock(ctx, empresaID, productoID, usuarioID, input.Tipo, input.Cantidad, motivo)
}

func (s *Service) ListarMovimientos(
	ctx context.Context,
	empresaID, productoID string,
	limit, offset int,
) ([]domain.StockMovimiento, int, error) {
	if empresaID == "" || productoID == "" {
		return nil, 0, fmt.Errorf("%w: empresa y producto requeridos", domain.ErrInvalido)
	}
	if _, err := s.ObtenerProducto(ctx, empresaID, productoID); err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	} else if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.stock.ListarMovimientos(ctx, empresaID, productoID, limit, offset)
}

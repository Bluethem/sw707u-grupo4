package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/redis"
	factapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/app"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	platformauth "github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/auth"
	"golang.org/x/crypto/bcrypt"
)

type SessionStore interface {
	Create(context.Context, string, redis.SessionData) error
	Get(context.Context, string) (*redis.SessionData, error)
	Delete(context.Context, string) error
}

type Dependencies struct {
	Gestion  *gestionapp.Service
	Usuarios gestionapp.UsuarioRepository
	Sessions SessionStore
	Emisor   *factapp.Emisor
	Query    ports.ComprobanteQuery
}

type API struct {
	gestion  *gestionapp.Service
	usuarios gestionapp.UsuarioRepository
	sessions SessionStore
	emisor   *factapp.Emisor
	query    ports.ComprobanteQuery
}

func NewAPIHandler(deps Dependencies) http.Handler {
	api := &API{
		gestion: deps.Gestion, usuarios: deps.Usuarios, sessions: deps.Sessions,
		emisor: deps.Emisor, query: deps.Query,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", api.health)
	mux.HandleFunc("POST /api/v1/auth/login", api.login)

	authenticated := platformauth.Authenticate(api.lookupSession)
	admin := func(handler http.HandlerFunc) http.Handler {
		return authenticated(platformauth.RequireRole(platformauth.RolAdministrador)(handler))
	}
	mux.Handle("POST /api/v1/auth/logout", authenticated(http.HandlerFunc(api.logout)))
	mux.Handle("GET /api/v1/auth/me", authenticated(http.HandlerFunc(api.me)))
	mux.Handle("POST /api/v1/comprobantes", authenticated(http.HandlerFunc(api.emitirComprobante)))
	mux.Handle("GET /api/v1/comprobantes/{id}", authenticated(http.HandlerFunc(api.obtenerComprobante)))
	mux.Handle("GET /api/v1/comprobantes", authenticated(http.HandlerFunc(api.listarComprobantes)))
	mux.Handle("POST /api/v1/clientes", authenticated(http.HandlerFunc(api.crearCliente)))
	mux.Handle("GET /api/v1/clientes", authenticated(http.HandlerFunc(api.listarClientes)))
	mux.Handle("GET /api/v1/clientes/{id}", authenticated(http.HandlerFunc(api.obtenerCliente)))
	mux.Handle("PATCH /api/v1/clientes/{id}", authenticated(http.HandlerFunc(api.actualizarCliente)))
	mux.Handle("GET /api/v1/empresas/me", admin(api.obtenerEmpresa))
	mux.Handle("PATCH /api/v1/empresas/me", admin(api.actualizarEmpresa))
	mux.Handle("POST /api/v1/usuarios", admin(api.crearUsuario))
	mux.Handle("GET /api/v1/usuarios", admin(api.listarUsuarios))
	mux.Handle("PATCH /api/v1/usuarios", admin(api.actualizarUsuario))
	mux.Handle("POST /api/v1/series", admin(api.crearSerie))
	mux.Handle("GET /api/v1/series", admin(api.listarSeries))
	mux.Handle("PATCH /api/v1/series", admin(api.actualizarSerie))
	return mux
}

func (a *API) lookupSession(ctx context.Context, token string) (platformauth.User, error) {
	session, err := a.sessions.Get(ctx, token)
	if err != nil {
		return platformauth.User{}, err
	}
	if session == nil {
		return platformauth.User{}, platformauth.ErrSessionNotFound
	}
	return platformauth.User{ID: session.UserID, EmpresaID: session.EmpresaID, Rol: session.Rol}, nil
}

func (a *API) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	usuario, err := a.usuarios.ObtenerPorEmail(r.Context(), gestiondomain.NormalizarEmail(input.Email))
	if errors.Is(err, gestiondomain.ErrNoEncontrado) {
		writeError(w, http.StatusUnauthorized, "credenciales invalidas")
		return
	}
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if !usuario.Activo {
		writeError(w, http.StatusUnauthorized, "credenciales invalidas")
		return
	}
	if err := platformauth.VerifyPassword(usuario.PasswordHash, input.Password); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			writeError(w, http.StatusUnauthorized, "credenciales invalidas")
		} else {
			writeDomainError(w, err)
		}
		return
	}
	empresa, err := a.gestion.ObtenerEmpresa(r.Context(), usuario.EmpresaID)
	if errors.Is(err, gestiondomain.ErrNoEncontrado) {
		writeError(w, http.StatusUnauthorized, "credenciales invalidas")
		return
	}
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if !empresa.Activa {
		writeError(w, http.StatusUnauthorized, "credenciales invalidas")
		return
	}
	token, err := redis.GenerateToken()
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if err := a.sessions.Create(r.Context(), token, redis.SessionData{
		UserID: usuario.ID, EmpresaID: usuario.EmpresaID, Rol: string(usuario.Rol),
	}); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	user, _ := platformauth.UserFromContext(r.Context())
	if err := a.sessions.Delete(r.Context(), user.Token); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) me(w http.ResponseWriter, r *http.Request) {
	current, _ := platformauth.UserFromContext(r.Context())
	usuario, err := a.usuarios.Obtener(r.Context(), current.EmpresaID, current.ID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	empresa, err := a.gestion.ObtenerEmpresa(r.Context(), current.EmpresaID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"usuario": usuarioResponse(usuario),
		"empresa": empresaResponse(empresa),
		"rol":     usuario.Rol,
	})
}

func (a *API) obtenerEmpresa(w http.ResponseWriter, r *http.Request) {
	current, _ := platformauth.UserFromContext(r.Context())
	empresa, err := a.gestion.ObtenerEmpresa(r.Context(), current.EmpresaID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, empresaResponse(empresa))
}

func (a *API) actualizarEmpresa(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RUC             *string `json:"ruc"`
		RazonSocial     *string `json:"razon_social"`
		DomicilioFiscal *string `json:"domicilio_fiscal"`
		Regimen         *string `json:"regimen"`
		Activa          *bool   `json:"activa"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	current, _ := platformauth.UserFromContext(r.Context())
	empresa, err := a.gestion.ActualizarEmpresaMe(r.Context(), current.EmpresaID, gestionapp.ActualizarEmpresaInput{
		RUC: input.RUC, RazonSocial: input.RazonSocial, DomicilioFiscal: input.DomicilioFiscal,
		Regimen: input.Regimen, Activa: input.Activa,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, empresaResponse(empresa))
}

func (a *API) crearUsuario(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string            `json:"email"`
		Password string            `json:"password"`
		Nombre   string            `json:"nombre"`
		Rol      gestiondomain.Rol `json:"rol"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	current, _ := platformauth.UserFromContext(r.Context())
	usuario, err := a.gestion.CrearUsuario(r.Context(), current.EmpresaID, gestionapp.CrearUsuarioInput{
		Email: input.Email, Password: input.Password, Nombre: input.Nombre, Rol: input.Rol,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, usuarioResponse(usuario))
}

func (a *API) listarUsuarios(w http.ResponseWriter, r *http.Request) {
	current, _ := platformauth.UserFromContext(r.Context())
	usuarios, err := a.gestion.ListarUsuarios(r.Context(), current.EmpresaID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response := make([]any, 0, len(usuarios))
	for _, usuario := range usuarios {
		response = append(response, usuarioResponse(usuario))
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *API) actualizarUsuario(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID       string             `json:"id"`
		Email    *string            `json:"email"`
		Password *string            `json:"password"`
		Nombre   *string            `json:"nombre"`
		Rol      *gestiondomain.Rol `json:"rol"`
		Activo   *bool              `json:"activo"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if strings.TrimSpace(input.ID) == "" {
		writeError(w, http.StatusBadRequest, "id requerido")
		return
	}
	current, _ := platformauth.UserFromContext(r.Context())
	usuario, err := a.gestion.ActualizarUsuario(r.Context(), current.EmpresaID, input.ID, gestionapp.ActualizarUsuarioInput{
		Email: input.Email, Password: input.Password, Nombre: input.Nombre, Rol: input.Rol, Activo: input.Activo,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, usuarioResponse(usuario))
}

func (a *API) crearSerie(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Tipo      gestiondomain.TipoSerie `json:"tipo"`
		Serie     string                  `json:"serie"`
		Siguiente int64                   `json:"siguiente"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	current, _ := platformauth.UserFromContext(r.Context())
	serie, err := a.gestion.CrearSerie(r.Context(), current.EmpresaID, gestionapp.CrearSerieInput(input))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, serieResponse(serie))
}

func (a *API) listarSeries(w http.ResponseWriter, r *http.Request) {
	current, _ := platformauth.UserFromContext(r.Context())
	series, err := a.gestion.ListarSeries(r.Context(), current.EmpresaID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response := make([]any, 0, len(series))
	for _, serie := range series {
		response = append(response, serieResponse(serie))
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *API) actualizarSerie(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Tipo      gestiondomain.TipoSerie `json:"tipo"`
		Serie     string                  `json:"serie"`
		Siguiente *int64                  `json:"siguiente"`
		Activa    *bool                   `json:"activa"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	current, _ := platformauth.UserFromContext(r.Context())
	serie, err := a.gestion.ActualizarSerie(r.Context(), current.EmpresaID, input.Tipo, input.Serie, gestionapp.ActualizarSerieInput{
		Siguiente: input.Siguiente, Activa: input.Activa,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, serieResponse(serie))
}

func empresaResponse(empresa gestiondomain.Empresa) map[string]any {
	return map[string]any{
		"id": empresa.ID, "ruc": empresa.RUC, "razon_social": empresa.RazonSocial,
		"domicilio_fiscal": empresa.DomicilioFiscal, "regimen": empresa.Regimen,
		"activa": empresa.Activa, "creado_en": empresa.CreadoEn,
	}
}

func usuarioResponse(usuario gestiondomain.Usuario) map[string]any {
	return map[string]any{
		"id": usuario.ID, "empresa_id": usuario.EmpresaID, "email": usuario.Email,
		"nombre": usuario.Nombre, "rol": usuario.Rol, "activo": usuario.Activo, "creado_en": usuario.CreadoEn,
	}
}

func serieResponse(serie gestiondomain.Serie) map[string]any {
	return map[string]any{
		"empresa_id": serie.EmpresaID, "tipo": serie.Tipo, "serie": serie.Serie,
		"siguiente": serie.Siguiente, "activa": serie.Activa,
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "json invalido")
		return false
	}
	return true
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, gestiondomain.ErrInvalido):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, gestiondomain.ErrNoEncontrado):
		writeError(w, http.StatusNotFound, "recurso no encontrado")
	case errors.Is(err, gestiondomain.ErrConflicto):
		writeError(w, http.StatusConflict, "recurso duplicado")
	default:
		writeError(w, http.StatusInternalServerError, "error interno")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

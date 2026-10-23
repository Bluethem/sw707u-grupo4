package httpapi

import (
	"io"
	"net/http"

	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	platformauth "github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/auth"
)

// cargaMasivaProductos maneja POST /api/v1/productos/carga-masiva
// Endpoint legacy: parsea, valida e inserta en un solo paso.
func (a *API) cargaMasivaProductos(w http.ResponseWriter, r *http.Request) {
	current, _ := platformauth.UserFromContext(r.Context())

	data, ok := leerArchivoMultipart(w, r, a, current.EmpresaID)
	if !ok {
		return
	}

	resultado, err := a.gestion.CargaMasivaProductos(r.Context(), current.EmpresaID, data)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if len(resultado.Errores) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errores": resultado.Errores})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"cantidad": resultado.Cantidad})
}

// previewCargaMasiva maneja POST /api/v1/productos/carga-masiva/preview
// Parsea y valida el Excel pero NO inserta en BD.
// Respuesta éxito: {"filas": [...]}
// Respuesta con errores: {"errores": [...]}  (HTTP 422)
func (a *API) previewCargaMasiva(w http.ResponseWriter, r *http.Request) {
	current, _ := platformauth.UserFromContext(r.Context())

	data, ok := leerArchivoMultipart(w, r, a, current.EmpresaID)
	if !ok {
		return
	}

	filas, errores, err := a.gestion.PreviewCargaMasiva(r.Context(), current.EmpresaID, data)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if len(errores) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errores": errores})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"filas": filas})
}

// confirmarCargaMasiva maneja POST /api/v1/productos/carga-masiva/confirmar
// Acepta JSON con las filas (editadas) y las inserta en BD tras re-validarlas.
// Body: {"filas": [...FilaProducto]}
func (a *API) confirmarCargaMasiva(w http.ResponseWriter, r *http.Request) {
	current, _ := platformauth.UserFromContext(r.Context())

	var body struct {
		Filas []gestionapp.FilaProducto `json:"filas"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if len(body.Filas) == 0 {
		writeError(w, http.StatusBadRequest, "filas requeridas")
		return
	}

	resultado, err := a.gestion.ConfirmarCargaMasiva(r.Context(), current.EmpresaID, body.Filas)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if len(resultado.Errores) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errores": resultado.Errores})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"cantidad": resultado.Cantidad})
}

// leerArchivoMultipart extrae el campo "archivo" del formulario multipart y lo retorna como bytes.
// Retorna (nil, false) si hubo un error y ya escribió la respuesta HTTP.
func leerArchivoMultipart(w http.ResponseWriter, r *http.Request, a *API, empresaID string) ([]byte, bool) {
	cfg, err := a.gestion.ObtenerConfigCatalogo(r.Context(), empresaID)
	if err != nil {
		writeDomainError(w, err)
		return nil, false
	}

	maxBytes := int64(cfg.CargaMaxMB) * 1024 * 1024
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		writeError(w, http.StatusBadRequest, "archivo demasiado grande o multipart invalido")
		return nil, false
	}

	file, _, err := r.FormFile("archivo")
	if err != nil {
		writeError(w, http.StatusBadRequest, "campo 'archivo' requerido en el formulario")
		return nil, false
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "error leyendo el archivo")
		return nil, false
	}
	if int64(len(data)) > maxBytes {
		writeError(w, http.StatusBadRequest, "archivo supera el tamaño máximo permitido")
		return nil, false
	}
	return data, true
}

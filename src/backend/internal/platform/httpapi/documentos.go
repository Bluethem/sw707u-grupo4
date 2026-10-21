package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	facturaciondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	platformauth "github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/auth"
)

func (a *API) descargarPDFComprobante(w http.ResponseWriter, r *http.Request) {
	a.descargarComprobante(w, r, "pdf", "application/pdf", a.documentos.PDF)
}

func (a *API) descargarXMLComprobante(w http.ResponseWriter, r *http.Request) {
	a.descargarComprobante(w, r, "xml", "application/xml", a.documentos.XML)
}

func (a *API) descargarCDRComprobante(w http.ResponseWriter, r *http.Request) {
	a.descargarComprobante(w, r, "cdr", "application/octet-stream", a.documentos.CDR)
}

type documentoLoader func(ctx context.Context, empresaID, id string) ([]byte, error)

func (a *API) descargarComprobante(
	w http.ResponseWriter,
	r *http.Request,
	ext, contentType string,
	load documentoLoader,
) {
	if a.documentos == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	user, _ := platformauth.UserFromContext(r.Context())
	id := r.PathValue("id")
	data, err := load(r.Context(), user.EmpresaID, id)
	if err != nil {
		writeDocumentosError(w, err)
		return
	}
	if a.query == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	c, err := a.query.GetByID(r.Context(), user.EmpresaID, id)
	if err != nil || c == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	filename := fmt.Sprintf("%s-%d.%s", c.Serie, c.Numero, ext)
	writeRaw(w, http.StatusOK, contentType, filename, data)
}

func writeDocumentosError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, facturaciondomain.ErrComprobanteNoEncontrado),
		errors.Is(err, facturaciondomain.ErrCDRNoEncontrado):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, facturaciondomain.ErrCDRNoDisponible):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "error interno")
	}
}

func writeRaw(w http.ResponseWriter, status int, contentType, filename string, data []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

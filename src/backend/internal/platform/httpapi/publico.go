package httpapi

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	factapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/app"
)

func (a *API) consultaPublica(w http.ResponseWriter, r *http.Request) {
	if a.consulta == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	var input struct {
		RUC               string `json:"ruc"`
		TipoDoc           string `json:"tipo_doc"`
		Numero            string `json:"numero"`
		Serie             string `json:"serie"`
		NumeroComprobante int64  `json:"numero_comprobante"`
		Fecha             string `json:"fecha"`
		Monto             string `json:"monto"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	result, err := a.consulta.Buscar(r.Context(), factapp.ConsultaInput{
		RUC: input.RUC, TipoDoc: input.TipoDoc, Numero: input.Numero, Serie: input.Serie,
		Fecha: input.Fecha, Monto: input.Monto, NumeroComprobante: input.NumeroComprobante,
	}, clientIP(r))
	if err != nil {
		writeConsultaError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id": result.ID, "tipo": result.Tipo, "serie": result.Serie, "numero": result.Numero,
		"fecha": result.Fecha, "total": result.Total, "estado": result.Estado,
		"razon_social_emisor": result.RazonSocialEmisor, "token": result.Token,
	})
}

func (a *API) descargarPDFPublico(w http.ResponseWriter, r *http.Request) {
	a.descargarDocumentoPublico(w, r, "pdf", "application/pdf", a.documentos.PDF)
}

func (a *API) descargarXMLPublico(w http.ResponseWriter, r *http.Request) {
	a.descargarDocumentoPublico(w, r, "xml", "application/xml", a.documentos.XML)
}

func (a *API) descargarDocumentoPublico(
	w http.ResponseWriter,
	r *http.Request,
	ext, contentType string,
	load documentoLoader,
) {
	if a.consulta == nil || a.documentos == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	id := r.PathValue("id")
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	empresaID, ok, err := a.consulta.ValidarToken(r.Context(), token, id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if !ok {
		writeError(w, http.StatusUnauthorized, "token invalido")
		return
	}
	data, err := load(r.Context(), empresaID, id)
	if err != nil {
		writeDocumentosError(w, err)
		return
	}
	if a.query == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	c, err := a.query.GetByID(r.Context(), empresaID, id)
	if err != nil || c == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	filename := fmt.Sprintf("%s-%d.%s", c.Serie, c.Numero, ext)
	writeRaw(w, http.StatusOK, contentType, filename, data)
}

func writeConsultaError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, factapp.ErrDatosNoCoinciden):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, factapp.ErrConsultaExpirada):
		writeError(w, http.StatusGone, err.Error())
	case errors.Is(err, factapp.ErrConsultaRateLimit):
		writeError(w, http.StatusTooManyRequests, err.Error())
	default:
		writeDomainError(w, err)
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

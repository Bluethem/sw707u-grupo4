package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	platformauth "github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/auth"
)

func (a *API) generarResumen(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Fecha string `json:"fecha"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	fecha, err := time.Parse("2006-01-02", strings.TrimSpace(input.Fecha))
	if err != nil {
		writeError(w, http.StatusBadRequest, domain.ErrFechaResumenInvalida.Error())
		return
	}
	user, _ := platformauth.UserFromContext(r.Context())
	if a.resumenes == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	resumen, err := a.resumenes.GenerarResumen(r.Context(), user.EmpresaID, fecha)
	if err != nil {
		writeFacturacionError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resumenResponse(resumen))
}

func (a *API) listarResumenes(w http.ResponseWriter, r *http.Request) {
	user, _ := platformauth.UserFromContext(r.Context())
	if a.resumenes == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	items, err := a.resumenes.Listar(r.Context(), user.EmpresaID)
	if err != nil {
		writeFacturacionError(w, err)
		return
	}
	response := make([]any, 0, len(items))
	for i := range items {
		response = append(response, resumenResponse(&items[i]))
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *API) obtenerResumen(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	user, _ := platformauth.UserFromContext(r.Context())
	if a.resumenes == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	resumen, err := a.resumenes.Obtener(r.Context(), user.EmpresaID, id)
	if err != nil {
		writeFacturacionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resumenResponse(resumen))
}

func resumenResponse(r *domain.ResumenDiario) map[string]any {
	comps := make([]map[string]any, 0, len(r.Comprobantes))
	for _, c := range r.Comprobantes {
		comps = append(comps, map[string]any{
			"comprobante_id": c.ComprobanteID,
			"anulado":        c.Anulado,
			"tipo":           c.Tipo,
			"serie":          c.Serie,
			"numero":         c.Numero,
			"estado":         c.Estado,
		})
	}
	return map[string]any{
		"id":           r.ID,
		"empresa_id":   r.EmpresaID,
		"fecha":        r.Fecha.Format("2006-01-02"),
		"correlativo":  r.Correlativo,
		"estado":       r.Estado,
		"ticket":       r.Ticket,
		"creado_en":    r.CreadoEn,
		"comprobantes": comps,
	}
}

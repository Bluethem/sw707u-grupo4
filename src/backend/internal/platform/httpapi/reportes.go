package httpapi

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	factapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/app"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	platformauth "github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/auth"
)

func (a *API) reporteVentasJSON(w http.ResponseWriter, r *http.Request) {
	user, _ := platformauth.UserFromContext(r.Context())
	if a.reportes == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	desde, hasta, err := parseReporteFechas(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	rep, err := a.reportes.Ventas(r.Context(), user.EmpresaID, desde, hasta)
	if err != nil {
		writeReporteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ventasReporteResponse(rep))
}

func (a *API) reporteVentasCSV(w http.ResponseWriter, r *http.Request) {
	user, _ := platformauth.UserFromContext(r.Context())
	if a.reportes == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	desde, hasta, err := parseReporteFechas(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	rep, err := a.reportes.Ventas(r.Context(), user.EmpresaID, desde, hasta)
	if err != nil {
		writeReporteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(
		`attachment; filename="ventas-%s.csv"`, rep.FechaHasta,
	))
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"tipo", "cantidad", "valor_venta", "igv", "total"})
	for _, row := range rep.PorTipo {
		_ = writer.Write(ventasTipoCSV(row))
	}
	_ = writer.Write(ventasTipoCSV(rep.Totales, "totales"))
	writer.Flush()
}

func parseReporteFechas(r *http.Request) (time.Time, time.Time, error) {
	today := time.Now().UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	q := r.URL.Query()
	desdeStr := strings.TrimSpace(q.Get("fecha_desde"))
	hastaStr := strings.TrimSpace(q.Get("fecha_hasta"))
	if desdeStr == "" && hastaStr == "" {
		return today, today, nil
	}
	if desdeStr == "" {
		desdeStr = hastaStr
	}
	if hastaStr == "" {
		hastaStr = desdeStr
	}
	desde, err := time.Parse(time.DateOnly, desdeStr)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("fecha_desde invalida")
	}
	hasta, err := time.Parse(time.DateOnly, hastaStr)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("fecha_hasta invalida")
	}
	return desde, hasta, nil
}

func ventasReporteResponse(rep factapp.VentasReporte) map[string]any {
	porTipo := make([]any, 0, len(rep.PorTipo))
	for _, row := range rep.PorTipo {
		porTipo = append(porTipo, ventasTipoJSON(row))
	}
	return map[string]any{
		"fecha_desde": rep.FechaDesde,
		"fecha_hasta": rep.FechaHasta,
		"por_tipo":    porTipo,
		"anulados": map[string]any{
			"cantidad": rep.Anulados.Cantidad,
			"total":    rep.Anulados.Total,
		},
		"totales": ventasTipoJSON(rep.Totales),
	}
}

func ventasTipoJSON(row factapp.VentasTipo) map[string]any {
	return map[string]any{
		"tipo":        row.Tipo,
		"cantidad":    row.Cantidad,
		"valor_venta": row.ValorVenta,
		"igv":         row.IGV,
		"total":       row.Total,
	}
}

func ventasTipoCSV(row factapp.VentasTipo, tipoOverride ...string) []string {
	tipo := row.Tipo
	if len(tipoOverride) > 0 && tipoOverride[0] != "" {
		tipo = tipoOverride[0]
	}
	return []string{
		tipo,
		fmt.Sprintf("%d", row.Cantidad),
		row.ValorVenta.StringFixed(2),
		row.IGV.StringFixed(2),
		row.Total.StringFixed(2),
	}
}

func writeReporteError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrFechaReporteInvalida):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrRangoReporteExcedido):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "error interno")
	}
}

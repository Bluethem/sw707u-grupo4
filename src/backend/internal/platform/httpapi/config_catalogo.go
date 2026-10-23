package httpapi

import (
	"net/http"

	platformauth "github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/auth"
	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
)

// GET /api/v1/configuracion/catalogo
func (a *API) getConfigCatalogo(w http.ResponseWriter, r *http.Request) {
	user, _ := platformauth.UserFromContext(r.Context())
	cfg, err := a.gestion.ObtenerConfigCatalogo(r.Context(), user.EmpresaID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}

	var precioMaximo *string
	if cfg.PrecioMaximo != nil {
		s := cfg.PrecioMaximo.String()
		precioMaximo = &s
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"empresa_id":              cfg.EmpresaID,
		"precio_minimo":           cfg.PrecioMinimo.String(),
		"precio_maximo":           precioMaximo,
		"carga_max_filas":         cfg.CargaMaxFilas,
		"carga_max_mb":            cfg.CargaMaxMB,
		"unidades_permitidas":     cfg.UnidadesPermitidas,
		"afectaciones_permitidas": cfg.AfectacionesPermitidas,
		"actualizado_en":          cfg.ActualizadoEn,
	})
}

// PATCH /api/v1/configuracion/catalogo
func (a *API) patchConfigCatalogo(w http.ResponseWriter, r *http.Request) {
	user, _ := platformauth.UserFromContext(r.Context())

	var body struct {
		PrecioMinimo           *string  `json:"precio_minimo"`
		PrecioMaximo           *string  `json:"precio_maximo"`
		CargaMaxFilas          *int     `json:"carga_max_filas"`
		CargaMaxMB             *int     `json:"carga_max_mb"`
		UnidadesPermitidas     []string `json:"unidades_permitidas"`
		AfectacionesPermitidas []string `json:"afectaciones_permitidas"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}

	cfg, err := a.gestion.ActualizarConfigCatalogo(r.Context(), user.EmpresaID,
		gestionapp.ActualizarConfigCatalogoInput{
			PrecioMinimo:           body.PrecioMinimo,
			PrecioMaximo:           body.PrecioMaximo,
			CargaMaxFilas:          body.CargaMaxFilas,
			CargaMaxMB:             body.CargaMaxMB,
			UnidadesPermitidas:     body.UnidadesPermitidas,
			AfectacionesPermitidas: body.AfectacionesPermitidas,
		},
	)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	var precioMaximo *string
	if cfg.PrecioMaximo != nil {
		s := cfg.PrecioMaximo.String()
		precioMaximo = &s
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"empresa_id":              cfg.EmpresaID,
		"precio_minimo":           cfg.PrecioMinimo.String(),
		"precio_maximo":           precioMaximo,
		"carga_max_filas":         cfg.CargaMaxFilas,
		"carga_max_mb":            cfg.CargaMaxMB,
		"unidades_permitidas":     cfg.UnidadesPermitidas,
		"afectaciones_permitidas": cfg.AfectacionesPermitidas,
		"actualizado_en":          cfg.ActualizadoEn,
	})
}

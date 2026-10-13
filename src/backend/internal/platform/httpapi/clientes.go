package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	platformauth "github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/auth"
)

func (a *API) crearCliente(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TipoDoc     string `json:"tipo_doc"`
		Numero      string `json:"numero"`
		RazonSocial string `json:"razon_social"`
		Direccion   string `json:"direccion"`
		Correo      string `json:"correo"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	current, _ := platformauth.UserFromContext(r.Context())
	if a.gestion == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	cliente, err := a.gestion.CrearCliente(r.Context(), current.EmpresaID, gestionapp.CrearClienteInput{
		TipoDoc: input.TipoDoc, Numero: input.Numero,
		RazonSocial: input.RazonSocial, Direccion: input.Direccion, Correo: input.Correo,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, clienteResponse(cliente))
}

func (a *API) listarClientes(w http.ResponseWriter, r *http.Request) {
	filtro, err := listarClientesFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	current, _ := platformauth.UserFromContext(r.Context())
	if a.gestion == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	items, total, err := a.gestion.ListarClientes(r.Context(), current.EmpresaID, filtro)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response := make([]any, 0, len(items))
	for _, c := range items {
		response = append(response, clienteResponse(c))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response, "total": total})
}

func (a *API) obtenerCliente(w http.ResponseWriter, r *http.Request) {
	current, _ := platformauth.UserFromContext(r.Context())
	if a.gestion == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	cliente, err := a.gestion.ObtenerCliente(r.Context(), current.EmpresaID, r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, clienteResponse(cliente))
}

func (a *API) actualizarCliente(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RazonSocial *string `json:"razon_social"`
		Direccion   *string `json:"direccion"`
		Correo      *string `json:"correo"`
		Activo      *bool   `json:"activo"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	current, _ := platformauth.UserFromContext(r.Context())
	if a.gestion == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	cliente, err := a.gestion.ActualizarCliente(r.Context(), current.EmpresaID, r.PathValue("id"), gestionapp.ActualizarClienteInput{
		RazonSocial: input.RazonSocial, Direccion: input.Direccion, Correo: input.Correo, Activo: input.Activo,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, clienteResponse(cliente))
}

func listarClientesFilter(r *http.Request) (gestionapp.ListarClientesFilter, error) {
	query := r.URL.Query()
	filtro := gestionapp.ListarClientesFilter{Q: query.Get("q")}
	var err error
	if value := query.Get("activo"); value != "" {
		switch value {
		case "true":
			v := true
			filtro.Activo = &v
		case "false":
			v := false
			filtro.Activo = &v
		default:
			return filtro, errors.New("activo invalido")
		}
	}
	if value := query.Get("limit"); value != "" {
		filtro.Limit, err = strconv.Atoi(value)
		if err != nil || filtro.Limit < 1 {
			return filtro, errors.New("limit invalido")
		}
	}
	if value := query.Get("offset"); value != "" {
		filtro.Offset, err = strconv.Atoi(value)
		if err != nil || filtro.Offset < 0 {
			return filtro, errors.New("offset invalido")
		}
	}
	return filtro, nil
}

func clienteResponse(c gestiondomain.Cliente) map[string]any {
	return map[string]any{
		"id": c.ID, "tipo_doc": c.TipoDoc, "numero": c.Numero,
		"razon_social": c.RazonSocial, "direccion": c.Direccion, "correo": c.Correo,
		"activo": c.Activo, "creado_en": c.CreadoEn, "actualizado_en": c.ActualizadoEn,
	}
}

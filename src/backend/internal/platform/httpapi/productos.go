package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	platformauth "github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/auth"
	"github.com/shopspring/decimal"
)

func (a *API) crearProducto(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Codigo        string `json:"codigo"`
		Descripcion   string `json:"descripcion"`
		Unidad        string `json:"unidad"`
		Precio        string `json:"precio"`
		Afectacion    string `json:"afectacion"`
		ControlaStock bool   `json:"controla_stock"`
		StockInicial  *int64 `json:"stock_inicial"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	precio, err := decimal.NewFromString(input.Precio)
	if err != nil {
		writeError(w, http.StatusBadRequest, "precio invalido")
		return
	}
	current, _ := platformauth.UserFromContext(r.Context())
	if a.gestion == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	detalle, err := a.gestion.CrearProducto(r.Context(), current.EmpresaID, gestionapp.CrearProductoInput{
		Codigo: input.Codigo, Descripcion: input.Descripcion, Unidad: input.Unidad,
		Precio: precio, Afectacion: input.Afectacion, ControlaStock: input.ControlaStock,
		StockInicial: input.StockInicial,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, productoDetalleResponse(detalle))
}

func (a *API) listarProductos(w http.ResponseWriter, r *http.Request) {
	filtro, err := listarProductosFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	current, _ := platformauth.UserFromContext(r.Context())
	if a.gestion == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	items, total, err := a.gestion.ListarProductos(r.Context(), current.EmpresaID, filtro)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response := make([]any, 0, len(items))
	for _, d := range items {
		response = append(response, productoDetalleResponse(d))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response, "total": total})
}

func (a *API) obtenerProducto(w http.ResponseWriter, r *http.Request) {
	current, _ := platformauth.UserFromContext(r.Context())
	if a.gestion == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	detalle, err := a.gestion.ObtenerProducto(r.Context(), current.EmpresaID, r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, productoDetalleResponse(detalle))
}

func (a *API) actualizarProducto(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Descripcion *string `json:"descripcion"`
		Unidad      *string `json:"unidad"`
		Precio      *string `json:"precio"`
		Afectacion  *string `json:"afectacion"`
		Activo      *bool   `json:"activo"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	var precio *decimal.Decimal
	if input.Precio != nil {
		p, err := decimal.NewFromString(*input.Precio)
		if err != nil {
			writeError(w, http.StatusBadRequest, "precio invalido")
			return
		}
		precio = &p
	}
	current, _ := platformauth.UserFromContext(r.Context())
	if a.gestion == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	detalle, err := a.gestion.ActualizarProducto(r.Context(), current.EmpresaID, r.PathValue("id"), gestionapp.ActualizarProductoInput{
		Descripcion: input.Descripcion, Unidad: input.Unidad, Precio: precio,
		Afectacion: input.Afectacion, Activo: input.Activo,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, productoDetalleResponse(detalle))
}

func (a *API) ajustarStockProducto(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Tipo     string `json:"tipo"`
		Cantidad int64  `json:"cantidad"`
		Motivo   string `json:"motivo"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	current, _ := platformauth.UserFromContext(r.Context())
	if a.gestion == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	result, err := a.gestion.AjustarStock(r.Context(), current.EmpresaID, r.PathValue("id"), current.ID, gestionapp.AjustarStockInput{
		Tipo: input.Tipo, Cantidad: input.Cantidad, Motivo: input.Motivo,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"producto_id": result.ProductoID, "cantidad": result.Cantidad, "movimiento_id": result.MovimientoID,
	})
}

func (a *API) listarMovimientosProducto(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := paginacionQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	current, _ := platformauth.UserFromContext(r.Context())
	if a.gestion == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	items, total, err := a.gestion.ListarMovimientos(r.Context(), current.EmpresaID, r.PathValue("id"), limit, offset)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response := make([]any, 0, len(items))
	for _, m := range items {
		response = append(response, stockMovimientoResponse(m))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response, "total": total})
}

func (a *API) dependenciasProducto(w http.ResponseWriter, r *http.Request) {
	current, _ := platformauth.UserFromContext(r.Context())
	if a.gestion == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	deps, err := a.gestion.DependenciasProducto(r.Context(), current.EmpresaID, r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	comprobantes := make([]any, 0, len(deps))
	for _, d := range deps {
		comprobantes = append(comprobantes, map[string]any{
			"id": d.ID, "tipo": d.Tipo, "serie": d.Serie, "numero": d.Numero, "estado": d.Estado,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total_comprobantes": len(deps),
		"comprobantes":       comprobantes,
	})
}

func (a *API) eliminarProducto(w http.ResponseWriter, r *http.Request) {
	current, _ := platformauth.UserFromContext(r.Context())
	if a.gestion == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	if err := a.gestion.EliminarProducto(r.Context(), current.EmpresaID, r.PathValue("id")); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func listarProductosFilter(r *http.Request) (gestionapp.ListarProductosFilter, error) {
	query := r.URL.Query()
	filtro := gestionapp.ListarProductosFilter{Q: query.Get("q")}
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

func paginacionQuery(r *http.Request) (limit, offset int, err error) {
	query := r.URL.Query()
	if value := query.Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit < 1 {
			return 0, 0, errors.New("limit invalido")
		}
	}
	if value := query.Get("offset"); value != "" {
		offset, err = strconv.Atoi(value)
		if err != nil || offset < 0 {
			return 0, 0, errors.New("offset invalido")
		}
	}
	return limit, offset, nil
}

func productoResponse(p gestiondomain.Producto) map[string]any {
	return map[string]any{
		"id": p.ID, "empresa_id": p.EmpresaID, "codigo": p.Codigo, "descripcion": p.Descripcion,
		"unidad": p.Unidad, "precio": p.Precio.StringFixed(2), "afectacion": p.Afectacion,
		"controla_stock": p.ControlaStock, "activo": p.Activo,
		"creado_en": p.CreadoEn, "actualizado_en": p.ActualizadoEn,
	}
}

func productoDetalleResponse(d gestionapp.ProductoDetalle) map[string]any {
	resp := productoResponse(d.Producto)
	if d.Stock != nil {
		resp["stock"] = *d.Stock
	}
	return resp
}

func stockMovimientoResponse(m gestiondomain.StockMovimiento) map[string]any {
	resp := map[string]any{
		"id": m.ID, "empresa_id": m.EmpresaID, "producto_id": m.ProductoID,
		"tipo": m.Tipo, "cantidad": m.Cantidad, "motivo": m.Motivo, "creado_en": m.CreadoEn,
	}
	if m.UsuarioID != nil {
		resp["usuario_id"] = *m.UsuarioID
	}
	return resp
}

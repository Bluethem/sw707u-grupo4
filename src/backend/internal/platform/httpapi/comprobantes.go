package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	factapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/app"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	platformauth "github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/auth"
	"github.com/shopspring/decimal"
)

type comprobanteInput struct {
	Tipo    domain.TipoComprobante `json:"tipo"`
	Serie   string                 `json:"serie"`
	Cliente *clienteInput          `json:"cliente"`
	Lineas  []lineaInput           `json:"lineas"`
}

type clienteInput struct {
	TipoDoc     string `json:"tipo_doc"`
	Numero      string `json:"numero"`
	RazonSocial string `json:"razon_social"`
}

type lineaInput struct {
	ProductoID     string            `json:"producto_id"`
	Cantidad       int64             `json:"cantidad"`
	PrecioUnitario string            `json:"precio_unitario"`
	Afectacion     domain.Afectacion `json:"afectacion"`
}

func (a *API) emitirComprobante(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		writeError(w, http.StatusBadRequest, "Idempotency-Key requerido")
		return
	}
	var input comprobanteInput
	if !decodeJSON(w, r, &input) {
		return
	}
	cmd, err := comandoDesdeInput(input, key)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, _ := platformauth.UserFromContext(r.Context())
	cmd.EmpresaID = user.EmpresaID
	if a.emisor == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	comprobante, created, err := a.emisor.Emitir(r.Context(), cmd)
	if err != nil {
		writeFacturacionError(w, err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, comprobanteResponse(comprobante))
}

func comandoDesdeInput(input comprobanteInput, key string) (factapp.EmitirComando, error) {
	if input.Tipo != domain.TipoBoleta && input.Tipo != domain.TipoFactura {
		return factapp.EmitirComando{}, errors.New("tipo comprobante invalido")
	}
	if strings.TrimSpace(input.Serie) == "" {
		return factapp.EmitirComando{}, errors.New("serie requerida")
	}
	if len(input.Lineas) == 0 {
		return factapp.EmitirComando{}, domain.ErrLineasVacias
	}
	cmd := factapp.EmitirComando{
		Tipo: input.Tipo, Serie: strings.TrimSpace(input.Serie),
		IdempotencyKey: key, Lineas: make([]factapp.EmitirLinea, 0, len(input.Lineas)),
	}
	if input.Cliente != nil {
		cmd.Cliente = &domain.DatosCliente{
			TipoDoc: input.Cliente.TipoDoc, Numero: input.Cliente.Numero,
			RazonSocial: input.Cliente.RazonSocial,
		}
	}
	for _, inputLinea := range input.Lineas {
		precio, err := decimal.NewFromString(inputLinea.PrecioUnitario)
		if err != nil {
			return factapp.EmitirComando{}, errors.New("precio_unitario invalido")
		}
		if strings.TrimSpace(inputLinea.ProductoID) == "" {
			return factapp.EmitirComando{}, errors.New("producto_id requerido")
		}
		if inputLinea.Cantidad <= 0 {
			return factapp.EmitirComando{}, errors.New("cantidad debe ser > 0")
		}
		if !precio.GreaterThan(decimal.Zero) {
			return factapp.EmitirComando{}, errors.New("precio debe ser > 0")
		}
		if inputLinea.Afectacion != domain.AfectacionGravado &&
			inputLinea.Afectacion != domain.AfectacionExonerado &&
			inputLinea.Afectacion != domain.AfectacionInafecto {
			return factapp.EmitirComando{}, errors.New("afectacion invalida")
		}
		cmd.Lineas = append(cmd.Lineas, factapp.EmitirLinea{
			ProductoID: strings.TrimSpace(inputLinea.ProductoID), Cantidad: inputLinea.Cantidad,
			PrecioUnitario: precio, Afectacion: inputLinea.Afectacion,
		})
	}
	return cmd, nil
}

func (a *API) obtenerComprobante(w http.ResponseWriter, r *http.Request) {
	user, _ := platformauth.UserFromContext(r.Context())
	if a.query == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	comprobante, err := a.query.GetByID(r.Context(), user.EmpresaID, r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	if comprobante == nil {
		writeError(w, http.StatusNotFound, "comprobante no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, comprobanteResponse(comprobante))
}

func (a *API) reenviarComprobante(w http.ResponseWriter, r *http.Request) {
	user, _ := platformauth.UserFromContext(r.Context())
	if a.comprobantes == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	c, err := a.comprobantes.ReenviarComprobante(r.Context(), user.EmpresaID, r.PathValue("id"))
	if err != nil {
		writeFacturacionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, comprobanteResponse(c))
}

func (a *API) solicitarBajaComprobante(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Motivo string `json:"motivo"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	user, _ := platformauth.UserFromContext(r.Context())
	if a.comprobantes == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	c, err := a.comprobantes.SolicitarBaja(r.Context(), user.EmpresaID, r.PathValue("id"), input.Motivo)
	if err != nil {
		writeFacturacionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, comprobanteResponse(c))
}

func (a *API) listarComprobantes(w http.ResponseWriter, r *http.Request) {
	filtro, err := listFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, _ := platformauth.UserFromContext(r.Context())
	if a.query == nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	items, total, err := a.query.List(r.Context(), user.EmpresaID, filtro)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	response := make([]any, 0, len(items))
	for i := range items {
		response = append(response, comprobanteResponse(&items[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response, "total": total})
}

func listFilter(r *http.Request) (ports.ListFilter, error) {
	query := r.URL.Query()
	filtro := ports.ListFilter{Tipo: query.Get("tipo"), Estado: query.Get("estado")}
	var err error
	if value := query.Get("fecha_desde"); value != "" {
		fecha, parseErr := time.Parse(time.DateOnly, value)
		if parseErr != nil {
			return filtro, errors.New("fecha_desde invalida")
		}
		filtro.FechaDesde = &fecha
	}
	if value := query.Get("fecha_hasta"); value != "" {
		fecha, parseErr := time.Parse(time.DateOnly, value)
		if parseErr != nil {
			return filtro, errors.New("fecha_hasta invalida")
		}
		filtro.FechaHasta = &fecha
	}
	if filtro.FechaDesde != nil && filtro.FechaHasta != nil {
		dias := filtro.FechaHasta.Sub(*filtro.FechaDesde) / (24 * time.Hour)
		if dias > 90 {
			return filtro, errors.New("rango de fechas excede 90 dias")
		}
	}
	filtro.Serie = strings.ToUpper(strings.TrimSpace(query.Get("serie")))
	if value := strings.TrimSpace(query.Get("numero")); value != "" {
		numero, parseErr := strconv.ParseInt(value, 10, 64)
		if parseErr != nil || numero < 1 {
			return filtro, errors.New("numero invalido")
		}
		filtro.Numero = &numero
	}
	filtro.Q = strings.TrimSpace(query.Get("q"))
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
	if filtro.Tipo != "" && filtro.Tipo != string(domain.TipoBoleta) && filtro.Tipo != string(domain.TipoFactura) {
		return filtro, errors.New("tipo invalido")
	}
	if filtro.Estado != "" && !estadoValido(filtro.Estado) {
		return filtro, errors.New("estado invalido")
	}
	return filtro, nil
}

func estadoValido(estado string) bool {
	switch domain.EstadoComprobante(estado) {
	case domain.EstadoPendiente, domain.EstadoEnviado, domain.EstadoAceptado,
		domain.EstadoObservado, domain.EstadoRechazado, domain.EstadoAnulada,
		domain.EstadoPendienteBaja:
		return true
	default:
		return false
	}
}

func comprobanteResponse(c *domain.Comprobante) map[string]any {
	response := map[string]any{
		"id": c.ID, "tipo": c.Tipo, "serie": c.Serie, "numero": c.Numero,
		"estado": c.Estado, "cliente": nil, "valor_venta": c.ValorVenta,
		"igv": c.IGV, "total": c.Total, "idempotency_key": c.IdempotencyKey,
		"baja_motivo": c.BajaMotivo,
	}
	if c.Cliente != nil {
		response["cliente"] = map[string]string{
			"tipo_doc": c.Cliente.TipoDoc, "numero": c.Cliente.Numero,
			"razon_social": c.Cliente.RazonSocial,
		}
	}
	lineas := make([]map[string]any, 0, len(c.Lineas))
	for _, linea := range c.Lineas {
		lineas = append(lineas, map[string]any{
			"producto_id": linea.ProductoID, "cantidad": linea.Cantidad,
			"precio_unitario": linea.PrecioUnit, "afectacion": linea.Afectacion,
			"valor_venta": linea.ValorVenta, "igv": linea.IGV, "total": linea.Total,
		})
	}
	response["lineas"] = lineas
	return response
}

func writeFacturacionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrStockInsuficiente):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrComprobanteNoEncontrado):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrOperacionNoPermitida):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrMotivoBajaInvalido):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrResumenNoEncontrado):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrResumenSinBoletas):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrFechaResumenInvalida):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrSerieNoEncontrada):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrClienteRequerido),
		errors.Is(err, domain.ErrRUCInvalido),
		errors.Is(err, domain.ErrLineasVacias),
		errors.Is(err, domain.ErrIdempotencyKeyVacia),
		errors.Is(err, domain.ErrProductoNoEncontrado),
		errors.Is(err, domain.ErrProductoInactivo),
		strings.Contains(err.Error(), "DNI invalido"),
		strings.Contains(err.Error(), "tipo comprobante invalido"),
		strings.Contains(err.Error(), "afectacion invalida"),
		strings.Contains(err.Error(), "cantidad debe ser > 0"),
		strings.Contains(err.Error(), "precio debe ser > 0"):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "error interno")
	}
}

package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

// ──────────────────────────────────────────────────────────────────────────────
// Tipos públicos
// ──────────────────────────────────────────────────────────────────────────────

// FilaProducto representa una fila procesada del Excel, lista para previsualización
// y edición en el frontend antes de confirmar la inserción en BD.
type FilaProducto struct {
	Codigo        string `json:"codigo"`
	Descripcion   string `json:"descripcion"`
	Unidad        string `json:"unidad"`
	Precio        string `json:"precio"`       // string decimal para facilitar edición
	Afectacion    string `json:"afectacion"`
	ControlaStock bool   `json:"controla_stock"`
	StockInicial  *int64 `json:"stock_inicial"` // nil si controla_stock=false
}

// ErrorCargaMasiva representa un error de validación de una fila.
type ErrorCargaMasiva struct {
	Fila    int    `json:"fila"`
	Columna string `json:"columna"`
	Mensaje string `json:"mensaje"`
}

// ResultadoCargaMasiva es el resultado de una inserción confirmada.
type ResultadoCargaMasiva struct {
	Cantidad int
	Errores  []ErrorCargaMasiva
}

// ──────────────────────────────────────────────────────────────────────────────
// Constantes internas
// ──────────────────────────────────────────────────────────────────────────────

type columnaRequerida struct {
	normalizada string
	canonica    string
}

var columnasRequeridas = []columnaRequerida{
	{"codigo", "codigo"},
	{"descripcion", "descripcion"},
	{"unidad", "unidad"},
	{"precio", "precio"},
	{"afectacion", "afectacion"},
	{"controlastock", "controla_stock"},
	{"stockinicial", "stock_inicial"},
}

func normalizeColHeader(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

// ──────────────────────────────────────────────────────────────────────────────
// PreviewCargaMasiva — parsea y valida el Excel SIN insertar en BD.
// Retorna filas si todo es válido, o lista de errores si hay problemas.
// ──────────────────────────────────────────────────────────────────────────────

func (s *Service) PreviewCargaMasiva(ctx context.Context, empresaID string, data []byte) ([]FilaProducto, []ErrorCargaMasiva, error) {
	if empresaID == "" {
		return nil, nil, fmt.Errorf("%w: empresa requerida", domain.ErrInvalido)
	}

	cfg, err := s.configCatalogo.GetByEmpresaID(ctx, empresaID)
	if err != nil {
		return nil, nil, err
	}

	// ── Abrir Excel ───────────────────────────────────────────────────────
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: archivo Excel invalido: %s", domain.ErrInvalido, err.Error())
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, fmt.Errorf("%w: el archivo no tiene hojas", domain.ErrInvalido)
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, nil, fmt.Errorf("%w: error leyendo la hoja", domain.ErrInvalido)
	}
	if len(rows) == 0 {
		return nil, nil, fmt.Errorf("%w: el archivo esta vacio", domain.ErrInvalido)
	}

	// ── Validación estructural: columnas presentes ─────────────────────
	header := rows[0]
	colIdx := make(map[string]int, len(header))
	for i, col := range header {
		colIdx[normalizeColHeader(col)] = i
	}
	for _, col := range columnasRequeridas {
		if _, ok := colIdx[col.normalizada]; !ok {
			return nil, nil, fmt.Errorf("%w: columna requerida faltante: %q", domain.ErrInvalido, col.canonica)
		}
	}

	getCel := func(row []string, normalizada string) string {
		idx, ok := colIdx[normalizada]
		if !ok || idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}

	dataRows := rows[1:]
	if len(dataRows) == 0 {
		return nil, nil, fmt.Errorf("%w: el archivo no tiene filas de datos", domain.ErrInvalido)
	}
	if len(dataRows) > cfg.CargaMaxFilas {
		return nil, nil, fmt.Errorf("%w: el archivo tiene %d filas, el maximo permitido es %d",
			domain.ErrInvalido, len(dataRows), cfg.CargaMaxFilas)
	}

	// ── Validación fila a fila ─────────────────────────────────────────
	var errores []ErrorCargaMasiva
	filas := make([]FilaProducto, 0, len(dataRows))
	codigosVistos := make(map[string]int, len(dataRows))

	for i, row := range dataRows {
		rowNum := i + 2 // fila Excel 1-indexada (cabecera en 1, datos desde 2)
		var fila FilaProducto
		var rowErrors []ErrorCargaMasiva

		addErr := func(col, msg string) {
			rowErrors = append(rowErrors, ErrorCargaMasiva{Fila: rowNum, Columna: col, Mensaje: msg})
		}

		// Codigo
		fila.Codigo = getCel(row, "codigo")
		if fila.Codigo == "" {
			addErr("codigo", "campo requerido")
		} else {
			upper := strings.ToUpper(fila.Codigo)
			if prevRow, seen := codigosVistos[upper]; seen {
				addErr("codigo", fmt.Sprintf("codigo duplicado, ya aparece en fila %d", prevRow))
			} else {
				codigosVistos[upper] = rowNum
			}
		}

		// Descripcion
		fila.Descripcion = getCel(row, "descripcion")
		if fila.Descripcion == "" {
			addErr("descripcion", "campo requerido")
		}

		// Unidad
		rawUnidad := getCel(row, "unidad")
		if rawUnidad == "" {
			addErr("unidad", "campo requerido")
		} else {
			fila.Unidad = strings.ToUpper(rawUnidad)
			if !cfg.UnidadPermitida(fila.Unidad) {
				addErr("unidad", fmt.Sprintf("unidad %q no permitida (permitidas: %s)",
					fila.Unidad, strings.Join(cfg.UnidadesPermitidas, ", ")))
			}
		}

		// Precio
		precioStr := getCel(row, "precio")
		if precioStr == "" {
			addErr("precio", "campo requerido")
		} else {
			precio, parseErr := decimal.NewFromString(precioStr)
			if parseErr != nil {
				addErr("precio", "debe ser un numero decimal valido (ej: 10.50)")
			} else {
				fila.Precio = precio.StringFixed(2)
				if rangoErr := cfg.PrecioEnRango(precio); rangoErr != nil {
					addErr("precio", strings.TrimPrefix(rangoErr.Error(), "invalido: "))
				}
			}
		}

		// Afectacion
		rawAfect := getCel(row, "afectacion")
		if rawAfect == "" {
			addErr("afectacion", "campo requerido")
		} else {
			fila.Afectacion = strings.ToLower(rawAfect)
			if !cfg.AfectacionPermitida(fila.Afectacion) {
				addErr("afectacion", fmt.Sprintf("afectacion %q no permitida (permitidas: %s)",
					fila.Afectacion, strings.Join(cfg.AfectacionesPermitidas, ", ")))
			}
		}

		// ControlaStock
		controlaStr := strings.ToUpper(strings.TrimSpace(getCel(row, "controlastock")))
		switch controlaStr {
		case "SI", "SÍ", "S", "TRUE", "1", "YES", "Y":
			fila.ControlaStock = true
		case "NO", "N", "FALSE", "0":
			fila.ControlaStock = false
		case "":
			addErr("controla_stock", "campo requerido (SI o NO)")
		default:
			addErr("controla_stock", fmt.Sprintf("valor %q no valido, use SI o NO", controlaStr))
		}

		// StockInicial
		stockStr := getCel(row, "stockinicial")
		if fila.ControlaStock && controlaStr != "" {
			if stockStr == "" {
				addErr("stock_inicial", "requerido cuando controla_stock es SI")
			} else {
				stockVal, parseErr := strconv.ParseInt(stockStr, 10, 64)
				if parseErr != nil {
					// Puede venir como float del Excel (ej: "100.0")
					f64, floatErr := strconv.ParseFloat(stockStr, 64)
					if floatErr != nil {
						addErr("stock_inicial", "debe ser un numero entero (ej: 100)")
					} else {
						v := int64(f64)
						if v < 0 {
							addErr("stock_inicial", "debe ser mayor o igual a 0")
						} else {
							fila.StockInicial = &v
						}
					}
				} else if stockVal < 0 {
					addErr("stock_inicial", "debe ser mayor o igual a 0")
				} else {
					fila.StockInicial = &stockVal
				}
			}
		}

		if len(rowErrors) > 0 {
			errores = append(errores, rowErrors...)
		} else {
			filas = append(filas, fila)
		}
	}

	if len(errores) > 0 {
		return nil, errores, nil
	}
	return filas, nil, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// ConfirmarCargaMasiva — re-valida filas editadas por el usuario y las inserta.
// Estrategia all-or-nothing: si hay errores de validación no se inserta nada.
// ──────────────────────────────────────────────────────────────────────────────

func (s *Service) ConfirmarCargaMasiva(ctx context.Context, empresaID string, filas []FilaProducto) (ResultadoCargaMasiva, error) {
	if empresaID == "" {
		return ResultadoCargaMasiva{}, fmt.Errorf("%w: empresa requerida", domain.ErrInvalido)
	}
	if len(filas) == 0 {
		return ResultadoCargaMasiva{}, fmt.Errorf("%w: no hay filas para insertar", domain.ErrInvalido)
	}

	cfg, err := s.configCatalogo.GetByEmpresaID(ctx, empresaID)
	if err != nil {
		return ResultadoCargaMasiva{}, err
	}

	var errores []ErrorCargaMasiva
	codigosVistos := make(map[string]int, len(filas))

	type filaParseada struct {
		Codigo        string
		Descripcion   string
		Unidad        string
		Precio        decimal.Decimal
		Afectacion    string
		ControlaStock bool
		StockInicial  *int64
	}
	parsed := make([]filaParseada, 0, len(filas))

	for i, fila := range filas {
		rowNum := i + 1
		var rowErrors []ErrorCargaMasiva

		addErr := func(col, msg string) {
			rowErrors = append(rowErrors, ErrorCargaMasiva{Fila: rowNum, Columna: col, Mensaje: msg})
		}

		fp := filaParseada{
			Codigo:        strings.TrimSpace(fila.Codigo),
			Descripcion:   strings.TrimSpace(fila.Descripcion),
			Unidad:        strings.ToUpper(strings.TrimSpace(fila.Unidad)),
			Afectacion:    strings.ToLower(strings.TrimSpace(fila.Afectacion)),
			ControlaStock: fila.ControlaStock,
			StockInicial:  fila.StockInicial,
		}

		if fp.Codigo == "" {
			addErr("codigo", "campo requerido")
		} else {
			upper := strings.ToUpper(fp.Codigo)
			if prevRow, seen := codigosVistos[upper]; seen {
				addErr("codigo", fmt.Sprintf("codigo duplicado, ya aparece en fila %d", prevRow))
			} else {
				codigosVistos[upper] = rowNum
			}
		}
		if fp.Descripcion == "" {
			addErr("descripcion", "campo requerido")
		}
		if fp.Unidad == "" {
			addErr("unidad", "campo requerido")
		} else if !cfg.UnidadPermitida(fp.Unidad) {
			addErr("unidad", fmt.Sprintf("unidad %q no permitida", fp.Unidad))
		}
		precioStr := strings.TrimSpace(fila.Precio)
		if precioStr == "" {
			addErr("precio", "campo requerido")
		} else {
			precio, parseErr := decimal.NewFromString(precioStr)
			if parseErr != nil {
				addErr("precio", "debe ser un numero decimal valido (ej: 10.50)")
			} else {
				fp.Precio = precio
				if rangoErr := cfg.PrecioEnRango(precio); rangoErr != nil {
					addErr("precio", strings.TrimPrefix(rangoErr.Error(), "invalido: "))
				}
			}
		}
		if fp.Afectacion == "" {
			addErr("afectacion", "campo requerido")
		} else if !cfg.AfectacionPermitida(fp.Afectacion) {
			addErr("afectacion", fmt.Sprintf("afectacion %q no permitida", fp.Afectacion))
		}
		if fp.ControlaStock && fp.StockInicial == nil {
			addErr("stock_inicial", "requerido cuando controla_stock es true")
		}
		if fp.StockInicial != nil && *fp.StockInicial < 0 {
			addErr("stock_inicial", "debe ser mayor o igual a 0")
		}

		if len(rowErrors) > 0 {
			errores = append(errores, rowErrors...)
		} else {
			parsed = append(parsed, fp)
		}
	}

	if len(errores) > 0 {
		return ResultadoCargaMasiva{Errores: errores}, nil
	}

	// ── Insertar todos ─────────────────────────────────────────────────
	for _, fp := range parsed {
		_, err := s.CrearProducto(ctx, empresaID, CrearProductoInput{
			Codigo:        fp.Codigo,
			Descripcion:   fp.Descripcion,
			Unidad:        fp.Unidad,
			Precio:        fp.Precio,
			Afectacion:    fp.Afectacion,
			ControlaStock: fp.ControlaStock,
			StockInicial:  fp.StockInicial,
		})
		if err != nil {
			if errors.Is(err, domain.ErrConflicto) {
				return ResultadoCargaMasiva{
					Errores: []ErrorCargaMasiva{
						{Fila: 0, Columna: "codigo", Mensaje: fmt.Sprintf("el codigo %q ya existe en el catalogo", fp.Codigo)},
					},
				}, nil
			}
			return ResultadoCargaMasiva{}, err
		}
	}

	return ResultadoCargaMasiva{Cantidad: len(parsed)}, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// CargaMasivaProductos — endpoint legacy: preview + confirmar en un solo paso.
// Se mantiene para compatibilidad con uso directo de la API.
// ──────────────────────────────────────────────────────────────────────────────

func (s *Service) CargaMasivaProductos(ctx context.Context, empresaID string, data []byte) (ResultadoCargaMasiva, error) {
	filas, errores, err := s.PreviewCargaMasiva(ctx, empresaID, data)
	if err != nil {
		return ResultadoCargaMasiva{}, err
	}
	if len(errores) > 0 {
		return ResultadoCargaMasiva{Errores: errores}, nil
	}
	return s.ConfirmarCargaMasiva(ctx, empresaID, filas)
}

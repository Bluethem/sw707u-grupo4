# Changelog

Todos los cambios notables de FlowFacture se documentan aquí.

El formato sigue [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/),
y el proyecto adhiere a [Versionado Semántico](https://semver.org/lang/es/) de forma informal por hitos de módulo (M1–M9).

## [0.9.0] — 2026-10-23 — M9 Mejoras catálogo

### Added

- **Ordenamiento de columnas** en la tabla de productos admin (`ProductosAdmin.tsx`): clic en cabecera ordena asc/desc; indicador visual ↑/↓ en columna activa. Columnas: Código, Descripción, Precio, Afectación, Stock, Activo.
- **Selector de orden** en el catálogo POS (`ProductCatalog.tsx`): Nombre A→Z (default), Nombre Z→A, Precio ↑, Precio ↓, Código A→Z. El atajo `Enter` respeta el orden seleccionado.
- **Eliminación definitiva de productos** con chequeo previo de dependencias:
  - `GET /api/v1/productos/{id}/dependencias` — lista comprobantes que referencian el producto (admin).
  - `DELETE /api/v1/productos/{id}` — borra producto, `productos_stock` y `stock_movimientos` en transacción (admin).
  - `ProductoEliminarModal.tsx` — modal con carga de dependencias, advertencia amarilla si tiene comprobantes y confirmación con botón rojo.

### Fixed

- **CORS:** `DELETE` no estaba incluido en `Access-Control-Allow-Methods`; el preflight bloqueaba la eliminación desde el browser (`cors.go`).
- **Modal eliminación:** estado `deleting` no se reseteaba al abrir el modal para un segundo producto, dejando el botón bloqueado en "Eliminando…" sin haber iniciado la operación (`ProductoEliminarModal.tsx`).

---

## [0.8.0] — 2026-10-22 — M8 Config paramétrica + Carga masiva Excel

### Added

- **Migración `000007_config_catalogo`:** tabla `config_catalogo` con PK `empresa_id`; campos `precio_min`, `precio_max`, `carga_max_filas`, `carga_max_mb`, `unidades_permitidas[]`, `afectaciones_permitidas[]`.
- **Backend config paramétrica:**
  - Dominio `ConfigCatalogo` con `DefaultConfigCatalogo()` y `ValidarConfigCatalogo()`.
  - Adapter Postgres: `GetByEmpresaID` (retorna defaults si no hay fila), `Upsert`.
  - Servicio: `ObtenerConfigCatalogo`, `ActualizarConfigCatalogo`.
  - Endpoints admin: `GET /api/v1/configuracion/catalogo`, `PATCH /api/v1/configuracion/catalogo`.
- **Frontend configuración:** página `/admin/configuracion` con formulario pre-llenado desde la BD; chips seleccionables para unidades y afectaciones; feedback "Guardado correctamente".
- **Backend carga masiva Excel:**
  - Dependencia `github.com/xuri/excelize/v2 v2.8.1`.
  - `POST /api/v1/productos/carga-masiva/preview` — valida el Excel y retorna filas parseadas sin insertar.
  - `POST /api/v1/productos/carga-masiva/confirmar` — inserta las filas confirmadas.
  - Validaciones: columnas requeridas, tipos, `controla_stock` (SI/NO/1/0), unidad y afectación contra `config_catalogo`, precio en rango, sin duplicados en el archivo ni en BD.
  - Estrategia all-or-nothing: si hay errores, retorna `{"errores":[{"fila":N,"columna":"…","mensaje":"…"}]}` sin insertar nada.
- **Frontend carga masiva:** `CargaMasivaModal.tsx` con flujo de dos pasos: previsualización editable de filas (tabla con celdas editables, filas eliminables) → confirmación de inserción.
- Archivos Excel de demo en `docs/3/`: `catalogo_pos_demo.xlsx` y `productos_testing.xlsx`.

---

## [Unreleased — promovido desde borrador]

### Added

- **M7** documentos, consulta pública y reporte de ventas.
  - Almacenamiento FS (`BLOB_DIR`), UBL/PDF con QR (`FRONTEND_PUBLIC_URL`), descargas autenticadas PDF/XML/CDR y regeneración si falta blob.
  - Listado ampliado: `serie`, `numero`, `q`, tope 90 días en fechas.
  - API pública: `POST /api/v1/publico/consulta`, descargas PDF/XML con token Redis (15 min, rate limit por IP).
  - Reporte admin: `GET /api/v1/reportes/ventas` (JSON) y `ventas.csv`.
  - UI: filtros en `/comprobantes`, portal `/consulta`, `/admin/reportes` con CSV y gráficos SVG.
  - E2E `TestE2E_Documentos_EmitirDescargarYConsultar`.
- **M6+** reenvío manual, baja de factura/boleta y resumen diario de boletas (Fake SUNAT).
  - API admin: `POST /comprobantes/{id}/reenviar`, `POST /comprobantes/{id}/baja`, `POST|GET /resumenes`, `GET /resumenes/{id}`.
  - Migración `000006_m6_plus`: `resumenes_diarios`, `resumen_comprobantes`, `baja_motivo`, outbox con `referencia_id` y `comprobante_id` nullable.
  - Worker despacha `envio_comprobante` \| `baja_factura` \| `resumen_diario`.
  - UI: botones en detalle de comprobante; página `/admin/resumenes`.
  - Tests API/E2E de reenvío, baja y resumen.
- Documentación de contratos HTTP (`docs/2/2-apis.md`), scripts (`scripts/README.md`) y tests (`tests/README.md`).

### Fixed

- Listado de productos incluye `stock` (antes la UI admin mostraba `—` aunque hubiera saldo).
- Baja de boleta y listado de resúmenes dejan de responder 404 cuando la API Docker está rebuildada con M6+.

## [0.6.0] — 2026-10-13 — M6 Worker / outbox

### Added

- Worker que consume `outbox_envios` con `FOR UPDATE SKIP LOCKED`.
- Fake SUNAT (`SUNAT_FAKE_MODE=accept|reject|error`) y actualización de estados de comprobante.
- Migración outbox worker; E2E emitir → aceptar.

## [0.5.0] — 2026-10-15 — Admin UI

### Added

- Pantallas admin: empresa, usuarios, series, clientes, productos (ajustes/movimientos).
- Navegación por rol en el shell frontend.

## [0.4.0] — 2026-10-13 — M4 Frontend caja

### Added

- Login, caja POS (`/caja`), listado y detalle de comprobantes.
- Diseño según `DESIGN.md`; polling de estado mientras `PENDIENTE`/`ENVIADO`.
- Docker Compose con servicio `web` y seed idempotente.

## [0.3.0] — 2026-10-12 — M3 Catálogo e inventario

### Added

- CRUD productos, control de stock, ajustes y movimientos.
- API `/api/v1/productos` (+ ajustes/movimientos); migración `000004`.

## [0.2.0] — 2026-10-12 — M2 Clientes

### Added

- CRUD clientes con validación de documento y soft-deactivate.
- API `/api/v1/clientes`; migración `000003`.

## [0.1.0] — 2026-10-12 — M1 Seguridad + emisión HTTP

### Added

- Auth por sesión Redis, RBAC admin/cajero, empresa, usuarios y series.
- Emisión HTTP `POST /api/v1/comprobantes` con `Idempotency-Key`.
- Dominio de facturación (ciclo 2): boleta/factura, cálculo IGV, stock atómico, outbox inicial.
- Seed `cmd/seed`, migraciones `000001`–`000002`, harness `internal/apitest`.

### Notes

- Integración tributaria real (greenter / SUNAT producción), PDF/XML/CDR públicos (M7), notas de crédito y alertas de plazo quedan fuera de este changelog hasta su hito.

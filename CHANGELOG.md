# Changelog

Todos los cambios notables de FlowFacture se documentan aquí.

El formato sigue [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/),
y el proyecto adhiere a [Versionado Semántico](https://semver.org/lang/es/) de forma informal por hitos de módulo (M1–M7).

## [Unreleased]

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

# Tests

Los tests automatizados viven en `src/backend` (Go). Esta carpeta documenta cómo ejecutarlos; no hay suite separada en `tests/` todavía.

## Requisitos

- Go 1.22+
- Postgres (Compose) y, para `apitest`, Redis
- `DATABASE_URL` apuntando a una base **descartable** (recomendado nombre `*_test` o DB de desarrollo que aceptes truncar)
- Redis DB distinta de `0` para apitest (ej. `/1`)

> **Destructivo:** el harness de `internal/apitest` hace `TRUNCATE … CASCADE` y `FLUSHDB` en Redis. No uses la DB de producción ni Redis `/0` con sesiones reales.

Si el host ya usa `:5432`, publica Postgres en `:55432` (override / `POSTGRES_HOST_PORT`) y usa ese puerto en `DATABASE_URL`.

```bash
docker compose up -d postgres redis
export PATH=$HOME/.local/go-install/go/bin:$PATH   # si aplica
export DATABASE_URL='postgres://ff:ff@127.0.0.1:55432/flowfacture_test?sslmode=disable'
export REDIS_URL='redis://127.0.0.1:6379/1'
cd src/backend
```

Sin `DATABASE_URL`, los tests de integración Postgres se omiten (`skip`); dominio y validaciones puras sí corren.

## Capas

| Capa | Paquetes | Servicios |
|---|---|---|
| Unitarios | `internal/*/domain`, `internal/platform/auth`, `internal/platform/config`, … | Ninguno |
| Integración adapters | `adapters/postgres`, `adapters/sunat`, `adapters/redis` | Postgres / Redis según test |
| App facturación | `internal/facturacion/app` | Postgres (procesador, emitir) |
| API HTTP | `internal/apitest` (`TestAPI_*`) | Postgres + Redis |
| E2E | `internal/apitest` (`TestE2E_*`) | Postgres + Redis (+ procesador en proceso) |

Usar `-p 1` en apitest/adapters para evitar carreras sobre la misma DB.

## Comandos útiles

### Unitarios rápidos

```bash
go test ./internal/platform/auth/ ./internal/facturacion/domain/ ./internal/gestion/domain/ ./internal/platform/config/ -count=1
```

### Integración + facturación + adapters

```bash
go test ./internal/facturacion/... ./adapters/postgres/... ./adapters/sunat/ -count=1 -p 1
```

### API (`TestAPI_*`)

```bash
go test ./internal/apitest/ -run 'TestAPI_' -count=1 -p 1
```

Incluye auth, usuarios, series, clientes, productos, emisión, descargas, consulta pública, reportes, reenvío/baja y resúmenes.

### E2E (`TestE2E_*`)

```bash
go test ./internal/apitest/ -run 'TestE2E_' -count=1 -p 1
```

Ejemplos: flujo multi-rol, catálogo, emisión, worker Fake SUNAT, resumen diario, documentos + consulta pública (`TestE2E_Documentos_EmitirDescargarYConsultar`).

### Suite M7 recomendada

```bash
go test ./internal/apitest/ ./internal/facturacion/... ./adapters/postgres/ ./adapters/storage/ ./adapters/ubl/ ./adapters/pdf/ -count=1 -p 1
go build ./cmd/api/ ./cmd/worker/
cd ../frontend && pnpm exec tsc --noEmit && pnpm build
```

Filtros M6+:

```bash
go test ./internal/apitest/ -run 'Reenviar|Baja|Resumen' -count=1 -p 1
go test ./internal/facturacion/app/ -run 'Resumen|Procesador' -count=1 -p 1
```

## Mapa de archivos (apitest)

| Archivo | Cobertura |
|---|---|
| `auth_api_test.go` | Login / me / logout |
| `usuarios_api_test.go` / `series_api_test.go` | M1 admin |
| `clientes_api_test.go` / `e2e_clientes_test.go` | M2 |
| `productos_api_test.go` / `e2e_catalogo_test.go` | M3 |
| `emitir_api_test.go` / `e2e_emitir_test.go` | M5 emisión HTTP |
| `e2e_worker_test.go` | M6 outbox → Fake SUNAT |
| `reenviar_baja_api_test.go` | M6+ reenvío y bajas |
| `resumen_api_test.go` / `e2e_resumen_test.go` | M6+ resúmenes |
| `descarga_api_test.go` | M7 descargas PDF/XML/CDR |
| `consulta_publica_api_test.go` | M7 consulta pública + token |
| `listado_api_test.go` | M7 filtros serie/numero/q y 90 días |
| `reporte_api_test.go` | M7 reporte ventas JSON/CSV |
| `e2e_documentos_test.go` | M7 E2E emitir → CDR → descargas → consulta |
| `e2e_flujo_test.go` | Flujo transversal multi-empresa |

## Frontend

No hay suite E2E de browser en el repo. Chequeo estático:

```bash
cd src/frontend
pnpm exec tsc --noEmit
pnpm build
```

## Detalle adicional

Ver también `src/backend/README.md` (seed, emisión curl, worker) y contratos en `docs/2/2-apis.md`.

# GitHub Actions

Dentro de la carpeta `.github/worflows` se tiene suites que se corre dentor de github actions para automatizar procesos CI/CD.

| Archivo | Cobertura |
|---|---|
| `test-api-backend` | Test de apis individuales |
| `test-e2e-backend` | Test e2e del backend |
| `test-unit-backend` | Test unitarios |

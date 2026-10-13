# Backend FlowFacture

## Seed local M1

El seed ejecuta las migraciones y crea una empresa, un administrador y las
series `B001`/`F001`. Requiere Postgres; no requiere Redis.

```bash
export DATABASE_URL='postgres://ff:ff@127.0.0.1:5432/flowfacture?sslmode=disable'
go run ./cmd/seed
```

Por defecto crea `admin@flowfacture.local` con contraseña `Admin123!` y muestra
las credenciales al terminar. Se pueden cambiar con `SEED_RUC`,
`SEED_RAZON_SOCIAL`, `SEED_ADMIN_EMAIL` y `SEED_ADMIN_PASSWORD`. El comando es
para una base nueva; una segunda ejecución falla para no sobrescribir datos.

## Tests M1: seguridad y API

Los tests unitarios no necesitan servicios:

```bash
go test ./internal/platform/auth/ ./internal/facturacion/domain/ ./internal/platform/config/ -count=1
```

Los tests API y E2E requieren Postgres y Redis:

```bash
docker compose up -d postgres redis
export DATABASE_URL='postgres://ff:ff@127.0.0.1:5432/flowfacture_test?sslmode=disable'
export REDIS_URL='redis://127.0.0.1:6379/1'
go test ./internal/apitest/ -run 'TestAPI_' -count=1
go test ./internal/apitest/ -run 'TestE2E_' -count=1
```

> **ADVERTENCIA: tests destructivos.** El harness ejecuta `TRUNCATE ... CASCADE`
> en Postgres y `FLUSHDB` en Redis. Usa una base Postgres descartable terminada
> en `_test` y una DB Redis distinta de `0` (por ejemplo `/1`).

## Tests de facturación (ciclo 2)

Requiere Postgres accesible (`docker compose up postgres -d`).

> **ADVERTENCIA: tests destructivos.** Los tests de integración ejecutan
> `TRUNCATE ... CASCADE` sobre las tablas de facturación. Usa una base de datos
> dedicada y descartable (se recomienda que `DATABASE_URL` apunte a una base
> cuyo nombre termine en `_test`) o acepta que sus datos serán eliminados.

Si el puerto **5432** ya está en uso en tu máquina, define `POSTGRES_HOST_PORT=55432` en `.env` (ver `.env.example`) y usa ese puerto en `DATABASE_URL`.

```bash
export DATABASE_URL='postgres://ff:ff@127.0.0.1:5432/flowfacture_test?sslmode=disable'
cd src/backend
go test ./internal/facturacion/... ./adapters/postgres/... -count=1
```

Sin `DATABASE_URL`, los tests de integración se omiten (skip); los de dominio y las validaciones de `Emitir` en paquete `app` siempre corren.

El DoD del ciclo 2 exige ejecutar la suite completa con base de datos.

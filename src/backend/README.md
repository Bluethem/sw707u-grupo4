# Backend FlowFacture

## Tests de facturación (ciclo 2)

Requiere Postgres accesible (`docker compose up postgres -d`).

Si el puerto **5432** ya está en uso en tu máquina, define `POSTGRES_HOST_PORT=55432` en `.env` (ver `.env.example`) y usa ese puerto en `DATABASE_URL`.

```bash
export DATABASE_URL='postgres://ff:ff@127.0.0.1:5432/flowfacture?sslmode=disable'
cd src/backend
go test ./internal/facturacion/... ./adapters/postgres/... -count=1
```

Sin `DATABASE_URL`, los tests de integración se omiten (skip); los de dominio y las validaciones de `Emitir` en paquete `app` siempre corren.

El DoD del ciclo 2 exige ejecutar la suite completa con base de datos.

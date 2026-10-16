# Scripts y comandos operativos

Esta carpeta reserva espacio para scripts auxiliares. Hoy el arranque y el seed se hacen con **Docker Compose** y los binarios Go bajo `src/backend/cmd/`.

## Arranque completo (recomendado)

Desde la raíz del repo:

```bash
cp .env.example .env   # si aún no existe
docker compose up -d --build
```

Servicios:

| Servicio | Puerto host | Rol |
|---|---|---|
| `postgres` | (interno; opcional vía override) | Datos + outbox |
| `redis` | `6379` | Sesiones |
| `seed` | — (one-shot) | Migraciones + demo idempotente |
| `api` | `8080` | HTTP API |
| `worker` | — | Outbox → Fake SUNAT |
| `web` | `3000` | Next.js |

- API health: http://127.0.0.1:8080/health  
- Web: http://127.0.0.1:3000  
- Credenciales demo: `admin@flowfacture.local` / `Admin123!`

### Puerto Postgres en el host

Compose **no** publica `5432` por defecto (evita choque con Postgres local). Para tests/seed desde el host, usa `docker-compose.override.yml` o `POSTGRES_HOST_PORT=55432` (ver `.env.example`) y:

```bash
export DATABASE_URL='postgres://ff:ff@127.0.0.1:55432/flowfacture?sslmode=disable'
```

## Seed

### Con Compose

```bash
docker compose run --rm seed
```

Idempotente: si la empresa demo ya existe, sale `0` sin pisar datos.

### En el host

```bash
cd src/backend
export DATABASE_URL='postgres://ff:ff@127.0.0.1:55432/flowfacture?sslmode=disable'
go run ./cmd/seed
```

Variables opcionales: `SEED_RUC`, `SEED_RAZON_SOCIAL`, `SEED_ADMIN_EMAIL`, `SEED_ADMIN_PASSWORD`.

## API / Worker en el host

```bash
cd src/backend
export DATABASE_URL='postgres://ff:ff@127.0.0.1:55432/flowfacture?sslmode=disable'
export REDIS_URL='redis://127.0.0.1:6379/0'
export CORS_ORIGINS='http://localhost:3000'
go run ./cmd/api

# otra terminal
export SUNAT_FAKE_MODE=accept   # accept | reject | error
go run ./cmd/worker
```

Variables del worker: `WORKER_POLL_INTERVAL`, `WORKER_BATCH_SIZE`, `WORKER_MAX_INTENTOS`, `WORKER_BACKOFF_BASE` (detalle en `src/backend/README.md`).

## Frontend en el host

```bash
cd src/frontend
pnpm install
pnpm dev
```

`NEXT_PUBLIC_API_URL` default: `http://localhost:8080`.

## Smoke curl (post-seed)

```bash
export API_BASE='http://127.0.0.1:8080'
TOKEN=$(curl -sS -X POST "$API_BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@flowfacture.local","password":"Admin123!"}' \
  | jq -r .token)

curl -sS "$API_BASE/api/v1/productos?limit=5" -H "Authorization: Bearer $TOKEN"
curl -sS "$API_BASE/api/v1/resumenes" -H "Authorization: Bearer $TOKEN"
```

Contratos completos: [`docs/2/2-apis.md`](../docs/2/2-apis.md).

## Rebuild tras cambios M6+

Si la UI muestra 404 en `/resumenes` o `/baja`, la imagen `api` está vieja:

```bash
docker compose build api worker && docker compose up -d api worker
```

# FlowFacture — Frontend (caja)

App Next.js 15 (App Router) para login, POS (`/caja`) y comprobantes.

## Variables de entorno

| Variable | Descripción |
|---|---|
| `NEXT_PUBLIC_API_URL` | Base de la API Go (default `http://localhost:8080`) |

Copia `.env.example` en la raíz del monorepo si usas Docker Compose; para desarrollo local en el host:

```bash
# src/frontend/.env.local (opcional)
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Arranque con Docker (recomendado)

Desde la raíz del repo. El servicio `seed` corre **una vez** (idempotente) antes de `api`/`worker`:

```bash
docker compose up -d --build
```

- Web: http://localhost:3000  
- API: http://localhost:8080  
- Login: `admin@flowfacture.local` / `Admin123!`  
- Producto demo: `SKU-1`

Si el puerto **5432 del host** ya lo ocupa otro Postgres, en `.env` pon `POSTGRES_HOST_PORT=55432`. Eso no afecta a api/seed dentro de Compose (siguen usando el hostname `postgres`).

Para forzar re-seed tras borrar datos: `docker compose run --rm seed`.

## Arranque local (API en el host)

1. `docker compose up -d postgres redis`
2. Seed (apunta al puerto publicado del contenedor):

   ```bash
   cd src/backend
   export DATABASE_URL='postgres://ff:ff@127.0.0.1:5432/flowfacture?sslmode=disable'
   go run ./cmd/seed
   ```

3. API / worker / web como antes (`go run ./cmd/api`, `pnpm dev`).

Credenciales demo: `admin@flowfacture.local` / `Admin123!`.

## Build

```bash
cd src/frontend
pnpm exec tsc --noEmit
pnpm build
```

## Flujo demo rápido

Login con el admin del seed → `/caja` → agregar productos → emitir boleta/factura → enlace al detalle en `/comprobantes/[id]`. Con el worker corriendo, el detalle puede pasar a `ACEPTADO` (polling cada 3s mientras `PENDIENTE`/`ENVIADO`).

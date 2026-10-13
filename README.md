# FlowFacture

> Sistema de gestión y facturación electronica

<div align="center">
  <img src="/assets/img/flowfacture_logo.png" alt="Logo FactuGest" width="250">
</div>

* **Curso:** Construcción de Software 2
* **Docente:** Anibal Javier Guillen Vasquez
* **Grupo:** 4
* **Alumnos:**
    - David Luza Ccorimanya
    - Henry Javier Medina Malpartida

## Arranque local (ciclo 1)

Requisitos: Docker y Docker Compose.

Con un solo comando levantas **Postgres + API + Worker + Web**:

```bash
cp .env.example .env
docker compose up --build
```

No hace falta inicializar el backend a mano: `api` y `worker` se construyen desde `src/backend` y arrancan dentro de Compose. Docker Compose carga `.env`; el `DATABASE_URL` de ejemplo usa el hostname interno `postgres`.

- Web: http://localhost:3000
- API health: http://localhost:8080/health

### Desarrollo sin Docker (parcial)

Útil si quieres hot-reload del frontend o depurar Go en el host. Postgres puede seguir en Compose (`docker compose up postgres`).

```bash
# Backend (en el host; necesita Go)
cd src/backend
export DATABASE_URL='postgres://ff:ff@localhost:5432/flowfacture?sslmode=disable'
go run ./cmd/api

# Frontend (pnpm)
cd src/frontend
pnpm install
pnpm dev
```

Documentación de producto: `docs/1/`. Arquitectura: `docs/2/`. Diseño de arranque: `docs/superpowers/specs/2026-10-10-flowfacture-monolito-modular-design.md`.

Tests de facturación (ciclo 2): ver [`src/backend/README.md`](src/backend/README.md).

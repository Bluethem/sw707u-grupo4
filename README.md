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

## Arranque

Requisitos: Docker, Docker Compose y pnpm.

Con un dos comando levantas **Postgres + Redis + API + Worker + Seed, a nivel local se levanta separado la web**:

```bash
cp .env.example .env
docker compose up --build
cd src/frontend
pnpm install && pnpm dev
```

No hace falta inicializar el backend a mano: `api` y `worker` se construyen desde `src/backend` y arrancan dentro de Compose. Docker Compose carga `.env`; el `DATABASE_URL` de ejemplo usa el hostname interno `postgres`.

- Web: http://localhost:3000
- API health: http://localhost:8080/health
- Login: `admin@flowfacture.local` / `Admin123!`

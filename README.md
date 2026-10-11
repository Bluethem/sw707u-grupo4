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

```bash
cp .env.example .env
docker compose up --build
```

- Web: http://localhost:3000
- API health: http://localhost:8080/health

### Desarrollo sin Docker (parcial)

```bash
# Backend
cd src/backend
export DATABASE_URL='postgres://ff:ff@localhost:5432/flowfacture?sslmode=disable'
go run ./cmd/api

# Frontend
cd src/frontend
npm install
npm run dev
```

Documentación de producto: `docs/1/`. Arquitectura: `docs/2/`. Diseño de arranque: `docs/superpowers/specs/2026-10-10-flowfacture-monolito-modular-design.md`.

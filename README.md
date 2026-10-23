# FlowFacture

> Sistema de gestión y facturación electrónica para SUNAT — Perú

<div align="center">
  <img src="/assets/img/flowfacture_logo.png" alt="Logo FlowFacture" width="250">
</div>

**Curso:** Construcción de Software II – SW707U
**Docente:** Anibal Javier Guillen Vasquez
**Grupo:** 4
**Integrantes:** David Luza Ccorimanya · Henry Javier Medina Malpartida

---

## Descripción

FlowFacture es un sistema multi-tenant de emisión de comprobantes electrónicos (boletas y facturas) orientado a pequeñas y medianas empresas peruanas. Implementa el flujo completo de facturación electrónica: desde la venta en el POS hasta la transmisión asíncrona a SUNAT, incluyendo control de stock, gestión de catálogo, resúmenes diarios y anulaciones.

**Stack:** Go 1.22 · Next.js 15 · PostgreSQL 16 · Redis 7 · Docker Compose

---

## Arranque rápido

**Requisitos:** Docker, Docker Compose y pnpm.

```bash
# 1. Variables de entorno
cp .env.example .env

# 2. Backend + infraestructura (Postgres, Redis, seed, API, Worker)
docker compose up --build -d

# 3. Frontend (desarrollo)
cd src/frontend
pnpm install && pnpm dev
```

El servicio `seed` ejecuta migraciones y carga datos demo de forma idempotente. No es necesario inicializar la base de datos manualmente.

| Servicio | URL |
|---|---|
| Frontend | http://localhost:3000 |
| API | http://localhost:8080 |
| Health check | http://localhost:8080/health |

---

## Credenciales de demo

| Rol | Email | Contraseña |
|---|---|---|
| Administrador | `admin@flowfacture.local` | `Admin123!` |
| Cajero | `cajero@flowfacture.local` | `Cajero123!` |

---

## Estructura del repositorio

```
sw707u-grupo4/
├── src/
│   ├── backend/          Go 1.22 — API + Worker + Seed
│   │   ├── cmd/
│   │   │   ├── api/      Servidor HTTP (:8080)
│   │   │   ├── worker/   Procesador outbox (SUNAT)
│   │   │   └── seed/     Migraciones + datos demo
│   │   ├── internal/
│   │   │   ├── facturacion/  Bounded context: emisión, SUNAT, CDR
│   │   │   ├── gestion/      Bounded context: clientes, productos, stock
│   │   │   ├── platform/     Auth, HTTP, config, middleware
│   │   │   └── apitest/      Tests de integración y E2E
│   │   └── adapters/
│   │       ├── postgres/     Repositorios PostgreSQL
│   │       ├── redis/        Store de sesiones
│   │       ├── sunat/        Adaptador SUNAT (modo fake disponible)
│   │       ├── pdf/          Generador PDF
│   │       └── firma/        Firma digital XML
│   └── frontend/         Next.js 15 + React 19 + TypeScript
│       ├── app/          App Router (páginas)
│       └── components/   Componentes reutilizables
├── docs/
│   ├── 1/                Descripción, requisitos y casos de uso
│   ├── 2/                Arquitectura C4, ADRs y contratos de API
│   └── 3/                Archivos de demo para carga masiva
├── assets/               Diagramas C4 (DSL Structurizr) e imágenes
├── deployment/           Guía de despliegue en VPS
├── scripts/              Scripts de prueba y utilidades
├── tests/                Documentación de la suite de tests
├── docker-compose.yml
├── .env.example
├── CHANGELOG.md
└── DESIGN.md
```

---

## Módulos implementados

| Módulo | Descripción |
|---|---|
| M1 — Seguridad y Auth | Sesiones Redis, RBAC Admin/Cajero, multi-tenant |
| M2 — Clientes | CRUD con validación DNI/RUC/CE/PAS, baja lógica |
| M3 — Catálogo e Inventario | CRUD productos, stock con auditoría, afectación IGV |
| M4 — Frontend POS | Caja registradora, emisión de comprobantes, polling de estado |
| M5 — Emisión HTTP | API idempotente, XML UBL 2.1, transacción atómica |
| M6 — Worker y Outbox | SUNAT Fake, reintentos con backoff exponencial |
| M6+ — Bajas y Resúmenes | Anulaciones, reenvío manual, resúmenes diarios de boletas |
| M7 — Documentos | Descarga PDF/XML/CDR, consulta pública, reportes de venta |
| M8 — Config paramétrica | Reglas por empresa, carga masiva Excel con previsualización |
| M9 — Mejoras catálogo | Ordenamiento de columnas, eliminación definitiva de productos |

---

## Tests

### Tests unitarios (sin base de datos)

```bash
cd src/backend
go test ./internal/platform/auth/ ./internal/facturacion/domain/ ./internal/platform/config/ -count=1 -v
```

### Tests de integración API (requiere PostgreSQL + Redis)

```bash
cd src/backend
export DATABASE_URL='postgres://ff:ff@127.0.0.1:5432/flowfacture_test?sslmode=disable'
export REDIS_URL='redis://127.0.0.1:6379/1'
go test ./internal/apitest/ -run 'TestAPI_' -count=1 -p 1 -v
```

### Tests E2E (incluye worker)

```bash
go test ./internal/apitest/ -run 'TestE2E_' -count=1 -p 1 -v
```

### Test individual

```bash
go test ./internal/apitest/ -run 'TestAPI_Productos' -count=1 -p 1
```

> Los tests de integración usan `TRUNCATE CASCADE` y `FLUSHDB`. Usar siempre la base `flowfacture_test` (nunca `flowfacture`) y Redis DB `/1`.

### Levantar infraestructura para tests locales

```bash
docker compose up -d postgres redis
cd src/backend && go run ./cmd/seed   # primera vez
```

---

## Variables de entorno clave

| Variable | Default | Descripción |
|---|---|---|
| `DATABASE_URL` | `postgres://ff:ff@postgres:5432/flowfacture` | Conexión PostgreSQL |
| `REDIS_URL` | `redis://redis:6379/0` | Conexión Redis |
| `SUNAT_FAKE_MODE` | `accept` | Simulador SUNAT: `accept` / `reject` / `error` |
| `SECRETS_MASTER_KEY` | *(cambiar)* | Clave AES-GCM 32 bytes para cifrar certificados |
| `CORS_ORIGINS` | `http://localhost:3000` | Orígenes permitidos (separados por coma) |
| `MIGRATE_ON_START` | `true` | Ejecutar migraciones al iniciar la API |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | URL de la API usada por el frontend |

Ver `.env.example` para la lista completa.

---

## Documentación

| Documento | Contenido |
|---|---|
| `docs/1/1-descripcion.md` | Procesos de negocio y actores |
| `docs/1/2-requerimientos.md` | 50 RF + 37 RNF con priorización MoSCoW |
| `docs/1/3-casos-uso.md` | 22 casos de uso detallados |
| `docs/2/1-arquitectura.md` | Arquitectura C4 (contexto, contenedores, componentes, despliegue) |
| `docs/2/2-apis.md` | Contratos de todos los endpoints REST |
| `docs/2/ADRS/` | 5 ADRs de decisiones arquitectónicas |
| `INFORME_GRUPO4.md` | Informe técnico del curso (Etapas I y II) |
| `CHANGELOG.md` | Historial de versiones por módulo |
| `DESIGN.md` | Sistema de diseño visual del frontend |
| `deployment/DEPLOY.md` | Guía de despliegue en VPS con Docker |

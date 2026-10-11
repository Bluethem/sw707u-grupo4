# ADR 001: Selección de tecnología

**Estado:** Accepted  
**Fecha:** 2026-10-10

## Contexto

FlowFacture necesita API síncrona, worker asíncrono para SUNAT, UI de caja/gestión y almacenamiento de archivos, con despliegue reproducible en un VPS.

## Decisión

| Pieza | Tecnología |
|---|---|
| API y Worker | Go |
| Frontend | Next.js (TypeScript, App Router) |
| Base de datos | PostgreSQL |
| Archivos | Cloudflare R2 (API S3); fake/local en desarrollo |
| Despliegue | Docker Compose en VPS |

## Consecuencias

- Un solo lenguaje en backend para API y Worker facilita compartir dominio.
- Next.js cubre UI autenticada y rutas públicas de consulta.
- R2 se abstrae detrás de `StoragePort`; en local se puede usar filesystem o MinIO más adelante.

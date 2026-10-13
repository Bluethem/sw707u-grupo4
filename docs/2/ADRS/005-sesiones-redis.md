# ADR 005: Sesiones de autenticación en Redis

**Estado:** Accepted  
**Fecha:** 2026-10-12

## Contexto

La API debe autenticar peticiones multi-tenant sin acoplar el estado de sesión a instancias concretas de la API. PostgreSQL ya concentra datos de negocio y la cola outbox (ADR 003); usar la misma base para sesiones añadiría carga de lectura/escritura frecuente y limpieza de filas expiradas.

## Decisión

Almacenar sesiones opacas en **Redis 7**:

- Tras login válido (credenciales en Postgres), la API genera un token opaco y guarda `SET session:{token}` con JSON `{user_id, empresa_id, rol}` y TTL configurable (`SESSION_TTL`, por defecto 8 horas).
- El middleware valida `Authorization: Bearer` con `GET`; si falta la clave o expiró, responde 401.
- Logout invalida la sesión con `DEL session:{token}`.
- Los datos de negocio (empresas, usuarios, series, comprobantes) siguen en PostgreSQL. La cola outbox y el patrón de envío a SUNAT **no cambian** (ADR 003).

## Consecuencias

- La API puede escalarse horizontalmente compartiendo el mismo Redis.
- Redis es volátil por diseño: caída o flush implica re-login; aceptable para sesiones de caja.
- Se añade un servicio más en Docker Compose y la variable `REDIS_URL` en configuración.
- TTL nativo evita jobs de limpieza de sesiones en Postgres.

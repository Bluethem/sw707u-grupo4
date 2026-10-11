# ADR 004: Secretos cifrados en base de datos

**Estado:** Accepted  
**Fecha:** 2026-10-10

## Contexto

Certificados digitales y credenciales SOL/OSE deben almacenarse cifrados; un gestor de secretos dedicado es excesivo para el despliegue inicial.

## Decisión

Cifrar en reposo con AES-GCM. La llave maestra vive solo en variable de entorno (`SECRETS_MASTER_KEY`). Nunca commitear secretos ni loguearlos.

## Consecuencias

- Cumple RNF-09 con poca infra.
- Rotación de llave requiere proceso de re-cifrado (fuera de ciclo 1).

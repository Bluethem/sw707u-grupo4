# ADR 003: Cola outbox en PostgreSQL

**Estado:** Accepted  
**Fecha:** 2026-10-10

## Contexto

Los envíos a SUNAT deben sobrevivir reinicios y no bloquear la caja. Un broker adicional (Rabbit/Kafka) aumenta la operación en un VPS pequeño.

## Decisión

Usar patrón outbox en PostgreSQL: el trabajo de envío se inserta en la misma transacción de emisión. El Worker consume con `FOR UPDATE SKIP LOCKED` y aplica backoff en reintentos.

## Consecuencias

- Atomicidad emisión↔encolado.
- Escalado horizontal del Worker sin broker.
- La cola no es tan rica como un message broker; suficiente para el alcance.

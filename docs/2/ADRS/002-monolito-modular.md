# ADR 002: Monolito modular (gestion / facturacion)

**Estado:** Accepted  
**Fecha:** 2026-10-10

## Contexto

Se necesita separar mentalmente la gestión PYME de la integración SUNAT, sin perder atomicidad (stock + correlativo + comprobante + outbox).

## Decisión

Un solo deployable API Go y un Worker, con dos bounded contexts en `internal/gestion` e `internal/facturacion`. Comunicación por puertos internos (sin HTTP). La orquestación de la transacción de emisión vive en `facturacion` (`EmitirComprobante`). El frontend Next.js consume una sola API.

## Consecuencias

- Preserva transacciones ACID y simplifica el curso.
- La frontera de paquetes permite extraer facturación a otro servicio más adelante.
- `facturacion` no importa `gestion`; usa `StockPort` implementado en adapters.

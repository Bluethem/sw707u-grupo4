---
marp: true
theme: default
paginate: true
headingDivider: false
---

# FlowFacture
## Aplicación de temas SW707 (PC3)
Construcción dirigida por datos · Cloud-native · Observabilidad
Grupo 4 — SW707U

---

# Agenda
1. Construcción dirigida por datos
2. Cloud-native / distribuido
3. Observabilidad + resiliencia

**Hilo:** emitir → outbox → worker → SUNAT → estado

---

# Contexto: FlowFacture
| Pieza | Rol |
|------|-----|
| Next.js | Caja / admin |
| API Go | Emisión síncrona |
| Worker Go | Envío asíncrono SUNAT |
| PostgreSQL | Datos + outbox |
| Redis | Sesiones |

---

# FSM del comprobante
Estados en `internal/facturacion/domain/comprobante.go`:

```
PENDIENTE → ENVIADO → ACEPTADO | OBSERVADO | RECHAZADO
                ↘ ANULADA / PENDIENTE_BAJA
```

Los resúmenes (`internal/facturacion/domain/resumen.go`) tienen FSM propia: `PENDIENTE` → `ENVIADO` → `ACEPTADO` | `RECHAZADO`.

---

# Configuración y late binding
`config.Load()` lee el entorno al arrancar: cambiar comportamiento sin recompilar.

- `WORKER_POLL_INTERVAL` — intervalo de sondeo del worker (def. `1s`)
- `WORKER_MAX_INTENTOS` — tope de reintentos (def. `5`)
- `WORKER_BACKOFF_BASE` — base del backoff (def. `1s`)
- `SUNAT_ENV` — `beta` o `production` (def. `beta`)
- `SESSION_TTL` — vida de la sesión Redis (def. `8h`)

---

# SUNAT_FAKE_MODE
Flag de demo y kill-switch ligero del simulador. Evidencia: `adapters/sunat/factory.go` (`NewFromConfig` elige `NewFake` si faltan credenciales).

| Valor | Efecto |
|------|--------|
| `accept` | respuesta aceptada (defecto) |
| `reject` | rechazo simulado |
| `error` | error de transporte; corta el camino feliz sin recompilar |

---

# Monolito modular
ADR 002: monolito modular. Un deployable API (`cmd/api`) y un Worker (`cmd/worker`). No hay microservicios.

| Contexto | Paquete | Qué vive ahí |
|------|-----|-----|
| `gestion` | `internal/gestion` | Catálogo, stock, usuarios |
| `facturacion` | `internal/facturacion` | Emisión, XML, outbox |

La emisión va por el caso de uso `Emisor.Emitir` (`internal/facturacion/app/emitir.go`). Los contextos se hablan por puertos internos (`StockPort`), sin HTTP. `facturacion` no importa `gestion`. Next.js consume una sola API.

---

# Outbox en PostgreSQL
ADR 003. La caja no espera a SUNAT; la cola vive en la misma base.

1. **TX local:** comprobante `PENDIENTE` + fila en `outbox_envios` (`EnqueueEnvio`, misma transacción).
2. **Worker:** reclama con `FOR UPDATE SKIP LOCKED` (`adapters/postgres/outbox_consumer.go`).
3. **SUNAT:** marca `ENVIADO`, envía el XML y recibe el CDR.
4. **Estado:** cierra en `ACEPTADO` o `RECHAZADO` (`PENDIENTE` → `ENVIADO` → `ACEPTADO` | `RECHAZADO`).

Consistencia eventual: la emisión es ACID; el resultado de SUNAT llega después. Sin broker (Rabbit/Kafka). Varios workers no toman el mismo trabajo.

---

# Middleware: puertas de control
`Authenticate` y `RequireRole` en `internal/platform/auth/middleware.go`, cableados en `internal/platform/httpapi/api.go`.

1. **Gate auth.** Bearer → sesión en Redis. Sin sesión, 401. La sesión guarda `empresa_id`.
2. **Gate rol.** `RequireRole` deja catálogo, bajas y reportes solo al administrador.
3. **Tenant.** `WithUser` mete el usuario (y su `empresa_id`) en el `context`. Las consultas filtran por esa empresa.

Puente a observabilidad: ese mismo `context` de la petición es por donde viajaría un trace id. Hoy propaga identidad y tenant.

---

# REST vs gRPC
Decisión consciente. Un solo backend Go; el frontend habla JSON. gRPC no aplica: no hay llamadas service-to-service tipadas.

| | REST/JSON (hoy) | gRPC (no usado) |
|------|-----|-----|
| Contrato | Rutas `/api/v1` | Archivos `.proto` |
| Cliente | Next.js (caja / admin) | Servicios internos |
| Cuándo encaja | Exposición al navegador | Alta velocidad S2S |

REST basta: el único cliente de la API es Next.js. API y Worker no se llaman por red; comparten el dominio en código. Extraer facturación más adelante no obliga a adoptar gRPC.

---

# Observabilidad hoy
**Hoy.** Logs de proceso, un health check y el estado del comprobante a la vista. Sin logs JSON, sin métricas y sin trazas.

| Señal | Qué existe |
|------|-----|
| Logs | `log.Printf` en API y Worker |
| Salud | `GET /health` → `{"status":"ok"}` |
| Dominio | badge de estado en el admin |

- API (`cmd/api/main.go`): `api listening` y `firma=%s` al arrancar.
- Worker (`cmd/worker/main.go`): `worker started (poll=… batch=… max_intentos=…)`, error de `RunOnce`, `worker shutting down`.
- Admin: últimos comprobantes con badge en `components/admin/DashboardAdmin.tsx`; listado en `app/comprobantes/page.tsx`.

`GET /health` (`internal/platform/httpapi/api.go`) no consulta Postgres ni Redis: solo dice que el proceso responde.

---

# Resiliencia del worker
La observabilidad detecta el fallo. La resiliencia lo sobrevive. **Hoy** el worker aguanta con configuración de arranque, no con un circuit breaker.

| Knob | Variable | Default | Código |
|------|----------|---------|--------|
| poll | `WORKER_POLL_INTERVAL` | `1s` | pausa del bucle en `cmd/worker/main.go` |
| batch | `WORKER_BATCH_SIZE` | `10` | `ProcesadorEnvio.RunOnce` |
| max_intentos | `WORKER_MAX_INTENTOS` | `5` | `reintentarOFallar` |
| backoff | `WORKER_BACKOFF_BASE` | `1s` | `CalcularBackoff` |

`CalcularBackoff` (`internal/facturacion/app/procesador_envio.go`) espera `base × 2^intentos`, con tope de 60×. `Retry` deja el job para más tarde; al llegar a `max_intentos`, `Fail`.

Un log `worker run once`, o un comprobante que sigue en `PENDIENTE`, es la señal. El outbox y el backoff son lo que no pierde el envío a SUNAT.

---

# OpenTelemetry y RED
**Próximo paso.** No está implementado: no hay SDK, exporter ni endpoint de métricas. Profiling de runtime tampoco.

Ejemplo conceptual. Dos spans de la misma emisión, unidos por un trace que hoy no existe.

| Span | Ancla en el código | Atributos |
|------|--------------------|-----------|
| `EmitirComprobante` | `Emisor.Emitir` (`internal/facturacion/app/emitir.go`) | `empresa_id`, tipo, monto |
| `EnviarSUNAT` | `ProcesadorEnvio` (`internal/facturacion/app/procesador_envio.go`) | comprobante, intento, resultado |

RED en la API, también **próximo paso**:

- **Rate** — peticiones por segundo a `/api/v1`.
- **Errors** — proporción de 5xx y de emisiones fallidas.
- **Duration** — latencia (p95) de emitir.

**Próximo paso:** la correlación API→Worker persistiría `traceparent` / `trace_id` en `outbox_envios`. No se llaman por HTTP. Hoy ese enlace no existe.

<!-- `EmitirComprobante` es el nombre de ADR 002; el método real es `Emisor.Emitir`. -->

---

# Mapa tema ↔ evidencia

| Tema | Ruta |
|------|------|
| Dominio (FSM) | `internal/facturacion/domain/comprobante.go` |
| Resúmenes | `internal/facturacion/domain/resumen.go` |
| Config | `internal/platform/config/config.go` |
| ADR monolito | `docs/2/ADRS/002-monolito-modular.md` |
| ADR outbox | `docs/2/ADRS/003-outbox-postgres.md` |
| Worker | `cmd/worker/main.go` |
| Reintentos | `internal/facturacion/app/procesador_envio.go` |
| Health | `internal/platform/httpapi/api.go` (`GET /health`) |

Rutas Go relativas a `src/backend/`.

---

# Cierre y Q&A
Tres próximos pasos, en este orden:

1. Logs JSON y un request/trace id en el `context` de la API y del Worker.
2. Métricas RED de la API (Rate, Errors, Duration).
3. Spans de OpenTelemetry en la emisión y en el envío a SUNAT.

**Hoy** el proceso se ve por logs, `GET /health` y el estado en el admin. **Próximo paso:** correlacionar una emisión con su CDR y medir la API.

Gracias. Preguntas.


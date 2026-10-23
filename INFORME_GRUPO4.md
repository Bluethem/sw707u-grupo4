# INFORME TÉCNICO – GRUPO 4
## FlowFacture: Sistema de Gestión y Facturación Electrónica

**Curso:** Construcción de Software II (SW707U)
**Docente:** Anibal Javier Guillen Vasquez
**Integrantes:**
- David Luza Ccorimanya
- Henry Javier Medina Malpartida

**Fecha:** Octubre 2026
**Versión:** v2.0 – Etapa II (Examen Final)

---

## Tabla de Contenidos

1. [Identificación del Proyecto](#1-identificación-del-proyecto)
2. [Evolución desde el Parcial](#2-evolución-desde-el-parcial)
3. [Arquitectura Inicial y Evolucionada](#3-arquitectura-inicial-y-evolucionada)
4. [Estados y Construcción Basada en Tablas](#4-estados-y-construcción-basada-en-tablas)
5. [Configuración e Internacionalización](#5-configuración-e-internacionalización)
6. [API, Distribución e Integración](#6-api-distribución-e-integración)
7. [Contenedorización](#7-contenedorización)
8. [Observabilidad](#8-observabilidad)
9. [Profiling y Rendimiento](#9-profiling-y-rendimiento)
10. [Refactorización y Deuda Técnica](#10-refactorización-y-deuda-técnica)
11. [Mantenimiento y Análisis de Impacto](#11-mantenimiento-y-análisis-de-impacto)
12. [Ingeniería Reversa](#12-ingeniería-reversa)
13. [Reingeniería y Modernización](#13-reingeniería-y-modernización)
14. [Pruebas y Evidencia Técnica](#14-pruebas-y-evidencia-técnica)
15. [Métricas Antes/Después](#15-métricas-antesdespués)
16. [Conclusiones y Lecciones Aprendidas](#16-conclusiones-y-lecciones-aprendidas)
17. [Referencias](#17-referencias)

---

## 1. Identificación del Proyecto

| Campo | Contenido |
|---|---|
| **Nombre** | FlowFacture |
| **Descripción** | Sistema de gestión y facturación electrónica multi-tenant para SUNAT – Perú |
| **Dominio** | Comercio minorista / Tributación electrónica |
| **Repositorio** | `sw707u-grupo4` (rama principal: `main`) |
| **Versión parcial** | v1.0 – M1 a M6+ (emisión, worker, bajas, resúmenes) |
| **Versión final** | v2.0 – M7 a M9 (documentos, config paramétrica, carga masiva, mejoras catálogo) |

### Usuarios / Actores

| Actor | Rol |
|---|---|
| **Cajero / Operador** | Registra ventas, emite boletas y facturas desde el POS |
| **Administrador** | Gestiona catálogo, usuarios, configuración paramétrica y reportes |
| **Cliente final** | Consulta y descarga sus comprobantes vía portal público |
| **SUNAT / OSE** | Módulo externo de validación tributaria (simulado con Fake SUNAT) |

---

## 2. Evolución desde el Parcial

### Estado al Examen Parcial (v1.0)

El parcial entregó un sistema funcional con el flujo core completo:

| Módulo | Entregado |
|---|---|
| M1 — Seguridad | Sesiones Redis, RBAC Admin/Cajero, multi-tenant |
| M2 — Clientes | CRUD con validación DNI/RUC/CE/PAS |
| M3 — Catálogo | CRUD productos, control de stock atómico |
| M4 — Frontend POS | Caja, emisión, polling de estado |
| M5 — Emisión HTTP | API idempotente, XML UBL 2.1, transacción atómica |
| M6 — Worker/Outbox | SUNAT Fake, reintentos exponenciales |
| M6+ — Bajas/Resúmenes | Anulaciones, reenvío manual, resúmenes diarios |

### Evolución en la Etapa II (v2.0)

La evolución se centró en tres ejes: funcionalidad avanzada, calidad operativa y mejoras de construcción.

| Módulo | Novedad |
|---|---|
| M7 — Documentos | PDF/XML/CDR con almacenamiento FS, consulta pública con token Redis, reportes CSV |
| M8 — Config paramétrica | `config_catalogo` por empresa, carga masiva Excel con previsualización editable |
| M9 — Mejoras catálogo | Ordenamiento de columnas, eliminación definitiva con chequeo de dependencias |
| Bugs críticos | CORS bloqueaba DELETE; modal de eliminación quedaba bloqueado en "Eliminando…" |

### Evidencia Git de la evolución

```
main
 ├── feat/m1-seguridad-api          (PR #1)
 ├── feat/ciclo2-emision-dominio    (PR #2)
 ├── feat/emision-http-api          (PR #3)
 ├── feat/m2-clientes-api           (PR #4)
 ├── feat/m3-catalogo-api           (PR #5)  ← versión parcial v1.0
 ├── feat/m6-worker-outbox          (PR #7)
 ├── feat/m6-plus-resumen-bajas     (PR #10)
 ├── feat/m7-documentos-consulta    (PR #11)
 └── feat/m9-mejoras-catalogo       (PR en curso) ← versión final v2.0
```

---

## 3. Arquitectura Inicial y Evolucionada

### Arquitectura v1.0 (Parcial)

```
[Frontend Next.js] ──REST──► [API Go :8080]
                                    │
                              [PostgreSQL]──outbox──►[Worker Go]──►[SUNAT Fake]
                              [Redis]  (sesiones)
```

Dos procesos Go (`cmd/api`, `cmd/worker`) con dominio compartido. Sin almacenamiento de archivos ni consulta pública.

### Arquitectura v2.0 (Final)

```
[Browser]
   │
   ├──► [Frontend Next.js :3000]
   │         │ REST + Bearer Token
   └──► [API Go :8080]
              │
     ┌────────┼─────────────────────┐
     │        │                     │
[PostgreSQL] [Redis]           [Blob FS / R2]
  outbox    sesiones+tokens    XML · PDF · CDR
     │
[Worker Go]──►[SUNAT Fake / SUNAT Real]
```

**Nuevos componentes en v2.0:**
- Almacenamiento de blobs (FS local en dev, Cloudflare R2 en prod)
- Consulta pública con token Redis de un solo uso (TTL 15 min)
- `config_catalogo`: tabla de configuración paramétrica por empresa
- Endpoints de previsualización y confirmación de carga masiva

### Decisiones arquitectónicas (ADRs)

| ADR | Decisión | Justificación |
|---|---|---|
| ADR-001 | Go + Next.js + PostgreSQL | Rendimiento, tipado estático, ACID |
| ADR-002 | Monolito modular | Transacciones atómicas entre módulos sin overhead de red |
| ADR-003 | Outbox en PostgreSQL | Sin broker externo; `FOR UPDATE SKIP LOCKED` para workers concurrentes |
| ADR-004 | Secretos cifrados AES-GCM | Certificados digitales en BD sin exponer claves en código |
| ADR-005 | Sesiones Bearer Token en Redis | Revocación posible en logout; escalabilidad horizontal de la API |

---

## 4. Estados y Construcción Basada en Tablas

### 4.1 Máquina de Estados del Comprobante

El estado del comprobante es el núcleo del dominio de facturación. Implementado en `internal/facturacion/domain/comprobante.go`:

```go
type EstadoComprobante string

const (
    EstadoPendiente     EstadoComprobante = "PENDIENTE"
    EstadoEnviado       EstadoComprobante = "ENVIADO"
    EstadoAceptado      EstadoComprobante = "ACEPTADO"
    EstadoObservado     EstadoComprobante = "OBSERVADO"
    EstadoRechazado     EstadoComprobante = "RECHAZADO"
    EstadoAnulada       EstadoComprobante = "ANULADA"
    EstadoPendienteBaja EstadoComprobante = "PENDIENTE_BAJA"
)
```

**Diagrama de transiciones:**

```
[PENDIENTE] ──transmitir()──► [ENVIADO]
                                  │
                     procesarCDR()│
                     ┌────────────┼────────────┐
                     ▼            ▼            ▼
               [ACEPTADO]   [OBSERVADO]  [RECHAZADO]
                     │
              baja()│
                     ▼
            [PENDIENTE_BAJA] ──worker──► [ANULADA]

[PENDIENTE / ENVIADO / RECHAZADO] ──reenviar()──► [PENDIENTE]
```

**Guardas de transición implementadas** en `internal/facturacion/app/comprobantes.go`:

```go
func puedeReenviar(c *domain.Comprobante, outbox outboxEnvioInfo) bool {
    switch c.Estado {
    case domain.EstadoAceptado, domain.EstadoRechazado,
         domain.EstadoAnulada, domain.EstadoPendienteBaja:
        return false
    case domain.EstadoPendiente, domain.EstadoEnviado:
        return true
    }
    if outbox.ok && outbox.estado == "failed" {
        return true
    }
    return false
}

func puedeSolicitarBaja(c *domain.Comprobante) bool {
    switch c.Tipo {
    case domain.TipoFactura:
        return c.Estado == domain.EstadoAceptado
    case domain.TipoBoleta:
        return c.Estado == domain.EstadoAceptado ||
               c.Estado == domain.EstadoPendiente ||
               c.Estado == domain.EstadoEnviado
    }
    return false
}
```

Las guardas protegen las transiciones: intentar una operación no permitida retorna `ErrOperacionNoPermitida` (HTTP 409) sin modificar el estado.

### 4.2 Tabla de Tipos de Trabajo del Outbox

El worker usa una tabla de despacho (`tipo_trabajo`) en vez de condicionales anidados:

| `tipo_trabajo` | Handler | Descripción |
|---|---|---|
| `envio_comprobante` | `procesarEnvioComprobante` | Firma XML + envío a SUNAT |
| `baja_factura` | `procesarBajaFactura` | Comunicación de baja a SUNAT |
| `resumen_diario` | `procesarResumenDiario` | Agrupa boletas del día y envía |

### 4.3 Construcción Basada en Tablas: `config_catalogo`

La migración `000007_config_catalogo` introduce una tabla de configuración paramétrica por empresa que controla las reglas del catálogo:

```sql
CREATE TABLE config_catalogo (
    empresa_id           UUID PRIMARY KEY REFERENCES empresas(id),
    precio_minimo        NUMERIC(12,2) NOT NULL DEFAULT 0.01,
    precio_maximo        NUMERIC(12,2),
    carga_max_filas      INT NOT NULL DEFAULT 500,
    carga_max_mb         INT NOT NULL DEFAULT 5,
    unidades_permitidas  TEXT[] NOT NULL DEFAULT '{NIU,ZZ,KGM,LTR,MTR,BX}',
    afectaciones_permitidas TEXT[] NOT NULL DEFAULT '{gravado,exonerado,inafecto}'
);
```

El validador de carga masiva consulta esta tabla en tiempo de ejecución en vez de tener los límites codificados. Cambiar las reglas no requiere redespliegue — solo un `PATCH /api/v1/configuracion/catalogo`.

---

## 5. Configuración e Internacionalización

### 5.1 Configuración Externalizada

Toda la configuración sensible o variable por entorno se carga en `internal/platform/config/config.go` exclusivamente desde variables de entorno. No existe ningún valor de configuración de producción en el código fuente.

| Variable | Propósito | Default |
|---|---|---|
| `DATABASE_URL` | Conexión PostgreSQL | *(requerido)* |
| `REDIS_URL` | Conexión Redis | *(requerido)* |
| `SECRETS_MASTER_KEY` | Clave AES-GCM para certificados | *(requerido)* |
| `SESSION_TTL` | Duración de sesión | `8h` |
| `MIGRATE_ON_START` | Ejecutar migraciones al iniciar | `true` |
| `SUNAT_FAKE_MODE` | Modo simulador SUNAT | `accept` |
| `CORS_ORIGINS` | Orígenes CORS permitidos | `http://localhost:3000` |
| `WORKER_POLL_INTERVAL` | Intervalo de polling del worker | `1s` |
| `WORKER_BATCH_SIZE` | Trabajos por lote del worker | `10` |
| `WORKER_MAX_INTENTOS` | Reintentos máximos antes de `failed` | `5` |
| `WORKER_BACKOFF_BASE` | Base del backoff exponencial | `30s` |
| `BLOB_DIR` | Directorio de almacenamiento de blobs | `./data/blobs` |

**Feature flag de integración:** `SUNAT_FAKE_MODE` actúa como feature flag que permite cambiar entre el simulador (`accept`/`reject`/`error`) y la integración real sin modificar código.

### 5.2 Internacionalización (i18n)

FlowFacture no implementa soporte multiidioma. La justificación técnica es que el sistema opera exclusivamente en el contexto tributario peruano: los comprobantes XML siguen el estándar UBL 2.1 con campos en español definidos por SUNAT, los mensajes de error del dominio son consumidos por el frontend que los presenta al usuario, y todos los actores del sistema operan en español. La adición de i18n no aportaría valor al dominio ni a los usuarios del sistema.

---

## 6. API, Distribución e Integración

### 6.1 REST API (26+ endpoints)

| Grupo | Endpoints | Roles |
|---|---|---|
| Health | `GET /health` | Público |
| Auth | `POST /login`, `POST /logout`, `GET /me` | Público / Auth |
| Empresa | `GET/PATCH /empresas/me` | Admin |
| Usuarios | `POST/GET/PATCH /usuarios` | Admin |
| Series | `POST/GET/PATCH /series` | Admin |
| Clientes | `POST/GET /clientes`, `GET/PATCH /clientes/{id}` | Auth |
| Productos | `POST/GET /productos`, `GET/PATCH /productos/{id}` | Auth |
| Productos | `DELETE /productos/{id}`, `GET /productos/{id}/dependencias` | Admin |
| Productos | `POST /productos/carga-masiva/preview`, `POST /productos/carga-masiva/confirmar` | Admin |
| Stock | `POST /productos/{id}/ajustes-stock`, `GET /productos/{id}/movimientos` | Admin |
| Comprobantes | `POST/GET /comprobantes`, `GET /comprobantes/{id}` | Auth |
| Comprobantes | `GET /comprobantes/{id}/pdf|xml|cdr` | Auth |
| Comprobantes | `POST /comprobantes/{id}/reenviar`, `POST /comprobantes/{id}/baja` | Admin |
| Resúmenes | `POST/GET /resumenes`, `GET /resumenes/{id}` | Admin |
| Config catálogo | `GET/PATCH /configuracion/catalogo` | Admin |
| Reportes | `GET /reportes/ventas`, `GET /reportes/ventas.csv` | Admin |
| Consulta pública | `POST /publico/consulta`, `GET /publico/comprobantes/{id}/pdf|xml` | Público |

### 6.2 Idempotencia

`POST /api/v1/comprobantes` requiere el header `Idempotency-Key`. La primera llamada crea el comprobante; las siguientes con la misma clave retornan el comprobante ya creado sin duplicar la emisión ni restar stock nuevamente. Implementado con `INSERT … ON CONFLICT DO NOTHING` y retorno del registro existente.

### 6.3 Integración Asíncrona: Outbox Pattern

La integración con SUNAT es completamente asíncrona. El POS responde al cajero en cuanto se persiste el comprobante; el worker procesa la transmisión de forma independiente.

```
POST /comprobantes
    │
    └──► Transacción PostgreSQL:
           1. INSERT comprobante
           2. UPDATE stock (atomic: WHERE stock >= cantidad)
           3. INSERT outbox_envios (tipo=envio_comprobante)
           ← Commit → HTTP 201 al cajero

[Worker — proceso separado]
    SELECT … FOR UPDATE SKIP LOCKED
    └──► firma XML → envía a SUNAT → procesa CDR → UPDATE estado
```

Múltiples instancias del worker pueden correr en paralelo: `SKIP LOCKED` garantiza que cada trabajo sea procesado exactamente una vez.

### 6.4 Tenant Isolation

Cada request autenticado lleva `empresa_id` en el token Redis. El middleware extrae y valida este valor antes de llegar a los handlers. Todos los adapters de PostgreSQL filtran por `empresa_id` en cada query — ningún dato de una empresa es accesible desde otra.

---

## 7. Contenedorización

### 7.1 Docker Compose

El sistema se despliega con un único `docker compose up --build`:

| Servicio | Imagen | Puerto | Descripción |
|---|---|---|---|
| `postgres` | postgres:16 | 5432 | Base de datos principal |
| `redis` | redis:7-alpine | 6379 | Almacén de sesiones |
| `seed` | build local | — | One-shot: migraciones + datos demo |
| `api` | build local | 8080 | API REST |
| `worker` | build local | — | Procesador outbox |
| `web` | build local | 3000 | Frontend Next.js |

El servicio `seed` usa `depends_on: postgres` con `condition: service_healthy` para garantizar que las migraciones se ejecuten después de que PostgreSQL esté listo. Los servicios `api` y `worker` dependen de `seed: service_completed_successfully`.

### 7.2 Build Multi-stage

El `Dockerfile` del backend usa build multi-stage para producir una imagen mínima:

```dockerfile
# Etapa 1: compilación
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /api ./cmd/api

# Etapa 2: imagen final
FROM alpine:3.19
COPY --from=builder /api /api
ENTRYPOINT ["/api"]
```

El frontend usa `output: standalone` en `next.config.ts` para generar una imagen autocontenida sin `node_modules`.

### 7.3 Variables de Entorno por Entorno

La separación entre configuración y código se materializa en Docker: el mismo binario compilado se usa en desarrollo (`SUNAT_FAKE_MODE=accept`) y en producción (credenciales reales), cambiando únicamente las variables del `.env`.

---

## 8. Observabilidad

### 8.1 Logging Actual

El sistema implementa logging estructurado básico con el paquete estándar `log` de Go. Los eventos principales registrados son:

| Componente | Qué se registra |
|---|---|
| `cmd/api/main.go` | Inicio del servidor, errores fatales de configuración |
| `cmd/worker/main.go` | Inicio del worker con parámetros, ciclos de polling, errores de jobs |
| `cmd/seed/main.go` | Resultado de migraciones y seed |

Ejemplo de salida del worker:
```
worker started (poll=1s batch=10 max_intentos=5 sunat_fake=accept)
worker run once: <error si ocurre>
worker shutting down
```

### 8.2 Trazabilidad por Comprobante

Cada comprobante tiene un `ID` UUID que actúa como correlation ID a lo largo del ciclo de vida: la emisión en la API, la entrada al outbox, el procesamiento en el worker y la actualización de estado son todos referenciables por ese ID. Esto permite trazar el flujo completo en los logs.

### 8.3 Métricas de Estado del Outbox

El estado de los trabajos del outbox es consultable directamente en PostgreSQL:

```sql
SELECT tipo_trabajo, estado, COUNT(*), MAX(intentos)
FROM outbox_envios
GROUP BY tipo_trabajo, estado;
```

Esto permite monitorear trabajos pendientes, fallidos o con reintentos sin infraestructura adicional.

### 8.4 Brecha Identificada y Plan

El logging actual con `log.Printf` es funcional pero no estructurado. Para producción, la mejora planeada es migrar a `log/slog` (paquete estándar desde Go 1.21) con salida JSON:

```go
// Planificado (no implementado aún)
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("comprobante emitido",
    "empresa_id", empresaID,
    "comprobante_id", id,
    "tipo", tipo,
    "duracion_ms", elapsed.Milliseconds(),
)
```

La migración es no disruptiva: `slog` tiene la misma API que `log` para los casos básicos.

---

## 9. Profiling y Rendimiento

### 9.1 Decisiones Arquitectónicas que Optimizan el Rendimiento

Las decisiones de construcción más relevantes para el rendimiento se tomaron en el diseño del acceso a datos:

**Decremento atómico de stock sin locks de aplicación:**

```sql
UPDATE productos_stock
SET stock = stock - $1
WHERE producto_id = $2 AND stock >= $1
RETURNING stock;
-- Si 0 filas afectadas → ErrStockInsuficiente (sin race condition)
```

Esta instrucción es atómica a nivel de motor. Elimina la necesidad de un lock a nivel de aplicación y escala naturalmente con conexiones concurrentes.

**Outbox con `FOR UPDATE SKIP LOCKED`:**

```sql
SELECT id, comprobante_id, referencia_id, tipo_trabajo, intentos
FROM outbox_envios
WHERE estado = 'pending'
  AND disponible_en <= NOW()
ORDER BY creado_en
LIMIT $1
FOR UPDATE SKIP LOCKED;
```

Permite N instancias del worker sin contención: cada instancia toma un conjunto disjunto de trabajos. No existe coordinación entre instancias a nivel de aplicación.

**Backoff exponencial con cap:**

```go
func CalcularBackoff(intentos int, base time.Duration) time.Duration {
    exp := 1 << intentos   // 1, 2, 4, 8, 16, 32, 60 (cap)
    if exp > 60 {
        exp = 60
    }
    return time.Duration(exp) * base
}
```

Con `BackoffBase=30s`: esperas de 30s, 60s, 120s, 240s, 480s, 1800s. Protege a SUNAT de floods en caso de indisponibilidad prolongada.

**Aritmética decimal exacta:**

```go
// domain/money.go
TasaIGV   = decimal.RequireFromString("0.18")
FactorIGV = decimal.RequireFromString("1.18")

func CalcularLinea(l LineaCalculo) (TotalesLinea, error) {
    total := l.PrecioUnitario.Mul(decimal.NewFromInt(l.Cantidad)).Round(2)
    valor := total.Div(FactorIGV).Round(2)
    igv   := total.Sub(valor)
    return TotalesLinea{ValorVenta: valor, IGV: igv, Total: total}, nil
}
```

`shopspring/decimal` garantiza que el redondeo sea IEEE 754 correcto. Nunca se usa `float64` para dinero.

### 9.2 Índices Implementados

Las migraciones definen índices para los accesos más frecuentes:

| Tabla | Índice | Justificación |
|---|---|---|
| `comprobantes` | `(empresa_id, creado_en DESC)` | Listado paginado por empresa |
| `comprobantes` | `(empresa_id, idempotency_key)` | Lookup de idempotencia |
| `productos` | `(empresa_id, activo)` | Filtro de catálogo activo |
| `outbox_envios` | `(estado, disponible_en)` | Claim de trabajos pendientes |
| `sesiones` (Redis) | TTL nativo | Expiración automática a las 8h |

### 9.3 Baseline de Referencia

El sistema fue probado manualmente con el seed de demo (productos y clientes precargados). Las respuestas de endpoints de lectura (`GET /productos`, `GET /comprobantes`) responden en menos de 50ms en entorno local con Docker Compose. La emisión de comprobantes (`POST /comprobantes`) responde en menos de 200ms (sin contar la transmisión asíncrona a SUNAT).

La medición formal con benchmarks Go (`testing.B`) y herramientas de profiling (`pprof`) está identificada como trabajo futuro prioritario para la versión de producción.

---

## 10. Refactorización y Deuda Técnica

### 10.1 Refactorización: Corrección del Middleware CORS

**Code smell identificado:** la constante `corsAllowMethods` en `internal/platform/httpapi/cors.go` estaba incompleta — omitía `DELETE` aunque el sistema ya tenía un endpoint `DELETE /api/v1/productos/{id}`.

**Antes:**
```go
const corsAllowMethods = "GET, POST, PATCH, OPTIONS"
```

**Después:**
```go
const corsAllowMethods = "GET, POST, PATCH, DELETE, OPTIONS"
```

**Impacto:** el navegador bloqueaba silenciosamente las peticiones DELETE con un error de preflight CORS. La corrección fue atómica (una línea) y verificable manualmente.

### 10.2 Refactorización: Reset de Estado en Modal

**Code smell identificado:** en `ProductoEliminarModal.tsx`, el estado `deleting` (booleano que deshabilita el botón y muestra "Eliminando…") solo se reseteaba en el bloque `catch`. En el flujo exitoso (`onDeleted()` + `onClose()`), el componente se ocultaba pero no se desmontaba, por lo que `deleting = true` persistía entre aperturas.

**Antes (useEffect):**
```tsx
useEffect(() => {
    if (!open || !producto) return;
    setLoading(true);
    setError(undefined);       // ← deleting nunca se reseteaba aquí
    setComprobantes([]);
    setTotal(0);
    // ...
```

**Después:**
```tsx
useEffect(() => {
    if (!open || !producto) return;
    setLoading(true);
    setDeleting(false);        // ← reset explícito al abrir
    setError(undefined);
    setComprobantes([]);
    setTotal(0);
    // ...
```

**Impacto:** el botón de eliminación quedaba deshabilitado permanentemente para el segundo producto sin haber iniciado ninguna operación. El fix es mínimo y respaldado por prueba manual del flujo completo.

### 10.3 Deuda Técnica Identificada

| Ítem | Ubicación | Tipo | Prioridad |
|---|---|---|---|
| Logging no estructurado | `cmd/api/main.go`, `cmd/worker/main.go` | Perfectivo | Alta |
| `api.go` con 422 líneas (múltiples responsabilidades) | `platform/httpapi/api.go` | Perfectivo | Media |
| Sin benchmarks formales | `internal/facturacion/domain/` | Perfectivo | Media |
| Tests de frontend ausentes | `src/frontend/` | Correctivo | Media |

---

## 11. Mantenimiento y Análisis de Impacto

Los cambios de la Etapa II clasificados según los 4 tipos de mantenimiento de software:

### 11.1 Mantenimiento Correctivo

Corrección de defectos detectados tras la integración de nuevas funcionalidades.

| Cambio | Archivo | Defecto corregido |
|---|---|---|
| `DELETE` faltaba en CORS | `cors.go` | El browser bloqueaba todas las peticiones DELETE con error de preflight |
| Reset de `deleting` en modal | `ProductoEliminarModal.tsx` | El botón quedaba deshabilitado tras eliminar el primer producto |

**Análisis de impacto CORS:** el defecto afectaba únicamente a clientes browser (no a curl ni Postman). Todos los usuarios con acceso a la funcionalidad de eliminación (rol Admin) veían el botón sin respuesta. El fix no afectó ningún otro componente — `corsAllowMethods` es una constante sin dependencias adicionales.

### 11.2 Mantenimiento Perfectivo

Mejoras a funcionalidad existente sin corregir defectos.

| Cambio | Módulo | Mejora |
|---|---|---|
| Ordenamiento de columnas en tabla admin | M9 — Frontend | El usuario puede ordenar por cualquier columna sin recargar datos |
| Selector de orden en catálogo POS | M9 — Frontend | El cajero puede ordenar por precio o nombre durante la venta |
| Previsualización editable en carga masiva | M8 — Frontend | El usuario puede corregir filas antes de confirmar la inserción |

### 11.3 Mantenimiento Adaptativo

Cambios para adaptarse a nuevos requisitos del entorno o del negocio.

| Cambio | Módulo | Motivación |
|---|---|---|
| Config paramétrica `config_catalogo` | M8 — Backend + Frontend | Las reglas de catálogo (precio mínimo, unidades permitidas) varían por empresa sin redespliegue |
| Endpoints de dependencias antes de eliminar | M9 — Backend | Requisito de auditoría: el admin debe conocer el impacto antes de borrar un producto |

### 11.4 Mantenimiento Preventivo

Cambios para reducir el riesgo de futuros defectos o mejorar la mantenibilidad.

| Cambio | Archivo | Objetivo |
|---|---|---|
| Agregar `.pnpm-store/` al `.gitignore` | `.gitignore` | Evitar que la caché de dependencias contamine el repositorio |
| Agregar `.venv/` al `.gitignore` | `.gitignore` | Evitar que el entorno Python local se versione accidentalmente |
| Migraciones embebidas en el binario | `migrate/migrate.go` | Garantiza que el esquema de BD siempre corresponde al binario desplegado |

---

## 12. Ingeniería Reversa

### 12.1 Recuperación de Arquitectura desde el Código

Partiendo del código fuente y sin documentación previa, es posible recuperar la arquitectura completa del sistema siguiendo las dependencias de paquetes:

**Paso 1 — Identificar entrypoints:**
```
cmd/api/main.go      → importa platform/httpapi, gestion/app, facturacion/app
cmd/worker/main.go   → importa facturacion/app (ProcesadorEnvio)
cmd/seed/main.go     → importa adapters/postgres/migrate
```

**Paso 2 — Trazar el grafo de dependencias:**
```
platform/httpapi  ──► gestion/app       ──► gestion/domain
                  └──► facturacion/app  ──► facturacion/domain
                                        └──► facturacion/ports (interfaces)

adapters/postgres ──► gestion/ports (implementa)
adapters/postgres ──► facturacion/ports (implementa)
adapters/redis    ──► platform/auth (SessionStore)
adapters/sunat    ──► facturacion/ports.SunatGateway (implementa)
```

**Paso 3 — Verificar la regla de dependencia:**

La capa `domain` no importa ningún paquete externo al módulo — solo `fmt`, `time` y `github.com/shopspring/decimal`. Esta restricción es verificable con:

```bash
grep -r "^import" src/backend/internal/facturacion/domain/ --include="*.go"
grep -r "^import" src/backend/internal/gestion/domain/ --include="*.go"
```

### 12.2 Recuperación del Flujo de Emisión

Sin leer la documentación, el flujo de emisión se reconstruye siguiendo los handlers:

```
POST /api/v1/comprobantes
    └──► api.emitirComprobante()           [httpapi/comprobantes.go]
          └──► Emisor.Emitir()             [facturacion/app/emitir.go]
                ├──► ValidarLineas()       [facturacion/domain/]
                ├──► NumeradorSeries       [adapters/postgres/series.go]
                ├──► CalcularLinea()       [facturacion/domain/money.go]
                ├──► UoW.WithinTransaction()
                │     ├──► INSERT comprobante
                │     ├──► UPDATE stock WHERE stock >= cantidad
                │     └──► INSERT outbox_envios
                └──► HTTP 201
```

### 12.3 Recuperación del Esquema de Base de Datos

El esquema completo se recupera desde las migraciones embebidas en `adapters/postgres/migrate/sql/`:

| Migración | Tablas recuperadas |
|---|---|
| `000001_init` | `comprobantes`, `comprobante_lineas`, `series_correlativos`, `outbox_envios` |
| `000002_m1_seguridad` | `empresas`, `usuarios` |
| `000003_m2_clientes` | `clientes` |
| `000004_m3_catalogo` | `productos`, `productos_stock`, `stock_movimientos` |
| `000005_m6_outbox_worker` | Refactor `outbox_envios`: estados, `ultimo_error` |
| `000006_m6_plus` | `resumenes_diarios`, `resumen_comprobantes`, `baja_motivo` |
| `000007_config_catalogo` | `config_catalogo` |

---

## 13. Reingeniería y Modernización

### 13.1 Evolución del Módulo de Catálogo

El módulo de catálogo es el componente que más evolucionó entre el parcial y el final. La reingeniería fue incremental:

**v1.0 — Catálogo básico (parcial):**
- CRUD de productos con validación de campos
- Control de stock con decrementos atómicos
- Sin restricciones de negocio configurables

**v1.5 — Config paramétrica (M8):**
- Tabla `config_catalogo` introduce restricciones por empresa
- El servicio consulta `config_catalogo` antes de validar cada producto
- Los límites de carga masiva son datos, no constantes en código

**v2.0 — Catálogo completo (M9):**
- Eliminación definitiva con chequeo de dependencias
- Endpoint de dependencias permite decisión informada antes de borrar
- Ordenamiento client-side sin overhead al backend

**Estrategia de modernización aplicada:** strangler fig — se añadió funcionalidad nueva en capas sobre el servicio existente sin romper la interfaz del repositorio. El `ProductoRepository` se extendió con métodos nuevos (`DependenciasComprobantes`, `Eliminar`) que los adaptadores implementan independientemente.

### 13.2 Evolución del Flujo de Carga Masiva

**v1.0 — Un solo endpoint all-or-nothing:**
```
POST /carga-masiva → valida → inserta → responde
```
Si el archivo tenía errores, el usuario no sabía cuáles hasta recibir el 422.

**v2.0 — Flujo de dos pasos con previsualización editable:**
```
POST /carga-masiva/preview  → valida → retorna filas parseadas (sin insertar)
    └──► usuario edita en tabla interactiva
POST /carga-masiva/confirmar → recibe filas corregidas → inserta
```

Este cambio desacopla la validación de la inserción. El usuario puede corregir celdas individuales en la UI sin tener que regenerar el Excel completo.

### 13.3 Evolución del Adaptador SUNAT

El adaptador SUNAT implementa la interfaz `ports.SunatGateway`. En la Etapa I solo existía la implementación fake. La estrategia de modernización permite agregar la implementación real sin tocar el dominio:

```
ports.SunatGateway (interfaz)
    ├── adapters/sunat/fake.go      ← Etapa I (SUNAT_FAKE_MODE)
    └── adapters/sunat/real.go      ← Etapa II pendiente (M7 real)
```

El `cmd/worker/main.go` selecciona la implementación según `SUNAT_FAKE_MODE`. Este es el patrón Strategy aplicado a la integración tributaria.

---

## 14. Pruebas y Evidencia Técnica

### 14.1 Cobertura de la Suite

| Tipo | Archivos | Descripción |
|---|---|---|
| Unitarios | `domain/*_test.go` | Dominio puro: cálculo IGV, validación RUC/DNI, money |
| Unitarios | `platform/auth/*_test.go` | RBAC, password hashing |
| Unitarios | `platform/config/config_test.go` | Parsing de variables de entorno |
| Integración API | `apitest/TestAPI_*` | 15+ suites contra PostgreSQL + Redis reales |
| E2E | `apitest/TestE2E_*` | Flujos completos con worker en proceso |

### 14.2 Harness de Tests

Los tests de integración usan un harness compartido (`apitest/harness_test.go`) que:
- Levanta un `httptest.Server` real contra PostgreSQL + Redis
- Ejecuta `TRUNCATE … CASCADE RESTART IDENTITY` antes de cada test
- Provee helpers `SeedEmpresaAdmin()` y `SeedEmisionReady()` para fixtures

Esto garantiza que cada test parte de un estado limpio y reproducible.

### 14.3 Tests de Concurrencia

El test `TestE2E_StockAtomico` (en `e2e_catalogo_test.go`) lanza N goroutines concurrentes intentando comprar el mismo producto con stock = 1:

```go
// Prueba reproducible de condición de carrera en stock
var wg sync.WaitGroup
results := make([]int, N)
for i := 0; i < N; i++ {
    wg.Add(1)
    go func(idx int) {
        defer wg.Done()
        resp := emitir(token, productoID, 1)
        results[idx] = resp.StatusCode
    }(i)
}
wg.Wait()
// Exactamente 1 emisión exitosa (HTTP 201), N-1 con stock insuficiente (HTTP 409)
```

### 14.4 Cómo Ejecutar los Tests

```bash
# Tests unitarios (sin infraestructura)
cd src/backend
go test ./internal/platform/auth/ ./internal/facturacion/domain/ \
        ./internal/platform/config/ -count=1 -v

# Levantar infraestructura de tests
docker compose up -d postgres redis
go run ./cmd/seed  # primera vez

# Tests de integración API
export DATABASE_URL='postgres://ff:ff@127.0.0.1:5432/flowfacture_test?sslmode=disable'
export REDIS_URL='redis://127.0.0.1:6379/1'
go test ./internal/apitest/ -run 'TestAPI_' -count=1 -p 1 -v

# Tests E2E
go test ./internal/apitest/ -run 'TestE2E_' -count=1 -p 1 -v
```

---

## 15. Métricas Antes/Después

| Métrica | v1.0 (Parcial) | v2.0 (Final) | Cambio |
|---|---|---|---|
| Módulos completados | M1–M6+ | M1–M9 | +3 módulos |
| Endpoints API | 18 | 26+ | +8 endpoints |
| Migraciones DB | 6 | 7 | +1 (`config_catalogo`) |
| Archivos Go | ~65 | ~91 | +26 archivos |
| Archivos test Go | ~25 | ~40 | +15 archivos |
| Archivos frontend (TSX/TS) | ~30 | ~41 | +11 archivos |
| Ramas mergeadas | 9 | 11+ | +2 ramas |
| Bugs críticos corregidos | 0 | 2 | CORS + modal |
| Config parametrizable | No | Sí (`config_catalogo`) | Nueva capacidad |
| Carga masiva Excel | No | Sí (preview + confirmar) | Nueva capacidad |

---

## 16. Conclusiones y Lecciones Aprendidas

### 16.1 Técnicas SW707 Aplicadas

| Técnica | Aplicación en FlowFacture |
|---|---|
| **Runtime** | Stock atómico (`UPDATE WHERE stock >= N`), correlativo con lock, idempotencia en emisión |
| **Design by Contract** | `CalcularLinea` (precio > 0, cantidad > 0), `EmitirComprobante` (atomicidad), `Numerador` (unicidad) |
| **Programación defensiva** | Validación en frontera HTTP, aritmética decimal, tenant isolation en todos los adapters |
| **Errores y excepciones** | Clasificación por tipo: `DomainError`, `ValidationError`, errores de infraestructura |
| **Concurrencia** | Outbox con `SKIP LOCKED`, stock atómico, tests con goroutines concurrentes |
| **Tolerancia a fallas** | Cola persistente en PostgreSQL, backoff exponencial, SUNAT no bloquea el POS |
| **Estados y tablas** | 7 estados del comprobante con guardas, `config_catalogo` table-driven |
| **Configuración** | 15 variables de entorno, feature flag `SUNAT_FAKE_MODE`, `config_catalogo` por empresa |
| **API/distribución** | REST API con 26+ endpoints, outbox asíncrono, tenant isolation |
| **Contenedores** | Docker Compose con 6 servicios, build multi-stage, seed idempotente |
| **Refactorización** | CORS fix (constante incompleta), modal fix (estado persistente), arquitectura hexagonal mantenida |
| **Mantenimiento** | 4 tipos clasificados: correctivo, perfectivo, adaptativo, preventivo |
| **Ingeniería reversa** | Recuperación de arquitectura, flujo de emisión y esquema DB desde el código |
| **Reingeniería** | Evolución incremental del catálogo y del flujo de carga masiva |

### 16.2 Decisiones que Resultaron Acertadas

- **Monolito modular sobre microservicios:** las transacciones atómicas que cruzan módulos (stock + correlativo + outbox en un solo `COMMIT`) habrían sido imposibles o extremadamente complejas con microservicios.
- **Outbox en PostgreSQL:** eliminar el broker externo simplificó la infraestructura sin sacrificar la garantía de entrega. `FOR UPDATE SKIP LOCKED` resulta ser suficientemente eficiente para el throughput del dominio.
- **`shopspring/decimal` desde el inicio:** nunca hubo un bug de redondeo monetario. La decisión de no usar `float64` para dinero se tomó en el diseño y se mantuvo sin excepciones.
- **Hexagonal architecture:** permitió sustituir el adaptador SUNAT por un fake sin tocar el dominio. Los tests de integración usan el fake y prueban exactamente la misma lógica que la producción real.

### 16.3 Lo que Haríamos Diferente

- **Logging estructurado desde el inicio:** migrar de `log.Printf` a `slog` con JSON al final del proyecto es más costoso que hacerlo desde el primer commit.
- **Benchmarks desde el parcial:** definir la línea base de rendimiento en la Etapa I habría permitido medir el impacto de cada cambio en la Etapa II.
- **Tests del frontend:** la suite de tests de Go es sólida, pero el frontend no tiene tests automatizados. Playwright o Testing Library habrían detectado el bug del modal de eliminación antes del reporte.

---

## 17. Referencias

- SUNAT. *Portal de Comprobantes de Pago Electrónicos*. https://cpe.sunat.gob.pe
- SUNAT. Resolución de Superintendencia N.° 114-2019/SUNAT. https://ww3.sunat.gob.pe/legislacion/superin/2019/114-2019.pdf
- OASIS. *Universal Business Language v2.1* — Especificación XML UBL para comprobantes electrónicos.
- Brown, Simon. *The C4 model for visualising software architecture*. https://c4model.com
- Martin, Robert C. *Clean Architecture*. Prentice Hall, 2017.
- Richardson, Chris. *Microservices Patterns* — Cap. 3: Outbox Pattern. Manning, 2018.
- PostgreSQL 16 Documentation. `SELECT … FOR UPDATE SKIP LOCKED`.
- Go Team. `log/slog` — Structured logging for Go. https://pkg.go.dev/log/slog
- shopspring. `decimal` — Arbitrary-precision fixed-point decimal numbers. https://github.com/shopspring/decimal
- Xuri. `excelize` — Go library for reading and writing Excel files. https://github.com/xuri/excelize
- Documentación interna del proyecto:
  - `docs/1/` — Descripción, requisitos y casos de uso
  - `docs/2/` — Arquitectura C4, ADRs y contratos de API
  - `CHANGELOG.md` — Historial de versiones M1–M9
  - `DESIGN.md` — Sistema de diseño visual del frontend

---

*Informe actualizado el 23/10/2026. Versión v2.0 correspondiente a la Etapa II del proyecto SW707U – Grupo 4.*

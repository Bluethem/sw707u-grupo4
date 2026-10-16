# Contratos de API

Base URL local: `http://127.0.0.1:8080`

Autenticación: header `Authorization: Bearer <token>` (sesión Redis). El login también puede dejar cookie según el cliente; el frontend usa Bearer.

Roles:

| Rol | Código | Uso típico |
|---|---|---|
| Administrador | `administrador` | Configuración, catálogo mutaciones, reenvío, bajas, resúmenes |
| Cajero | `cajero` | Emisión, consulta de clientes/productos/comprobantes |

Errores comunes (JSON `{ "error": "…" }`):

| HTTP | Cuándo |
|---|---|
| 400 | Validación / cuerpo inválido |
| 401 | Sin sesión o credenciales inválidas |
| 403 | Rol insuficiente |
| 404 | Recurso no encontrado (tenant-safe) |
| 409 | Conflicto (stock, código duplicado, operación no permitida) |
| 500 | Error interno |

Multi-tenant: el `empresa_id` sale del token; no se acepta en el body para cambiar de empresa.

---

## Salud

### `GET /health`

Sin auth. Respuesta `200`: `{ "status": "ok" }`.

---

## Auth (M1)

### `POST /api/v1/auth/login`

Body:

```json
{ "email": "admin@flowfacture.local", "password": "Admin123!" }
```

Respuesta `200`: `{ "token": "<sesión>" }`.

### `POST /api/v1/auth/logout`

Auth requerida. Invalida la sesión.

### `GET /api/v1/auth/me`

Auth requerida. Devuelve usuario + empresa del token (`id`, `email`, `nombre`, `rol`, `empresa_id`, …).

---

## Empresa, usuarios y series (M1) — admin

### `GET /api/v1/empresas/me` / `PATCH /api/v1/empresas/me`

Consulta / actualiza datos de la empresa del token.

`PATCH` (campos opcionales): `ruc`, `razon_social`, `domicilio_fiscal`, `regimen`, `activa`.

### `POST /api/v1/usuarios` / `GET /api/v1/usuarios` / `PATCH /api/v1/usuarios`

Crear, listar y actualizar usuarios de la empresa.

`POST`: `email`, `password`, `nombre`, `rol` (`administrador` \| `cajero`).

`PATCH`: `id` obligatorio; opcionales `email`, `password`, `nombre`, `rol`, `activo`.

### `POST /api/v1/series` / `GET /api/v1/series` / `PATCH /api/v1/series`

Series de facturación (`boleta` / `factura`).

`POST`: `tipo`, `serie`, `siguiente`.

`PATCH`: `tipo`, `serie`; opcionales `siguiente`, `activa`.

---

## Clientes (M2)

Auth: cajero o admin.

### `POST /api/v1/clientes`

```json
{
  "tipo_doc": "DNI",
  "numero": "12345678",
  "razon_social": "Juan Pérez",
  "direccion": "Av. Lima 100",
  "correo": "juan@example.com"
}
```

`tipo_doc`: `DNI` \| `RUC` \| `CE` \| `PAS` (según dominio).

### `GET /api/v1/clientes`

Query: `q`, `activo` (`true`/`false`), `limit`, `offset`.

Respuesta: `{ "items": […], "total": N }`.

### `GET /api/v1/clientes/{id}`

### `PATCH /api/v1/clientes/{id}`

No cambia `tipo_doc` ni `numero`. Puede desactivar con `activo: false`.

---

## Productos e inventario (M3)

| Método | Ruta | Rol |
|---|---|---|
| `POST` | `/api/v1/productos` | admin |
| `GET` | `/api/v1/productos` | auth |
| `GET` | `/api/v1/productos/{id}` | auth |
| `PATCH` | `/api/v1/productos/{id}` | admin |
| `POST` | `/api/v1/productos/{id}/ajustes-stock` | admin |
| `GET` | `/api/v1/productos/{id}/movimientos` | admin |

### Crear

```json
{
  "codigo": "SKU-001",
  "descripcion": "Producto demo",
  "unidad": "NIU",
  "precio": "118.00",
  "afectacion": "gravado",
  "controla_stock": true,
  "stock_inicial": 10
}
```

`afectacion`: `gravado` \| `exonerado` \| `inafecto`.

### Listar / obtener

Incluyen `stock` (número o ausente/`null` si no controla stock). Query listado: `q`, `activo`, `limit`, `offset`.

### `PATCH`

No cambia `codigo` ni `controla_stock`. Campos: `descripcion`, `unidad`, `precio`, `afectacion`, `activo`.

### Ajuste de stock

```json
{ "tipo": "entrada", "cantidad": 5, "motivo": "compra" }
```

`tipo`: `entrada` \| `salida` \| `ajuste`. Salida mayor al saldo → `409`.

---

## Comprobantes — emisión y consulta (M5 / M4)

Auth: cajero o admin.

### `POST /api/v1/comprobantes`

Header obligatorio: `Idempotency-Key`.

```json
{
  "tipo": "boleta",
  "serie": "B001",
  "cliente": {
    "tipo_doc": "DNI",
    "numero": "12345678",
    "razon_social": "Juan Pérez"
  },
  "lineas": [
    {
      "producto_id": "<uuid>",
      "cantidad": 1,
      "precio_unitario": "118.00",
      "afectacion": "gravado"
    }
  ]
}
```

- `tipo`: `boleta` \| `factura`.
- Misma `Idempotency-Key` → mismo comprobante, HTTP `200`; primera vez `201`.
- Encola outbox `envio_comprobante` (worker M6).

Estados: `PENDIENTE`, `ENVIADO`, `ACEPTADO`, `OBSERVADO`, `RECHAZADO`, `ANULADA`, `PENDIENTE_BAJA`.

### `GET /api/v1/comprobantes/{id}`

Detalle con líneas y `baja_motivo` si aplica.

### `GET /api/v1/comprobantes`

Query: `tipo`, `estado`, `fecha_desde`, `fecha_hasta` (`YYYY-MM-DD`), `limit`, `offset`.

Respuesta: `{ "items": […], "total": N }`.

---

## Reenvío y bajas (M6+) — admin

### `POST /api/v1/comprobantes/{id}/reenviar`

Sin body. Reencola `envio_comprobante` si el comprobante está en `PENDIENTE`/`ENVIADO` o el outbox de envío falló / quedó pendiente con error. No aplica a `ACEPTADO`, `RECHAZADO`, `ANULADA`, `PENDIENTE_BAJA`.

### `POST /api/v1/comprobantes/{id}/baja`

```json
{ "motivo": "Error de emisión" }
```

| Tipo | Condición | Efecto |
|---|---|---|
| `factura` | `ACEPTADO` | `PENDIENTE_BAJA` + outbox `baja_factura` |
| `boleta` | `ACEPTADO`, `PENDIENTE` o `ENVIADO` | `PENDIENTE_BAJA` + motivo; **sin** job inmediato (entra en resumen diario) |

Respuesta: comprobante actualizado (`200`).

---

## Resúmenes diarios (M6+) — admin

### `POST /api/v1/resumenes`

```json
{ "fecha": "2026-10-16" }
```

Agrupa boletas informables del día, crea resumen y encola `resumen_diario`. Respuesta `201`.

### `GET /api/v1/resumenes`

Lista de la empresa (array).

### `GET /api/v1/resumenes/{id}`

Detalle:

```json
{
  "id": "…",
  "empresa_id": "…",
  "fecha": "2026-10-16",
  "correlativo": 1,
  "estado": "PENDIENTE",
  "ticket": "",
  "creado_en": "…",
  "comprobantes": [
    { "comprobante_id": "…", "anulado": false }
  ]
}
```

`anulado: true` si la boleta iba en `PENDIENTE_BAJA`. Tras aceptación del worker: boletas normales → `ACEPTADO`; anuladas pendientes → `ANULADA`.

---

## Worker y Fake SUNAT (referencia)

No es HTTP pública; consume `outbox_envios` con `tipo_trabajo`:

| `tipo_trabajo` | Referencia | Fake |
|---|---|---|
| `envio_comprobante` | `comprobante_id` | `Enviar` |
| `baja_factura` | `comprobante_id` | `EnviarBaja` |
| `resumen_diario` | `referencia_id` (resumen) | `EnviarResumen` |

Modo local: `SUNAT_FAKE_MODE=accept|reject|error` (ver `scripts/README.md` y `src/backend/README.md`).

---

## CORS

Orígenes configurables con `CORS_ORIGINS` (default Compose: `http://localhost:3000`).

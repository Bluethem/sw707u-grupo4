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

Query: `tipo`, `estado`, `fecha_desde`, `fecha_hasta` (`YYYY-MM-DD`), `serie`, `numero`, `q` (razón social o documento del cliente), `limit`, `offset`.

- `serie` se normaliza a mayúsculas; `numero` es el correlativo del comprobante (entero ≥ 1).
- Si `fecha_desde` y `fecha_hasta` abarcan más de **90 días** → `400` `{ "error": "rango de fechas excede 90 dias" }`.

Respuesta: `{ "items": […], "total": N }`.

### Descargas (M7) — auth cajero/admin

| Método | Ruta | Content-Type | Notas |
|---|---|---|---|
| `GET` | `/api/v1/comprobantes/{id}/pdf` | `application/pdf` | Si falta blob, se regenera desde datos del comprobante |
| `GET` | `/api/v1/comprobantes/{id}/xml` | `application/xml` | Idem |
| `GET` | `/api/v1/comprobantes/{id}/cdr` | `application/octet-stream` | Solo tras estado final con CDR en almacenamiento |

Errores documentos:

| HTTP | Cuándo |
|---|---|
| 404 | Comprobante no encontrado (tenant) o CDR inexistente con estado final (`ACEPTADO` u otro distinto de `PENDIENTE`/`ENVIADO`) sin archivo en almacenamiento |
| 409 | CDR aún no disponible: comprobante en `PENDIENTE` o `ENVIADO`, o aún no hay CDR persistido tras el envío |

Respuesta binaria con `Content-Disposition: attachment; filename="{serie}-{numero}.{ext}"`.

Archivos en disco local vía `BLOB_DIR` (clave `{comprobante_id}.pdf|xml|cdr`). Ver `src/backend/README.md`.

---

## Consulta pública (M7)

Sin autenticación de usuario interno. Rate limit por IP (`RemoteAddr` del request): tras 10 intentos fallidos en 15 min → `429` `{ "error": "demasiados intentos" }`.

### `POST /api/v1/publico/consulta`

```json
{
  "ruc": "20100070970",
  "tipo_doc": "DNI",
  "numero": "12345678",
  "serie": "B001",
  "numero_comprobante": 1,
  "fecha": "2026-10-20",
  "monto": "236.00"
}
```

- `fecha` = fecha UTC de `creado_en` del comprobante (`YYYY-MM-DD`).
- `monto` = total con 2 decimales, debe coincidir exactamente.
- Comprobante con más de **365 días** desde `creado_en` → `410` `{ "error": "consulta expirada; solicite el archivo a la empresa" }` (cuenta como intento fallido para rate limit).
- Datos que no coinciden → `404` `{ "error": "datos no coinciden" }`.

Respuesta `200`:

```json
{
  "id": "<uuid>",
  "tipo": "boleta",
  "serie": "B001",
  "numero": 1,
  "fecha": "2026-10-20",
  "total": "236.00",
  "estado": "ACEPTADO",
  "razon_social_emisor": "Empresa API SAC",
  "token": "<sesión consulta>"
}
```

El `token` dura **15 minutos** en Redis y autoriza descargas públicas del mismo comprobante.

### `GET /api/v1/publico/comprobantes/{id}/pdf?token=…`

### `GET /api/v1/publico/comprobantes/{id}/xml?token=…`

Token inválido o comprobante distinto al del token → `401` `{ "error": "token invalido" }`.

---

## Reporte de ventas (M7) — admin

Query obligatoria en la práctica: `fecha_desde`, `fecha_hasta` (`YYYY-MM-DD`). Si ambas faltan, el backend usa el día UTC actual. Rango máximo **90 días** (mismo mensaje que listado).

### `GET /api/v1/reportes/ventas`

Respuesta `200`:

```json
{
  "fecha_desde": "2026-10-20",
  "fecha_hasta": "2026-10-20",
  "por_tipo": [
    { "tipo": "boleta", "cantidad": 2, "valor_venta": "200.00", "igv": "36.00", "total": "236.00" }
  ],
  "anulados": { "cantidad": 0, "total": "0.00" },
  "totales": { "tipo": "", "cantidad": 2, "valor_venta": "200.00", "igv": "36.00", "total": "236.00" }
}
```

Cuenta comprobantes no anulados en totales; bloque `anulados` resume comprobantes en `ANULADA` o `PENDIENTE_BAJA` del período.

### `GET /api/v1/reportes/ventas.csv`

Misma query de fechas. CSV UTF-8 con cabecera `tipo,cantidad,valor_venta,igv,total` y fila final `totales,…`.

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

## Configuración paramétrica del catálogo (M8) — admin

### `GET /api/v1/configuracion/catalogo`

Devuelve la configuración de validación del catálogo para la empresa del token. Si no hay fila en BD, retorna los valores por defecto.

Respuesta `200`:

```json
{
  "empresa_id": "…",
  "precio_minimo": "0.01",
  "precio_maximo": null,
  "carga_max_filas": 500,
  "carga_max_mb": 5,
  "unidades_permitidas": ["NIU", "ZZ", "KGM", "LTR", "MTR", "BX"],
  "afectaciones_permitidas": ["gravado", "exonerado", "inafecto"]
}
```

`precio_maximo: null` significa sin límite superior.

### `PATCH /api/v1/configuracion/catalogo`

Actualiza la configuración. Todos los campos son opcionales.

Body:

```json
{
  "precio_minimo": "0.50",
  "precio_maximo": "9999.99",
  "carga_max_filas": 200,
  "carga_max_mb": 10,
  "unidades_permitidas": ["NIU", "ZZ"],
  "afectaciones_permitidas": ["gravado", "exonerado"]
}
```

Restricciones: `precio_minimo > 0`; `precio_maximo >= precio_minimo` si no es null; `carga_max_filas` 1–2000; `carga_max_mb` 1–20; al menos un valor en cada array.

Respuesta `200`: configuración actualizada.

---

## Carga masiva de productos (M8) — admin

Flujo de dos pasos: previsualización editable → confirmación de inserción.

### `POST /api/v1/productos/carga-masiva/preview`

`multipart/form-data`, campo `archivo` (`.xlsx`).

Valida el archivo contra `config_catalogo` y devuelve las filas parseadas sin insertar nada.

Columnas esperadas en la primera hoja fila 1 (case-insensitive):
`codigo`, `descripcion`, `unidad`, `precio`, `afectacion`, `controla_stock`, `stock_inicial`

Respuesta `200`:

```json
{
  "filas": [
    { "codigo": "PROD-01", "descripcion": "Producto demo", "unidad": "NIU",
      "precio": "10.50", "afectacion": "gravado", "controla_stock": true, "stock_inicial": 100 }
  ]
}
```

Respuesta `422` (errores de validación, nada insertado):

```json
{
  "errores": [
    { "fila": 3, "columna": "unidad", "mensaje": "unidad 'KG' no está en la lista permitida" },
    { "fila": 5, "columna": "precio", "mensaje": "precio 0.20 está por debajo del mínimo 0.50" }
  ]
}
```

### `POST /api/v1/productos/carga-masiva/confirmar`

Recibe el arreglo de filas (posiblemente editado tras la previsualización) e inserta en transacción.

Body: `{ "filas": [ { …misma estructura… } ] }`

Respuesta `200`: `{ "cantidad": 15 }`

Respuesta `409` si algún código ya existe en el catálogo de la empresa.

---

## Dependencias y eliminación de productos (M9) — admin

### `GET /api/v1/productos/{id}/dependencias`

Lista los comprobantes que referencian el producto vía `comprobante_lineas`.

Respuesta `200`:

```json
{
  "total_comprobantes": 3,
  "comprobantes": [
    { "id": "…", "tipo": "boleta", "serie": "B001", "numero": 42, "estado": "ACEPTADO" }
  ]
}
```

Si `total_comprobantes = 0` el producto puede eliminarse sin impacto en registros históricos.

### `DELETE /api/v1/productos/{id}`

Elimina el producto en transacción atómica: `stock_movimientos` → `productos_stock` → `productos`.

Las `comprobante_lineas` conservan el `producto_id` por auditoría (sin FK constraint).

Respuesta `204` sin body. Respuesta `404` si no existe o pertenece a otra empresa.

---

## CORS

Orígenes configurables con `CORS_ORIGINS` (default Compose: `http://localhost:3000`).

Métodos permitidos: `GET, POST, PATCH, DELETE, OPTIONS`.

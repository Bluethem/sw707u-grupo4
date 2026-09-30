# Requerimientos

> Sistema de Gestión y Facturación Electrónica

**Convenciones**
- Prioridad MoSCoW: **M** (Must), **S** (Should), **C** (Could). Los requerimientos de evolución prevista van en la sección 3.
- Trazabilidad: `RN-xx` remite a las reglas de negocio de la sección 1 (Descripción del Proceso).
- Las metas numéricas de los RNF son **propuestas** y deben validarse con el docente/cliente y medirse en pruebas.

---

## 1. Requerimientos funcionales

### 1.1 Acceso, roles y multiempresa

| ID | Requerimiento | Prioridad | Trazab. |
|---|---|---|---|
| RF-01 | El sistema autentica usuarios con credenciales y mantiene sesión con expiración configurable. | M | RN-10 |
| RF-02 | El sistema aplica control de acceso por rol: **Cajero** (vender, emitir, consultar) y **Administrador** (todo lo anterior + catálogo, precios, stock, series, usuarios, anulaciones). | M | RN-10 |
| RF-03 | El sistema soporta múltiples empresas (tenants) aisladas: cada usuario solo accede a datos de su empresa. | M | RN-09 |
| RF-04 | El administrador registra los datos de la empresa emisora: RUC, razón social, domicilio fiscal, logo, régimen. | M | RN-06 |
| RF-05 | El administrador carga y actualiza el certificado digital y las credenciales de envío (SOL/OSE/PSE) de su empresa; el sistema valida vigencia del certificado y avisa antes de su vencimiento. | M | RN-06 |
| RF-06 | El administrador configura series por tipo de comprobante (F### facturas, B### boletas) y por punto de venta. | M | RN-02 |

### 1.2 Clientes

| ID | Requerimiento | Prioridad | Trazab. |
|---|---|---|---|
| RF-07 | Registrar, editar, buscar y desactivar clientes (tipo y número de documento, nombre/razón social, dirección, correo). | M | RN-01 |
| RF-08 | Validar formato y dígito verificador de RUC (11 dígitos) y DNI (8 dígitos). | M | RN-01 |
| RF-09 | Permitir venta a "cliente genérico" (sin identificar) cuando la boleta no supera S/ 700.00 y el cliente no lo solicita. | M | RN-01 |
| RF-10 | Autocompletar datos del cliente por RUC/DNI desde un servicio de consulta externo. | C | — |

### 1.3 Catálogo e inventario

| ID | Requerimiento | Prioridad | Trazab. |
|---|---|---|---|
| RF-11 | Registrar, editar, buscar y desactivar productos: código, descripción, unidad de medida, precio, afectación de IGV (gravado/exonerado/inafecto), control de stock sí/no. | M | RN-03 |
| RF-12 | Mantener stock por producto (y por almacén/sede si aplica) con movimientos trazables (venta, ajuste, anulación). | M | RN-05 |
| RF-13 | El administrador registra ajustes manuales de stock con motivo obligatorio. | S | RN-05, RN-10 |
| RF-14 | Alertar productos bajo un stock mínimo configurable. | C | — |
| RF-15 | Carga masiva de catálogo (CSV/Excel) con validación y reporte de errores. | S | — |

### 1.4 Ventas y caja

| ID | Requerimiento | Prioridad | Trazab. |
|---|---|---|---|
| RF-16 | El cajero arma una venta agregando productos por búsqueda o código; el sistema calcula subtotal, IGV y total en tiempo real. | M | RN-03 |
| RF-17 | Editar cantidades y quitar líneas antes de emitir; aplicar descuentos por línea o global con límite configurable. | S | — |
| RF-18 | Al emitir, el sistema reserva/descuenta stock de forma atómica; si no hay stock suficiente, rechaza la línea y notifica al cajero sin afectar las demás ventas. | M | RN-05 |
| RF-19 | Registrar el medio de pago de la venta (efectivo, tarjeta, transferencia, billetera digital), solo como dato informativo en esta fase. | S | — |
| RF-20 | Permitir guardar una venta en espera y retomarla. | C | — |
| RF-21 | Idempotencia: reintentar una emisión por fallo de red no genera un comprobante duplicado. | M | RN-02 |

### 1.5 Emisión de comprobantes

| ID | Requerimiento | Prioridad | Trazab. |
|---|---|---|---|
| RF-22 | Emitir **boletas de venta electrónicas** y **facturas electrónicas**. | M | RN-01 |
| RF-23 | Para facturas, exigir RUC válido del cliente. Para boletas, exigir tipo y número de documento cuando el total **supera** S/ 700.00 o el cliente lo solicita. | M | RN-01 |
| RF-24 | Asignar serie y correlativo únicos y consecutivos por empresa, tipo y serie, sin duplicados ni saltos. | M | RN-02 |
| RF-25 | Calcular impuestos por afectación de cada línea (gravado, exonerado, inafecto) con aritmética decimal y regla de redondeo documentada. | M | RN-03 |
| RF-26 | Generar el XML del comprobante según UBL 2.1 y los catálogos de SUNAT. | M | RN-06 |
| RF-27 | Firmar digitalmente el XML con el certificado vigente de la empresa. | M | RN-06 |
| RF-28 | Impedir la modificación de un comprobante ya emitido. | M | RN-04 |
| RF-29 | Consolidar ventas hasta S/ 5.00 en una boleta (opcional por empresa). | C | RN-11 |

### 1.6 Envío, validación y estados

| ID | Requerimiento | Prioridad | Trazab. |
|---|---|---|---|
| RF-30 | Enviar facturas a SUNAT/OSE dentro del plazo de 3 días calendario, de forma asíncrona (sin bloquear la caja). | M | RN-07 |
| RF-31 | Generar y enviar el **resumen diario** de boletas (agrupadas por día de emisión) dentro del plazo de hasta 7 días calendario. | M | RN-07 |
| RF-32 | Procesar la constancia de recepción (CDR) y actualizar el estado del comprobante: pendiente, enviado, aceptado, observado, rechazado. | M | RN-07 |
| RF-33 | Reintentar envíos fallidos automáticamente con backoff y permitir reenvío manual, dejando trazabilidad de cada intento. | M | RN-07 |
| RF-34 | Mostrar al usuario el estado de cada comprobante y los motivos de rechazo u observación. | M | RN-07 |
| RF-35 | Alertar comprobantes próximos a vencer su plazo de envío. | S | RN-07 |
| RF-36 | Permitir emitir aunque SUNAT/OSE no esté disponible: el comprobante queda en cola y se envía al restablecerse el servicio. | M | RN-07 |

### 1.7 Anulación y correcciones

| ID | Requerimiento | Prioridad | Trazab. |
|---|---|---|---|
| RF-37 | Dar de baja **facturas** mediante comunicación de baja. | M | RN-04 |
| RF-38 | Dar de baja **boletas** incluyéndolas con estado "anulado" en el resumen diario del día de emisión. | M | RN-04 |
| RF-39 | Emitir notas de crédito y de débito vinculadas a un comprobante. | S | RN-04 |
| RF-40 | Al anular o emitir nota de crédito por devolución, revertir stock según corresponda. | S | RN-05 |
| RF-41 | Solo el rol Administrador puede anular; el sistema exige motivo. | M | RN-10 |

### 1.8 Entrega, consulta y conservación

| ID | Requerimiento | Prioridad | Trazab. |
|---|---|---|---|
| RF-42 | Generar la representación impresa en **PDF** (con código QR y datos requeridos) y permitir su descarga. | M | — |
| RF-43 | Permitir la descarga del **XML firmado** y del **CDR**. | M | RN-08 |
| RF-44 | Enviar el comprobante al correo del cliente. | S | — |
| RF-45 | Ofrecer una página web de consulta segura para que el cliente acceda a sus comprobantes por al menos un año desde la emisión. | M | RN-08 |
| RF-46 | Almacenar comprobantes, resúmenes diarios, comunicaciones de baja y constancias de rechazo durante el plazo de conservación aplicable. | M | RN-08 |
| RF-47 | Buscar y filtrar comprobantes por fecha, tipo, serie, número, cliente y estado. | M | — |

### 1.9 Auditoría y reportes básicos

| ID | Requerimiento | Prioridad | Trazab. |
|---|---|---|---|
| RF-48 | Registrar bitácora de auditoría de acciones sensibles (emisión, anulación, cambio de precio, ajuste de stock, cambio de certificado) con usuario, fecha y detalle. | M | RN-10 |
| RF-49 | Reporte diario de ventas por tipo de comprobante y medio de pago. | S | — |
| RF-50 | Exportar listados de comprobantes a CSV/Excel. | S | — |

---

## 2. Requerimientos no funcionales

### 2.1 Rendimiento

| ID | Requerimiento | Meta propuesta |
|---|---|---|
| RNF-01 | Tiempo de respuesta de la emisión en caja (hasta persistir el comprobante y devolver el PDF), sin contar la validación de SUNAT, que es asíncrona. | p95 ≤ 2 s bajo carga normal |
| RNF-02 | Búsqueda de productos en caja. | p95 ≤ 300 ms con catálogo de 50 000 productos |
| RNF-03 | Capacidad de ventas simultáneas por empresa sin degradación. | ≥ 20 emisiones/s en pruebas de carga |

### 2.2 Integridad y concurrencia

| ID | Requerimiento | Criterio de verificación |
|---|---|---|
| RNF-04 | El stock nunca queda negativo ante ventas simultáneas del mismo producto. | Prueba con N hilos: exactamente `stock_inicial` ventas exitosas |
| RNF-05 | Los correlativos son únicos y consecutivos aun con emisiones concurrentes y fallos parciales. | Sin huecos ni duplicados tras pruebas de concurrencia y rollback |
| RNF-06 | La emisión (correlativo, stock, comprobante) es atómica: o se completa todo o nada. | Pruebas de fallo inyectado en cada paso |
| RNF-07 | Los montos se manejan con tipos decimales exactos, nunca punto flotante. | Revisión de código y pruebas de redondeo |

### 2.3 Seguridad

| ID | Requerimiento |
|---|---|
| RNF-08 | Toda comunicación cliente-servidor y con servicios externos usa TLS. |
| RNF-09 | Certificados digitales y credenciales SOL/OSE se almacenan cifrados en reposo; nunca en el repositorio ni en logs. |
| RNF-10 | Las contraseñas se almacenan con hash adaptativo (Argon2/bcrypt) con sal. |
| RNF-11 | Aislamiento entre empresas garantizado a nivel de datos (`empresa_id` obligatorio y, de ser posible, Row-Level Security). |
| RNF-12 | Protección contra las vulnerabilidades más comunes (inyección SQL, XSS, CSRF, control de acceso roto), verificada con revisión o escaneo. |
| RNF-13 | La página de consulta del cliente exige un mecanismo que garantice que solo el receptor accede a sus comprobantes. |

### 2.4 Disponibilidad y tolerancia a fallos

| ID | Requerimiento | Meta propuesta |
|---|---|---|
| RNF-14 | El fallo o lentitud de SUNAT/OSE no bloquea la venta ni la emisión en caja. | 0 ventas bloqueadas por indisponibilidad externa |
| RNF-15 | Los envíos pendientes sobreviven a reinicios del sistema (cola persistente). | Sin pérdida de envíos tras reinicio |
| RNF-16 | Disponibilidad del sistema en horario comercial. | ≥ 99 % mensual |
| RNF-17 | Respaldo automático de la base de datos con restauración probada. | RPO ≤ 24 h; restauración verificada |

### 2.5 Cumplimiento normativo

| ID | Requerimiento |
|---|---|
| RNF-18 | El XML cumple UBL 2.1 y las validaciones de SUNAT; se verifica contra el ambiente de pruebas (beta). |
| RNF-19 | Los plazos de envío (3 días facturas, 7 días boletas vía resumen diario) se controlan y se alerta antes de su vencimiento. |
| RNF-20 | Las reglas normativas (umbrales, plazos, tasas) son parametrizables y no están codificadas de forma dispersa, para absorber cambios de SUNAT. |

### 2.6 Usabilidad

| ID | Requerimiento | Meta propuesta |
|---|---|---|
| RNF-21 | La pantalla de caja se opera con teclado y con pantalla táctil. | Venta simple completa sin usar el mouse |
| RNF-22 | Un cajero nuevo puede emitir su primer comprobante sin manual. | ≤ 15 min de inducción |
| RNF-23 | Los mensajes de error indican causa y acción correctiva en lenguaje del usuario (ej. "RUC inválido", "stock insuficiente: quedan 3"). | Revisión heurística |
| RNF-24 | Interfaz responsive y accesible (contraste, tamaños táctiles, etiquetas). | Nivel WCAG AA como objetivo |

### 2.7 Mantenibilidad y calidad

| ID | Requerimiento | Meta propuesta |
|---|---|---|
| RNF-25 | Arquitectura limpia con dependencias hacia el dominio; el dominio no depende de frameworks, BD ni servicios externos. | Verificable con análisis de dependencias |
| RNF-26 | Puertos y adaptadores para firma, SUNAT/OSE, PDF y persistencia, con implementación simulada (fake) para pruebas. | Flujo de emisión probado de extremo a extremo con fake |
| RNF-27 | Cobertura de pruebas unitarias alta en cálculo de impuestos y reglas del dominio. | ≥ 90 % líneas/ramas en dominio de facturación |
| RNF-28 | Pruebas de propiedades sobre los totales (`total = subtotal + IGV`, independencia del orden de líneas). | Suite property-based |
| RNF-29 | Integración continua que ejecuta pruebas y análisis estático en cada cambio. | Pipeline obligatorio antes de fusionar |

### 2.8 Observabilidad y auditoría

| ID | Requerimiento |
|---|---|
| RNF-30 | Logs estructurados con identificador de correlación por venta/comprobante, sin datos sensibles. |
| RNF-31 | Métricas de negocio y operación: comprobantes por estado, tasa de rechazo, latencia de envío a CDR, cola pendiente. |
| RNF-32 | Trazabilidad completa del ciclo de vida de cada comprobante (quién, cuándo, qué estado, qué respuesta de SUNAT). |

### 2.9 Escalabilidad y portabilidad

| ID | Requerimiento |
|---|---|
| RNF-33 | Alta de una nueva empresa sin cambios de código ni despliegue. |
| RNF-34 | Componentes sin estado (o con estado externalizado) para permitir escalado horizontal. |
| RNF-35 | Despliegue reproducible con contenedores y configuración por variables de entorno. |

### 2.10 Datos y pruebas

| ID | Requerimiento |
|---|---|
| RNF-36 | Los datos de prueba (productos, precios, clientes) se generan con scripts automatizados (ej. Faker), con RUC/DNI de dígito verificador válido. |
| RNF-37 | Los datos sintéticos son reproducibles (semilla fija) y nunca contienen datos reales de personas ni empresas. |

---

## 3. Evolución prevista (fuera del alcance inicial)

| ID | Requerimiento | Prioridad |
|---|---|---|
| RF-51 | Integración con pasarelas de pago electrónicas y conciliación con la venta. | C |
| RF-52 | Módulo de reportes analíticos de ventas con gráficos dinámicos (por producto, período, cajero, sede). | C |
| RF-53 | Facturación recurrente (suscripciones) con generación programada de comprobantes. | C |
| RF-54 | Guías de remisión electrónicas. | C |
| RF-55 | Emisión offline en punto de venta con sincronización posterior. | C |

---

## 4. Matriz resumen

| Categoría | Cantidad |
|---|---|
| RF Must | 36 |
| RF Should | 10 |
| RF Could (alcance inicial) | 4 |
| RF de evolución (fuera de alcance) | 5 |
| RNF | 37 |

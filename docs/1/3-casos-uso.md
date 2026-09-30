# Módulos y Casos de Uso

Formato elegido: **casos de uso** (aplicado de forma consistente a todos los módulos).
Trazabilidad: `RF-xx` (requerimientos funcionales) y `RN-xx` (reglas de negocio).

---

## 1. Módulos del sistema

| ID | Módulo | Descripción | RF cubiertos | Casos de uso |
|---|---|---|---|---|
| M1 | **Seguridad y Configuración** | Autenticación, roles, usuarios, datos de la empresa emisora, series, certificado digital, credenciales de envío y bitácora de auditoría. Sostiene el aislamiento entre empresas. | RF-01 a RF-06, RF-48 | CU-01 a CU-05 |
| M2 | **Clientes** | Registro, validación (RUC/DNI), búsqueda y mantenimiento de clientes. | RF-07 a RF-10 | CU-06 |
| M3 | **Catálogo e Inventario** | Gestión de productos, afectación de IGV, stock, movimientos, ajustes y carga masiva. | RF-11 a RF-15 | CU-07 a CU-09 |
| M4 | **Ventas y Caja** | Interfaz de caja: armado de la venta, cálculo en tiempo real, identificación del cliente, medio de pago y confirmación. | RF-16 a RF-21 | CU-10 |
| M5 | **Emisión de Comprobantes** | Núcleo de dominio de facturación: reglas de identificación, cálculo de impuestos, numeración, descuento atómico de stock, generación y firma del XML UBL 2.1 y notas de crédito. | RF-22 a RF-29, RF-39, RF-40 | CU-11 a CU-13 |
| M6 | **Integración Tributaria** | Envío asíncrono a SUNAT/OSE, resumen diario de boletas, procesamiento de CDR, bajas, reintentos, estados y alertas de plazo. | RF-30 a RF-38, RF-41 | CU-14 a CU-18 |
| M7 | **Documentos, Consulta y Reportes** | Representación PDF, descarga de XML/CDR, búsqueda de comprobantes, consulta web del cliente, conservación y reportes básicos. | RF-42 a RF-47, RF-49, RF-50 | CU-19 a CU-22 |

> Si el grupo tiene menos de 7 integrantes, se pueden fusionar M2 con M3 (datos maestros) y M7 con M4 sin romper las dependencias.

**Dependencias principales:** M4 → M5 (`<<include>>` de emisión) → M3 (stock) y M2 (cliente); M5 → M6 (cola de envío); M6 → M7 (estados, CDR); M1 es transversal.

### Resumen de casos de uso

| CU | Nombre | Módulo | Prioridad |
|---|---|---|---|
| CU-01 | Iniciar sesión | M1 | M |
| CU-02 | Gestionar usuarios y roles | M1 | M |
| CU-03 | Configurar empresa y series | M1 | M |
| CU-04 | Registrar certificado digital y credenciales de envío | M1 | M |
| CU-05 | Consultar bitácora de auditoría | M1 | M |
| CU-06 | Gestionar clientes | M2 | M |
| CU-07 | Gestionar productos | M3 | M |
| CU-08 | Ajustar stock | M3 | S |
| CU-09 | Cargar catálogo masivo | M3 | S |
| CU-10 | Registrar venta en caja | M4 | M |
| CU-11 | Emitir boleta de venta electrónica | M5 | M |
| CU-12 | Emitir factura electrónica | M5 | M |
| CU-13 | Emitir nota de crédito | M5 | S |
| CU-14 | Enviar factura a SUNAT/OSE | M6 | M |
| CU-15 | Enviar resumen diario de boletas | M6 | M |
| CU-16 | Dar de baja factura | M6 | M |
| CU-17 | Dar de baja boleta | M6 | M |
| CU-18 | Monitorear y reenviar comprobantes | M6 | M |
| CU-19 | Descargar PDF, XML y CDR | M7 | M |
| CU-20 | Buscar comprobantes | M7 | M |
| CU-21 | Consultar comprobante (cliente final) | M7 | M |
| CU-22 | Generar reporte diario de ventas | M7 | S |

---

## 2. Casos de uso

## Módulo M1: Seguridad y Configuración

### CU-01: Iniciar sesión

- **Actor(es):** Cajero, Administrador.
- **Objetivo:** Acceder al sistema con las funciones correspondientes a su rol y empresa.
- **Precondiciones:** El usuario existe, está activo y pertenece a una empresa registrada.
- **Disparador:** El usuario abre la aplicación e ingresa a la pantalla de acceso.
- **Flujo principal:**
  1. El usuario ingresa su identificador y contraseña.
  2. El sistema valida las credenciales.
  3. El sistema identifica la empresa y el rol del usuario.
  4. El sistema crea la sesión con tiempo de expiración configurable.
  5. El sistema muestra la pantalla inicial según el rol (caja para cajero, panel para administrador).
- **Flujos alternativos:**
  - A1. El usuario cierra sesión manualmente; el sistema invalida la sesión.
  - A2. La sesión expira por inactividad; el sistema solicita autenticarse de nuevo.
- **Postcondiciones:** Existe una sesión activa asociada a una empresa y un rol; el acceso queda en bitácora.
- **Excepciones:**
  - E1. Credenciales inválidas: el sistema muestra un mensaje genérico (sin indicar cuál dato falló).
  - E2. Usuario desactivado: el sistema deniega el acceso.
  - E3. Superado el máximo de intentos fallidos: el sistema bloquea temporalmente la cuenta.
- **Trazabilidad:** RF-01, RF-02, RF-03, RNF-10.

### CU-02: Gestionar usuarios y roles

- **Actor(es):** Administrador.
- **Objetivo:** Crear y mantener los usuarios de su empresa con el rol adecuado.
- **Precondiciones:** Sesión activa de Administrador.
- **Disparador:** El administrador ingresa al módulo de usuarios.
- **Flujo principal:**
  1. El sistema lista los usuarios de la empresa.
  2. El administrador selecciona "Nuevo usuario".
  3. Ingresa nombre, correo y rol (Cajero o Administrador).
  4. El sistema valida los datos y la unicidad del correo dentro de la empresa.
  5. El sistema crea el usuario con una contraseña temporal y registra la acción en bitácora.
- **Flujos alternativos:**
  - A1. Editar datos o cambiar el rol de un usuario.
  - A2. Desactivar o reactivar un usuario.
  - A3. Restablecer contraseña.
- **Postcondiciones:** El usuario queda registrado o actualizado y el cambio queda auditado.
- **Excepciones:**
  - E1. Correo duplicado: el sistema rechaza el registro.
  - E2. Se intenta desactivar o degradar al único administrador: el sistema lo impide.
- **Trazabilidad:** RF-02, RF-48, RN-10.

### CU-03: Configurar empresa y series

- **Actor(es):** Administrador.
- **Objetivo:** Definir los datos del emisor y las series de comprobantes.
- **Precondiciones:** Sesión activa de Administrador.
- **Disparador:** El administrador ingresa a la configuración de empresa.
- **Flujo principal:**
  1. El administrador ingresa RUC, razón social, domicilio fiscal, régimen y logo.
  2. El sistema valida el RUC (formato y dígito verificador).
  3. El administrador crea una serie indicando tipo de comprobante y código (F### para facturas, B### para boletas).
  4. El sistema valida el formato y la unicidad de la serie dentro de la empresa.
  5. El sistema guarda la serie con correlativo inicial en 1 y la deja activa.
- **Flujos alternativos:**
  - A1. Desactivar una serie sin comprobantes pendientes.
  - A2. Asignar series a distintos puntos de venta.
- **Postcondiciones:** La empresa tiene datos fiscales y al menos una serie activa por tipo de comprobante.
- **Excepciones:**
  - E1. RUC inválido: el sistema rechaza el dato.
  - E2. Serie duplicada o con formato incorrecto: el sistema lo notifica.
  - E3. Se intenta modificar o eliminar una serie con comprobantes emitidos: el sistema lo impide.
- **Trazabilidad:** RF-04, RF-06, RN-02.

### CU-04: Registrar certificado digital y credenciales de envío

- **Actor(es):** Administrador.
- **Objetivo:** Habilitar la firma digital y el envío a SUNAT/OSE para su empresa.
- **Precondiciones:** Sesión activa de Administrador; empresa con RUC registrado.
- **Disparador:** El administrador ingresa a la sección de certificado y credenciales.
- **Flujo principal:**
  1. El administrador carga el certificado digital y su clave.
  2. El sistema verifica que el certificado esté vigente y corresponda al RUC de la empresa.
  3. El administrador registra las credenciales de envío (SOL, OSE o PSE).
  4. El sistema cifra y almacena certificado y credenciales.
  5. El sistema ejecuta una prueba de conexión contra el servicio de validación.
  6. El sistema marca la empresa como habilitada para emitir.
- **Flujos alternativos:**
  - A1. Reemplazar un certificado por otro vigente.
  - A2. Seleccionar un PSE en lugar de firma propia.
  - A3. El sistema alerta con antelación configurable el vencimiento del certificado.
- **Postcondiciones:** La empresa queda habilitada para firmar y enviar comprobantes; la acción queda en bitácora.
- **Excepciones:**
  - E1. Clave incorrecta o archivo dañado: el sistema rechaza la carga.
  - E2. Certificado vencido o de otro RUC: el sistema rechaza la carga.
  - E3. Falla la prueba de conexión: la empresa queda "pendiente de verificación" y no puede emitir.
- **Trazabilidad:** RF-05, RN-06, RNF-09.

### CU-05: Consultar bitácora de auditoría

- **Actor(es):** Administrador.
- **Objetivo:** Revisar las acciones sensibles realizadas en el sistema.
- **Precondiciones:** Sesión activa de Administrador.
- **Disparador:** El administrador ingresa al módulo de auditoría.
- **Flujo principal:**
  1. El administrador define filtros (rango de fechas, usuario, tipo de acción).
  2. El sistema muestra los eventos (emisión, anulación, cambio de precio, ajuste de stock, cambio de certificado) con usuario, fecha y detalle.
  3. El administrador abre un evento para ver el detalle completo.
- **Flujos alternativos:**
  - A1. Exportar el resultado a CSV.
- **Postcondiciones:** Ninguna modificación de datos; la consulta no altera la bitácora.
- **Excepciones:**
  - E1. Sin resultados: el sistema informa que no hay eventos para los filtros.
  - E2. Rango demasiado amplio: el sistema pide acotarlo.
- **Trazabilidad:** RF-48, RNF-32.

---

## Módulo M2: Clientes

### CU-06: Gestionar clientes

- **Actor(es):** Cajero, Administrador; servicio externo de consulta de RUC/DNI (opcional).
- **Objetivo:** Mantener el registro de clientes para identificarlos en los comprobantes.
- **Precondiciones:** Sesión activa.
- **Disparador:** El usuario ingresa al módulo de clientes o registra uno desde la caja.
- **Flujo principal:**
  1. El usuario selecciona "Nuevo cliente".
  2. Ingresa tipo y número de documento (DNI o RUC).
  3. El sistema valida formato y dígito verificador.
  4. El usuario completa nombre o razón social, dirección y correo.
  5. El sistema verifica que no exista otro cliente con el mismo documento en la empresa.
  6. El sistema guarda el cliente.
- **Flujos alternativos:**
  - A1. Buscar clientes por documento o nombre.
  - A2. Editar datos de un cliente existente.
  - A3. Desactivar un cliente (no se elimina si tiene comprobantes).
  - A4. Autocompletar datos desde el servicio externo por RUC/DNI.
- **Postcondiciones:** El cliente queda disponible para seleccionarlo en ventas.
- **Excepciones:**
  - E1. Documento con formato o dígito verificador inválido: el sistema lo rechaza.
  - E2. Documento duplicado en la empresa: el sistema ofrece abrir el existente.
  - E3. Servicio externo no disponible: el sistema permite ingreso manual.
- **Trazabilidad:** RF-07, RF-08, RF-09, RF-10, RN-01.

---

## Módulo M3: Catálogo e Inventario

### CU-07: Gestionar productos

- **Actor(es):** Administrador.
- **Objetivo:** Mantener el catálogo con precios, afectación de IGV y control de stock.
- **Precondiciones:** Sesión activa de Administrador.
- **Disparador:** El administrador ingresa al catálogo.
- **Flujo principal:**
  1. El administrador selecciona "Nuevo producto".
  2. Ingresa código, descripción, unidad de medida, precio, afectación de IGV (gravado, exonerado o inafecto) e indica si controla stock (con stock inicial).
  3. El sistema valida unicidad del código, precio mayor a cero y afectación válida.
  4. El sistema guarda el producto y, si corresponde, registra el movimiento de stock inicial.
- **Flujos alternativos:**
  - A1. Editar un producto: el cambio de precio queda en bitácora y no afecta comprobantes ya emitidos.
  - A2. Desactivar un producto: deja de aparecer en caja pero se conserva en el historial.
  - A3. Buscar productos por código o descripción.
- **Postcondiciones:** El producto queda disponible para su venta con los datos tributarios definidos.
- **Excepciones:**
  - E1. Código duplicado: el sistema lo rechaza.
  - E2. Precio o stock inicial inválido (cero o negativo): el sistema lo rechaza.
- **Trazabilidad:** RF-11, RF-12, RN-03, RN-10.

### CU-08: Ajustar stock

- **Actor(es):** Administrador.
- **Objetivo:** Corregir el stock por mermas, conteos o ingresos, con trazabilidad.
- **Precondiciones:** Sesión activa de Administrador; el producto controla stock.
- **Disparador:** El administrador selecciona "Ajustar stock" en un producto.
- **Flujo principal:**
  1. El administrador elige tipo de ajuste (entrada o salida) e ingresa cantidad.
  2. Ingresa el motivo (obligatorio).
  3. El sistema verifica que el stock resultante no sea negativo.
  4. El sistema aplica el ajuste y registra el movimiento con usuario, fecha y motivo.
- **Flujos alternativos:**
  - A1. Consultar el historial de movimientos del producto.
- **Postcondiciones:** Stock actualizado y movimiento trazable en bitácora.
- **Excepciones:**
  - E1. Stock resultante negativo: el sistema rechaza el ajuste.
  - E2. Motivo vacío: el sistema exige completarlo.
  - E3. Conflicto por venta simultánea sobre el mismo producto: el sistema reintenta o solicita repetir la operación.
- **Trazabilidad:** RF-12, RF-13, RN-05.

### CU-09: Cargar catálogo masivo

- **Actor(es):** Administrador.
- **Objetivo:** Incorporar muchos productos de una sola vez desde un archivo.
- **Precondiciones:** Sesión activa de Administrador; archivo CSV/Excel con la plantilla vigente.
- **Disparador:** El administrador selecciona "Importar productos".
- **Flujo principal:**
  1. El administrador carga el archivo.
  2. El sistema valida la estructura de columnas.
  3. El sistema valida cada fila (código único, precio, afectación de IGV, unidad).
  4. El sistema muestra una vista previa con filas válidas y con error.
  5. El administrador confirma la importación.
  6. El sistema inserta las filas válidas y entrega un reporte de resultados.
- **Flujos alternativos:**
  - A1. Descargar la plantilla del archivo.
  - A2. Cancelar tras la vista previa; no se guarda nada.
- **Postcondiciones:** Productos válidos incorporados al catálogo; reporte de errores disponible.
- **Excepciones:**
  - E1. Estructura o formato de archivo inválido: el sistema rechaza la carga.
  - E2. Archivo que excede el tamaño máximo: el sistema pide dividirlo.
  - E3. Filas con error: se omiten y se reportan con el motivo.
- **Trazabilidad:** RF-15, RNF-36.

---

## Módulo M4: Ventas y Caja

### CU-10: Registrar venta en caja

- **Actor(es):** Cajero.
- **Objetivo:** Armar y confirmar una venta para emitir su comprobante en el menor tiempo posible.
- **Precondiciones:** Sesión activa de Cajero; empresa habilitada para emitir (CU-04); serie activa; catálogo cargado.
- **Disparador:** El cajero inicia una nueva venta.
- **Flujo principal:**
  1. El cajero agrega productos por búsqueda, código o escáner, con su cantidad.
  2. El sistema calcula subtotal, IGV y total en tiempo real.
  3. El cajero selecciona el tipo de comprobante (boleta o factura).
  4. El cajero identifica al cliente (opcional en boleta hasta S/ 700.00; obligatorio en factura).
  5. El cajero selecciona el medio de pago.
  6. El cajero confirma la venta.
  7. El sistema ejecuta la emisión (`<<include>>` CU-11 o CU-12).
  8. El sistema muestra el comprobante emitido para su impresión o descarga.
- **Flujos alternativos:**
  - A1. Editar cantidades o quitar líneas antes de confirmar.
  - A2. Aplicar descuento por línea o global dentro del límite configurado.
  - A3. Dejar la venta en espera y retomarla luego.
  - A4. Cancelar la venta antes de confirmarla; no se consume stock ni correlativo.
- **Postcondiciones:** Venta registrada y vinculada a un comprobante emitido; stock descontado.
- **Excepciones:**
  - E1. Stock insuficiente en una línea: el sistema informa el disponible y no permite confirmar esa línea.
  - E2. Boleta con total mayor a S/ 700.00 sin documento del cliente: el sistema exige identificarlo.
  - E3. Empresa sin certificado vigente o sin serie activa: el sistema bloquea la emisión y avisa al administrador.
  - E4. Fallo o reintento al confirmar: la clave de idempotencia evita emitir un comprobante duplicado.
- **Trazabilidad:** RF-16 a RF-21, RN-01, RN-03, RN-05.

---

## Módulo M5: Emisión de Comprobantes

### CU-11: Emitir boleta de venta electrónica

- **Actor(es):** Cajero (a través de CU-10), Sistema.
- **Objetivo:** Generar una boleta válida, con numeración correcta y stock consistente.
- **Precondiciones:** Venta confirmada con al menos una línea; serie de boletas activa; certificado vigente.
- **Disparador:** Confirmación de la venta con tipo "boleta" (CU-10).
- **Flujo principal:**
  1. El sistema valida la regla de identificación: si el total supera S/ 700.00 o el cliente lo solicita, exige tipo y número de documento.
  2. El sistema calcula los impuestos por línea según su afectación.
  3. El sistema abre una transacción única.
  4. El sistema descuenta el stock de forma atómica por cada línea.
  5. El sistema toma el siguiente correlativo de la serie con bloqueo.
  6. El sistema genera el XML UBL 2.1 y lo firma digitalmente.
  7. El sistema persiste el comprobante en estado PENDIENTE y lo registra para el resumen diario.
  8. El sistema confirma la transacción.
  9. El sistema genera la representación PDF con código QR y la devuelve.
- **Flujos alternativos:**
  - A1. Ventas de hasta S/ 5.00 consolidadas en una boleta (si la empresa lo habilita).
  - A2. El cliente pide el comprobante por correo electrónico.
- **Postcondiciones:** Boleta emitida, correlativo consumido, stock descontado y comprobante pendiente de informarse en el resumen diario.
- **Excepciones:**
  - E1. Stock insuficiente en algún producto: se revierte toda la transacción; no se consume correlativo.
  - E2. Falla en la firma (certificado inválido): se revierte la transacción; no se consume correlativo.
  - E3. Falla al generar el PDF: el comprobante queda emitido y el PDF se regenera luego.
  - E4. Reintento con la misma clave de idempotencia: el sistema devuelve la boleta ya emitida.
- **Trazabilidad:** RF-22 a RF-28, RN-01 a RN-06, RNF-04 a RNF-07.

### CU-12: Emitir factura electrónica

- **Actor(es):** Cajero (a través de CU-10), Sistema.
- **Objetivo:** Generar una factura válida para un cliente con RUC.
- **Precondiciones:** Venta confirmada con al menos una línea; cliente con RUC válido; serie de facturas activa; certificado vigente.
- **Disparador:** Confirmación de la venta con tipo "factura" (CU-10).
- **Flujo principal:**
  1. El sistema valida que el cliente tenga RUC válido.
  2. El sistema calcula los impuestos por línea según su afectación.
  3. El sistema abre una transacción única.
  4. El sistema descuenta el stock de forma atómica por cada línea.
  5. El sistema toma el siguiente correlativo de la serie con bloqueo.
  6. El sistema genera el XML UBL 2.1 y lo firma digitalmente.
  7. El sistema persiste la factura en estado PENDIENTE y la encola para su envío (CU-14).
  8. El sistema confirma la transacción.
  9. El sistema genera el PDF con código QR y lo devuelve.
- **Flujos alternativos:**
  - A1. El cliente pide la factura por correo electrónico.
- **Postcondiciones:** Factura emitida y encolada para envío a SUNAT/OSE.
- **Excepciones:**
  - E1. Cliente sin RUC o RUC inválido: el sistema no permite emitir factura.
  - E2. Stock insuficiente o falla en la firma: se revierte toda la transacción; no se consume correlativo.
  - E3. Reintento con la misma clave de idempotencia: el sistema devuelve la factura ya emitida.
- **Trazabilidad:** RF-22 a RF-28, RN-01 a RN-06.

### CU-13: Emitir nota de crédito

- **Actor(es):** Administrador, Sistema.
- **Objetivo:** Corregir o anular parcial o totalmente un comprobante ya emitido sin modificarlo.
- **Precondiciones:** Comprobante origen emitido y no rechazado; serie de notas activa.
- **Disparador:** El administrador selecciona "Emitir nota de crédito" sobre un comprobante.
- **Flujo principal:**
  1. El administrador busca y selecciona el comprobante origen.
  2. Selecciona el motivo (catálogo de SUNAT) e indica ítems o monto a acreditar.
  3. El sistema valida que lo acreditado no exceda el comprobante original.
  4. El sistema calcula los impuestos de la nota.
  5. El sistema toma el correlativo de la serie de notas.
  6. El sistema genera y firma el XML, y persiste la nota vinculada al comprobante origen.
  7. Si el motivo es devolución, el sistema revierte el stock correspondiente.
  8. El sistema encola la nota para su envío (CU-14 o CU-15 según el origen).
- **Flujos alternativos:**
  - A1. Nota de crédito total que sustituye la anulación del comprobante.
- **Postcondiciones:** Nota emitida y vinculada; stock ajustado si aplica.
- **Excepciones:**
  - E1. Monto o cantidad mayor al original: el sistema rechaza la nota.
  - E2. Comprobante origen rechazado por SUNAT: el sistema indica emitir un comprobante nuevo.
  - E3. Falla de firma: se revierte y no se consume correlativo.
- **Trazabilidad:** RF-39, RF-40, RN-04.

---

## Módulo M6: Integración Tributaria

### CU-14: Enviar factura a SUNAT/OSE

- **Actor(es):** Sistema (proceso automático); SUNAT/OSE (actor externo).
- **Objetivo:** Informar la factura dentro del plazo y registrar el resultado de validación.
- **Precondiciones:** Factura firmada en estado PENDIENTE; empresa habilitada para el envío.
- **Disparador:** Ingreso de la factura a la cola de envío o ejecución del planificador.
- **Flujo principal:**
  1. El sistema toma las facturas pendientes de la cola.
  2. El sistema comprime el XML firmado y lo envía a SUNAT/OSE.
  3. SUNAT/OSE valida y responde con la constancia de recepción (CDR).
  4. El sistema almacena el CDR.
  5. El sistema actualiza el estado de la factura (ACEPTADO, OBSERVADO o RECHAZADO).
  6. El sistema registra el intento en la trazabilidad del comprobante.
- **Flujos alternativos:**
  - A1. Aceptada con observaciones: el estado queda OBSERVADO y se muestran las advertencias.
  - A2. El administrador dispara el envío manual desde CU-18.
- **Postcondiciones:** La factura tiene un estado definitivo o sigue pendiente de reintento; el CDR queda almacenado.
- **Excepciones:**
  - E1. Servicio no disponible o tiempo de espera agotado: el sistema reintenta con espera progresiva y la factura sigue PENDIENTE.
  - E2. Rechazo: estado RECHAZADO con el motivo, y alerta al administrador.
  - E3. Plazo de envío próximo a vencer (3 días calendario): el sistema genera alerta prioritaria.
  - E4. Credenciales o certificado inválidos: el sistema marca error de configuración y notifica al administrador.
- **Trazabilidad:** RF-30, RF-32 a RF-36, RN-07, RNF-14, RNF-15.

### CU-15: Enviar resumen diario de boletas

- **Actor(es):** Sistema (proceso programado), Administrador (envío manual); SUNAT (actor externo).
- **Objetivo:** Informar las boletas emitidas en un día mediante resumen diario dentro del plazo.
- **Precondiciones:** Existen boletas o notas asociadas emitidas ese día aún no informadas.
- **Disparador:** Ejecución diaria programada o solicitud manual del administrador.
- **Flujo principal:**
  1. El sistema agrupa las boletas y notas asociadas por día de emisión.
  2. El sistema genera el XML del resumen con su identificador correlativo.
  3. El sistema firma el resumen y lo envía a SUNAT.
  4. El sistema obtiene el ticket y consulta el resultado.
  5. SUNAT responde con la constancia del resumen (ACEPTADA o RECHAZADA).
  6. El sistema almacena la constancia y actualiza el estado de las boletas incluidas.
- **Flujos alternativos:**
  - A1. Enviar más de un resumen para el mismo día de emisión.
  - A2. Modificar lo informado incluyendo las boletas corregidas en otro resumen del mismo día de emisión.
  - A3. Incluir boletas dadas de baja con estado "anulado" (ver CU-17).
- **Postcondiciones:** Boletas del día informadas con estado definido; constancia almacenada.
- **Excepciones:**
  - E1. Resumen rechazado: el sistema registra el motivo, permite corregir y generar uno nuevo.
  - E2. SUNAT no disponible: el sistema reintenta hasta cumplir el plazo (7 días calendario desde el día siguiente).
  - E3. Plazo próximo a vencer: el sistema alerta al administrador.
- **Trazabilidad:** RF-31, RF-33 a RF-35, RN-07.

### CU-16: Dar de baja factura

- **Actor(es):** Administrador; SUNAT/OSE (actor externo).
- **Objetivo:** Anular una factura mediante comunicación de baja.
- **Precondiciones:** Factura aceptada por SUNAT, no dada de baja y dentro del plazo normativo aplicable.
- **Disparador:** El administrador selecciona "Dar de baja" en una factura.
- **Flujo principal:**
  1. El administrador busca la factura e ingresa el motivo de la baja.
  2. El sistema valida que se pueda dar de baja.
  3. El sistema genera y firma la comunicación de baja y la envía.
  4. SUNAT/OSE responde con la constancia.
  5. El sistema actualiza la factura a estado ANULADA y registra la acción en bitácora.
- **Flujos alternativos:**
  - A1. Si la baja no corresponde por plazo o condición, el sistema sugiere emitir una nota de crédito (CU-13).
- **Postcondiciones:** Factura en estado ANULADA con trazabilidad completa.
- **Excepciones:**
  - E1. Factura no aceptada o ya anulada: el sistema rechaza la solicitud.
  - E2. Motivo vacío: el sistema exige completarlo.
  - E3. Rechazo de SUNAT: la factura conserva su estado y se muestra el motivo.
- **Trazabilidad:** RF-37, RF-41, RN-04, RN-10.

### CU-17: Dar de baja boleta

- **Actor(es):** Administrador, Sistema; SUNAT (actor externo).
- **Objetivo:** Anular una boleta incluyéndola con estado "anulado" en el resumen diario.
- **Precondiciones:** Boleta emitida, no dada de baja y dentro del plazo (hasta 7 días calendario, contados desde el día siguiente a su envío o generación).
- **Disparador:** El administrador selecciona "Dar de baja" en una boleta.
- **Flujo principal:**
  1. El administrador busca la boleta e ingresa el motivo.
  2. El sistema valida que sea posible la baja.
  3. El sistema marca la boleta como PENDIENTE DE BAJA.
  4. El sistema la incluye con estado "anulado" en el resumen diario correspondiente a su día de emisión (CU-15).
  5. Al aceptarse el resumen, el sistema actualiza la boleta a ANULADA y registra la acción en bitácora.
- **Flujos alternativos:**
  - A1. Si la baja no corresponde, el sistema sugiere una nota de crédito (CU-13).
- **Postcondiciones:** Boleta en estado ANULADA con trazabilidad completa.
- **Excepciones:**
  - E1. Fuera de plazo o ya anulada: el sistema rechaza la solicitud.
  - E2. Boleta con nota de crédito vinculada: el sistema informa la restricción.
  - E3. Resumen rechazado: la boleta vuelve a su estado previo y se notifica.
- **Trazabilidad:** RF-38, RF-41, RN-04, RN-10.

### CU-18: Monitorear y reenviar comprobantes

- **Actor(es):** Administrador (Cajero solo consulta).
- **Objetivo:** Supervisar el estado de los envíos, resolver fallos y evitar vencimientos de plazo.
- **Precondiciones:** Sesión activa.
- **Disparador:** El usuario ingresa al panel de estados o recibe una alerta.
- **Flujo principal:**
  1. El sistema muestra los comprobantes agrupados por estado (pendiente, enviado, aceptado, observado, rechazado).
  2. El administrador filtra por estado, fecha o tipo.
  3. El administrador abre un comprobante para ver intentos de envío, respuestas y motivo de rechazo u observación.
  4. El administrador selecciona "Reenviar" en un comprobante pendiente o con error de envío.
  5. El sistema reintenta el envío y actualiza el estado.
- **Flujos alternativos:**
  - A1. Para un comprobante rechazado, el sistema guía a emitir un comprobante nuevo o una nota (CU-12, CU-13).
  - A2. El sistema muestra alertas de comprobantes próximos a vencer su plazo.
- **Postcondiciones:** Estado actualizado tras el reenvío; toda acción queda trazada.
- **Excepciones:**
  - E1. Comprobante ya aceptado: el sistema no permite reenviarlo.
  - E2. Servicio externo no disponible: el sistema mantiene el estado y programa un nuevo intento.
- **Trazabilidad:** RF-33, RF-34, RF-35, RNF-32.

---

## Módulo M7: Documentos, Consulta y Reportes

### CU-19: Descargar PDF, XML y CDR

- **Actor(es):** Cajero, Administrador.
- **Objetivo:** Obtener la representación impresa y los archivos oficiales de un comprobante.
- **Precondiciones:** Comprobante emitido.
- **Disparador:** El usuario selecciona el comprobante y elige el formato.
- **Flujo principal:**
  1. El usuario localiza el comprobante (CU-20).
  2. Elige el formato: PDF, XML firmado o CDR.
  3. El sistema recupera el archivo almacenado.
  4. El sistema entrega la descarga.
- **Flujos alternativos:**
  - A1. Reimprimir el PDF.
  - A2. Enviar el comprobante por correo al cliente.
- **Postcondiciones:** Archivo entregado sin modificar el comprobante.
- **Excepciones:**
  - E1. CDR aún no disponible (comprobante pendiente): el sistema informa el estado actual.
  - E2. PDF no encontrado: el sistema lo regenera desde los datos del comprobante.
- **Trazabilidad:** RF-42, RF-43, RF-44, RN-08.

### CU-20: Buscar comprobantes

- **Actor(es):** Cajero, Administrador.
- **Objetivo:** Localizar comprobantes emitidos rápidamente.
- **Precondiciones:** Sesión activa.
- **Disparador:** El usuario ingresa al listado de comprobantes.
- **Flujo principal:**
  1. El usuario define filtros: fecha, tipo, serie, número, cliente y estado.
  2. El sistema muestra los resultados paginados con su estado.
  3. El usuario abre el detalle de un comprobante.
- **Flujos alternativos:**
  - A1. Exportar el listado a CSV/Excel.
- **Postcondiciones:** Ninguna modificación de datos.
- **Excepciones:**
  - E1. Sin resultados: el sistema lo informa.
  - E2. Rango de fechas excesivo: el sistema solicita acotarlo.
- **Trazabilidad:** RF-47, RF-50.

### CU-21: Consultar comprobante (cliente final)

- **Actor(es):** Cliente final.
- **Objetivo:** Consultar y descargar un comprobante recibido desde una página web segura.
- **Precondiciones:** Comprobante emitido dentro del período de disponibilidad (al menos un año desde la emisión).
- **Disparador:** El cliente ingresa al enlace o escanea el QR del comprobante.
- **Flujo principal:**
  1. El cliente ingresa los datos de verificación (tipo y número de documento, serie, número, fecha y monto).
  2. El sistema valida que los datos coincidan con un comprobante emitido a ese cliente.
  3. El sistema muestra el comprobante.
  4. El cliente descarga el PDF o el XML.
- **Flujos alternativos:**
  - A1. El cliente consulta otro comprobante.
- **Postcondiciones:** Acceso registrado; el cliente solo accedió a su propia información.
- **Excepciones:**
  - E1. Datos que no coinciden: el sistema muestra un mensaje genérico y limita los intentos.
  - E2. Comprobante fuera del período de disponibilidad: el sistema indica cómo solicitarlo a la empresa.
  - E3. Exceso de intentos fallidos: el sistema bloquea temporalmente la consulta.
- **Trazabilidad:** RF-45, RN-08, RNF-13.

### CU-22: Generar reporte diario de ventas

- **Actor(es):** Administrador.
- **Objetivo:** Conocer el resumen de ventas de un día o rango.
- **Precondiciones:** Sesión activa de Administrador.
- **Disparador:** El administrador ingresa a reportes.
- **Flujo principal:**
  1. El administrador selecciona fecha o rango y filtros opcionales (cajero, sede).
  2. El sistema agrega totales por tipo de comprobante, medio de pago y cajero.
  3. El sistema muestra el reporte con subtotales, IGV y total.
- **Flujos alternativos:**
  - A1. Exportar el reporte a CSV/Excel.
- **Postcondiciones:** Ninguna modificación de datos.
- **Excepciones:**
  - E1. Sin ventas en el período: el sistema muestra el reporte vacío.
  - E2. Comprobantes anulados: se presentan separados y no suman al total vendido.
- **Trazabilidad:** RF-49, RF-50.

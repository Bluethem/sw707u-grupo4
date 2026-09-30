# 1. Descripción del Proceso

<div align="center">
  <img src="../../assets/img/flowfacture_logo.png" alt="Logo FactuGest" width="250">
</div>

**Proceso elegido:** Emisión de comprobantes de pago electrónicos (boletas y facturas) en punto de venta, con control básico de inventario.

## 1.1 Contexto

El proceso opera en el **área de ventas/caja** de un pequeño o mediano comercio (bodegas, ferreterías, boticas, tiendas de ropa, restaurantes). Depende de la **administración del negocio**, que define catálogo, precios y series de comprobantes. Su rol es formalizar cada venta ante la entidad tributaria (SUNAT) y sostener el cumplimiento fiscal. Hoy se ejecuta de forma manual (talonarios, Excel) o con sistemas heredados lentos.

## 1.2 Propósito

- **Problema que resuelve:** demoras en caja y errores humanos (RUC incorrecto, IGV mal calculado, correlativos duplicados o con huecos) que generan rechazos fiscales y riesgo de sanciones.
- **Valor generado:** menor tiempo de atención, comprobantes válidos a la primera, trazabilidad de ventas e inventario consistente, y cumplimiento fiscal oportuno.

## 1.3 Alcance

| | Detalle |
|---|---|
| **Inicio** | El cajero registra una venta (selección de productos y datos del cliente). |
| **Fin** | El comprobante queda en estado final (aceptado/rechazado por SUNAT) y el cliente recibe el PDF/XML. |
| **Incluye** | Registro de clientes y catálogo, cálculo de impuestos, reserva/descuento de stock, generación de XML UBL, firma digital, envío y procesamiento de la constancia de recepción (CDR), generación y descarga de PDF/XML. |
| **Excluye (fase inicial)** | Contabilidad y libros electrónicos, planilla, compras/proveedores, logística de despacho, pasarelas de pago, reportes analíticos, facturación recurrente, guías de remisión. |

## 1.4 Entradas y salidas

| Entradas | Salidas |
|---|---|
| Productos y cantidades vendidas | Comprobante electrónico (XML firmado + PDF) |
| Datos del cliente (DNI/RUC, nombre, dirección) | Constancia de recepción (CDR) de SUNAT |
| Catálogo: precios, afectación de IGV, stock | Stock actualizado |
| Series y correlativos por empresa | Registro de venta con estado y trazabilidad |
| Certificado digital y credenciales de la empresa | Errores/observaciones para reproceso |

## 1.5 Actores y roles

| Actor | Rol en el proceso |
|---|---|
| **Cajero/operador de venta** | Registra la venta y emite el comprobante. |
| **Administrador del negocio** | Gestiona catálogo, precios, stock, series, usuarios; anula y corrige comprobantes. |
| **Cliente final** | Proporciona sus datos, recibe y descarga el comprobante. |
| **Entidad tributaria (SUNAT/OSE)** | Valida el XML y devuelve el CDR (aceptado, rechazado u observado). |
| **Servicio de firma digital** | Firma el XML con el certificado del emisor. |
| **Sistema FlowFlacture** | Orquesta cálculo, numeración, firma, envío, almacenamiento y generación del PDF. |

## 1.6 Reglas de negocio clave

1. **RN-01.** La boleta puede emitirse sin identificar al cliente hasta S/ 700.00 inclusive. Si el total supera S/ 700.00, o el cliente lo solicita, es obligatorio registrar tipo y número de documento (DNI/RUC).
2. **RN-02.** Cada comprobante tiene **serie + correlativo únicos y consecutivos** por empresa y tipo (F### facturas, B### boletas); no se permiten duplicados ni saltos.
3. **RN-03.** El **IGV (18%)** se calcula según la afectación de cada producto (gravado, exonerado, inafecto); los montos usan aritmética decimal con regla de redondeo definida.
4. **RN-04.** Las facturas se anulan por comunicación de baja. Las boletas se dan de baja incluyendo el documento con estado "anulado" en el resumen diario del día de emisión, dentro del plazo de 7 días calendario.
5. **RN-05.** **No se vende sin stock suficiente**: el stock nunca es negativo, incluso con ventas simultáneas.
6. **RN-06.** El XML debe firmarse con el certificado vigente de la empresa y cumplir el estándar UBL 2.1 antes del envío.
7. **RN-07.** Facturas: envío en 3 días calendario. Boletas: resumen diario con plazo de hasta 7 días calendario.
8. **RN-08.** Almacenar comprobantes, resúmenes, bajas y constancias de rechazo, y ofrecer consulta web segura al cliente por al menos un año.
9. **RN-09.** Cada empresa (tenant) solo accede a sus propios datos, series y certificados.
10. **RN-10.** Solo el rol administrador puede anular comprobantes o modificar precios y catálogo.
11. **RN-11.** Las ventas que no excedan S/ 5.00 pueden consolidarse en una boleta (opcional, útil para bodegas).
12. **RN-12.** La boleta no da derecho a crédito fiscal ni sustenta gasto, salvo que se identifique al adquirente con RUC (o DNI/RUC en el caso de deducciones de cuarta y quinta categoría).

## 1.7 Problemas actuales y oportunidades

| Problema | Consecuencia | Oportunidad |
|---|---|---|
| Digitación manual de datos y cálculos | Errores de RUC, IGV y totales | Validación automática y cálculo en dominio |
| Sistemas heredados lentos y poco intuitivos | Colas en caja, capacitación costosa | UI de caja optimizada para pocas interacciones |
| Numeración manual o sin control de concurrencia | Duplicados o huecos en correlativos | Generación transaccional de correlativos |
| Inventario desactualizado o sin bloqueo | Sobreventa y quiebres de stock | Descuento atómico de stock por venta |
| Rechazos de SUNAT gestionados a mano | Reenvíos tardíos, riesgo fiscal | Cola de envío con reintentos y estados visibles |
| Comprobantes dispersos (papel, correo) | Difícil reimprimir o auditar | Repositorio centralizado con descarga PDF/XML |

## 1.8 Indicadores (KPIs)

Las líneas base se miden en el proceso actual; no se inventan aquí.

| KPI | Definición |
|---|---|
| Tiempo promedio de atención en caja | Desde el inicio de la venta hasta la entrega del comprobante |
| Tasa de comprobantes rechazados | Rechazados / total enviados |
| Tasa de errores de emisión | Comprobantes con nota de crédito por error / total |
| Tiempo de emisión a CDR | Desde la emisión hasta la respuesta de SUNAT |
| Incidencias de sobreventa | Ventas fallidas o inconsistentes por stock |
| Volumen diario/mensual | Comprobantes emitidos por periodo |

# Referencias

**Normativa y portales de SUNAT**

1. SUNAT. *Boleta de Venta* (Tipos de comprobantes, Portal CPE). https://cpe.sunat.gob.pe/tipos_de_comprobantes/boleta
2. SUNAT. *Comprobantes desde los sistemas del contribuyente* (SEE-Del Contribuyente). https://cpe.sunat.gob.pe/noticias/comprobantes-desde-los-sistemas-del-contribuyente
3. SUNAT. *Operatividad* (Orientación SUNAT). https://orientacion.sunat.gob.pe/3529-operatividad
4. SUNAT. *Concepto y características: Boleta de Venta Electrónica desde SEE Contribuyente*. https://orientacion.sunat.gob.pe/node/711
5. SUNAT. *Concepto y características: Emisión Electrónica desde los Sistemas del Contribuyente*. https://orientacion.sunat.gob.pe/node/708
6. SUNAT. Resolución de Superintendencia N.° 114-2019/SUNAT (modifica la normativa sobre la boleta de venta electrónica y las notas vinculadas). https://ww3.sunat.gob.pe/legislacion/superin/2019/114-2019.pdf
Aquí tienes la transcripción organizada y formateada en Markdown de los tres documentos proporcionados. Se ha limpiado el ruido típico del OCR (como errores de caracteres) para garantizar una lectura clara y profesional, manteniendo la estructura original, tablas y bloques de código.

---

# Documento 1: Observabilidad y Rendimiento en Sistemas Distribuidos

## Observabilidad y Rendimiento en la Construcción
**Semana 10 | Unidad 3: Técnicas de Construcción Avanzadas y Sistemas Distribuidos (SW707)**

### Los Tres Pilares de la Observabilidad
1. **LOGS** (Registros): Eventos discretos, detallados y con contexto. *(El historial clínico)*.
2. **TRACES** (Trazas): Flujo causal de una petición a través de la red. *(La resonancia magnética)*.
3. **METRICS** (Métricas): Agregaciones numéricas en el tiempo. *(Los signos vitales)*.

> **La Observabilidad se Construye, No se Instala**  
> La observabilidad traslada el enfoque del monitoreo reactivo (*¿Qué se rompió?*) al diagnóstico proactivo (*¿Por qué se rompió en tiempo de ejecución?*). Es una característica intrínseca del código.

---

### Registros Estructurados e Hiper-contextualización
Ejemplo de log plano:  
`[INFO] 2023-10-24 10:15:32 - User 4592 logged in from IP 192.168.1.5`

Ejemplo de log estructurado (JSON):
```json
{
  "timestamp": "2023-10-24T10:15:32Z",
  "level": "INFO",
  "event": "user_login",
  "user_id": 4592,
  "tenant": "latam_corp",
  "ip_address": "192.168.1.5"
}
```
**MDC (Mapped Diagnostic Context):** Funciona como un pasaporte de datos que viaja adherido al hilo de ejecución (Thread). Inyecta automáticamente metadatos (como el `user_id` o `trace_id`) en cada log generado por ese hilo, permitiendo búsquedas indexadas y correlación sin ensuciar la firma de los métodos.

---

### Métricas de Sistema y Dominio: Modelos RED y USE

#### Modelo RED (Para Servicios/APIs)
- **Rate (Tasa):** Peticiones por segundo.
- **Errors (Errores):** Tasa de fallas (HTTP 5xx, excepciones).
- **Duration (Duración):** Latencia y distribución temporal (p95, p99).

#### Modelo USE (Para Recursos Físicos/Sistema)
- **Utilization (Utilización):** Porcentaje de tiempo que el recurso (CPU, Disco, Red) está ocupado.
- **Saturation (Saturación):** Medida de la cola de trabajo extra que no puede ser procesada.
- **Errors (Errores):** Fallas de hardware, red o memoria.

---

### Trazado Distribuido y Propagación de Contexto
- **Trace ID global:** `8a3c60f7...` (El Hilo de Ariadna que abraza y correlaciona múltiples Span IDs a medida que una única petición cruza la frontera de diferentes procesos y servidores).
  - *Span A:* API Gateway (150ms)
  - *Span B:* Auth Microservice (45ms)
- **El Estándar: OpenTelemetry.** Provee la especificación unificada para generar y propagar este contexto (mediante la cabecera `traceparent` en HTTP/gRPC) evitando el bloqueo de proveedor (*vendor lock-in*).

---

### Análisis de Rendimiento: Profiling de Runtime
**Lectura del Flame Graph:** Busca las mesetas anchas en la cima. Esas funciones son las que monopolizan la CPU en tiempo de ejecución.
- **Sampling (Muestreo):** Toma fotografías periódicas del stack. Bajo impacto estadístico, bajo overhead. Seguro para entornos de producción.
- **Instrumentación:** Inyecta código para medir cada entrada/salida de función. Alta precisión, alto overhead. Exclusivo para entornos de desarrollo y pruebas.

#### Detección de Cuellos de Botella y Optimización de Recursos
- **¿CPU Saturada (USE alto)?** Revisar Flame Graphs. Buscar bucles ineficientes, serialización JSON pesada, o algoritmos subóptimos.
- **¿Consumo de Memoria Creciendo (OOM Inminente)?** Buscar Memory Leaks. Analizar Heap Dumps por retención de objetos y referencias estáticas no liberadas.
- **¿Latencia Alta (RED Alto) pero CPU Baja?** Cuello de Botella en I/O. Revisar consultas lentas a Base de Datos, latencia de red, o Thread Starvation (bloqueos de concurrencia y deadlocks).

---

### Instrumentación en la Práctica (OpenTelemetry)
```python
from opentelemetry import trace
from mi_app.database import guardar_pago

# Obtener el tracer para este módulo
tracer = trace.get_tracer(__name__)

def procesar_transaccion(pago):
    # Crear un Span personalizado para medir esta operación específica
    with tracer.start_as_current_span("procesar_pago") as span:
        # Hiper-contextualización: Inyectar datos del dominio
        span.set_attribute("usuario_id", pago.user_id)
        span.set_attribute("monto_total", pago.monto)
        
        try:
            guardar_pago(pago)
            span.set_attribute("estado", "completado")
        except Exception as e:
            # Registro estructurado del error en la traza
            span.record_exception(e)
            span.set_status(trace.Status(trace.StatusCode.ERROR))
            raise
```

> **Relación con la Resiliencia:**  
> La observabilidad (Semana 10) detecta la anomalía. La resiliencia (Semana 7) activa el patrón (ej. Circuit Breaker) y sobrevive al fallo.  
> *La observabilidad es el sistema nervioso; la resiliencia es el sistema inmunológico. Juntos garantizan la supervivencia del software en tiempo de ejecución.*

---

### Checklist para la Práctica Calificada 3 (PC3)
1. **Servicio Distribuido:** Construir un microservicio (API REST o gRPC).
2. **Logs Estructurados:** Implementar registros en JSON inyectando contexto vía MDC.
3. **Telemetría OTel:** Exponer un endpoint de métricas y generar Spans manuales.
4. **Resiliencia Activa:** Demostrar una estrategia de tolerancia a fallas basada en configuración dinámica (tiempo de ejecución).
5. **Análisis de Runtime:** Entregar reporte de Profiling documentando la identificación y afinamiento de cuellos de botella.

---
---

# Documento 2: Sistemas Distribuidos Cloud-Native

## Construcción | Semana 11
**Temas Clave:**
- **Arquitectura:** De Monolitos a Microservicios.
- **APIs:** REST VS. gRPC.
- **Middleware:** Observabilidad e Intercepción.
- **Enrutamiento:** Service Discovery.
- **Estado:** Consistencia Eventual y Sagas.
- **Implementación:** Código en Python.

---

### El Caos de la Red: Cliente-Servidor Clásico vs. Microservicios
Dividir el código introduce el caos de la red.
- **Cliente-Servidor Clásico:** Fuerte acoplamiento, latencia predecible, llamadas locales.
- **Microservicios y Bounded Contexts:** 
  - *Red no confiable:* Introduce retrasos, los paquetes se pierden, los nodos mueren.
  - *Topología dinámica:* Las direcciones IP cambian constantemente.
  - *Falla parcial:* Un servicio dependiente puede colapsar mientras el resto vive.

---

### Protocolos de Comunicación Interna y Externa

| Característica | REST (HTTP/1.1 + JSON) | gRPC (HTTP/2 + Protocol Buffers) |
| :--- | :--- | :--- |
| **Protocolo Subyacente** | HTTP/1.1 (Peticiones secuenciales) | HTTP/2 (Multiplexación bidireccional) |
| **Formato de Payload** | JSON (Texto, legible por humanos, pesado) | Protobuf (Binario, comprimido, tipado estricto) |
| **Contrato** | OpenAPI/Swagger (Opcional) | Archivos `.proto` (Obligatorio e inmutable) |
| **Caso de Uso** | Exposición pública y clientes web | Comunicación de alta velocidad Service-to-Service |

---

### Intercepción de Peticiones en el Middleware
El middleware actúa como puertas de control (*Gates*) antes de llegar a la lógica de negocio:
1. **Gate 1 - Autenticación (JWT):** Validación criptográfica de identidad sin consultar la base de datos de estado.
2. **Gate 2 - Validación Defensiva:** Barrera de seguridad Fail-Fast. Rechazo inmediato de payloads malformados antes de consumir ciclos de CPU.
3. **Gate 3 - Propagación de Trace ID:** Inyección de un identificador único. Fundamental para la observabilidad (logs, métricas y trazas). El Trace ID es el único hilo conductor para diagnosticar errores de tiempo de ejecución a través de múltiples contenedores.

---

### Evolución de Esquemas sin Romper Contratos
- **Backward Compatibility (Cliente Antiguo -> Servidor Nuevo):** El servidor lee datos antiguos proporcionando valores por defecto para campos nuevos.
- **Forward Compatibility (Servidor Nuevo -> Cliente Antiguo):** El cliente ignora con seguridad los campos nuevos en el flujo binario.
- **Regla de Oro:** Nunca cambies el tipo de un campo existente ni elimines un campo requerido. Añade solo campos opcionales.
- **Ventaja de Protobuf:** La serialización binaria maneja la evolución del esquema por diseño, a diferencia de JSON donde las claves faltantes suelen provocar excepciones de tiempo de ejecución.

---

### Enrutamiento Dinámico en un Entorno Efímero
- **El Problema:** Hardcodear direcciones IP es una falacia. Los contenedores mueren y nacen constantemente por escalado o fallos.
- **La Solución (Service Discovery):** Mantiene un registro dinámico (latido de vida / health checks) de las instancias saludables en tiempo de ejecución.
- **Balanceo de Carga:** Distribución de peticiones (Round-Robin, Least-Connections) para optimizar el rendimiento y evitar cuellos de botella.

---

### Gestión de Estado y Consistencia Eventual
**Teorema CAP:** Los sistemas Cloud-Native priorizan **AP** (Disponibilidad y Tolerancia a Particiones), aceptando la Consistencia Eventual.

**Transactional Outbox Pattern:**
1. **Transacción Local (ACID):** El microservicio guarda la entidad de negocio Y un Evento Outbox en la misma base de datos local simultáneamente.
2. **Relé de Mensajes (Message Relay):** Un proceso asíncrono lee la tabla Outbox.
3. **Colas de Mensajes:** El evento se publica en el Message Broker garantizando entrega (At-Least-Once).
4. **Patrón Saga:** Los servicios suscriptores reaccionan al evento y aplican compensaciones si ocurre un fallo.

---

### Anatomía de un Servicio Distribuido (Python)

#### Middleware (FastAPI) - Inyección y propagación del Trace ID
```python
from fastapi import Request, FastAPI
import uuid

app = FastAPI()

@app.middleware("http")
async def add_process_time_header(request: Request, call_next):
    # Inyecta el Trace ID en el contexto de la petición
    trace_id = request.headers.get("X-Trace-Id", str(uuid.uuid4()))
    request.state.trace_id = trace_id
    
    response = await call_next(request)
    
    # Propaga el Trace ID en la respuesta
    response.headers["X-Trace-Id"] = trace_id
    return response
```

#### Invocación gRPC (Cliente)
```python
import grpc
from my_service_pb2 import ...

try:
    # Llamada RPC sincrónica (A. Invocación tipada mediante stubs generados)
    response = stub.GetData(request)
    return response.data
except grpc.RpcError as e:
    # B. Validación defensiva fail-fast y manejo de excepciones de red
    if e.code() == grpc.StatusCode.NOT_FOUND:
        print(f"Datos no encontrados: {e.details()}")
    else:
        print(f"Error de red: {e.code()} - {e.details()}")
    raise
```

---

### 3 Reglas de Oro de la Construcción Cloud-Native
1. **Diseñar para el fallo:** Asumir que la red fallará, implementar time-outs y circuit breakers.
2. **Trazar absolutamente todo:** Sin observabilidad (logs, métricas, trace IDs), el sistema es una caja negra.
3. **Automatizar la recuperación:** Delegar el manejo de estado distribuido a patrones como Sagas y Outbox.

> *Hemos construido un sistema distribuido complejo. Pero, ¿qué pasa cuando este código envejece, los requerimientos cambian o el diseño se degrada? Próxima semana: Refactorización y gestión de la Deuda Técnica.*

---
---

# Documento 3: Construcción Dirigida por Datos

## Universidad Nacional de Ingeniería - FIIS
### Construcción Dirigida por Datos: Estados, Tablas y Configuración
**Semana 9 | Unidad 3: Técnicas de Construcción Avanzadas (SW707 Construcción de Software II)**

- **Unidades 1 y 2 (El Pasado):** Robustez y Concurrencia. Programación defensiva, contratos, manejo de excepciones y tolerancia a fallas (Semanas 1-8).
- **Unidad 3 / Semana 9 (El Presente):** Lógica Dinámica. Máquinas de estado, tablas, i18n y configuración en tiempo de ejecución.
- **Unidades 3 y 4 (El Futuro):** Distribución y Evolución. Sistemas cloud-native, observabilidad, refactorización e ingeniería reversa (Semanas 10-15).

> Hemos construido sistemas seguros a prueba de fallos; ahora aprenderemos a alterar su comportamiento sin necesidad de recompilar el código.

---

### El Futuro Dinámico: Datos Estructurados vs. El Pasado Rígido
- **El Pasado Rígido (Lógica Imperativa):** `if/else/switch`. Alta complejidad ciclomática. Difícil de mantener. Requiere redespliegue para cada cambio de regla.
- **El Futuro Dinámico (Datos Estructurados):** Control de flujo delegado a estructuras de datos. Reglas claras, escalables y modificables en tiempo de ejecución.  
  *La construcción dirigida por datos cambia la pregunta fundamental: en lugar de "¿Qué código debemos escribir?", nos preguntamos "¿Qué configuración debemos leer?".*

---

### Máquinas de Estado Finito (FSM) como Mecanismo de Control
Ejemplo de flujo: `[PENDIENTE]` → (1. recibir_pago) → `[PAGADO]` → (2. despachar) → `[ENVIADO]`
- **Comportamiento por Estado:** Las operaciones permitidas están dictadas estrictamente por el estado actual del sistema.
- **Restricción Matemática:** Previene transiciones ilegales por diseño, eliminando la necesidad de validar el estado previo en cada función.
- **Construcción Segura:** El modelo de autómatas convierte reglas de negocio ambiguas en topologías exactas.

#### Implementando FSM mediante POO: El Patrón State
- **Polimorfismo sobre Condicionales:** Reemplazo de sentencias `switch` masivas delegando el comportamiento a clases de estado específicas.
- **Principio Abierto/Cerrado (OCP):** Nuevos estados pueden añadirse creando nuevas clases, sin modificar el contexto original.
- **Encapsulamiento:** Cada estado gestiona sus propias transiciones y reglas de validación de forma autónoma.

---

### Construcción Dirigida por Tablas (Table-Driven Design)
Reemplazar lógica de control con búsquedas de datos.
- **Acceso Directo (O(1)):** En lugar de evaluar múltiples condiciones booleanas secuenciales (`if plan==A && user==B`), se indexa directamente la matriz: `Tabla[Usuario][Plan]`.
- **Desacoplamiento:** Separa las reglas de negocio (datos) de la máquina de ejecución (código base).

#### Intérpretes y Parsing Basado en Gramáticas
- **El Motor de Parsing:** Los compiladores y gramáticas modernas procesan el código fuente utilizando analizadores dirigidos por tablas (Table-Driven Parsers).
- **Gramática como Datos:** Las reglas sintácticas se compilan en una tabla de transición de estados masiva (Abstract Syntax Tree).
- **Ejecución Segura:** El intérprete navega por la tabla consumiendo tokens; si no hay transición válida definida, lanza un error de sintaxis instantáneo.

---

### Configuración Externa y Enlace Tardío (Late Binding)
- **Externalización:** Mover constantes, umbrales y cadenas de conexión fuera del código binario compilado (JSON, Environment Variables). Evita la recompilación para cambios operacionales. Permite adaptar una misma construcción de software a múltiples entornos (Desarrollo, QA, Producción) instantáneamente.
- **Enlace Tardío:** Retrasar la asignación de valores y parámetros hasta el último momento absoluto (Runtime).

#### Feature Flags (Indicadores de Funcionalidad)
Desacoplar el Despliegue del Lanzamiento (Release).
- **Pruebas en Producción:** Habilitar funcionalidades ocultas solo para usuarios internos antes del lanzamiento oficial (Feature A Estable, Feature B Beta).
- **Lanzamientos Canary:** Desplegar nuevas características gradualmente a un subconjunto estadístico de usuarios para monitorear impacto.
- **Rollbacks Instantáneos (Kill Switches):** Apagar una característica defectuosa en milisegundos manipulando una variable, sin necesidad de revertir el código fuente.

---

### Diagnóstico Estructural: i18n vs. l10n
| Concepto | Definición | Técnica | Naturaleza |
| :--- | :--- | :--- | :--- |
| **i18n (Internacionalización)** | Preparar la base de código para ser global. | Externalización de cadenas, soporte UTF-8, lógica de formateo de fechas. | Ingeniería de Software pura (Arquitectura). |
| **l10n (Localización)** | Adaptar el software a una región específica. | Archivos de recursos específicos (es_PE, en_US), monedas locales, matices culturales. | Datos e inyección de contenido (Implementación). |

#### i18n como Construcción Dirigida por Tablas
- **Mecanismo:** En lugar de codificar `if (lang == 'es')`, el software busca la clave en un diccionario estructurado (Tabla) cargado en la memoria durante el Runtime.
- **Resultado:** Separación total entre la lógica estática de la interfaz de usuario y el catálogo dinámico de textos.

---

### Integración Nativa con la Programación Orientada a Objetos
- **Construcción por Lógica Basada en Estados:** Patrón State & Factory. Resolviendo implementaciones en tiempo de ejecución desde un diccionario de clases instanciables.
- **Configuración Tablas & Externa:** Patrón Strategy. Polimorfismo estructural que reemplaza la complejidad condicional y encapsula las transiciones de las FSMs.
- **Enlace Tardío:** Inyección de Dependencias (IoC). Inyectar los parámetros de los Feature Flags y Locales (i18n) directamente en los constructores de los objetos.

---

### El Mapa Arquitectónico: Síntesis en Clean Architecture
- **Entidades:** Máquinas de Estado Finito (FSM) modelando los ciclos de vida inmutables del negocio.
- **Casos de Uso:** Tablas de Reglas de Negocio que calculan algoritmos de forma desacoplada.
- **Adaptadores de Interfaz:** Parsers dirigidos por tablas transformando los payloads HTTP entrantes.
- **Frameworks y Drivers:** Configuración, Feature Flags, i18n Resource Bundles y Late Binding inyectando variables operativas.

---

### El Triángulo de Valor de la Lógica Dirigida por Datos
1. **Flexibilidad Extrema:** Alteración instantánea del comportamiento en producción mediante configuración y enlace tardío, sin downtime.
2. **Mantenibilidad Pura:** Reducción dramática de la complejidad ciclomática al eliminar cascadas interminables de código condicional.
3. **Seguridad Operacional:** Las restricciones estructurales de las FSMs y gramáticas hacen que los estados y flujos ilegales sean matemáticamente imposibles.

> **Menos Código Imperativo. Mejores Estructuras de Datos.**  
> El verdadero dominio de la construcción de software en SW707 no radica en saber cómo escribir un bloque condicional complejo, sino en saber cómo abstraerlo en datos.

---

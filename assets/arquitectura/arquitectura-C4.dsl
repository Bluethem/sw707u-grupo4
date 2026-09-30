workspace "FlowFlacture" "Sistema de Gestión y Facturación Electrónica (boletas y facturas) con control de inventario - Modelo C4 (Go + Next.js + PostgreSQL + Cloudflare R2 en VPS)" {

    model {

        # ---------------------------------------------------------------
        # Personas
        # ---------------------------------------------------------------
        cajero = person "Cajero" "Registra ventas, emite boletas/facturas y consulta comprobantes (RF-02, CU-10)." "Interno"
        admin = person "Administrador del negocio" "Gestiona catálogo, precios, stock, series, usuarios, certificado, anulaciones y reportes." "Interno"
        cliente = person "Cliente final" "Recibe y consulta/descarga sus comprobantes (PDF/XML) desde una web segura (CU-21)." "Externo"

        # ---------------------------------------------------------------
        # Sistemas externos
        # ---------------------------------------------------------------
        sunat = softwareSystem "SUNAT / OSE" "Valida el XML UBL 2.1 y devuelve la constancia de recepción (CDR); recibe resúmenes diarios y comunicaciones de baja." "Externo"
        consultaDoc = softwareSystem "Servicio de consulta RUC/DNI" "Autocompleta datos de clientes por RUC/DNI (opcional, RF-10)." "Externo"
        correo = softwareSystem "Servicio de correo (SMTP)" "Envía comprobantes por correo al cliente (RF-44)." "Externo"

        # ---------------------------------------------------------------
        # Sistema principal
        # ---------------------------------------------------------------
        flow = softwareSystem "FlowFlacture" "Orquesta cálculo de impuestos, numeración, descuento de stock, firma, envío a SUNAT, almacenamiento y generación de PDF. Multiempresa (tenants aislados)." {

            web = container "Aplicación Web" "Pantallas de caja (teclado/táctil), catálogo, clientes, series, monitoreo de estados, reportes y auditoría, más las rutas públicas del portal de consulta del cliente (mín. 1 año)." "Next.js (TypeScript), responsive WCAG AA" "Web Browser"

            api = container "API Backend" "Casos de uso síncronos: seguridad, clientes, catálogo, ventas y emisión atómica (correlativo + stock + XML firmado + trabajo de envío en una sola transacción). Arquitectura limpia / puertos y adaptadores. Sin estado." "Go (cmd/api)" {

                # --- M1 Seguridad y configuración
                authController = component "Auth Controller" "Login, logout, expiración de sesión, bloqueo por intentos fallidos (CU-01)." "HTTP Handler"
                authService = component "Servicio de Autenticación y RBAC" "Valida credenciales (Argon2/bcrypt), roles Cajero/Administrador (RF-01, RF-02)." "Application Service"
                tenantContext = component "Contexto de Tenant" "Resuelve empresa_id por request y lo impone en toda consulta (RN-09, RNF-11)." "Middleware"
                configController = component "Configuración Controller" "Empresa emisora, series, certificado digital y credenciales SOL/OSE/PSE (CU-03, CU-04)." "HTTP Handler"
                configService = component "Servicio de Configuración" "Valida RUC, series únicas, vigencia del certificado." "Application Service"
                secretosService = component "Servicio de Cifrado de Secretos" "Cifra/descifra certificados y credenciales con AES-GCM; llave maestra por variable de entorno (RNF-09)." "Adapter"
                auditoriaService = component "Servicio de Auditoría" "Bitácora de acciones sensibles: emisión, anulación, precios, stock, certificado (RF-48, CU-05)." "Application Service"

                # --- M2 / M3
                clientesController = component "Clientes Controller" "CRUD/búsqueda de clientes y autocompletado (CU-06)." "HTTP Handler"
                clientesService = component "Servicio de Clientes" "Valida RUC/DNI (dígito verificador), unicidad por empresa." "Application Service"
                catalogoController = component "Catálogo e Inventario Controller" "Productos, ajustes de stock y carga masiva (CU-07, CU-08, CU-09)." "HTTP Handler"
                catalogoService = component "Servicio de Catálogo" "Productos, afectación IGV, importación CSV/Excel con validación." "Application Service"
                stockService = component "Servicio de Stock" "Descuento/reversa atómica de stock con movimientos trazables; nunca negativo (RN-05, RNF-04)." "Domain/Application Service"

                # --- M4 / M5
                ventasController = component "Ventas y Caja Controller" "Arma venta, calcula totales, recibe confirmación con clave de idempotencia (CU-10)." "HTTP Handler"
                emisionService = component "Servicio de Emisión de Comprobantes" "Orquesta emisión (boleta/factura/nota de crédito) en una transacción única (CU-11, CU-12, CU-13)." "Application Service (Use Case)"
                reglasIdentificacion = component "Reglas de Identificación y Validación" "Umbral S/ 700, RUC obligatorio en factura, parametrizable (RN-01, RNF-20)." "Domain"
                calculadoraImpuestos = component "Calculadora de Impuestos" "IGV 18% por afectación; aritmética decimal exacta (shopspring/decimal) y redondeo definido (RN-03, RNF-07)." "Domain"
                numeracionService = component "Servicio de Numeración" "Serie + correlativo único y consecutivo con bloqueo transaccional (RN-02, RNF-05)." "Domain/Application Service"
                idempotenciaService = component "Control de Idempotencia" "Evita comprobantes duplicados ante reintentos (RF-21)." "Application Service"
                xmlBuilder = component "Generador XML UBL 2.1" "Construye el XML según catálogos SUNAT (RF-26)." "Adapter"
                firmaAdapter = component "Adaptador de Firma Digital" "Firma XML con el certificado de la empresa, propio o vía PSE (RF-27)." "Adapter (Puerto: FirmaPort)"
                pdfGenerator = component "Generador de PDF" "Representación impresa con código QR; regenerable (RF-42)." "Adapter (Puerto: PdfPort)"
                comprobanteRepo = component "Repositorio de Comprobantes" "Persistencia, estados y trazabilidad; comprobantes inmutables (RF-28, RNF-32)." "Repository (Puerto de persistencia)"
                outboxPublisher = component "Publicador de Cola de Envío (Outbox)" "Inserta el trabajo de envío PENDIENTE en la misma transacción de emisión." "Adapter"
                archivoAdapter = component "Adaptador de Almacenamiento" "Sube/descarga XML, PDF y CDR (Puerto: StoragePort)." "Adapter (cliente S3)"

                # --- M6 (consulta/acciones) y M7
                estadosController = component "Estados y Reenvío Controller" "Panel de estados, reenvío manual, bajas (CU-16, CU-17, CU-18)." "HTTP Handler"
                documentosController = component "Documentos y Búsqueda Controller" "Búsqueda, descarga PDF/XML/CDR, envío por correo (CU-19, CU-20)." "HTTP Handler"
                consultaPublicaController = component "Consulta Pública Controller" "Valida datos de verificación, limita intentos (CU-21, RNF-13)." "HTTP Handler"
                reportesService = component "Servicio de Reportes" "Reporte diario de ventas y exportaciones CSV/Excel (CU-22, RF-49, RF-50)." "Application Service"
                notificacionAdapter = component "Adaptador de Correo" "Envío de comprobantes por SMTP." "Adapter"
                consultaDocAdapter = component "Adaptador de Consulta RUC/DNI" "Cliente del servicio externo con degradación a ingreso manual." "Adapter"
            }

            worker = container "Worker de Integración Tributaria" "Procesos asíncronos: envío de facturas, resumen diario de boletas, bajas, CDR, reintentos y alertas de plazo. No bloquea la caja (RNF-14). Comparte dominio y adaptadores con la API." "Go (cmd/worker)" {
                planificador = component "Planificador" "Dispara resumen diario y revisión de plazos (3 y 7 días calendario)." "Scheduler"
                colaConsumer = component "Consumidor de Cola (Outbox)" "Toma trabajos pendientes con FOR UPDATE SKIP LOCKED; sobrevive a reinicios (RNF-15)." "Job Consumer"
                envioFacturas = component "Servicio de Envío de Facturas" "Comprime y envía XML firmado; procesa estado ACEPTADO/OBSERVADO/RECHAZADO (CU-14)." "Application Service"
                resumenDiario = component "Servicio de Resumen Diario" "Agrupa boletas y notas por día, firma, envía, consulta ticket (CU-15)." "Application Service"
                bajasService = component "Servicio de Bajas" "Comunicación de baja de facturas y boletas 'anuladas' en resumen (CU-16, CU-17)." "Application Service"
                cdrProcessor = component "Procesador de CDR" "Interpreta y almacena constancias; actualiza estados." "Application Service"
                retryPolicy = component "Política de Reintentos" "Backoff progresivo con trazabilidad de cada intento (RF-33)." "Domain Service"
                alertaPlazos = component "Alerta de Plazos y Certificado" "Alerta comprobantes próximos a vencer y vencimiento del certificado (RF-35, RNF-19)." "Application Service"
                sunatGateway = component "Adaptador SUNAT/OSE" "Cliente SOAP/REST hacia SUNAT/OSE (Puerto: SunatPort; con Fake para pruebas)." "Adapter"
                firmaWorker = component "Adaptador de Firma Digital" "Firma resúmenes y comunicaciones de baja." "Adapter (Puerto: FirmaPort)"
                secretosWorker = component "Servicio de Cifrado de Secretos" "Descifra certificado y credenciales SOL/OSE (AES-GCM)." "Adapter"
                archivoWorker = component "Adaptador de Almacenamiento" "Guarda CDR, resúmenes y constancias (Puerto: StoragePort)." "Adapter (cliente S3)"
                estadoRepo = component "Repositorio de Estados y Envíos" "Actualiza estado, intentos y constancias." "Repository"
            }

            db = container "Base de Datos Relacional" "Empresas, usuarios, clientes, productos, stock, series/correlativos, comprobantes, resúmenes, bitácora. Además: tabla de cola (outbox) y certificados/credenciales cifrados. Aislamiento por empresa_id / Row-Level Security." "PostgreSQL" "Database"
            storage = container "Almacén de Archivos" "XML firmados, CDR, PDF, resúmenes, bajas y constancias de rechazo; bucket privado; conservación ≥ 1 año (RN-08). También recibe los respaldos diarios de la BD." "Cloudflare R2 (API compatible con S3)" "Database"
        }

        # ---------------------------------------------------------------
        # Relaciones - Contexto
        # ---------------------------------------------------------------
        cajero -> flow "Registra ventas y emite comprobantes"
        admin -> flow "Administra catálogo, series, usuarios, certificado; anula y reporta"
        cliente -> flow "Consulta y descarga sus comprobantes"
        flow -> sunat "Envía comprobantes, resúmenes y bajas; recibe CDR" "HTTPS/SOAP"
        flow -> consultaDoc "Consulta datos por RUC/DNI" "HTTPS"
        flow -> correo "Envía comprobantes por correo" "SMTP/TLS"

        # ---------------------------------------------------------------
        # Relaciones - Contenedores
        # ---------------------------------------------------------------
        cajero -> web "Usa" "HTTPS"
        admin -> web "Usa" "HTTPS"
        cliente -> web "Consulta comprobantes (rutas públicas)" "HTTPS"
        web -> api "Invoca" "JSON/HTTPS"
        api -> db "Lee/escribe (transacciones de emisión y outbox)" "SQL/TLS"
        api -> storage "Guarda/recupera XML y PDF" "S3 API/TLS"
        worker -> db "Toma trabajos de envío y actualiza estados" "SQL/TLS"
        worker -> storage "Guarda CDR, resúmenes y bajas" "S3 API/TLS"
        worker -> sunat "Envía y consulta constancias" "HTTPS/SOAP"
        db -> storage "Respaldo diario (pg_dump)" "S3 API/TLS"
        api -> consultaDoc "Autocompleta clientes" "HTTPS"
        api -> correo "Envía comprobantes" "SMTP/TLS"

        # ---------------------------------------------------------------
        # Relaciones - Componentes API
        # ---------------------------------------------------------------
        web -> authController "Login" "JSON/HTTPS"
        web -> ventasController "Confirma venta" "JSON/HTTPS"
        web -> clientesController "Gestiona clientes" "JSON/HTTPS"
        web -> catalogoController "Gestiona catálogo y stock" "JSON/HTTPS"
        web -> configController "Configura empresa/series/certificado" "JSON/HTTPS"
        web -> estadosController "Monitorea, reenvía y da de baja" "JSON/HTTPS"
        web -> documentosController "Busca y descarga" "JSON/HTTPS"
        web -> consultaPublicaController "Consulta pública" "JSON/HTTPS"

        authController -> authService "Usa"
        authService -> db "Usuarios y sesiones" "SQL"
        tenantContext -> authService "Obtiene empresa y rol"
        configController -> configService "Usa"
        configService -> db "Empresa y series" "SQL"
        configService -> secretosService "Cifra certificado y credenciales"
        secretosService -> db "Guarda/lee secretos cifrados" "SQL"
        clientesController -> clientesService "Usa"
        clientesService -> db "Clientes" "SQL"
        clientesService -> consultaDocAdapter "Autocompleta"
        consultaDocAdapter -> consultaDoc "Consulta" "HTTPS"
        catalogoController -> catalogoService "Usa"
        catalogoController -> stockService "Ajustes de stock"
        catalogoService -> db "Productos" "SQL"
        stockService -> db "Stock y movimientos (bloqueo atómico)" "SQL"

        ventasController -> emisionService "Ejecuta emisión"
        ventasController -> idempotenciaService "Verifica clave"
        idempotenciaService -> db "Claves de idempotencia" "SQL"
        emisionService -> reglasIdentificacion "Valida identificación"
        emisionService -> calculadoraImpuestos "Calcula IGV y totales"
        emisionService -> stockService "Descuenta stock"
        emisionService -> numeracionService "Toma correlativo"
        emisionService -> xmlBuilder "Genera XML UBL 2.1"
        emisionService -> firmaAdapter "Firma XML"
        emisionService -> comprobanteRepo "Persiste comprobante PENDIENTE"
        emisionService -> outboxPublisher "Registra trabajo de envío"
        emisionService -> pdfGenerator "Genera PDF con QR"
        emisionService -> auditoriaService "Registra emisión"
        numeracionService -> db "Series/correlativos con bloqueo" "SQL"
        comprobanteRepo -> db "Comprobantes y trazabilidad" "SQL"
        comprobanteRepo -> archivoAdapter "Guarda XML firmado"
        firmaAdapter -> secretosService "Obtiene certificado descifrado"
        outboxPublisher -> db "Inserta trabajo (misma transacción)" "SQL"
        pdfGenerator -> archivoAdapter "Guarda PDF"
        archivoAdapter -> storage "Sube/descarga archivos" "S3 API"
        auditoriaService -> db "Bitácora" "SQL"

        estadosController -> comprobanteRepo "Estados e intentos"
        estadosController -> outboxPublisher "Reenvío manual"
        documentosController -> comprobanteRepo "Consulta comprobantes"
        documentosController -> pdfGenerator "Regenera PDF"
        documentosController -> archivoAdapter "Descarga XML/PDF/CDR"
        documentosController -> notificacionAdapter "Envía por correo"
        notificacionAdapter -> correo "SMTP" "SMTP/TLS"
        consultaPublicaController -> comprobanteRepo "Valida y recupera comprobante"
        reportesService -> db "Agregaciones de ventas" "SQL"

        # ---------------------------------------------------------------
        # Relaciones - Componentes Worker
        # ---------------------------------------------------------------
        planificador -> resumenDiario "Dispara resumen diario"
        planificador -> alertaPlazos "Revisa plazos y certificado"
        colaConsumer -> db "Toma trabajos (FOR UPDATE SKIP LOCKED)" "SQL"
        colaConsumer -> envioFacturas "Entrega factura pendiente"
        colaConsumer -> bajasService "Entrega bajas"
        envioFacturas -> sunatGateway "Envía factura"
        envioFacturas -> retryPolicy "Aplica reintentos"
        envioFacturas -> cdrProcessor "Procesa respuesta"
        resumenDiario -> firmaWorker "Firma resumen"
        resumenDiario -> sunatGateway "Envía resumen y consulta ticket"
        resumenDiario -> cdrProcessor "Procesa constancia"
        resumenDiario -> estadoRepo "Lee boletas pendientes"
        bajasService -> firmaWorker "Firma baja"
        bajasService -> sunatGateway "Envía baja"
        bajasService -> cdrProcessor "Procesa constancia"
        cdrProcessor -> estadoRepo "Actualiza estados"
        cdrProcessor -> archivoWorker "Guarda CDR/constancias"
        archivoWorker -> storage "Sube archivos" "S3 API"
        estadoRepo -> db "Estados e intentos" "SQL"
        alertaPlazos -> estadoRepo "Consulta pendientes"
        firmaWorker -> secretosWorker "Obtiene certificado descifrado"
        secretosWorker -> db "Lee secretos cifrados" "SQL"
        sunatGateway -> secretosWorker "Obtiene credenciales SOL/OSE"
        sunatGateway -> sunat "HTTPS/SOAP"

        # ---------------------------------------------------------------
        # Despliegue
        # ---------------------------------------------------------------
        deploymentEnvironment "Producción" {

            deploymentNode "Equipo del usuario" "" "PC / tablet / móvil" {
                deploymentNode "Navegador" "" "Chrome / Edge / Firefox" {
                    navegador = infrastructureNode "Navegador web" "Renderiza la aplicación Next.js." "HTML/JS"
                }
            }

            deploymentNode "VPS" "Servidor virtual privado" "Ubuntu Server" {

                deploymentNode "Docker Compose" "Configuración por variables de entorno (RNF-35)" "Docker" {

                    proxy = infrastructureNode "Reverse Proxy" "Terminación TLS, rate limiting, enrutamiento." "Caddy / Nginx"

                    deploymentNode "Contenedor web" "" "Node.js" {
                        webInstance = containerInstance web
                    }
                    deploymentNode "Contenedor api" "" "Go (binario único)" {
                        apiInstance = containerInstance api
                    }
                    deploymentNode "Contenedor worker" "" "Go (binario único)" {
                        workerInstance = containerInstance worker
                    }
                    deploymentNode "Contenedor postgres" "Volumen persistente; pg_dump diario a R2 (RPO ≤ 24 h)" "PostgreSQL" {
                        dbInstance = containerInstance db
                    }
                }
            }

            deploymentNode "Cloudflare" "" "Cloud" {
                deploymentNode "Bucket R2 privado" "Acceso por API, sin exposición pública de objetos" "Cloudflare R2" {
                    storageInstance = containerInstance storage
                }
            }

            deploymentNode "SUNAT / OSE" "" "Servicios externos" {
                softwareSystemInstance sunat
            }
            deploymentNode "Servicios de terceros" "" "Servicios externos" {
                softwareSystemInstance consultaDoc
                softwareSystemInstance correo
            }

            navegador -> proxy "HTTPS"
            proxy -> webInstance "Enruta páginas" "HTTP"
            proxy -> apiInstance "Enruta /api" "HTTP"
        }
    }

    views {

        systemContext flow "C1-Contexto" "Diagrama de contexto de FlowFlacture" {
            include *
            autoLayout lr
        }

        container flow "C2-Contenedores" "Contenedores de FlowFlacture" {
            include *
            autoLayout lr
        }

        component api "C3-Componentes-API" "Componentes de la API Backend (Go)" {
            include *
            autoLayout tb
        }

        component worker "C3-Componentes-Worker" "Componentes del Worker de Integración Tributaria (Go)" {
            include *
            autoLayout tb
        }

        dynamic flow "Flujo-Emision-Factura" "Emisión de comprobante y envío asíncrono a SUNAT" {
            cajero -> web "1. Confirma la venta"
            web -> api "2. POST /ventas (clave de idempotencia)"
            api -> db "3. Transacción: stock, correlativo, comprobante PENDIENTE y trabajo de envío"
            api -> storage "4. Guarda XML firmado y PDF"
            worker -> db "5. Toma trabajo pendiente (SKIP LOCKED)"
            worker -> sunat "6. Envía XML y recibe CDR"
            worker -> storage "7. Guarda CDR"
            worker -> db "8. Actualiza estado (ACEPTADO/OBSERVADO/RECHAZADO)"
            autoLayout lr
        }

        deployment flow "Producción" "D1-Despliegue" "Despliegue en VPS con Docker Compose + Cloudflare R2" {
            include *
            autoLayout lr
        }

        styles {
            element "Person" {
                shape Person
                background #08427b
                color #ffffff
            }
            element "Externo" {
                background #999999
                color #ffffff
            }
            element "Software System" {
                background #1168bd
                color #ffffff
            }
            element "Container" {
                background #438dd5
                color #ffffff
            }
            element "Component" {
                background #85bbf0
                color #000000
            }
            element "Web Browser" {
                shape WebBrowser
            }
            element "Database" {
                shape Cylinder
            }
            element "Infrastructure Node" {
                background #f5a623
                color #000000
            }
        }
    }
}

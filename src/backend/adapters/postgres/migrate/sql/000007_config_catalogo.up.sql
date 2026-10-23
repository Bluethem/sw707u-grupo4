-- Configuración paramétrica de validación del catálogo por empresa.
-- Si una empresa no tiene fila, los adaptadores usan los valores DEFAULT de esta tabla.
-- Solo el administrador de la empresa puede leer y modificar su propia fila.

CREATE TABLE IF NOT EXISTS config_catalogo (
    empresa_id             TEXT        PRIMARY KEY REFERENCES empresas(id) ON DELETE CASCADE,

    -- Límites de precio por producto
    precio_minimo          NUMERIC(18,2) NOT NULL DEFAULT 0.01
                               CHECK (precio_minimo > 0),
    precio_maximo          NUMERIC(18,2)          DEFAULT NULL
                               CHECK (precio_maximo IS NULL OR precio_maximo >= precio_minimo),

    -- Límites de la carga masiva
    carga_max_filas        INTEGER     NOT NULL DEFAULT 500
                               CHECK (carga_max_filas BETWEEN 1 AND 2000),
    carga_max_mb           INTEGER     NOT NULL DEFAULT 2
                               CHECK (carga_max_mb BETWEEN 1 AND 20),

    -- Conjuntos de valores permitidos
    -- Subset de los códigos SUNAT: NIU=unidad, ZZ=servicio, KGM=kg, LTR=litro, MTR=metro, BX=caja
    unidades_permitidas    TEXT[]      NOT NULL DEFAULT ARRAY['NIU','ZZ','KGM','LTR','MTR','BX'],
    afectaciones_permitidas TEXT[]     NOT NULL DEFAULT ARRAY['gravado','exonerado','inafecto'],

    -- Trazabilidad
    actualizado_en         TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Índice para búsqueda directa por empresa (la PK ya cubre esto,
-- pero lo dejamos explícito para que el plan de ejecución sea legible en auditorías).
CREATE INDEX IF NOT EXISTS idx_config_catalogo_empresa
    ON config_catalogo (empresa_id);

COMMENT ON TABLE config_catalogo IS
    'Reglas de validación del catálogo de productos, configurables por empresa. '
    'Aislamiento multi-tenant: cada fila pertenece a una empresa mediante empresa_id FK.';

COMMENT ON COLUMN config_catalogo.precio_minimo          IS 'Precio unitario mínimo aceptado en alta y carga masiva (S/).';
COMMENT ON COLUMN config_catalogo.precio_maximo          IS 'Precio unitario máximo aceptado. NULL = sin límite superior.';
COMMENT ON COLUMN config_catalogo.carga_max_filas        IS 'Máximo de filas permitidas en una carga masiva (1–2000).';
COMMENT ON COLUMN config_catalogo.carga_max_mb           IS 'Tamaño máximo del archivo Excel en MB (1–20).';
COMMENT ON COLUMN config_catalogo.unidades_permitidas    IS 'Códigos SUNAT de unidad de medida habilitados para esta empresa.';
COMMENT ON COLUMN config_catalogo.afectaciones_permitidas IS 'Tipos de afectación IGV habilitados (gravado, exonerado, inafecto).';

CREATE TABLE IF NOT EXISTS productos_stock (
  empresa_id  TEXT NOT NULL,
  producto_id TEXT NOT NULL,
  cantidad    BIGINT NOT NULL CHECK (cantidad >= 0),
  PRIMARY KEY (empresa_id, producto_id)
);

CREATE TABLE IF NOT EXISTS series_correlativos (
  empresa_id TEXT NOT NULL,
  tipo       TEXT NOT NULL CHECK (tipo IN ('boleta', 'factura')),
  serie      TEXT NOT NULL,
  siguiente  BIGINT NOT NULL CHECK (siguiente >= 1),
  PRIMARY KEY (empresa_id, tipo, serie)
);

CREATE TABLE IF NOT EXISTS comprobantes (
  id               TEXT PRIMARY KEY,
  empresa_id       TEXT NOT NULL,
  tipo             TEXT NOT NULL CHECK (tipo IN ('boleta', 'factura')),
  serie            TEXT NOT NULL,
  numero           BIGINT NOT NULL,
  estado           TEXT NOT NULL CHECK (estado IN (
    'PENDIENTE','ENVIADO','ACEPTADO','OBSERVADO','RECHAZADO','ANULADA','PENDIENTE_BAJA'
  )),
  cliente_tipo     TEXT,
  cliente_numero   TEXT,
  cliente_razon    TEXT,
  valor_venta      NUMERIC(18,2) NOT NULL,
  igv              NUMERIC(18,2) NOT NULL,
  total            NUMERIC(18,2) NOT NULL,
  idempotency_key  TEXT NOT NULL,
  creado_en        TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (empresa_id, tipo, serie, numero),
  UNIQUE (empresa_id, idempotency_key)
);

CREATE TABLE IF NOT EXISTS comprobante_lineas (
  id             BIGSERIAL PRIMARY KEY,
  comprobante_id TEXT NOT NULL REFERENCES comprobantes(id),
  producto_id    TEXT NOT NULL,
  cantidad       BIGINT NOT NULL,
  precio_unit    NUMERIC(18,2) NOT NULL,
  afectacion     TEXT NOT NULL,
  valor_venta    NUMERIC(18,2) NOT NULL,
  igv            NUMERIC(18,2) NOT NULL,
  total          NUMERIC(18,2) NOT NULL
);

CREATE TABLE IF NOT EXISTS outbox_envios (
  id             BIGSERIAL PRIMARY KEY,
  comprobante_id TEXT NOT NULL REFERENCES comprobantes(id),
  tipo_trabajo   TEXT NOT NULL DEFAULT 'envio_comprobante',
  estado         TEXT NOT NULL DEFAULT 'pending',
  intentos       INT NOT NULL DEFAULT 0,
  disponible_en  TIMESTAMPTZ NOT NULL DEFAULT now(),
  creado_en      TIMESTAMPTZ NOT NULL DEFAULT now()
);

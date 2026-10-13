CREATE TABLE IF NOT EXISTS productos (
  id              TEXT PRIMARY KEY,
  empresa_id      TEXT NOT NULL REFERENCES empresas(id),
  codigo          TEXT NOT NULL,
  descripcion     TEXT NOT NULL,
  unidad          TEXT NOT NULL DEFAULT 'NIU',
  precio          NUMERIC(18,2) NOT NULL CHECK (precio > 0),
  afectacion      TEXT NOT NULL CHECK (afectacion IN ('gravado','exonerado','inafecto')),
  controla_stock  BOOLEAN NOT NULL DEFAULT true,
  activo          BOOLEAN NOT NULL DEFAULT true,
  creado_en       TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (empresa_id, codigo)
);

CREATE TABLE IF NOT EXISTS stock_movimientos (
  id           BIGSERIAL PRIMARY KEY,
  empresa_id   TEXT NOT NULL REFERENCES empresas(id),
  producto_id  TEXT NOT NULL REFERENCES productos(id),
  tipo         TEXT NOT NULL CHECK (tipo IN ('inicial','entrada','salida','venta')),
  cantidad     BIGINT NOT NULL CHECK (cantidad > 0),
  motivo       TEXT NOT NULL DEFAULT '',
  usuario_id   TEXT,
  creado_en    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_productos_empresa_desc
  ON productos (empresa_id, descripcion);
CREATE INDEX IF NOT EXISTS idx_stock_mov_producto
  ON stock_movimientos (empresa_id, producto_id, creado_en DESC);

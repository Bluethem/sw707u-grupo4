CREATE TABLE IF NOT EXISTS clientes (
  id             TEXT PRIMARY KEY,
  empresa_id     TEXT NOT NULL REFERENCES empresas(id),
  tipo_doc       TEXT NOT NULL CHECK (tipo_doc IN ('DNI', 'RUC')),
  numero         TEXT NOT NULL,
  razon_social   TEXT NOT NULL,
  direccion      TEXT NOT NULL DEFAULT '',
  correo         TEXT NOT NULL DEFAULT '',
  activo         BOOLEAN NOT NULL DEFAULT true,
  creado_en      TIMESTAMPTZ NOT NULL DEFAULT now(),
  actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (empresa_id, tipo_doc, numero)
);

CREATE INDEX IF NOT EXISTS idx_clientes_empresa_razon
  ON clientes (empresa_id, razon_social);

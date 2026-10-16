-- M6+ resumen diario, bajas y outbox extendido

ALTER TABLE comprobantes
  ADD COLUMN IF NOT EXISTS baja_motivo TEXT NOT NULL DEFAULT '';

ALTER TABLE outbox_envios
  ALTER COLUMN comprobante_id DROP NOT NULL;

ALTER TABLE outbox_envios
  ADD COLUMN IF NOT EXISTS referencia_id TEXT;

CREATE TABLE IF NOT EXISTS resumenes_diarios (
  id            TEXT PRIMARY KEY,
  empresa_id    TEXT NOT NULL REFERENCES empresas(id),
  fecha         DATE NOT NULL,
  correlativo   INT NOT NULL DEFAULT 1,
  estado        TEXT NOT NULL DEFAULT 'PENDIENTE'
                CHECK (estado IN ('PENDIENTE','ENVIADO','ACEPTADO','RECHAZADO')),
  ticket        TEXT NOT NULL DEFAULT '',
  creado_en     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (empresa_id, fecha, correlativo)
);

CREATE TABLE IF NOT EXISTS resumen_comprobantes (
  resumen_id     TEXT NOT NULL REFERENCES resumenes_diarios(id) ON DELETE CASCADE,
  comprobante_id TEXT NOT NULL REFERENCES comprobantes(id),
  anulado        BOOLEAN NOT NULL DEFAULT false,
  PRIMARY KEY (resumen_id, comprobante_id)
);

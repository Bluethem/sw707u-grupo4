-- M6 worker: columnas de seguimiento, estados normalizados e índice para claim.
-- Estados: pending | processing | done | failed

ALTER TABLE outbox_envios
  DROP CONSTRAINT IF EXISTS outbox_envios_estado_check;

ALTER TABLE outbox_envios
  ADD CONSTRAINT outbox_envios_estado_check
  CHECK (estado IN ('pending', 'processing', 'done', 'failed'));

ALTER TABLE outbox_envios
  ADD COLUMN IF NOT EXISTS ultimo_error TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE INDEX IF NOT EXISTS idx_outbox_claim
  ON outbox_envios (estado, disponible_en)
  WHERE estado = 'pending';

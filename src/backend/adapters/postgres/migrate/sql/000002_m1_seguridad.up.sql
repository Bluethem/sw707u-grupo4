CREATE TABLE IF NOT EXISTS empresas (
  id               TEXT PRIMARY KEY,
  ruc              TEXT NOT NULL UNIQUE,
  razon_social     TEXT NOT NULL,
  domicilio_fiscal TEXT NOT NULL DEFAULT '',
  regimen          TEXT NOT NULL DEFAULT '',
  activa           BOOLEAN NOT NULL DEFAULT true,
  creado_en        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS usuarios (
  id            TEXT PRIMARY KEY,
  empresa_id    TEXT NOT NULL REFERENCES empresas(id),
  email         TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  nombre        TEXT NOT NULL DEFAULT '',
  rol           TEXT NOT NULL CHECK (rol IN ('administrador', 'cajero')),
  activo        BOOLEAN NOT NULL DEFAULT true,
  creado_en     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (empresa_id, email)
);

ALTER TABLE series_correlativos
  ADD COLUMN IF NOT EXISTS activa BOOLEAN NOT NULL DEFAULT true;

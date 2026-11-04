"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Card, Input } from "@/components/ui";
import { ApiError, apiFetch } from "@/lib/api";
import {
  AFECTACIONES_IGV,
  UNIDADES_SUNAT,
  type ConfigCatalogo,
} from "@/lib/catalogo-config";

const UNIT_HINTS: Record<string, string> = {
  NIU: "Unidad",
  ZZ: "Servicio",
  KGM: "Kilogramo",
  LTR: "Litro",
  MTR: "Metro",
  BX: "Caja",
};

const sectionTitle: React.CSSProperties = {
  margin: "0 0 var(--spacing-12)",
  fontSize: "var(--text-body)",
  fontWeight: "var(--font-weight-medium)",
  color: "var(--color-ink-black)",
};

const sectionGap: React.CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: "var(--spacing-16)",
};

const fieldRow: React.CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--spacing-16)",
};

const fieldCol: React.CSSProperties = {
  flex: "1 1 180px",
  minWidth: 0,
  maxWidth: 280,
};

function applyConfigToForm(
  data: ConfigCatalogo,
  setters: {
    setPrecioMinimo: (v: string) => void;
    setPrecioMaximo: (v: string) => void;
    setCargaMaxFilas: (v: string) => void;
    setCargaMaxMB: (v: string) => void;
    setUnidades: (v: string[]) => void;
    setAfectaciones: (v: string[]) => void;
  },
) {
  setters.setPrecioMinimo(data.precio_minimo);
  setters.setPrecioMaximo(data.precio_maximo ?? "");
  setters.setCargaMaxFilas(String(data.carga_max_filas));
  setters.setCargaMaxMB(String(data.carga_max_mb));
  setters.setUnidades([...(data.unidades_permitidas ?? [])]);
  setters.setAfectaciones([...(data.afectaciones_permitidas ?? [])]);
}

function validateForm(input: {
  precioMinimo: string;
  precioMaximo: string;
  cargaMaxFilas: string;
  cargaMaxMB: string;
  unidades: string[];
  afectaciones: string[];
}): string | undefined {
  const min = Number(input.precioMinimo);
  if (!Number.isFinite(min) || min <= 0) {
    return "El precio mínimo debe ser mayor que 0";
  }
  if (input.precioMaximo.trim() !== "") {
    const max = Number(input.precioMaximo);
    if (!Number.isFinite(max) || max < min) {
      return "El precio máximo debe ser mayor o igual al mínimo";
    }
  }
  const filas = Number(input.cargaMaxFilas);
  if (!Number.isInteger(filas) || filas < 1 || filas > 2000) {
    return "Máximo de filas debe estar entre 1 y 2000";
  }
  const mb = Number(input.cargaMaxMB);
  if (!Number.isInteger(mb) || mb < 1 || mb > 20) {
    return "Tamaño máximo debe estar entre 1 y 20 MB";
  }
  if (input.unidades.length === 0) {
    return "Selecciona al menos una unidad";
  }
  if (input.afectaciones.length === 0) {
    return "Selecciona al menos una afectación";
  }
  return undefined;
}

export function ConfigCatalogoAdmin() {
  const router = useRouter();
  const [cfg, setCfg] = useState<ConfigCatalogo | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const [success, setSuccess] = useState(false);

  const [precioMinimo, setPrecioMinimo] = useState("");
  const [precioMaximo, setPrecioMaximo] = useState("");
  const [cargaMaxFilas, setCargaMaxFilas] = useState("");
  const [cargaMaxMB, setCargaMaxMB] = useState("");
  const [unidades, setUnidades] = useState<string[]>([]);
  const [afectaciones, setAfectaciones] = useState<string[]>([]);

  const formSetters = {
    setPrecioMinimo,
    setPrecioMaximo,
    setCargaMaxFilas,
    setCargaMaxMB,
    setUnidades,
    setAfectaciones,
  };

  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    try {
      const data = await apiFetch<ConfigCatalogo>("/api/v1/configuracion/catalogo");
      setCfg(data);
      applyConfigToForm(data, formSetters);
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        router.replace("/login");
        return;
      }
      setError(e instanceof Error ? e.message : "Error al cargar configuración");
    } finally {
      setLoading(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- setters are stable setState
  }, [router]);

  useEffect(() => {
    void load();
  }, [load]);

  function toggleUnidad(u: string) {
    setUnidades((prev) => (prev.includes(u) ? prev.filter((x) => x !== u) : [...prev, u]));
  }

  function toggleAfectacion(a: string) {
    setAfectaciones((prev) => (prev.includes(a) ? prev.filter((x) => x !== a) : [...prev, a]));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const localErr = validateForm({
      precioMinimo,
      precioMaximo,
      cargaMaxFilas,
      cargaMaxMB,
      unidades,
      afectaciones,
    });
    if (localErr) {
      setError(localErr);
      setSuccess(false);
      return;
    }

    setSaving(true);
    setError(undefined);
    setSuccess(false);
    try {
      const updated = await apiFetch<ConfigCatalogo>("/api/v1/configuracion/catalogo", {
        method: "PATCH",
        body: {
          precio_minimo: precioMinimo,
          precio_maximo: precioMaximo.trim(),
          carga_max_filas: Number(cargaMaxFilas),
          carga_max_mb: Number(cargaMaxMB),
          unidades_permitidas: unidades,
          afectaciones_permitidas: afectaciones,
        },
      });
      setCfg(updated);
      applyConfigToForm(updated, formSetters);
      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Error al guardar");
    } finally {
      setSaving(false);
    }
  }

  const canSave =
    !saving &&
    !loading &&
    unidades.length > 0 &&
    afectaciones.length > 0 &&
    precioMinimo.trim() !== "" &&
    cargaMaxFilas.trim() !== "" &&
    cargaMaxMB.trim() !== "";

  return (
    <>
      <div
        style={{
          display: "flex",
          flexWrap: "wrap",
          alignItems: "center",
          justifyContent: "space-between",
          gap: "var(--spacing-12)",
          marginBottom: "var(--spacing-16)",
        }}
      >
        <h1
          style={{
            margin: 0,
            fontSize: "var(--text-subheading)",
            fontFamily: "var(--font-family-display)",
            fontWeight: "var(--font-weight-regular)",
          }}
        >
          Configuración
        </h1>
        {cfg?.actualizado_en ? (
          <span style={{ fontSize: "13px", color: "var(--color-warm-gray)" }}>
            Actualizado {new Date(cfg.actualizado_en).toLocaleString("es-PE")}
          </span>
        ) : null}
      </div>

      <Card style={{ padding: "var(--spacing-16)", boxShadow: "none" }}>
        {loading ? (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Cargando…</p>
        ) : (
          <form onSubmit={(e) => void handleSubmit(e)} style={sectionGap}>
            <p style={{ margin: 0, fontSize: "var(--text-body)", color: "var(--color-warm-gray)" }}>
              Reglas para alta/edición de productos y carga masiva. No modifica productos ya
              guardados.
            </p>

            <div>
              <h2 style={sectionTitle}>Rango de precios (S/)</h2>
              <div style={fieldRow}>
                <div style={fieldCol}>
                  <Input
                    label="Precio mínimo"
                    type="number"
                    step="0.01"
                    min="0.01"
                    value={precioMinimo}
                    onChange={(e) => setPrecioMinimo(e.target.value)}
                    required
                  />
                </div>
                <div style={fieldCol}>
                  <Input
                    label="Precio máximo (vacío = sin límite)"
                    type="number"
                    step="0.01"
                    min="0.01"
                    value={precioMaximo}
                    onChange={(e) => setPrecioMaximo(e.target.value)}
                    placeholder="Sin límite"
                  />
                </div>
              </div>
            </div>

            <div
              style={{
                borderTop: "1px solid var(--color-stone-border)",
                paddingTop: "var(--spacing-16)",
              }}
            >
              <h2 style={sectionTitle}>Límites de carga masiva</h2>
              <div style={fieldRow}>
                <div style={fieldCol}>
                  <Input
                    label="Máximo de filas"
                    type="number"
                    min="1"
                    max="2000"
                    value={cargaMaxFilas}
                    onChange={(e) => setCargaMaxFilas(e.target.value)}
                    required
                  />
                </div>
                <div style={fieldCol}>
                  <Input
                    label="Tamaño máximo (MB)"
                    type="number"
                    min="1"
                    max="20"
                    value={cargaMaxMB}
                    onChange={(e) => setCargaMaxMB(e.target.value)}
                    required
                  />
                </div>
              </div>
            </div>

            <div
              style={{
                borderTop: "1px solid var(--color-stone-border)",
                paddingTop: "var(--spacing-16)",
              }}
            >
              <h2 style={sectionTitle}>Unidades de medida permitidas</h2>
              <ChipRow
                options={UNIDADES_SUNAT}
                selected={unidades}
                onToggle={toggleUnidad}
                labelize={(u) => (UNIT_HINTS[u] ? `${u} · ${UNIT_HINTS[u]}` : u)}
                emptyLabel="Selecciona al menos una unidad"
              />
            </div>

            <div
              style={{
                borderTop: "1px solid var(--color-stone-border)",
                paddingTop: "var(--spacing-16)",
              }}
            >
              <h2 style={sectionTitle}>Afectaciones IGV permitidas</h2>
              <ChipRow
                options={AFECTACIONES_IGV}
                selected={afectaciones}
                onToggle={toggleAfectacion}
                labelize={(a) => a.charAt(0).toUpperCase() + a.slice(1)}
                emptyLabel="Selecciona al menos una afectación"
              />
            </div>

            <div
              style={{
                borderTop: "1px solid var(--color-stone-border)",
                paddingTop: "var(--spacing-16)",
                display: "flex",
                flexWrap: "wrap",
                alignItems: "center",
                gap: "var(--spacing-12)",
              }}
            >
              <Button type="submit" disabled={!canSave}>
                {saving ? "Guardando…" : "Guardar cambios"}
              </Button>
              {success ? (
                <span style={{ fontSize: "var(--text-body)", color: "var(--color-cyan-edge)" }}>
                  Guardado correctamente
                </span>
              ) : null}
              {error ? (
                <span style={{ fontSize: "var(--text-body)", color: "var(--color-error-text)" }}>
                  {error}
                </span>
              ) : null}
            </div>
          </form>
        )}
      </Card>
    </>
  );
}

function ChipRow({
  options,
  selected,
  onToggle,
  labelize,
  emptyLabel,
}: {
  options: readonly string[];
  selected: string[];
  onToggle: (value: string) => void;
  labelize?: (value: string) => string;
  emptyLabel: string;
}) {
  return (
    <div>
      <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--spacing-8)" }}>
        {options.map((opt) => {
          const active = selected.includes(opt);
          return (
            <label
              key={opt}
              style={{
                display: "inline-flex",
                alignItems: "center",
                padding: "6px 12px",
                borderRadius: "var(--radius-buttons)",
                fontSize: "var(--text-body)",
                cursor: "pointer",
                userSelect: "none",
                border: `1px solid ${active ? "var(--color-ink-black)" : "var(--color-stone-border)"}`,
                background: active ? "var(--color-sky-wash)" : "var(--color-pure-white)",
                color: active ? "var(--color-ink-black)" : "var(--color-warm-gray)",
                fontWeight: active ? "var(--font-weight-medium)" : "var(--font-weight-regular)",
              }}
            >
              <input
                type="checkbox"
                style={{ display: "none" }}
                checked={active}
                onChange={() => onToggle(opt)}
              />
              {labelize ? labelize(opt) : opt}
            </label>
          );
        })}
      </div>
      {selected.length === 0 ? (
        <span
          style={{
            fontSize: "13px",
            color: "var(--color-error-text)",
            marginTop: "var(--spacing-8)",
            display: "block",
          }}
        >
          {emptyLabel}
        </span>
      ) : null}
    </div>
  );
}

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

const sectionLabel: React.CSSProperties = {
  fontSize: "var(--text-body)",
  color: "var(--color-warm-gray)",
  fontWeight: "var(--font-weight-medium)",
  marginBottom: "var(--spacing-8)",
  display: "block",
};

const chipBase: React.CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  padding: "4px 12px",
  borderRadius: "var(--radius-buttons)",
  fontSize: "var(--text-body)",
  cursor: "pointer",
  border: "1px solid var(--color-stone-border)",
  userSelect: "none",
};

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

  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    try {
      const data = await apiFetch<ConfigCatalogo>("/api/v1/configuracion/catalogo");
      setCfg(data);
      setPrecioMinimo(data.precio_minimo);
      setPrecioMaximo(data.precio_maximo ?? "");
      setCargaMaxFilas(String(data.carga_max_filas));
      setCargaMaxMB(String(data.carga_max_mb));
      setUnidades(data.unidades_permitidas ?? []);
      setAfectaciones(data.afectaciones_permitidas ?? []);
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        router.replace("/login");
        return;
      }
      setError(e instanceof Error ? e.message : "Error al cargar configuración");
    } finally {
      setLoading(false);
    }
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
    setSaving(true);
    setError(undefined);
    setSuccess(false);
    try {
      const updated = await apiFetch<ConfigCatalogo>("/api/v1/configuracion/catalogo", {
        method: "PATCH",
        body: {
          precio_minimo: precioMinimo || undefined,
          precio_maximo: precioMaximo,
          carga_max_filas: cargaMaxFilas ? Number(cargaMaxFilas) : undefined,
          carga_max_mb: cargaMaxMB ? Number(cargaMaxMB) : undefined,
          unidades_permitidas: unidades,
          afectaciones_permitidas: afectaciones,
        },
      });
      setCfg(updated);
      setPrecioMaximo(updated.precio_maximo ?? "");
      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Error al guardar");
    } finally {
      setSaving(false);
    }
  }

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
          <form onSubmit={(e) => void handleSubmit(e)}>
            <div style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-20)" }}>
              <div>
                <span style={sectionLabel}>Rango de precios (S/)</span>
                <div style={{ display: "flex", gap: "var(--spacing-16)", flexWrap: "wrap" }}>
                  <div style={{ flex: "1 1 200px" }}>
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
                  <div style={{ flex: "1 1 200px" }}>
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

              <div>
                <span style={sectionLabel}>Límites de carga masiva</span>
                <div style={{ display: "flex", gap: "var(--spacing-16)", flexWrap: "wrap" }}>
                  <div style={{ flex: "1 1 200px" }}>
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
                  <div style={{ flex: "1 1 200px" }}>
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

              <div>
                <span style={sectionLabel}>Unidades de medida permitidas</span>
                <ChipRow
                  options={UNIDADES_SUNAT}
                  selected={unidades}
                  onToggle={toggleUnidad}
                  emptyLabel="Selecciona al menos una unidad"
                />
              </div>

              <div>
                <span style={sectionLabel}>Afectaciones IGV permitidas</span>
                <ChipRow
                  options={AFECTACIONES_IGV}
                  selected={afectaciones}
                  onToggle={toggleAfectacion}
                  labelize={(a) => a.charAt(0).toUpperCase() + a.slice(1)}
                  emptyLabel="Selecciona al menos una afectación"
                />
              </div>

              <div style={{ display: "flex", alignItems: "center", gap: "var(--spacing-12)" }}>
                <Button
                  type="submit"
                  disabled={saving || unidades.length === 0 || afectaciones.length === 0}
                >
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
                ...chipBase,
                background: active ? "var(--color-sky-wash)" : "var(--color-pure-white)",
                color: active ? "var(--color-ink-black)" : "var(--color-warm-gray)",
                borderColor: active ? "var(--color-ink-black)" : "var(--color-stone-border)",
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
        <span style={{ fontSize: "13px", color: "var(--color-error-text)", marginTop: 4, display: "block" }}>
          {emptyLabel}
        </span>
      ) : null}
    </div>
  );
}

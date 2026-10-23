"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Card, Input } from "@/components/ui";
import { ApiError, apiFetch } from "@/lib/api";

type ConfigCatalogo = {
  empresa_id: string;
  precio_minimo: string;
  precio_maximo: string | null;
  carga_max_filas: number;
  carga_max_mb: number;
  unidades_permitidas: string[];
  afectaciones_permitidas: string[];
  actualizado_en: string;
};

const UNIDADES_SUNAT = ["NIU", "ZZ", "KGM", "LTR", "MTR", "BX"];
const AFECTACIONES = ["gravado", "exonerado", "inafecto"];

export function ConfigCatalogoAdmin() {
  const router = useRouter();
  const [cfg, setCfg] = useState<ConfigCatalogo | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const [success, setSuccess] = useState(false);

  // form state
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
    setUnidades((prev) =>
      prev.includes(u) ? prev.filter((x) => x !== u) : [...prev, u]
    );
  }

  function toggleAfectacion(a: string) {
    setAfectaciones((prev) =>
      prev.includes(a) ? prev.filter((x) => x !== a) : [...prev, a]
    );
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(undefined);
    setSuccess(false);
    try {
      const body: Record<string, unknown> = {
        precio_minimo: precioMinimo || undefined,
        precio_maximo: precioMaximo !== "" ? precioMaximo : null,
        carga_max_filas: cargaMaxFilas ? Number(cargaMaxFilas) : undefined,
        carga_max_mb: cargaMaxMB ? Number(cargaMaxMB) : undefined,
        unidades_permitidas: unidades,
        afectaciones_permitidas: afectaciones,
      };
      const updated = await apiFetch<ConfigCatalogo>("/api/v1/configuracion/catalogo", {
        method: "PATCH",
        body,
      });
      setCfg(updated);
      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Error al guardar");
    } finally {
      setSaving(false);
    }
  }

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
    gap: "6px",
    padding: "4px 12px",
    borderRadius: "var(--radius-buttons)",
    fontSize: "var(--text-body)",
    cursor: "pointer",
    border: "1px solid var(--color-stone-border)",
    userSelect: "none",
  };

  return (
    <>
      <div
        style={{
          display: "flex",
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
          Configuración del catálogo
        </h1>
        {cfg ? (
          <span style={{ fontSize: "13px", color: "var(--color-warm-gray)" }}>
            Actualizado: {new Date(cfg.actualizado_en).toLocaleString("es-PE")}
          </span>
        ) : null}
      </div>

      <Card style={{ padding: "var(--spacing-24)", boxShadow: "none", maxWidth: 640 }}>
        {loading ? (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Cargando…</p>
        ) : (
          <form onSubmit={(e) => void handleSubmit(e)}>
            <div style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-20)" }}>

              {/* Precios */}
              <div>
                <span style={sectionLabel}>Rango de precios (S/)</span>
                <div style={{ display: "flex", gap: "var(--spacing-16)" }}>
                  <div style={{ flex: 1 }}>
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
                  <div style={{ flex: 1 }}>
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

              {/* Carga masiva */}
              <div>
                <span style={sectionLabel}>Límites de carga masiva</span>
                <div style={{ display: "flex", gap: "var(--spacing-16)" }}>
                  <div style={{ flex: 1 }}>
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
                  <div style={{ flex: 1 }}>
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

              {/* Unidades */}
              <div>
                <span style={sectionLabel}>Unidades de medida permitidas</span>
                <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--spacing-8)" }}>
                  {UNIDADES_SUNAT.map((u) => {
                    const active = unidades.includes(u);
                    return (
                      <label
                        key={u}
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
                          onChange={() => toggleUnidad(u)}
                        />
                        {u}
                      </label>
                    );
                  })}
                </div>
                {unidades.length === 0 ? (
                  <span style={{ fontSize: "13px", color: "var(--color-error-text)", marginTop: "4px", display: "block" }}>
                    Selecciona al menos una unidad
                  </span>
                ) : null}
              </div>

              {/* Afectaciones */}
              <div>
                <span style={sectionLabel}>Afectaciones IGV permitidas</span>
                <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--spacing-8)" }}>
                  {AFECTACIONES.map((a) => {
                    const active = afectaciones.includes(a);
                    return (
                      <label
                        key={a}
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
                          onChange={() => toggleAfectacion(a)}
                        />
                        {a.charAt(0).toUpperCase() + a.slice(1)}
                      </label>
                    );
                  })}
                </div>
                {afectaciones.length === 0 ? (
                  <span style={{ fontSize: "13px", color: "var(--color-error-text)", marginTop: "4px", display: "block" }}>
                    Selecciona al menos una afectación
                  </span>
                ) : null}
              </div>

              {/* Actions */}
              <div style={{ display: "flex", alignItems: "center", gap: "var(--spacing-12)" }}>
                <Button
                  type="submit"
                  disabled={saving || unidades.length === 0 || afectaciones.length === 0}
                >
                  {saving ? "Guardando…" : "Guardar cambios"}
                </Button>
                {success ? (
                  <span style={{ fontSize: "var(--text-body)", color: "var(--color-success-text, green)" }}>
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

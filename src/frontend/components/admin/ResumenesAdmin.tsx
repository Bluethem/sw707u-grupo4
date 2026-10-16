"use client";

import Link from "next/link";
import { Fragment, useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Card, Input } from "@/components/ui";
import { Badge, isBadgeStatus } from "@/components/ui/Badge";
import { ApiError, apiFetch } from "@/lib/api";
import type { ResumenDiarioItem } from "@/lib/caja-types";
import { tdStyle, thStyle } from "./table-styles";

function todayISO(): string {
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

export function ResumenesAdmin() {
  const router = useRouter();
  const [items, setItems] = useState<ResumenDiarioItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();
  const [fecha, setFecha] = useState(todayISO);
  const [generating, setGenerating] = useState(false);
  const [notice, setNotice] = useState<string | undefined>();
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    try {
      const data = await apiFetch<ResumenDiarioItem[]>("/api/v1/resumenes");
      setItems(Array.isArray(data) ? data : []);
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        router.replace("/login");
        return;
      }
      setError(e instanceof Error ? e.message : "Error al cargar resúmenes");
      setItems([]);
    } finally {
      setLoading(false);
    }
  }, [router]);

  useEffect(() => {
    void load();
  }, [load]);

  async function generar() {
    const f = fecha.trim();
    if (!f) {
      setError("Indica una fecha");
      return;
    }
    setGenerating(true);
    setError(undefined);
    setNotice(undefined);
    try {
      const created = await apiFetch<ResumenDiarioItem>("/api/v1/resumenes", {
        method: "POST",
        body: { fecha: f },
      });
      setNotice(`Resumen ${created.fecha} generado (${created.estado}).`);
      void load();
      setExpandedId(created.id);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "No se pudo generar el resumen");
    } finally {
      setGenerating(false);
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
          Resúmenes diarios
        </h1>
      </div>

      <Card style={{ padding: "var(--spacing-16)", boxShadow: "none", marginBottom: "var(--spacing-16)" }}>
        <div
          style={{
            display: "flex",
            flexWrap: "wrap",
            gap: "var(--spacing-12)",
            alignItems: "flex-end",
          }}
        >
          <div style={{ flex: "1 1 180px", maxWidth: "240px" }}>
            <Input
              label="Fecha de emisión"
              type="date"
              value={fecha}
              onChange={(ev) => setFecha(ev.target.value)}
            />
          </div>
          <Button type="button" disabled={generating} onClick={() => void generar()}>
            {generating ? "Generando…" : "Generar resumen"}
          </Button>
        </div>
        {notice ? (
          <p style={{ margin: "var(--spacing-12) 0 0", color: "var(--color-cyan-edge)", fontSize: "var(--text-body)" }}>
            {notice}
          </p>
        ) : null}
      </Card>

      <Card style={{ padding: "var(--spacing-16)", boxShadow: "none" }}>
        {error ? (
          <p style={{ margin: "0 0 var(--spacing-12)", color: "var(--color-error-text)", fontSize: "var(--text-body)" }}>
            {error}
          </p>
        ) : null}

        {loading ? (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Cargando…</p>
        ) : items.length === 0 ? (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>No hay resúmenes.</p>
        ) : (
          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <thead>
                <tr>
                  <th style={thStyle}>Fecha</th>
                  <th style={thStyle}>Correlativo</th>
                  <th style={thStyle}>Estado</th>
                  <th style={thStyle}>Boletas</th>
                  <th style={thStyle}>Acciones</th>
                </tr>
              </thead>
              <tbody>
                {items.map((r) => (
                  <Fragment key={r.id}>
                    <tr>
                      <td style={tdStyle}>{r.fecha}</td>
                      <td style={tdStyle}>{r.correlativo}</td>
                      <td style={tdStyle}>
                        {isBadgeStatus(r.estado) ? (
                          <Badge status={r.estado} />
                        ) : (
                          r.estado
                        )}
                      </td>
                      <td style={tdStyle}>{r.comprobantes?.length ?? 0}</td>
                      <td style={tdStyle}>
                        <Button
                          type="button"
                          variant="ghost"
                          style={{ padding: "4px 10px" }}
                          onClick={() => setExpandedId((cur) => (cur === r.id ? null : r.id))}
                        >
                          {expandedId === r.id ? "Ocultar" : "Detalle"}
                        </Button>
                      </td>
                    </tr>
                    {expandedId === r.id ? (
                      <tr>
                        <td colSpan={5} style={{ ...tdStyle, background: "var(--color-stone-canvas)" }}>
                          {(r.comprobantes ?? []).length === 0 ? (
                            <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Sin comprobantes.</p>
                          ) : (
                            <ul style={{ margin: 0, paddingLeft: "1.25rem", fontSize: "var(--text-body)" }}>
                              {r.comprobantes.map((c) => (
                                <li key={c.comprobante_id}>
                                  <Link
                                    href={`/comprobantes/${encodeURIComponent(c.comprobante_id)}`}
                                    style={{ color: "var(--color-cyan-edge)" }}
                                  >
                                    {c.comprobante_id}
                                  </Link>
                                  {c.anulado ? (
                                    <span style={{ color: "var(--color-warm-gray)", marginLeft: "8px" }}>
                                      (anulado en resumen)
                                    </span>
                                  ) : null}
                                </li>
                              ))}
                            </ul>
                          )}
                        </td>
                      </tr>
                    ) : null}
                  </Fragment>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </>
  );
}

"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Card } from "@/components/ui";
import { ApiError, apiFetch } from "@/lib/api";
import type { SerieItem } from "@/lib/caja-types";
import { SerieFormModal } from "./SerieFormModal";
import { tdStyle, thStyle } from "./table-styles";

export function SeriesAdmin() {
  const router = useRouter();
  const [items, setItems] = useState<SerieItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();
  const [formOpen, setFormOpen] = useState(false);
  const [formMode, setFormMode] = useState<"create" | "edit">("create");
  const [editRow, setEditRow] = useState<SerieItem | null>(null);
  const [togglingKey, setTogglingKey] = useState<string | null>(null);

  const rowKey = (s: SerieItem) => `${s.tipo}:${s.serie}`;

  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    try {
      const data = await apiFetch<SerieItem[]>("/api/v1/series");
      setItems(Array.isArray(data) ? data : []);
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        router.replace("/login");
        return;
      }
      setError(e instanceof Error ? e.message : "Error al cargar series");
      setItems([]);
    } finally {
      setLoading(false);
    }
  }, [router]);

  useEffect(() => {
    void load();
  }, [load]);

  function openCreate() {
    setFormMode("create");
    setEditRow(null);
    setFormOpen(true);
  }

  function openEdit(s: SerieItem) {
    setFormMode("edit");
    setEditRow(s);
    setFormOpen(true);
  }

  async function toggleActiva(s: SerieItem) {
    const key = rowKey(s);
    setTogglingKey(key);
    try {
      await apiFetch("/api/v1/series", {
        method: "PATCH",
        body: { tipo: s.tipo, serie: s.serie, activa: !s.activa },
      });
      void load();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "No se pudo actualizar");
    } finally {
      setTogglingKey(null);
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
          Series
        </h1>
        <Button type="button" onClick={openCreate}>
          Nueva serie
        </Button>
      </div>

      <Card style={{ padding: "var(--spacing-16)", boxShadow: "none" }}>
        {error ? (
          <p style={{ margin: 0, marginBottom: "var(--spacing-12)", color: "var(--color-error-text)", fontSize: "var(--text-body)" }}>
            {error}
          </p>
        ) : null}

        {loading ? (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Cargando…</p>
        ) : items.length === 0 ? (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>No hay series configuradas.</p>
        ) : (
          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <thead>
                <tr>
                  <th style={thStyle}>Tipo</th>
                  <th style={thStyle}>Serie</th>
                  <th style={thStyle}>Siguiente</th>
                  <th style={thStyle}>Activa</th>
                  <th style={thStyle}>Acciones</th>
                </tr>
              </thead>
              <tbody>
                {items.map((s) => (
                  <tr key={rowKey(s)}>
                    <td style={tdStyle}>{s.tipo}</td>
                    <td style={tdStyle}>{s.serie}</td>
                    <td style={tdStyle}>{s.siguiente}</td>
                    <td style={tdStyle}>{s.activa ? "Sí" : "No"}</td>
                    <td style={{ ...tdStyle, whiteSpace: "nowrap" }}>
                      <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--spacing-4)" }}>
                        <Button
                          type="button"
                          variant="ghost"
                          onClick={() => openEdit(s)}
                          style={{ padding: "4px 10px" }}
                        >
                          Editar
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          disabled={togglingKey === rowKey(s)}
                          onClick={() => void toggleActiva(s)}
                          style={{ padding: "4px 10px" }}
                        >
                          {s.activa ? "Desactivar" : "Activar"}
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <SerieFormModal
        open={formOpen}
        mode={formMode}
        serieRow={editRow}
        onClose={() => setFormOpen(false)}
        onSaved={() => void load()}
      />
    </>
  );
}

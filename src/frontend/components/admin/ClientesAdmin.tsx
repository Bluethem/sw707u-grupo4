"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Card, Input } from "@/components/ui";
import { ApiError, apiFetch } from "@/lib/api";
import type { ClienteItem, Paginated } from "@/lib/caja-types";
import { ClienteFormModal } from "./ClienteFormModal";
import { selectFieldStyle, tdStyle, thStyle } from "./table-styles";

type ActivoFilter = "" | "true" | "false";

export function ClientesAdmin() {
  const router = useRouter();
  const [query, setQuery] = useState("");
  const [activoFilter, setActivoFilter] = useState<ActivoFilter>("");
  const [items, setItems] = useState<ClienteItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();

  const [formOpen, setFormOpen] = useState(false);
  const [formMode, setFormMode] = useState<"create" | "edit">("create");
  const [editId, setEditId] = useState<string | null>(null);
  const [deactivatingId, setDeactivatingId] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    const params = new URLSearchParams({ limit: "100" });
    const q = query.trim();
    if (q) params.set("q", q);
    if (activoFilter) params.set("activo", activoFilter);
    try {
      const data = await apiFetch<Paginated<ClienteItem>>(`/api/v1/clientes?${params}`);
      setItems(data.items ?? []);
      setTotal(data.total ?? 0);
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        router.replace("/login");
        return;
      }
      setError(e instanceof Error ? e.message : "Error al cargar clientes");
      setItems([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, [query, activoFilter, router]);

  useEffect(() => {
    const t = window.setTimeout(() => {
      void load();
    }, 280);
    return () => window.clearTimeout(t);
  }, [load]);

  function openCreate() {
    setFormMode("create");
    setEditId(null);
    setFormOpen(true);
  }

  function openEdit(c: ClienteItem) {
    setFormMode("edit");
    setEditId(c.id);
    setFormOpen(true);
  }

  async function deactivate(c: ClienteItem) {
    if (!c.activo) return;
    if (!window.confirm(`¿Desactivar ${c.tipo_doc} ${c.numero}?`)) return;
    setDeactivatingId(c.id);
    try {
      await apiFetch(`/api/v1/clientes/${c.id}`, {
        method: "PATCH",
        body: { activo: false },
      });
      void load();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "No se pudo desactivar");
    } finally {
      setDeactivatingId(null);
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
          Clientes
        </h1>
        <Button type="button" onClick={openCreate}>
          Nuevo cliente
        </Button>
      </div>

      <Card style={{ padding: "var(--spacing-16)", boxShadow: "none" }}>
        <div
          style={{
            display: "flex",
            flexWrap: "wrap",
            gap: "var(--spacing-12)",
            marginBottom: "var(--spacing-16)",
            alignItems: "flex-end",
          }}
        >
          <div style={{ flex: "1 1 200px", maxWidth: "320px" }}>
            <Input
              label="Buscar"
              placeholder="Documento o razón social"
              value={query}
              onChange={(ev) => setQuery(ev.target.value)}
            />
          </div>
          <label style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-4)" }}>
            <span style={{ fontSize: "var(--text-body)", color: "var(--color-warm-gray)" }}>Estado</span>
            <select
              value={activoFilter}
              onChange={(ev) => setActivoFilter(ev.target.value as ActivoFilter)}
              style={selectFieldStyle}
            >
              <option value="">Todos</option>
              <option value="true">Activos</option>
              <option value="false">Inactivos</option>
            </select>
          </label>
          <span style={{ fontSize: "13px", color: "var(--color-warm-gray)", paddingBottom: "6px" }}>
            {total} registro{total === 1 ? "" : "s"}
          </span>
        </div>

        {error ? (
          <p style={{ margin: 0, marginBottom: "var(--spacing-12)", color: "var(--color-error-text)", fontSize: "var(--text-body)" }}>
            {error}
          </p>
        ) : null}

        {loading ? (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Cargando…</p>
        ) : items.length === 0 ? (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>No hay clientes.</p>
        ) : (
          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <thead>
                <tr>
                  <th style={thStyle}>Tipo doc.</th>
                  <th style={thStyle}>Número</th>
                  <th style={thStyle}>Razón social</th>
                  <th style={thStyle}>Activo</th>
                  <th style={thStyle}>Acciones</th>
                </tr>
              </thead>
              <tbody>
                {items.map((c) => (
                  <tr key={c.id}>
                    <td style={tdStyle}>{c.tipo_doc}</td>
                    <td style={tdStyle}>{c.numero}</td>
                    <td style={tdStyle}>{c.razon_social}</td>
                    <td style={tdStyle}>{c.activo ? "Sí" : "No"}</td>
                    <td style={{ ...tdStyle, whiteSpace: "nowrap" }}>
                      <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--spacing-4)" }}>
                        <Button
                          type="button"
                          variant="ghost"
                          onClick={() => openEdit(c)}
                          style={{ padding: "4px 10px" }}
                        >
                          Editar
                        </Button>
                        {c.activo ? (
                          <Button
                            type="button"
                            variant="ghost"
                            disabled={deactivatingId === c.id}
                            onClick={() => void deactivate(c)}
                            style={{ padding: "4px 10px", color: "var(--color-error-text)" }}
                          >
                            Desactivar
                          </Button>
                        ) : null}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <ClienteFormModal
        open={formOpen}
        mode={formMode}
        clienteId={editId}
        onClose={() => setFormOpen(false)}
        onSaved={() => void load()}
      />
    </>
  );
}

"use client";

import { Fragment, useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Card, Input } from "@/components/ui";
import { ApiError, apiFetch } from "@/lib/api";
import type { Paginated, ProductoItem } from "@/lib/caja-types";
import { formatMoneyDecimal } from "@/lib/money";
import { ProductoAjusteStockModal } from "./ProductoAjusteStockModal";
import { ProductoFormModal } from "./ProductoFormModal";
import { ProductosMovimientosPanel } from "./ProductosMovimientosPanel";
import { selectFieldStyle, tdStyle, thStyle } from "./table-styles";

type ActivoFilter = "" | "true" | "false";

export function ProductosAdmin() {
  const router = useRouter();
  const [query, setQuery] = useState("");
  const [activoFilter, setActivoFilter] = useState<ActivoFilter>("");
  const [items, setItems] = useState<ProductoItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();

  const [formOpen, setFormOpen] = useState(false);
  const [formMode, setFormMode] = useState<"create" | "edit">("create");
  const [editId, setEditId] = useState<string | null>(null);

  const [ajusteOpen, setAjusteOpen] = useState(false);
  const [ajusteProducto, setAjusteProducto] = useState<ProductoItem | null>(null);

  const [expandedMovimientos, setExpandedMovimientos] = useState<string | null>(null);
  const [deactivatingId, setDeactivatingId] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    const params = new URLSearchParams({ limit: "100" });
    const q = query.trim();
    if (q) params.set("q", q);
    if (activoFilter) params.set("activo", activoFilter);
    try {
      const data = await apiFetch<Paginated<ProductoItem>>(`/api/v1/productos?${params}`);
      setItems(data.items ?? []);
      setTotal(data.total ?? 0);
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        router.replace("/login");
        return;
      }
      setError(e instanceof Error ? e.message : "Error al cargar productos");
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

  function openEdit(p: ProductoItem) {
    setFormMode("edit");
    setEditId(p.id);
    setFormOpen(true);
  }

  function openAjuste(p: ProductoItem) {
    setAjusteProducto(p);
    setAjusteOpen(true);
  }

  async function deactivate(p: ProductoItem) {
    if (!p.activo) return;
    if (!window.confirm(`¿Desactivar ${p.codigo}?`)) return;
    setDeactivatingId(p.id);
    try {
      await apiFetch(`/api/v1/productos/${p.id}`, {
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
          Productos
        </h1>
        <Button type="button" onClick={openCreate}>
          Nuevo producto
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
              placeholder="Código o descripción"
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
          <p style={{ margin: "0 0 var(--spacing-12)", color: "var(--color-error-text)", fontSize: "var(--text-body)" }}>
            {error}
          </p>
        ) : null}

        {loading ? (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Cargando…</p>
        ) : items.length === 0 ? (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>No hay productos.</p>
        ) : (
          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <thead>
                <tr>
                  <th style={thStyle}>Código</th>
                  <th style={thStyle}>Descripción</th>
                  <th style={thStyle}>Precio</th>
                  <th style={thStyle}>Afectación</th>
                  <th style={thStyle}>Stock</th>
                  <th style={thStyle}>Activo</th>
                  <th style={thStyle}>Acciones</th>
                </tr>
              </thead>
              <tbody>
                {items.map((p) => (
                  <Fragment key={p.id}>
                    <tr>
                      <td style={tdStyle}>{p.codigo}</td>
                      <td style={tdStyle}>{p.descripcion}</td>
                      <td style={tdStyle}>{formatMoneyDecimal(p.precio)}</td>
                      <td style={tdStyle}>{p.afectacion}</td>
                      <td style={tdStyle}>
                        {p.controla_stock ? (p.stock != null ? String(p.stock) : "—") : "—"}
                      </td>
                      <td style={tdStyle}>{p.activo ? "Sí" : "No"}</td>
                      <td style={{ ...tdStyle, whiteSpace: "nowrap" }}>
                        <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--spacing-4)" }}>
                          <Button
                            type="button"
                            variant="ghost"
                            onClick={() => openEdit(p)}
                            style={{ padding: "4px 10px" }}
                          >
                            Editar
                          </Button>
                          {p.controla_stock ? (
                            <Button
                              type="button"
                              variant="ghost"
                              onClick={() => openAjuste(p)}
                              style={{ padding: "4px 10px" }}
                            >
                              Stock
                            </Button>
                          ) : null}
                          <Button
                            type="button"
                            variant="ghost"
                            onClick={() =>
                              setExpandedMovimientos((cur) => (cur === p.id ? null : p.id))
                            }
                            style={{ padding: "4px 10px" }}
                          >
                            {expandedMovimientos === p.id ? "Ocultar mov." : "Movimientos"}
                          </Button>
                          {p.activo ? (
                            <Button
                              type="button"
                              variant="ghost"
                              disabled={deactivatingId === p.id}
                              onClick={() => void deactivate(p)}
                              style={{ padding: "4px 10px", color: "var(--color-error-text)" }}
                            >
                              Desactivar
                            </Button>
                          ) : null}
                        </div>
                      </td>
                    </tr>
                    {expandedMovimientos === p.id ? (
                      <tr>
                        <td colSpan={7} style={{ ...tdStyle, background: "var(--color-stone-canvas)" }}>
                          <ProductosMovimientosPanel productoId={p.id} />
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

      <ProductoFormModal
        open={formOpen}
        mode={formMode}
        productoId={editId}
        onClose={() => setFormOpen(false)}
        onSaved={() => void load()}
      />
      <ProductoAjusteStockModal
        open={ajusteOpen}
        productoId={ajusteProducto?.id ?? null}
        productoLabel={
          ajusteProducto ? `${ajusteProducto.codigo} — ${ajusteProducto.descripcion}` : ""
        }
        onClose={() => setAjusteOpen(false)}
        onSaved={() => void load()}
      />
    </>
  );
}

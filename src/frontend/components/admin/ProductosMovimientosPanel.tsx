"use client";

import { useEffect, useState } from "react";
import { ApiError, apiFetch } from "@/lib/api";
import type { Paginated, StockMovimiento } from "./producto-types";
import { tdStyle, thStyle } from "./table-styles";

type Props = {
  productoId: string;
};

export function ProductosMovimientosPanel({ productoId }: Props) {
  const [items, setItems] = useState<StockMovimiento[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();

  useEffect(() => {
    setLoading(true);
    setError(undefined);
    const params = new URLSearchParams({ limit: "10", offset: "0" });
    void apiFetch<Paginated<StockMovimiento>>(`/api/v1/productos/${productoId}/movimientos?${params}`)
      .then((res) => setItems(res.items ?? []))
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : "Error al cargar movimientos");
        setItems([]);
      })
      .finally(() => setLoading(false));
  }, [productoId]);

  if (loading) {
    return <p style={{ margin: "var(--spacing-8) 0 0", fontSize: "13px", color: "var(--color-warm-gray)" }}>Cargando movimientos…</p>;
  }
  if (error) {
    return <p style={{ margin: "var(--spacing-8) 0 0", fontSize: "13px", color: "var(--color-error-text)" }}>{error}</p>;
  }
  if (items.length === 0) {
    return <p style={{ margin: "var(--spacing-8) 0 0", fontSize: "13px", color: "var(--color-warm-gray)" }}>Sin movimientos recientes.</p>;
  }

  return (
    <div style={{ marginTop: "var(--spacing-8)", overflowX: "auto" }}>
      <table style={{ width: "100%", borderCollapse: "collapse" }}>
        <thead>
          <tr>
            <th style={thStyle}>Fecha</th>
            <th style={thStyle}>Tipo</th>
            <th style={thStyle}>Cantidad</th>
            <th style={thStyle}>Motivo</th>
          </tr>
        </thead>
        <tbody>
          {items.map((m) => (
            <tr key={m.id}>
              <td style={tdStyle}>{new Date(m.creado_en).toLocaleString("es-PE")}</td>
              <td style={tdStyle}>{m.tipo}</td>
              <td style={tdStyle}>{m.cantidad}</td>
              <td style={tdStyle}>{m.motivo || "—"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

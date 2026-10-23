"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui";
import { ApiError, apiFetch } from "@/lib/api";
import type { ProductoItem } from "@/lib/caja-types";
import { ModalShell } from "./ModalShell";
import { tdStyle, thStyle } from "./table-styles";

type Comprobante = {
  id: string;
  tipo: string;
  serie: string;
  numero: number;
  estado: string;
};

type Props = {
  open: boolean;
  producto: ProductoItem | null;
  onClose: () => void;
  onDeleted: () => void;
};

export function ProductoEliminarModal({ open, producto, onClose, onDeleted }: Props) {
  const [loading, setLoading] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const [comprobantes, setComprobantes] = useState<Comprobante[]>([]);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    if (!open || !producto) return;
    setLoading(true);
    setDeleting(false);
    setError(undefined);
    setComprobantes([]);
    setTotal(0);
    apiFetch<{ total_comprobantes: number; comprobantes: Comprobante[] }>(
      `/api/v1/productos/${producto.id}/dependencias`,
    )
      .then((data) => {
        setTotal(data.total_comprobantes);
        setComprobantes(data.comprobantes);
      })
      .catch((e) => {
        setError(e instanceof Error ? e.message : "Error al obtener dependencias");
      })
      .finally(() => setLoading(false));
  }, [open, producto]);

  async function handleEliminar() {
    if (!producto) return;
    setDeleting(true);
    setError(undefined);
    try {
      await apiFetch(`/api/v1/productos/${producto.id}`, { method: "DELETE" });
      onDeleted();
      onClose();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "No se pudo eliminar el producto");
      setDeleting(false);
    }
  }

  if (!producto) return null;

  return (
    <ModalShell open={open} title="Eliminar producto" onClose={onClose} maxWidth={600}>
      <p style={{ margin: "0 0 var(--spacing-12)", fontSize: "var(--text-body)" }}>
        Vas a eliminar permanentemente{" "}
        <strong>{producto.codigo} — {producto.descripcion}</strong>.
        Esta acción no se puede deshacer.
      </p>

      {loading && (
        <p style={{ color: "var(--color-warm-gray)", fontSize: "var(--text-body)" }}>
          Verificando dependencias…
        </p>
      )}

      {!loading && total === 0 && (
        <p
          style={{
            margin: "0 0 var(--spacing-16)",
            padding: "var(--spacing-12)",
            background: "var(--color-stone-canvas)",
            borderRadius: "var(--radius-cards)",
            fontSize: "var(--text-body)",
          }}
        >
          Este producto no tiene boletas ni facturas asociadas. Puede eliminarse sin
          afectar registros históricos.
        </p>
      )}

      {!loading && total > 0 && (
        <div style={{ marginBottom: "var(--spacing-16)" }}>
          <p
            style={{
              margin: "0 0 var(--spacing-8)",
              padding: "var(--spacing-12)",
              background: "#fff8e1",
              border: "1px solid #f9a825",
              borderRadius: "var(--radius-cards)",
              fontSize: "var(--text-body)",
              color: "#7a5800",
            }}
          >
            <strong>Advertencia:</strong> Este producto aparece en{" "}
            <strong>{total} comprobante{total === 1 ? "" : "s"}</strong>. Por auditoría,
            las líneas de esos comprobantes conservarán el producto_id pero el producto
            ya no estará disponible en el catálogo.
          </p>
          <div style={{ overflowX: "auto", maxHeight: 240, overflowY: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <thead>
                <tr>
                  <th style={thStyle}>Tipo</th>
                  <th style={thStyle}>Serie</th>
                  <th style={thStyle}>Número</th>
                  <th style={thStyle}>Estado</th>
                </tr>
              </thead>
              <tbody>
                {comprobantes.map((c) => (
                  <tr key={c.id}>
                    <td style={tdStyle}>{c.tipo}</td>
                    <td style={tdStyle}>{c.serie}</td>
                    <td style={tdStyle}>{c.numero}</td>
                    <td style={tdStyle}>{c.estado}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {error && (
        <p
          style={{
            margin: "0 0 var(--spacing-12)",
            color: "var(--color-error-text)",
            fontSize: "var(--text-body)",
          }}
        >
          {error}
        </p>
      )}

      {!loading && (
        <div style={{ display: "flex", justifyContent: "flex-end", gap: "var(--spacing-8)" }}>
          <Button type="button" variant="ghost" onClick={onClose} disabled={deleting}>
            Cancelar
          </Button>
          <Button
            type="button"
            disabled={deleting}
            onClick={() => void handleEliminar()}
            style={{ background: "var(--color-error-text)", color: "#fff", border: "none" }}
          >
            {deleting ? "Eliminando…" : "Eliminar definitivamente"}
          </Button>
        </div>
      )}
    </ModalShell>
  );
}

"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { AppNav } from "@/components/shell/AppNav";
import { RequireAuth } from "@/components/shell/RequireAuth";
import { Badge, isBadgeStatus } from "@/components/ui/Badge";
import { ApiError, apiFetch } from "@/lib/api";
import type { ComprobanteItem, Paginated, TipoComprobante } from "@/lib/caja-types";
import { formatMoneyDecimal } from "@/lib/money";

const cardSurface: React.CSSProperties = {
  background: "var(--color-pure-white)",
  borderRadius: "var(--radius-cards)",
  border: "1px solid var(--color-stone-border)",
  padding: "var(--spacing-16)",
  boxShadow: "none",
};

const thStyle: React.CSSProperties = {
  textAlign: "left",
  padding: "8px 12px",
  fontSize: "12px",
  fontWeight: "var(--font-weight-medium)",
  color: "var(--color-warm-gray)",
  borderBottom: "1px solid var(--color-stone-border)",
};

const tdStyle: React.CSSProperties = {
  padding: "10px 12px",
  fontSize: "var(--text-body)",
  borderBottom: "1px solid var(--color-stone-border)",
  verticalAlign: "middle",
};

function clienteLabel(c: ComprobanteItem["cliente"]): string {
  if (!c) return "—";
  const name = c.razon_social?.trim();
  if (name) return name;
  return c.numero || "—";
}

export default function ComprobantesPage() {
  const router = useRouter();
  const [tipo, setTipo] = useState<"" | TipoComprobante>("");
  const [estado, setEstado] = useState("");
  const [items, setItems] = useState<ComprobanteItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();

  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    const qs = new URLSearchParams();
    if (tipo) qs.set("tipo", tipo);
    if (estado) qs.set("estado", estado);
    qs.set("limit", "50");
    const path = `/api/v1/comprobantes?${qs.toString()}`;
    try {
      const data = await apiFetch<Paginated<ComprobanteItem>>(path);
      setItems(data.items ?? []);
      setTotal(data.total ?? 0);
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        router.replace("/login");
        return;
      }
      setError(e instanceof Error ? e.message : "Error al cargar comprobantes");
      setItems([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, [tipo, estado, router]);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <RequireAuth>
      <AppNav />
      <main
        style={{
          display: "block",
          minHeight: "calc(100vh - 57px)",
          padding: "var(--spacing-16)",
          maxWidth: "var(--page-max-width)",
          margin: "0 auto",
        }}
      >
        <h1
          style={{
            margin: "0 0 var(--spacing-16)",
            fontSize: "var(--text-subheading)",
            fontFamily: "var(--font-family-display)",
            fontWeight: "var(--font-weight-regular)",
          }}
        >
          Comprobantes
        </h1>

        <div style={{ ...cardSurface, marginBottom: "var(--spacing-16)" }}>
          <div
            style={{
              display: "flex",
              flexWrap: "wrap",
              gap: "var(--spacing-12)",
              alignItems: "flex-end",
            }}
          >
            <label style={{ display: "flex", flexDirection: "column", gap: 4, fontSize: "12px" }}>
              <span style={{ color: "var(--color-warm-gray)" }}>Tipo</span>
              <select
                value={tipo}
                onChange={(e) => setTipo(e.target.value as "" | TipoComprobante)}
                style={selectStyle}
              >
                <option value="">Todos</option>
                <option value="boleta">Boleta</option>
                <option value="factura">Factura</option>
              </select>
            </label>
            <label style={{ display: "flex", flexDirection: "column", gap: 4, fontSize: "12px" }}>
              <span style={{ color: "var(--color-warm-gray)" }}>Estado</span>
              <select value={estado} onChange={(e) => setEstado(e.target.value)} style={selectStyle}>
                <option value="">Todos</option>
                <option value="PENDIENTE">PENDIENTE</option>
                <option value="ENVIADO">ENVIADO</option>
                <option value="ACEPTADO">ACEPTADO</option>
                <option value="RECHAZADO">RECHAZADO</option>
                <option value="OBSERVADO">OBSERVADO</option>
                <option value="ANULADA">ANULADA</option>
                <option value="PENDIENTE_BAJA">PENDIENTE_BAJA</option>
              </select>
            </label>
            <button type="button" onClick={() => void load()} style={refreshBtnStyle}>
              Actualizar
            </button>
          </div>
        </div>

        {error ? (
          <p style={{ color: "var(--color-error-text)", margin: "0 0 var(--spacing-12)" }}>{error}</p>
        ) : null}

        <div style={{ ...cardSurface, overflowX: "auto" }}>
          {loading ? (
            <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Cargando…</p>
          ) : items.length === 0 ? (
            <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Sin comprobantes.</p>
          ) : (
            <>
              <p style={{ margin: "0 0 var(--spacing-12)", fontSize: "12px", color: "var(--color-warm-gray)" }}>
                {total} registro{total === 1 ? "" : "s"}
              </p>
              <table style={{ width: "100%", borderCollapse: "collapse", minWidth: 640 }}>
                <thead>
                  <tr>
                    <th style={thStyle}>Tipo</th>
                    <th style={thStyle}>Serie-Núm.</th>
                    <th style={thStyle}>Cliente</th>
                    <th style={{ ...thStyle, textAlign: "right" }}>Total</th>
                    <th style={thStyle}>Estado</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((row) => (
                    <tr key={row.id}>
                      <td style={tdStyle}>
                        <Link
                          href={`/comprobantes/${row.id}`}
                          style={{ color: "var(--color-cyan-edge)", textTransform: "capitalize" }}
                        >
                          {row.tipo}
                        </Link>
                      </td>
                      <td style={tdStyle}>
                        <Link href={`/comprobantes/${row.id}`} style={{ color: "var(--color-ink-black)" }}>
                          {row.serie}-{row.numero}
                        </Link>
                      </td>
                      <td style={tdStyle}>{clienteLabel(row.cliente)}</td>
                      <td style={{ ...tdStyle, textAlign: "right", fontWeight: "var(--font-weight-medium)" }}>
                        {formatMoneyDecimal(row.total)}
                      </td>
                      <td style={tdStyle}>
                        {isBadgeStatus(row.estado) ? (
                          <Badge status={row.estado} />
                        ) : (
                          <span style={{ fontSize: "12px" }}>{row.estado}</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </>
          )}
        </div>
      </main>
    </RequireAuth>
  );
}

const selectStyle: React.CSSProperties = {
  padding: "8px 10px",
  borderRadius: "var(--radius-inputs)",
  border: "1px solid var(--color-stone-border)",
  background: "var(--color-pure-white)",
  fontSize: "var(--text-body)",
  minWidth: 140,
};

const refreshBtnStyle: React.CSSProperties = {
  padding: "8px 14px",
  borderRadius: "var(--radius-buttons)",
  border: "1px solid var(--color-stone-border)",
  background: "var(--color-pure-white)",
  cursor: "pointer",
  fontSize: "var(--text-body)",
};

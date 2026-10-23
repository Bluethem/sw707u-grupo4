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

const PAGE_SIZE = 50;

function formatDateInput(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

function defaultFechaDesde(): string {
  const d = new Date();
  d.setDate(d.getDate() - 30);
  return formatDateInput(d);
}

function defaultFechaHasta(): string {
  return formatDateInput(new Date());
}

export default function ComprobantesPage() {
  const router = useRouter();
  const [tipo, setTipo] = useState<"" | TipoComprobante>("");
  const [estado, setEstado] = useState("");
  const [fechaDesde, setFechaDesde] = useState(defaultFechaDesde);
  const [fechaHasta, setFechaHasta] = useState(defaultFechaHasta);
  const [serie, setSerie] = useState("");
  const [numero, setNumero] = useState("");
  const [q, setQ] = useState("");
  const [offset, setOffset] = useState(0);
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
    if (fechaDesde) qs.set("fecha_desde", fechaDesde);
    if (fechaHasta) qs.set("fecha_hasta", fechaHasta);
    if (serie.trim()) qs.set("serie", serie.trim());
    if (numero.trim()) qs.set("numero", numero.trim());
    if (q.trim()) qs.set("q", q.trim());
    qs.set("limit", String(PAGE_SIZE));
    qs.set("offset", String(offset));
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
  }, [tipo, estado, fechaDesde, fechaHasta, serie, numero, q, offset, router]);

  useEffect(() => {
    void load();
  }, [load]);

  const applyFilters = () => {
    if (offset === 0) {
      void load();
    } else {
      setOffset(0);
    }
  };

  const page = Math.floor(offset / PAGE_SIZE) + 1;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const canPrev = offset > 0;
  const canNext = offset + PAGE_SIZE < total;

  return (
    <RequireAuth>
      <AppNav />
      <main
        style={{
          display: "block",
          width: "100%",
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
              <span style={{ color: "var(--color-warm-gray)" }}>Desde</span>
              <input
                type="date"
                value={fechaDesde}
                onChange={(e) => setFechaDesde(e.target.value)}
                style={inputStyle}
              />
            </label>
            <label style={{ display: "flex", flexDirection: "column", gap: 4, fontSize: "12px" }}>
              <span style={{ color: "var(--color-warm-gray)" }}>Hasta</span>
              <input
                type="date"
                value={fechaHasta}
                onChange={(e) => setFechaHasta(e.target.value)}
                style={inputStyle}
              />
            </label>
            <label style={{ display: "flex", flexDirection: "column", gap: 4, fontSize: "12px" }}>
              <span style={{ color: "var(--color-warm-gray)" }}>Serie</span>
              <input
                type="text"
                value={serie}
                onChange={(e) => setSerie(e.target.value)}
                placeholder="B001"
                style={inputStyle}
              />
            </label>
            <label style={{ display: "flex", flexDirection: "column", gap: 4, fontSize: "12px" }}>
              <span style={{ color: "var(--color-warm-gray)" }}>Número</span>
              <input
                type="text"
                inputMode="numeric"
                value={numero}
                onChange={(e) => setNumero(e.target.value.replace(/\D/g, ""))}
                placeholder="1"
                style={{ ...inputStyle, minWidth: 100 }}
              />
            </label>
            <label style={{ display: "flex", flexDirection: "column", gap: 4, fontSize: "12px", flex: "1 1 160px" }}>
              <span style={{ color: "var(--color-warm-gray)" }}>Cliente</span>
              <input
                type="search"
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder="Razón social o documento"
                style={{ ...inputStyle, minWidth: 180 }}
              />
            </label>
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
            <button type="button" onClick={applyFilters} style={refreshBtnStyle}>
              Buscar
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
              <div
                style={{
                  display: "flex",
                  flexWrap: "wrap",
                  alignItems: "center",
                  justifyContent: "space-between",
                  gap: "var(--spacing-12)",
                  marginBottom: "var(--spacing-12)",
                }}
              >
                <p style={{ margin: 0, fontSize: "12px", color: "var(--color-warm-gray)" }}>
                  {total} registro{total === 1 ? "" : "s"}
                  {total > PAGE_SIZE ? ` · página ${page} de ${totalPages}` : ""}
                </p>
                {total > PAGE_SIZE ? (
                  <div style={{ display: "flex", gap: 8 }}>
                    <button
                      type="button"
                      disabled={!canPrev}
                      onClick={() => setOffset((o) => Math.max(0, o - PAGE_SIZE))}
                      style={paginationBtnStyle(!canPrev)}
                    >
                      Anterior
                    </button>
                    <button
                      type="button"
                      disabled={!canNext}
                      onClick={() => setOffset((o) => o + PAGE_SIZE)}
                      style={paginationBtnStyle(!canNext)}
                    >
                      Siguiente
                    </button>
                  </div>
                ) : null}
              </div>
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

const inputStyle: React.CSSProperties = {
  padding: "8px 10px",
  borderRadius: "var(--radius-inputs)",
  border: "1px solid var(--color-stone-border)",
  background: "var(--color-pure-white)",
  fontSize: "var(--text-body)",
  minWidth: 140,
};

const paginationBtnStyle = (disabled: boolean): React.CSSProperties => ({
  padding: "6px 12px",
  borderRadius: "var(--radius-buttons)",
  border: "1px solid var(--color-stone-border)",
  background: "var(--color-pure-white)",
  cursor: disabled ? "not-allowed" : "pointer",
  opacity: disabled ? 0.5 : 1,
  fontSize: "12px",
});

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

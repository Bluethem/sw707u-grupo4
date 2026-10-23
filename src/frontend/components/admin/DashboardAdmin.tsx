"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Card } from "@/components/ui";
import { Badge, isBadgeStatus } from "@/components/ui/Badge";
import { ApiError, apiFetch } from "@/lib/api";
import type {
  ClienteItem,
  ComprobanteItem,
  Paginated,
  ProductoItem,
  ResumenDiarioItem,
  SerieItem,
} from "@/lib/caja-types";
import { formatMoneyDecimal } from "@/lib/money";
import { ReportCharts, type VentasTipoRow } from "./ReportCharts";
import { tdStyle, thStyle } from "./table-styles";

type Empresa = {
  razon_social: string;
  ruc: string;
};

type ReporteVentas = {
  por_tipo: VentasTipoRow[];
  anulados: { cantidad: number; total: string };
  totales: VentasTipoRow;
};

function todayISO(): string {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function monthStartISO(): string {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-01`;
}

export function DashboardAdmin() {
  const router = useRouter();
  const [empresa, setEmpresa] = useState<Empresa | null>(null);
  const [hoy, setHoy] = useState<ReporteVentas | null>(null);
  const [mes, setMes] = useState<ReporteVentas | null>(null);
  const [recientes, setRecientes] = useState<ComprobanteItem[]>([]);
  const [productos, setProductos] = useState(0);
  const [clientes, setClientes] = useState(0);
  const [resumenesPendientes, setResumenesPendientes] = useState(0);
  const [seriesActivas, setSeriesActivas] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();

  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    const dia = todayISO();
    const inicioMes = monthStartISO();
    try {
      const [emp, ventasHoy, ventasMes, comps, prods, clis, resums, series] = await Promise.all([
        apiFetch<Empresa>("/api/v1/empresas/me"),
        apiFetch<ReporteVentas>(`/api/v1/reportes/ventas?fecha_desde=${dia}&fecha_hasta=${dia}`),
        apiFetch<ReporteVentas>(`/api/v1/reportes/ventas?fecha_desde=${inicioMes}&fecha_hasta=${dia}`),
        apiFetch<Paginated<ComprobanteItem>>("/api/v1/comprobantes?limit=8"),
        apiFetch<Paginated<ProductoItem>>("/api/v1/productos?limit=1"),
        apiFetch<Paginated<ClienteItem>>("/api/v1/clientes?limit=1"),
        apiFetch<ResumenDiarioItem[]>("/api/v1/resumenes"),
        apiFetch<SerieItem[]>("/api/v1/series"),
      ]);
      setEmpresa(emp);
      setHoy(ventasHoy);
      setMes(ventasMes);
      setRecientes(comps.items ?? []);
      setProductos(prods.total ?? 0);
      setClientes(clis.total ?? 0);
      setResumenesPendientes((resums ?? []).filter((r) => r.estado === "PENDIENTE" || r.estado === "ENVIADO").length);
      setSeriesActivas((series ?? []).filter((s) => s.activa).length);
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        router.replace("/login");
        return;
      }
      setError(e instanceof Error ? e.message : "No se pudo cargar el resumen");
    } finally {
      setLoading(false);
    }
  }, [router]);

  useEffect(() => {
    void load();
  }, [load]);

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
        <div>
          <h1
            style={{
              margin: 0,
              fontSize: "var(--text-subheading)",
              fontFamily: "var(--font-family-display)",
              fontWeight: "var(--font-weight-regular)",
            }}
          >
            Resumen
          </h1>
          {empresa ? (
            <p style={{ margin: "4px 0 0", fontSize: "13px", color: "var(--color-warm-gray)" }}>
              {empresa.razon_social} · RUC {empresa.ruc}
            </p>
          ) : null}
        </div>
        <Button type="button" onClick={() => router.push("/caja")}>
          Ir a caja
        </Button>
      </div>

      {error ? (
        <p style={{ color: "var(--color-error-text)", margin: "0 0 var(--spacing-16)" }}>{error}</p>
      ) : null}

      <Card style={{ padding: "var(--spacing-16)", boxShadow: "none", marginBottom: "var(--spacing-16)" }}>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(160px, 1fr))",
            gap: "var(--spacing-16)",
            textAlign: "left",
          }}
        >
          <Kpi
            label="Ventas de hoy"
            value={loading ? "…" : formatMoneyDecimal(hoy?.totales.total ?? "0")}
            detail={loading ? "" : `${hoy?.totales.cantidad ?? 0} comprobantes`}
          />
          <Kpi
            label="Ventas del mes"
            value={loading ? "…" : formatMoneyDecimal(mes?.totales.total ?? "0")}
            detail={loading ? "" : `${mes?.totales.cantidad ?? 0} documentos`}
          />
          <Kpi
            label="Anulados (mes)"
            value={loading ? "…" : String(mes?.anulados.cantidad ?? 0)}
            detail={loading ? "" : formatMoneyDecimal(mes?.anulados.total ?? "0")}
          />
          <Kpi label="Productos" value={loading ? "…" : String(productos)} detail="En catálogo" href="/admin/productos" />
          <Kpi label="Clientes" value={loading ? "…" : String(clientes)} detail="Registrados" href="/admin/clientes" />
          <Kpi
            label="Resúmenes en curso"
            value={loading ? "…" : String(resumenesPendientes)}
            detail={`${seriesActivas} series activas`}
            href="/admin/resumenes"
          />
        </div>
      </Card>

      <Card style={{ padding: "var(--spacing-16)", boxShadow: "none", marginBottom: "var(--spacing-16)" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "var(--spacing-12)",
          }}
        >
          <h2
            style={{
              margin: 0,
              fontSize: "var(--text-body)",
              fontWeight: "var(--font-weight-medium)",
            }}
          >
            Ventas del mes
          </h2>
          <Link href="/admin/reportes" style={quietLink}>
            Ver reporte
          </Link>
        </div>
        {loading ? (
          <p style={{ color: "var(--color-warm-gray)" }}>Cargando…</p>
        ) : mes ? (
          <ReportCharts porTipo={mes.por_tipo} totales={mes.totales} />
        ) : (
          <p style={{ color: "var(--color-warm-gray)" }}>Sin datos.</p>
        )}
      </Card>

      <Card style={{ padding: "var(--spacing-16)", boxShadow: "none", marginBottom: "var(--spacing-16)" }}>
        <h2
          style={{
            margin: "0 0 var(--spacing-12)",
            fontSize: "var(--text-body)",
            fontWeight: "var(--font-weight-medium)",
          }}
        >
          Accesos
        </h2>
        <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--spacing-8)" }}>
          <Quick href="/admin/productos" label="Productos" />
          <Quick href="/admin/clientes" label="Clientes" />
          <Quick href="/admin/series" label="Series" />
          <Quick href="/admin/resumenes" label="Resúmenes" />
          <Quick href="/admin/reportes" label="Reportes" />
          <Quick href="/admin/configuracion" label="Configuración" />
          <Quick href="/comprobantes" label="Comprobantes" />
        </div>
      </Card>

      <Card style={{ padding: "var(--spacing-16)", boxShadow: "none" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "var(--spacing-12)",
          }}
        >
          <h2
            style={{
              margin: 0,
              fontSize: "var(--text-body)",
              fontWeight: "var(--font-weight-medium)",
            }}
          >
            Últimos comprobantes
          </h2>
          <Link href="/comprobantes" style={quietLink}>
            Ver todos
          </Link>
        </div>
        {loading ? (
          <p style={{ color: "var(--color-warm-gray)" }}>Cargando…</p>
        ) : recientes.length === 0 ? (
          <p style={{ color: "var(--color-warm-gray)", margin: 0 }}>Aún no hay comprobantes emitidos.</p>
        ) : (
          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <thead>
                <tr>
                  <th style={thStyle}>Documento</th>
                  <th style={thStyle}>Cliente</th>
                  <th style={thStyle}>Estado</th>
                  <th style={{ ...thStyle, textAlign: "right" }}>Total</th>
                </tr>
              </thead>
              <tbody>
                {recientes.map((c) => (
                  <tr key={c.id}>
                    <td style={tdStyle}>
                      <Link href={`/comprobantes/${c.id}`} style={{ color: "var(--color-ink-black)", textDecoration: "none" }}>
                        {c.tipo === "factura" ? "F" : "B"} {c.serie}-{c.numero}
                      </Link>
                    </td>
                    <td style={tdStyle}>{c.cliente?.razon_social || "—"}</td>
                    <td style={tdStyle}>
                      {isBadgeStatus(c.estado) ? <Badge status={c.estado} /> : c.estado}
                    </td>
                    <td style={{ ...tdStyle, textAlign: "right", fontVariantNumeric: "tabular-nums" }}>
                      {formatMoneyDecimal(c.total)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </>
  );
}

const quietLink: React.CSSProperties = {
  fontSize: "13px",
  color: "var(--color-cyan-edge)",
  textDecoration: "none",
};

function Kpi({
  label,
  value,
  detail,
  href,
}: {
  label: string;
  value: string;
  detail: string;
  href?: string;
}) {
  const body = (
    <>
      <p style={{ margin: 0, fontSize: "13px", color: "var(--color-warm-gray)" }}>{label}</p>
      <p
        style={{
          margin: "4px 0 0",
          fontSize: "var(--text-subheading)",
          fontFamily: "var(--font-family-display)",
          color: "var(--color-ink-black)",
        }}
      >
        {value}
      </p>
      {detail ? (
        <p style={{ margin: "2px 0 0", fontSize: "12px", color: "var(--color-ash-gray)" }}>{detail}</p>
      ) : null}
    </>
  );
  if (!href) return <div>{body}</div>;
  return (
    <Link href={href} style={{ textDecoration: "none", color: "inherit" }}>
      {body}
    </Link>
  );
}

function Quick({ href, label }: { href: string; label: string }) {
  return (
    <Link
      href={href}
      style={{
        display: "inline-flex",
        alignItems: "center",
        padding: "6px 12px",
        borderRadius: "var(--radius-buttons)",
        border: "1px solid var(--color-stone-border)",
        textDecoration: "none",
        color: "var(--color-ink-black)",
        fontSize: "var(--text-body)",
        background: "var(--color-pure-white)",
      }}
    >
      {label}
    </Link>
  );
}

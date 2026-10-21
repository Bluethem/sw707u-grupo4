"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Card, Input } from "@/components/ui";
import { ApiError, apiFetch, getApiBaseUrl } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { formatMoneyDecimal } from "@/lib/money";
import { ReportCharts, type VentasTipoRow } from "./ReportCharts";
import { tdStyle, thStyle } from "./table-styles";

function todayISO(): string {
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

type ReporteVentas = {
  fecha_desde: string;
  fecha_hasta: string;
  por_tipo: VentasTipoRow[];
  anulados: { cantidad: number; total: string };
  totales: VentasTipoRow;
};

export function ReportesAdmin() {
  const router = useRouter();
  const [desde, setDesde] = useState(todayISO);
  const [hasta, setHasta] = useState(todayISO);
  const [reporte, setReporte] = useState<ReporteVentas | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();
  const [exporting, setExporting] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    const qs = new URLSearchParams({ fecha_desde: desde.trim(), fecha_hasta: hasta.trim() });
    try {
      const data = await apiFetch<ReporteVentas>(`/api/v1/reportes/ventas?${qs}`);
      setReporte(data);
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        router.replace("/login");
        return;
      }
      setError(e instanceof ApiError ? e.message : "Error al cargar reporte");
      setReporte(null);
    } finally {
      setLoading(false);
    }
  }, [desde, hasta, router]);

  useEffect(() => {
    void load();
  }, [load]);

  async function exportCSV() {
    setExporting(true);
    setError(undefined);
    const qs = new URLSearchParams({ fecha_desde: desde.trim(), fecha_hasta: hasta.trim() });
    const token = getToken();
    try {
      const res = await fetch(`${getApiBaseUrl()}/api/v1/reportes/ventas.csv?${qs}`, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      });
      if (!res.ok) {
        throw new ApiError("No se pudo exportar CSV", res.status);
      }
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `ventas-${hasta.trim()}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Error al exportar");
    } finally {
      setExporting(false);
    }
  }

  const rows = reporte?.por_tipo ?? [];

  return (
    <>
      <div style={{ marginBottom: "var(--spacing-16)" }}>
        <h1
          style={{
            margin: 0,
            fontSize: "var(--text-subheading)",
            fontFamily: "var(--font-family-display)",
            fontWeight: "var(--font-weight-regular)",
          }}
        >
          Reporte de ventas
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
          <div style={{ flex: "1 1 160px", maxWidth: "200px" }}>
            <Input label="Desde" type="date" value={desde} onChange={(ev) => setDesde(ev.target.value)} />
          </div>
          <div style={{ flex: "1 1 160px", maxWidth: "200px" }}>
            <Input label="Hasta" type="date" value={hasta} onChange={(ev) => setHasta(ev.target.value)} />
          </div>
          <Button type="button" disabled={loading} onClick={() => void load()}>
            {loading ? "Cargando…" : "Actualizar"}
          </Button>
          <Button type="button" variant="ghost" disabled={exporting} onClick={() => void exportCSV()}>
            {exporting ? "Exportando…" : "Exportar CSV"}
          </Button>
        </div>
      </Card>

      <Card style={{ padding: "var(--spacing-16)", boxShadow: "none" }}>
        {error ? (
          <p style={{ margin: "0 0 var(--spacing-12)", color: "var(--color-error-text)" }}>{error}</p>
        ) : null}
        {loading ? (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Cargando…</p>
        ) : reporte ? (
          <>
            <p style={{ margin: "0 0 var(--spacing-12)", color: "var(--color-warm-gray)", fontSize: "var(--text-body)" }}>
              Período: {reporte.fecha_desde} — {reporte.fecha_hasta}
              {" · "}
              Anulados: {reporte.anulados.cantidad} ({formatMoneyDecimal(reporte.anulados.total)})
            </p>
            <div style={{ overflowX: "auto" }}>
              <table style={{ width: "100%", borderCollapse: "collapse" }}>
                <thead>
                  <tr>
                    <th style={thStyle}>Tipo</th>
                    <th style={thStyle}>Cantidad</th>
                    <th style={thStyle}>Valor venta</th>
                    <th style={thStyle}>IGV</th>
                    <th style={thStyle}>Total</th>
                  </tr>
                </thead>
                <tbody>
                  {rows.length === 0 ? (
                    <tr>
                      <td colSpan={5} style={tdStyle}>
                        Sin ventas vigentes en el período.
                      </td>
                    </tr>
                  ) : (
                    rows.map((r) => (
                      <tr key={r.tipo}>
                        <td style={tdStyle}>{r.tipo}</td>
                        <td style={tdStyle}>{r.cantidad}</td>
                        <td style={tdStyle}>{formatMoneyDecimal(r.valor_venta)}</td>
                        <td style={tdStyle}>{formatMoneyDecimal(r.igv)}</td>
                        <td style={tdStyle}>{formatMoneyDecimal(r.total)}</td>
                      </tr>
                    ))
                  )}
                  <tr>
                    <td style={{ ...tdStyle, fontWeight: "var(--font-weight-medium)" }}>Totales</td>
                    <td style={tdStyle}>{reporte.totales.cantidad}</td>
                    <td style={tdStyle}>{formatMoneyDecimal(reporte.totales.valor_venta)}</td>
                    <td style={tdStyle}>{formatMoneyDecimal(reporte.totales.igv)}</td>
                    <td style={tdStyle}>{formatMoneyDecimal(reporte.totales.total)}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <ReportCharts porTipo={rows} totales={reporte.totales} />
          </>
        ) : null}
      </Card>
    </>
  );
}

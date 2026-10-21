"use client";

import { formatMoneyDecimal } from "@/lib/money";

export type VentasTipoRow = {
  tipo: string;
  cantidad: number;
  valor_venta: string;
  igv: string;
  total: string;
};

type Props = {
  porTipo: VentasTipoRow[];
  totales: VentasTipoRow;
};

function parseAmount(value: string): number {
  const n = parseFloat(value);
  return Number.isFinite(n) ? n : 0;
}

export function ReportCharts({ porTipo, totales }: Props) {
  const maxTotal = Math.max(...porTipo.map((r) => parseAmount(r.total)), 1);
  const sumCantidad = porTipo.reduce((acc, r) => acc + r.cantidad, 0);

  if (porTipo.length === 0) {
    return (
      <p style={{ margin: 0, color: "var(--color-warm-gray)", fontSize: "var(--text-body)" }}>
        Sin datos para gráficos en el período.
      </p>
    );
  }

  let pieOffset = 0;
  const colors = ["var(--color-cyan-edge)", "var(--color-ink-black)", "var(--color-warm-gray)"];

  return (
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "repeat(auto-fit, minmax(240px, 1fr))",
        gap: "var(--spacing-24)",
        marginTop: "var(--spacing-16)",
      }}
    >
      <div>
        <h2
          style={{
            margin: "0 0 var(--spacing-12)",
            fontSize: "var(--text-body)",
            fontWeight: "var(--font-weight-medium)",
          }}
        >
          Ventas por tipo (total)
        </h2>
        <svg width="100%" height={40 + porTipo.length * 36} role="img" aria-label="Barras por tipo">
          {porTipo.map((row, i) => {
            const w = (parseAmount(row.total) / maxTotal) * 100;
            const y = 8 + i * 36;
            return (
              <g key={row.tipo}>
                <text x={0} y={y + 14} fontSize={12} fill="var(--color-warm-gray)">
                  {row.tipo}
                </text>
                <rect
                  x={72}
                  y={y}
                  width={`${Math.max(w, 2)}%`}
                  height={20}
                  rx={4}
                  fill="var(--color-cyan-edge)"
                />
                <text x="98%" y={y + 14} fontSize={11} textAnchor="end" fill="var(--color-ink-black)">
                  {formatMoneyDecimal(row.total)}
                </text>
              </g>
            );
          })}
        </svg>
      </div>

      <div>
        <h2
          style={{
            margin: "0 0 var(--spacing-12)",
            fontSize: "var(--text-body)",
            fontWeight: "var(--font-weight-medium)",
          }}
        >
          Proporción por tipo
        </h2>
        <svg width={160} height={160} viewBox="0 0 32 32" role="img" aria-label="Proporción por cantidad de comprobantes">
          <circle cx={16} cy={16} r={16} fill="var(--color-stone-canvas)" />
          {porTipo.map((row, i) => {
            const frac = sumCantidad > 0 ? row.cantidad / sumCantidad : 0;
            const dash = frac * 100;
            const circle = (
              <circle
                key={row.tipo}
                cx={16}
                cy={16}
                r={8}
                fill="transparent"
                stroke={colors[i % colors.length]}
                strokeWidth={16}
                pathLength={100}
                strokeDasharray={`${dash} ${100 - dash}`}
                transform={`rotate(${pieOffset * 3.6 - 90} 16 16)`}
              />
            );
            pieOffset += dash;
            return circle;
          })}
        </svg>
        <ul
          style={{
            margin: "var(--spacing-8) 0 0",
            padding: 0,
            listStyle: "none",
            fontSize: "var(--text-body)",
          }}
        >
          {porTipo.map((row, i) => (
            <li key={row.tipo} style={{ display: "flex", gap: 8, alignItems: "center" }}>
              <span
                style={{
                  width: 10,
                  height: 10,
                  borderRadius: 2,
                  background: colors[i % colors.length],
                }}
              />
              {row.tipo} — {row.cantidad}
            </li>
          ))}
        </ul>
        <p style={{ margin: "var(--spacing-8) 0 0", color: "var(--color-warm-gray)", fontSize: 12 }}>
          Total período: {formatMoneyDecimal(totales.total)}
        </p>
      </div>
    </div>
  );
}

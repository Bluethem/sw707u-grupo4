/** IGV rate for gravado lines (preview; API is source of truth on emit). */
export const IGV_RATE = 0.18;

export type SaleLinePreview = {
  cantidad: number;
  precioUnitario: number;
  /** true = gravado (IGV applies) */
  gravado: boolean;
};

export type SaleTotalsPreview = {
  valorVenta: number;
  igv: number;
  total: number;
};

function roundMoney(n: number): number {
  return Math.round(n * 100) / 100;
}

export function previewLineSubtotal(line: SaleLinePreview): number {
  return roundMoney(line.cantidad * line.precioUnitario);
}

/** Client-side totals aligned to 18% IGV on gravado base (valor venta = base gravada). */
export function previewSaleTotals(lines: SaleLinePreview[]): SaleTotalsPreview {
  let valorVenta = 0;
  for (const line of lines) {
    if (line.gravado) {
      valorVenta += previewLineSubtotal(line);
    }
  }
  valorVenta = roundMoney(valorVenta);
  const igv = roundMoney(valorVenta * IGV_RATE);
  const total = roundMoney(valorVenta + igv);
  return { valorVenta, igv, total };
}

export function formatMoney(amount: number, currency = "PEN"): string {
  return new Intl.NumberFormat("es-PE", {
    style: "currency",
    currency,
    minimumFractionDigits: 2,
  }).format(amount);
}

/** API amounts are decimal strings (shopspring/decimal JSON). */
export function formatMoneyDecimal(value: string, currency = "PEN"): string {
  const n = parseFloat(value);
  return formatMoney(Number.isFinite(n) ? n : 0, currency);
}

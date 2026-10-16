import type { CSSProperties } from "react";

export const thStyle: CSSProperties = {
  textAlign: "left",
  padding: "8px 12px",
  fontSize: "12px",
  fontWeight: "var(--font-weight-medium)",
  color: "var(--color-warm-gray)",
  borderBottom: "1px solid var(--color-stone-border)",
};

export const tdStyle: CSSProperties = {
  padding: "10px 12px",
  fontSize: "var(--text-body)",
  borderBottom: "1px solid var(--color-stone-border)",
  verticalAlign: "middle",
};

export const selectFieldStyle: CSSProperties = {
  padding: "4px 12px",
  borderRadius: "var(--radius-inputs)",
  border: "1px solid var(--color-stone-muted)",
  fontFamily: "var(--font-family-body)",
  fontSize: "var(--text-body)",
  background: "var(--color-pure-white)",
  color: "var(--color-ink-black)",
};

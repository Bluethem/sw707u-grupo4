import type { InputHTMLAttributes } from "react";

export type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
  error?: string;
};

const fieldStyle: React.CSSProperties = {
  width: "100%",
  padding: "4px 12px",
  borderRadius: "var(--radius-inputs)",
  border: "1px solid var(--color-stone-muted)",
  background: "var(--color-pure-white)",
  color: "var(--color-ink-black)",
  fontFamily: "var(--font-family-body)",
  fontSize: "var(--text-body)",
  lineHeight: "var(--leading-body)",
  outline: "none",
};

export function Input({ label, error, id, style, ...rest }: InputProps) {
  const inputId = id ?? (label ? `input-${label.replace(/\s+/g, "-")}` : undefined);

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-4)" }}>
      {label ? (
        <label
          htmlFor={inputId}
          style={{
            fontSize: "var(--text-body)",
            color: "var(--color-warm-gray)",
            fontWeight: "var(--font-weight-medium)",
          }}
        >
          {label}
        </label>
      ) : null}
      <input
        id={inputId}
        style={{
          ...fieldStyle,
          borderColor: error ? "var(--color-error-text)" : undefined,
          ...style,
        }}
        {...rest}
      />
      {error ? (
        <span style={{ fontSize: "13px", color: "var(--color-error-text)" }}>{error}</span>
      ) : null}
    </div>
  );
}

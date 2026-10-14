import type { HTMLAttributes, ReactNode } from "react";

/** Comprobante / queue states per M4 spec §2.1 */
export type BadgeStatus =
  | "PENDIENTE"
  | "ENVIADO"
  | "ACEPTADO"
  | "RECHAZADO"
  | "OBSERVADO";

export function isBadgeStatus(s: string): s is BadgeStatus {
  return ["PENDIENTE", "ENVIADO", "ACEPTADO", "RECHAZADO", "OBSERVADO"].includes(s);
}

export type BadgeProps = HTMLAttributes<HTMLSpanElement> & {
  status: BadgeStatus;
  children?: ReactNode;
};

const statusStyles: Record<BadgeStatus, React.CSSProperties> = {
  PENDIENTE: {
    background: "transparent",
    border: "1px solid var(--color-stone-border)",
    color: "var(--color-warm-gray)",
  },
  ENVIADO: {
    background: "var(--color-sky-wash)",
    border: "1px solid var(--color-sky-wash)",
    color: "var(--color-cyan-edge)",
  },
  ACEPTADO: {
    background: "var(--color-pure-white)",
    border: "1px solid var(--color-stone-border)",
    color: "var(--color-ink-black)",
  },
  RECHAZADO: {
    background: "var(--color-error-bg)",
    border: "1px solid var(--color-error-bg)",
    color: "var(--color-error-text)",
  },
  OBSERVADO: {
    background: "var(--color-warning-bg)",
    border: "1px solid var(--color-warning-bg)",
    color: "var(--color-warning-text)",
  },
};

export function Badge({ status, children, style, ...rest }: BadgeProps) {
  const label = children ?? status;

  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: "4px",
        padding: "4px 10px",
        borderRadius: "var(--radius-buttons)",
        fontSize: "12px",
        fontWeight: "var(--font-weight-medium)",
        lineHeight: 1.33,
        ...statusStyles[status],
        ...style,
      }}
      {...rest}
    >
      {status === "ACEPTADO" ? (
        <svg
          width="12"
          height="12"
          viewBox="0 0 12 12"
          fill="none"
          aria-hidden
          style={{ flexShrink: 0 }}
        >
          <path
            d="M2 6l3 3 5-6"
            stroke="var(--color-cyan-signal)"
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      ) : null}
      {label}
    </span>
  );
}

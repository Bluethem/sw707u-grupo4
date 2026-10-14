import type { ButtonHTMLAttributes, ReactNode } from "react";

export type ButtonVariant = "primary" | "ghost";

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  children: ReactNode;
};

const base: React.CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "var(--spacing-8)",
  padding: "8px 16px",
  borderRadius: "var(--radius-buttons)",
  fontFamily: "var(--font-family-body)",
  fontSize: "var(--text-body)",
  lineHeight: "var(--leading-body)",
  cursor: "pointer",
  transition: "opacity 0.15s ease, background 0.15s ease",
};

const variants: Record<ButtonVariant, React.CSSProperties> = {
  primary: {
    background: "var(--color-cyan-signal)",
    color: "#ffffff",
    border: "1px solid var(--color-cyan-edge)",
    fontWeight: "var(--font-weight-medium)",
  },
  ghost: {
    background: "transparent",
    color: "var(--color-ink-black)",
    border: "1px solid var(--color-stone-border)",
    fontWeight: "var(--font-weight-regular)",
  },
};

export function Button({
  variant = "primary",
  style,
  disabled,
  children,
  ...rest
}: ButtonProps) {
  return (
    <button
      type="button"
      disabled={disabled}
      style={{
        ...base,
        ...variants[variant],
        opacity: disabled ? 0.5 : 1,
        ...style,
      }}
      {...rest}
    >
      {children}
    </button>
  );
}

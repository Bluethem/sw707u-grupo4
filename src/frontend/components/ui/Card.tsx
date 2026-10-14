import type { HTMLAttributes, ReactNode } from "react";

export type CardProps = HTMLAttributes<HTMLDivElement> & {
  children: ReactNode;
  /** Elevated product-preview style (spec: use sparingly) */
  floating?: boolean;
};

export function Card({ children, floating, style, ...rest }: CardProps) {
  return (
    <div
      style={{
        background: "var(--color-pure-white)",
        borderRadius: floating ? "16px" : "var(--radius-cards)",
        border: "1px solid var(--color-stone-border)",
        padding: "var(--card-padding)",
        boxShadow: floating
          ? "rgba(17, 12, 46, 0.12) 0px 12px 45px 0px"
          : "var(--shadow-md)",
        ...style,
      }}
      {...rest}
    >
      {children}
    </div>
  );
}

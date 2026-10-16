"use client";

import type { ReactNode } from "react";

type Props = {
  open: boolean;
  title: string;
  onClose: () => void;
  children: ReactNode;
  maxWidth?: number;
};

export function ModalShell({ open, title, onClose, children, maxWidth = 420 }: Props) {
  if (!open) return null;

  return (
    <div
      role="dialog"
      aria-modal
      style={{
        position: "fixed",
        inset: 0,
        zIndex: 50,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "rgba(12, 10, 9, 0.35)",
        padding: "var(--spacing-16)",
      }}
      onClick={onClose}
    >
      <div
        onClick={(ev) => ev.stopPropagation()}
        style={{
          width: "100%",
          maxWidth,
          maxHeight: "min(90vh, 720px)",
          overflowY: "auto",
          background: "var(--color-pure-white)",
          borderRadius: "var(--radius-cards)",
          border: "1px solid var(--color-stone-border)",
          padding: "var(--spacing-16)",
          boxShadow: "var(--shadow-subtle)",
        }}
      >
        <h2
          style={{
            margin: "0 0 var(--spacing-12)",
            fontSize: "var(--text-subheading)",
            fontFamily: "var(--font-family-display)",
            fontWeight: "var(--font-weight-regular)",
          }}
        >
          {title}
        </h2>
        {children}
      </div>
    </div>
  );
}

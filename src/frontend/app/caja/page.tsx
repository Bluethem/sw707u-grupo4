"use client";

import { AppNav } from "@/components/shell/AppNav";
import { RequireAuth } from "@/components/shell/RequireAuth";
import { CajaPos } from "@/components/caja/CajaPos";

export default function CajaPage() {
  return (
    <RequireAuth>
      <AppNav />
      <main
        style={{
          display: "block",
          minHeight: "calc(100vh - 57px)",
          padding: "var(--spacing-16)",
          maxWidth: "var(--page-max-width)",
          margin: "0 auto",
        }}
      >
        <h1
          style={{
            margin: "0 0 var(--spacing-16)",
            fontSize: "var(--text-subheading)",
            fontFamily: "var(--font-family-display)",
            fontWeight: "var(--font-weight-regular)",
          }}
        >
          Caja
        </h1>
        <CajaPos />
      </main>
    </RequireAuth>
  );
}

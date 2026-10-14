"use client";

import type { KeyboardEvent } from "react";
import { Input } from "@/components/ui";
import { formatMoney } from "@/lib/money";
import type { ProductoItem } from "@/lib/caja-types";

const cardSurface: React.CSSProperties = {
  background: "var(--color-pure-white)",
  borderRadius: "var(--radius-cards)",
  border: "1px solid var(--color-stone-border)",
  padding: "var(--spacing-16)",
  boxShadow: "none",
};

type Props = {
  query: string;
  onQueryChange: (q: string) => void;
  loading: boolean;
  products: ProductoItem[];
  onAdd: (p: ProductoItem) => void;
};

export function ProductCatalog({ query, onQueryChange, loading, products, onAdd }: Props) {
  function onSearchKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Enter" && products[0]) {
      e.preventDefault();
      onAdd(products[0]);
    }
  }

  return (
    <section style={{ ...cardSurface, display: "flex", flexDirection: "column", gap: "var(--spacing-12)" }}>
      <h2
        style={{
          margin: 0,
          fontSize: "var(--text-body-lg)",
          fontFamily: "var(--font-family-display)",
          fontWeight: "var(--font-weight-medium)",
        }}
      >
        Productos
      </h2>
      <Input
        placeholder="Buscar código o descripción…"
        value={query}
        onChange={(ev) => onQueryChange(ev.target.value)}
        onKeyDown={onSearchKeyDown}
        autoComplete="off"
      />
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          gap: "var(--spacing-4)",
          maxHeight: "min(60vh, 520px)",
          overflowY: "auto",
        }}
      >
        {loading ? (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Buscando…</p>
        ) : products.length === 0 ? (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Sin resultados</p>
        ) : (
          products.map((p) => (
            <button
              key={p.id}
              type="button"
              onClick={() => onAdd(p)}
              style={{
                display: "grid",
                gridTemplateColumns: "72px 1fr auto",
                gap: "var(--spacing-8)",
                alignItems: "center",
                textAlign: "left",
                padding: "var(--spacing-8) var(--spacing-12)",
                borderRadius: "var(--radius-inputs)",
                border: "1px solid var(--color-stone-border)",
                background: "var(--color-pure-white)",
                cursor: "pointer",
                fontFamily: "var(--font-family-body)",
                fontSize: "var(--text-body)",
              }}
            >
              <span style={{ color: "var(--color-warm-gray)", fontSize: "12px" }}>{p.codigo}</span>
              <span style={{ color: "var(--color-ink-black)" }}>{p.descripcion}</span>
              <span style={{ fontWeight: "var(--font-weight-medium)", whiteSpace: "nowrap" }}>
                {formatMoney(parseFloat(p.precio))}
                {p.controla_stock && p.stock !== undefined ? (
                  <span style={{ color: "var(--color-warm-gray)", fontWeight: 400, marginLeft: 6 }}>
                    · {p.stock}
                  </span>
                ) : null}
              </span>
            </button>
          ))
        )}
      </div>
    </section>
  );
}

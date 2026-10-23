"use client";

import type { KeyboardEvent } from "react";
import { useState } from "react";
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

type SortKey = "descripcion_asc" | "descripcion_desc" | "precio_asc" | "precio_desc" | "codigo_asc";

function sortProducts(items: ProductoItem[], key: SortKey): ProductoItem[] {
  return [...items].sort((a, b) => {
    switch (key) {
      case "descripcion_asc":
        return a.descripcion.localeCompare(b.descripcion, "es", { sensitivity: "base" });
      case "descripcion_desc":
        return b.descripcion.localeCompare(a.descripcion, "es", { sensitivity: "base" });
      case "precio_asc":
        return parseFloat(a.precio) - parseFloat(b.precio);
      case "precio_desc":
        return parseFloat(b.precio) - parseFloat(a.precio);
      case "codigo_asc":
        return a.codigo.localeCompare(b.codigo, "es", { sensitivity: "base" });
    }
  });
}

type Props = {
  query: string;
  onQueryChange: (q: string) => void;
  loading: boolean;
  products: ProductoItem[];
  onAdd: (p: ProductoItem) => void;
};

export function ProductCatalog({ query, onQueryChange, loading, products, onAdd }: Props) {
  const [sortKey, setSortKey] = useState<SortKey>("descripcion_asc");

  function onSearchKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Enter" && products[0]) {
      e.preventDefault();
      onAdd(sortProducts(products, sortKey)[0]);
    }
  }

  const sorted = sortProducts(products, sortKey);

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
      <div style={{ display: "flex", gap: "var(--spacing-8)", alignItems: "flex-end" }}>
        <div style={{ flex: "1 1 0" }}>
          <Input
            placeholder="Buscar código o descripción…"
            value={query}
            onChange={(ev) => onQueryChange(ev.target.value)}
            onKeyDown={onSearchKeyDown}
            autoComplete="off"
          />
        </div>
        <label style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-4)" }}>
          <span style={{ fontSize: "12px", color: "var(--color-warm-gray)" }}>Ordenar</span>
          <select
            value={sortKey}
            onChange={(ev) => setSortKey(ev.target.value as SortKey)}
            style={{
              padding: "5px 10px",
              borderRadius: "var(--radius-inputs)",
              border: "1px solid var(--color-stone-muted)",
              fontFamily: "var(--font-family-body)",
              fontSize: "var(--text-body)",
              background: "var(--color-pure-white)",
              cursor: "pointer",
            }}
          >
            <option value="descripcion_asc">Nombre A→Z</option>
            <option value="descripcion_desc">Nombre Z→A</option>
            <option value="precio_asc">Precio ↑</option>
            <option value="precio_desc">Precio ↓</option>
            <option value="codigo_asc">Código A→Z</option>
          </select>
        </label>
      </div>
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
        ) : sorted.length === 0 ? (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Sin resultados</p>
        ) : (
          sorted.map((p) => (
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

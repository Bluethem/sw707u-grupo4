"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui";
import { Badge, isBadgeStatus } from "@/components/ui/Badge";
import { ApiError, apiFetch } from "@/lib/api";
import { validarClienteParaEmitir } from "@/lib/caja-validation";
import type {
  CartLine,
  ClienteItem,
  ComprobanteEmitido,
  Paginated,
  ProductoItem,
  SerieItem,
  TipoComprobante,
} from "@/lib/caja-types";
import { formatMoney, previewLineSubtotal, previewSaleTotals } from "@/lib/money";
import { ProductCatalog } from "./ProductCatalog";
import { QuickClienteModal } from "./QuickClienteModal";

const layoutGrid: React.CSSProperties = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr) minmax(320px, 420px)",
  gap: "var(--spacing-16)",
  alignItems: "start",
};

const cardSurface: React.CSSProperties = {
  background: "var(--color-pure-white)",
  borderRadius: "var(--radius-cards)",
  border: "1px solid var(--color-stone-border)",
  padding: "var(--spacing-16)",
  boxShadow: "none",
};

function parsePrecio(precio: string): number {
  const n = parseFloat(precio);
  return Number.isFinite(n) ? n : 0;
}

export function CajaPos() {
  const router = useRouter();
  const [productQuery, setProductQuery] = useState("");
  const [products, setProducts] = useState<ProductoItem[]>([]);
  const [productsLoading, setProductsLoading] = useState(false);

  const [tipo, setTipo] = useState<TipoComprobante>("boleta");
  const [series, setSeries] = useState<SerieItem[]>([]);
  const [serie, setSerie] = useState("");

  const [clienteQuery, setClienteQuery] = useState("");
  const [clientes, setClientes] = useState<ClienteItem[]>([]);
  const [cliente, setCliente] = useState<ClienteItem | null>(null);
  const [clientesLoading, setClientesLoading] = useState(false);
  const [modalCliente, setModalCliente] = useState(false);

  const [lines, setLines] = useState<CartLine[]>([]);
  const [emitError, setEmitError] = useState<string | undefined>();
  const [emitting, setEmitting] = useState(false);
  const [success, setSuccess] = useState<ComprobanteEmitido | null>(null);
  const [idempotencyKey, setIdempotencyKey] = useState<string | null>(null);

  const previewLines = useMemo(
    () =>
      lines.map((l) => ({
        cantidad: l.cantidad,
        precioUnitario: parsePrecio(l.precioUnitario),
        gravado: l.afectacion === "gravado",
      })),
    [lines],
  );
  const totals = useMemo(() => previewSaleTotals(previewLines), [previewLines]);

  const seriesForTipo = useMemo(
    () => series.filter((s) => s.tipo === tipo && s.activa),
    [series, tipo],
  );

  useEffect(() => {
    void apiFetch<SerieItem[]>("/api/v1/series")
      .then(setSeries)
      .catch(() => setSeries([]));
  }, []);

  useEffect(() => {
    if (seriesForTipo.some((s) => s.serie === serie)) return;
    setSerie(seriesForTipo[0]?.serie ?? "");
  }, [seriesForTipo, serie]);

  useEffect(() => {
    const q = productQuery.trim();
    const t = window.setTimeout(() => {
      setProductsLoading(true);
      const params = new URLSearchParams({ activo: "true", limit: "30" });
      if (q) params.set("q", q);
      void apiFetch<Paginated<ProductoItem>>(`/api/v1/productos?${params}`)
        .then((res) => setProducts(res.items))
        .catch(() => setProducts([]))
        .finally(() => setProductsLoading(false));
    }, 280);
    return () => window.clearTimeout(t);
  }, [productQuery]);

  useEffect(() => {
    const q = clienteQuery.trim();
    if (q.length < 2) {
      setClientes([]);
      return;
    }
    const t = window.setTimeout(() => {
      setClientesLoading(true);
      const params = new URLSearchParams({ q, limit: "15", activo: "true" });
      void apiFetch<Paginated<ClienteItem>>(`/api/v1/clientes?${params}`)
        .then((res) => setClientes(res.items))
        .catch(() => setClientes([]))
        .finally(() => setClientesLoading(false));
    }, 280);
    return () => window.clearTimeout(t);
  }, [clienteQuery]);

  const addProduct = useCallback((p: ProductoItem) => {
    setSuccess(null);
    setLines((prev) => {
      const idx = prev.findIndex((l) => l.productoId === p.id);
      if (idx >= 0) {
        const next = [...prev];
        next[idx] = { ...next[idx], cantidad: next[idx].cantidad + 1 };
        return next;
      }
      return [
        ...prev,
        {
          productoId: p.id,
          codigo: p.codigo,
          descripcion: p.descripcion,
          precioUnitario: p.precio,
          afectacion: p.afectacion,
          cantidad: 1,
        },
      ];
    });
  }, []);

  function changeQty(productoId: string, delta: number) {
    setSuccess(null);
    setLines((prev) =>
      prev
        .map((l) =>
          l.productoId === productoId ? { ...l, cantidad: l.cantidad + delta } : l,
        )
        .filter((l) => l.cantidad > 0),
    );
  }

  function removeLine(productoId: string) {
    setSuccess(null);
    setLines((prev) => prev.filter((l) => l.productoId !== productoId));
  }

  async function emitir() {
    setEmitError(undefined);
    const clienteErr = validarClienteParaEmitir(tipo, totals.total, cliente);
    if (clienteErr) {
      setEmitError(clienteErr);
      return;
    }
    if (!serie) {
      setEmitError("Seleccione una serie activa");
      return;
    }
    if (lines.length === 0) {
      setEmitError("Agregue al menos un producto");
      return;
    }

    const key = idempotencyKey ?? crypto.randomUUID();
    if (!idempotencyKey) setIdempotencyKey(key);

    setEmitting(true);
    try {
      const body: Record<string, unknown> = {
        tipo,
        serie,
        lineas: lines.map((l) => ({
          producto_id: l.productoId,
          cantidad: l.cantidad,
          precio_unitario: l.precioUnitario,
          afectacion: l.afectacion,
        })),
      };
      if (cliente) {
        body.cliente = {
          tipo_doc: cliente.tipo_doc,
          numero: cliente.numero,
          razon_social: cliente.razon_social,
        };
      }

      const res = await apiFetch<ComprobanteEmitido>("/api/v1/comprobantes", {
        method: "POST",
        body,
        idempotencyKey: key,
      });
      setSuccess(res);
      setLines([]);
      setCliente(null);
      setClienteQuery("");
      setIdempotencyKey(null);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        router.replace("/login");
        return;
      }
      setEmitError(err instanceof ApiError ? err.message : "Error al emitir");
    } finally {
      setEmitting(false);
    }
  }

  const tabStyle = (active: boolean): React.CSSProperties => ({
    padding: "6px 14px",
    borderRadius: "var(--radius-buttons)",
    border: active ? "1px solid var(--color-cyan-edge)" : "1px solid var(--color-stone-border)",
    background: active ? "var(--color-sky-wash)" : "transparent",
    color: active ? "var(--color-cyan-edge)" : "var(--color-ink-black)",
    fontFamily: "var(--font-family-body)",
    fontSize: "var(--text-body)",
    cursor: "pointer",
  });

  return (
    <>
      <div className="caja-grid" style={layoutGrid}>
        <ProductCatalog
          query={productQuery}
          onQueryChange={setProductQuery}
          loading={productsLoading}
          products={products}
          onAdd={addProduct}
        />

        <section style={{ ...cardSurface, display: "flex", flexDirection: "column", gap: "var(--spacing-12)" }}>
          <div style={{ display: "flex", gap: "var(--spacing-8)" }}>
            <button type="button" style={tabStyle(tipo === "boleta")} onClick={() => setTipo("boleta")}>
              Boleta
            </button>
            <button type="button" style={tabStyle(tipo === "factura")} onClick={() => setTipo("factura")}>
              Factura
            </button>
          </div>

          <label style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-4)" }}>
            <span style={{ fontSize: "var(--text-body)", color: "var(--color-warm-gray)" }}>Serie</span>
            <select
              value={serie}
              onChange={(ev) => setSerie(ev.target.value)}
              style={{
                padding: "4px 12px",
                borderRadius: "var(--radius-inputs)",
                border: "1px solid var(--color-stone-muted)",
                fontFamily: "var(--font-family-body)",
              }}
            >
              {seriesForTipo.length === 0 ? (
                <option value="">Sin series</option>
              ) : (
                seriesForTipo.map((s) => (
                  <option key={s.serie} value={s.serie}>
                    {s.serie}
                  </option>
                ))
              )}
            </select>
          </label>

          <div>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                marginBottom: "var(--spacing-8)",
              }}
            >
              <span style={{ fontSize: "var(--text-body)", color: "var(--color-warm-gray)" }}>Cliente</span>
              <Button type="button" variant="ghost" onClick={() => setModalCliente(true)} style={{ padding: "4px 12px" }}>
                + Cliente rápido
              </Button>
            </div>
            {cliente ? (
              <div
                style={{
                  padding: "var(--spacing-8) var(--spacing-12)",
                  border: "1px solid var(--color-stone-border)",
                  borderRadius: "var(--radius-inputs)",
                  display: "flex",
                  justifyContent: "space-between",
                  gap: "var(--spacing-8)",
                }}
              >
                <div>
                  <div style={{ fontWeight: "var(--font-weight-medium)" }}>{cliente.razon_social}</div>
                  <div style={{ fontSize: "12px", color: "var(--color-warm-gray)" }}>
                    {cliente.tipo_doc} {cliente.numero}
                  </div>
                </div>
                <button
                  type="button"
                  onClick={() => setCliente(null)}
                  style={{
                    border: "none",
                    background: "transparent",
                    color: "var(--color-warm-gray)",
                    cursor: "pointer",
                    fontSize: "12px",
                  }}
                >
                  Quitar
                </button>
              </div>
            ) : (
              <>
                <input
                  placeholder="Buscar cliente…"
                  value={clienteQuery}
                  onChange={(ev) => setClienteQuery(ev.target.value)}
                  style={{
                    width: "100%",
                    padding: "4px 12px",
                    borderRadius: "var(--radius-inputs)",
                    border: "1px solid var(--color-stone-muted)",
                    fontFamily: "var(--font-family-body)",
                    fontSize: "var(--text-body)",
                  }}
                />
                {clientesLoading ? (
                  <p style={{ margin: "8px 0 0", color: "var(--color-warm-gray)", fontSize: "12px" }}>
                    Buscando…
                  </p>
                ) : null}
                {clientes.length > 0 ? (
                  <ul
                    style={{
                      listStyle: "none",
                      margin: "var(--spacing-8) 0 0",
                      padding: 0,
                      border: "1px solid var(--color-stone-border)",
                      borderRadius: "var(--radius-inputs)",
                      maxHeight: "140px",
                      overflowY: "auto",
                    }}
                  >
                    {clientes.map((c) => (
                      <li key={c.id}>
                        <button
                          type="button"
                          onClick={() => {
                            setCliente(c);
                            setClientes([]);
                            setClienteQuery("");
                          }}
                          style={{
                            width: "100%",
                            textAlign: "left",
                            padding: "8px 12px",
                            border: "none",
                            borderBottom: "1px solid var(--color-stone-border)",
                            background: "var(--color-pure-white)",
                            cursor: "pointer",
                            fontFamily: "var(--font-family-body)",
                            fontSize: "var(--text-body)",
                          }}
                        >
                          {c.razon_social}{" "}
                          <span style={{ color: "var(--color-warm-gray)" }}>
                            ({c.tipo_doc} {c.numero})
                          </span>
                        </button>
                      </li>
                    ))}
                  </ul>
                ) : null}
              </>
            )}
          </div>

          <div style={{ borderTop: "1px solid var(--color-stone-border)", paddingTop: "var(--spacing-12)" }}>
            <h3
              style={{
                margin: "0 0 var(--spacing-8)",
                fontSize: "var(--text-body)",
                fontWeight: "var(--font-weight-medium)",
              }}
            >
              Líneas
            </h3>
            {lines.length === 0 ? (
              <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Carrito vacío</p>
            ) : (
              <ul style={{ listStyle: "none", margin: 0, padding: 0, display: "flex", flexDirection: "column", gap: 8 }}>
                {lines.map((l) => (
                  <li
                    key={l.productoId}
                    style={{
                      display: "grid",
                      gridTemplateColumns: "1fr auto auto",
                      gap: "var(--spacing-8)",
                      alignItems: "center",
                      fontSize: "var(--text-body)",
                    }}
                  >
                    <div>
                      <div>{l.descripcion}</div>
                      <div style={{ fontSize: "12px", color: "var(--color-warm-gray)" }}>
                        {formatMoney(parsePrecio(l.precioUnitario))} c/u
                      </div>
                    </div>
                    <div style={{ display: "flex", alignItems: "center", gap: 4 }}>
                      <button
                        type="button"
                        aria-label="Menos"
                        onClick={() => changeQty(l.productoId, -1)}
                        style={qtyBtn}
                      >
                        −
                      </button>
                      <span style={{ minWidth: 24, textAlign: "center" }}>{l.cantidad}</span>
                      <button
                        type="button"
                        aria-label="Más"
                        onClick={() => changeQty(l.productoId, 1)}
                        style={qtyBtn}
                      >
                        +
                      </button>
                    </div>
                    <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                      <span style={{ fontWeight: "var(--font-weight-medium)", whiteSpace: "nowrap" }}>
                        {formatMoney(
                          previewLineSubtotal({
                            cantidad: l.cantidad,
                            precioUnitario: parsePrecio(l.precioUnitario),
                            gravado: l.afectacion === "gravado",
                          }),
                        )}
                      </span>
                      <button type="button" onClick={() => removeLine(l.productoId)} style={removeBtn}>
                        ×
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </div>

          <div
            style={{
              borderTop: "1px solid var(--color-stone-border)",
              paddingTop: "var(--spacing-12)",
              display: "flex",
              flexDirection: "column",
              gap: "var(--spacing-4)",
              fontSize: "var(--text-body)",
            }}
          >
            <div style={{ display: "flex", justifyContent: "space-between" }}>
              <span style={{ color: "var(--color-warm-gray)" }}>Valor venta</span>
              <span>{formatMoney(totals.valorVenta)}</span>
            </div>
            <div style={{ display: "flex", justifyContent: "space-between" }}>
              <span style={{ color: "var(--color-warm-gray)" }}>IGV (18%)</span>
              <span>{formatMoney(totals.igv)}</span>
            </div>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                fontWeight: "var(--font-weight-semibold)",
                fontSize: "var(--text-body-lg)",
              }}
            >
              <span>Total</span>
              <span>{formatMoney(totals.total)}</span>
            </div>
          </div>

          {success ? (
            <div
              style={{
                padding: "var(--spacing-12)",
                borderRadius: "var(--radius-inputs)",
                border: "1px solid var(--color-stone-border)",
                background: "var(--surface-canvas)",
              }}
            >
              <p style={{ margin: "0 0 var(--spacing-8)", fontWeight: "var(--font-weight-medium)" }}>
                Emitido {success.serie}-{success.numero}
              </p>
              {isBadgeStatus(success.estado) ? <Badge status={success.estado} /> : null}
              <p style={{ margin: "var(--spacing-8) 0 0" }}>
                <Link href={`/comprobantes/${success.id}`} style={{ color: "var(--color-cyan-edge)" }}>
                  Ver comprobante
                </Link>
              </p>
              <Button type="button" variant="ghost" onClick={() => setSuccess(null)} style={{ marginTop: 8 }}>
                Nueva venta
              </Button>
            </div>
          ) : null}

          {emitError ? (
            <p style={{ margin: 0, color: "var(--color-error-text)", fontSize: "var(--text-body)" }}>{emitError}</p>
          ) : null}

          <Button
            type="button"
            disabled={emitting || lines.length === 0}
            onClick={() => void emitir()}
            style={{ width: "100%" }}
          >
            {emitting ? "Emitiendo…" : "Emitir"}
          </Button>
        </section>
      </div>

      <QuickClienteModal
        open={modalCliente}
        onClose={() => setModalCliente(false)}
        onCreated={(c) => {
          setCliente(c);
          setSuccess(null);
        }}
      />

      <style jsx global>{`
        @media (max-width: 860px) {
          .caja-grid {
            grid-template-columns: 1fr !important;
          }
        }
      `}</style>
    </>
  );
}

const qtyBtn: React.CSSProperties = {
  width: 28,
  height: 28,
  borderRadius: "var(--radius-inputs)",
  border: "1px solid var(--color-stone-border)",
  background: "var(--color-pure-white)",
  cursor: "pointer",
  fontSize: "16px",
  lineHeight: 1,
};

const removeBtn: React.CSSProperties = {
  border: "none",
  background: "transparent",
  color: "var(--color-warm-gray)",
  cursor: "pointer",
  fontSize: "18px",
  lineHeight: 1,
};

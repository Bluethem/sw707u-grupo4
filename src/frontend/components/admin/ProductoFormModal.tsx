"use client";

import { FormEvent, useEffect, useState } from "react";
import { Button, Input } from "@/components/ui";
import { ApiError, apiFetch } from "@/lib/api";
import { ModalShell } from "./ModalShell";
import { AFECTACIONES, type Afectacion, type ProductoDetalle } from "./producto-types";
import { selectFieldStyle } from "./table-styles";

type Mode = "create" | "edit";

type Props = {
  open: boolean;
  mode: Mode;
  productoId: string | null;
  onClose: () => void;
  onSaved: () => void;
};

const labelStyle: React.CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: "var(--spacing-4)",
};

const labelTextStyle: React.CSSProperties = {
  fontSize: "var(--text-body)",
  color: "var(--color-warm-gray)",
};

export function ProductoFormModal({ open, mode, productoId, onClose, onSaved }: Props) {
  const [codigo, setCodigo] = useState("");
  const [descripcion, setDescripcion] = useState("");
  const [unidad, setUnidad] = useState("NIU");
  const [precio, setPrecio] = useState("");
  const [afectacion, setAfectacion] = useState<Afectacion>("gravado");
  const [controlaStock, setControlaStock] = useState(false);
  const [stockInicial, setStockInicial] = useState("");
  const [activo, setActivo] = useState(true);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!open) return;
    setError(undefined);
    if (mode === "create") {
      setCodigo("");
      setDescripcion("");
      setUnidad("NIU");
      setPrecio("");
      setAfectacion("gravado");
      setControlaStock(false);
      setStockInicial("");
      setActivo(true);
      return;
    }
    if (!productoId) return;
    setLoadingDetail(true);
    void apiFetch<ProductoDetalle>(`/api/v1/productos/${productoId}`)
      .then((p) => {
        setCodigo(p.codigo);
        setDescripcion(p.descripcion);
        setUnidad(p.unidad);
        setPrecio(p.precio);
        setAfectacion(
          AFECTACIONES.includes(p.afectacion as Afectacion)
            ? (p.afectacion as Afectacion)
            : "gravado",
        );
        setControlaStock(p.controla_stock);
        setActivo(p.activo);
        setStockInicial("");
      })
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : "No se pudo cargar el producto");
      })
      .finally(() => setLoadingDetail(false));
  }, [open, mode, productoId]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(undefined);
    setSubmitting(true);
    try {
      if (mode === "create") {
        const body: Record<string, unknown> = {
          codigo: codigo.trim(),
          descripcion: descripcion.trim(),
          unidad: unidad.trim(),
          precio: precio.trim(),
          afectacion,
          controla_stock: controlaStock,
        };
        if (controlaStock && stockInicial.trim() !== "") {
          body.stock_inicial = parseInt(stockInicial, 10);
        }
        await apiFetch<ProductoDetalle>("/api/v1/productos", { method: "POST", body });
      } else if (productoId) {
        await apiFetch<ProductoDetalle>(`/api/v1/productos/${productoId}`, {
          method: "PATCH",
          body: {
            descripcion: descripcion.trim(),
            unidad: unidad.trim(),
            precio: precio.trim(),
            afectacion,
            activo,
          },
        });
      }
      onSaved();
      onClose();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "No se pudo guardar");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <ModalShell
      open={open}
      title={mode === "create" ? "Nuevo producto" : "Editar producto"}
      onClose={onClose}
      maxWidth={480}
    >
      {loadingDetail ? (
        <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Cargando…</p>
      ) : (
        <form
          onSubmit={(e) => void onSubmit(e)}
          style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-12)" }}
        >
          {mode === "create" ? (
            <Input label="Código" value={codigo} onChange={(ev) => setCodigo(ev.target.value)} required />
          ) : (
            <Input label="Código" value={codigo} readOnly disabled />
          )}
          <Input
            label="Descripción"
            value={descripcion}
            onChange={(ev) => setDescripcion(ev.target.value)}
            required
          />
          <Input label="Unidad" value={unidad} onChange={(ev) => setUnidad(ev.target.value)} required />
          <Input
            label="Precio"
            value={precio}
            onChange={(ev) => setPrecio(ev.target.value)}
            inputMode="decimal"
            required
          />
          <label style={labelStyle}>
            <span style={labelTextStyle}>Afectación IGV</span>
            <select
              value={afectacion}
              onChange={(ev) => setAfectacion(ev.target.value as Afectacion)}
              style={selectFieldStyle}
            >
              {AFECTACIONES.map((a) => (
                <option key={a} value={a}>
                  {a}
                </option>
              ))}
            </select>
          </label>
          {mode === "create" ? (
            <>
              <label style={{ ...labelStyle, flexDirection: "row", alignItems: "center", gap: "var(--spacing-8)" }}>
                <input
                  type="checkbox"
                  checked={controlaStock}
                  onChange={(ev) => setControlaStock(ev.target.checked)}
                />
                <span style={labelTextStyle}>Controla stock</span>
              </label>
              {controlaStock ? (
                <Input
                  label="Stock inicial"
                  value={stockInicial}
                  onChange={(ev) => setStockInicial(ev.target.value)}
                  inputMode="numeric"
                />
              ) : null}
            </>
          ) : (
            <>
              <p style={{ margin: 0, fontSize: "var(--text-body)", color: "var(--color-warm-gray)" }}>
                Controla stock: {controlaStock ? "sí" : "no"} (no editable)
              </p>
              <label style={{ ...labelStyle, flexDirection: "row", alignItems: "center", gap: "var(--spacing-8)" }}>
                <input type="checkbox" checked={activo} onChange={(ev) => setActivo(ev.target.checked)} />
                <span style={labelTextStyle}>Activo</span>
              </label>
            </>
          )}
          {error ? (
            <p style={{ margin: 0, color: "var(--color-error-text)", fontSize: "var(--text-body)" }}>{error}</p>
          ) : null}
          <div style={{ display: "flex", gap: "var(--spacing-8)", justifyContent: "flex-end" }}>
            <Button type="button" variant="ghost" onClick={onClose}>
              Cancelar
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? "Guardando…" : mode === "create" ? "Crear" : "Guardar"}
            </Button>
          </div>
        </form>
      )}
    </ModalShell>
  );
}

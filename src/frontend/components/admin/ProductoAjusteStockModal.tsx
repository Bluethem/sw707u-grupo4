"use client";

import { FormEvent, useState } from "react";
import { Button, Input } from "@/components/ui";
import { ApiError, apiFetch } from "@/lib/api";
import { ModalShell } from "./ModalShell";
import { selectFieldStyle } from "./table-styles";

type Props = {
  open: boolean;
  productoId: string | null;
  productoLabel: string;
  onClose: () => void;
  onSaved: () => void;
};

export function ProductoAjusteStockModal({ open, productoId, productoLabel, onClose, onSaved }: Props) {
  const [tipo, setTipo] = useState<"entrada" | "salida">("entrada");
  const [cantidad, setCantidad] = useState("");
  const [motivo, setMotivo] = useState("");
  const [error, setError] = useState<string | undefined>();
  const [submitting, setSubmitting] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (!productoId) return;
    setError(undefined);
    setSubmitting(true);
    try {
      const qty = parseInt(cantidad, 10);
      if (!Number.isFinite(qty) || qty < 1) {
        setError("Cantidad inválida");
        return;
      }
      await apiFetch(`/api/v1/productos/${productoId}/ajustes-stock`, {
        method: "POST",
        body: { tipo, cantidad: qty, motivo: motivo.trim() },
      });
      setCantidad("");
      setMotivo("");
      onSaved();
      onClose();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "No se pudo ajustar stock");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <ModalShell open={open} title="Ajuste de stock" onClose={onClose}>
      <p style={{ margin: "0 0 var(--spacing-12)", color: "var(--color-warm-gray)", fontSize: "var(--text-body)" }}>
        {productoLabel}
      </p>
      <form
        onSubmit={(e) => void onSubmit(e)}
        style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-12)" }}
      >
        <label style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-4)" }}>
          <span style={{ fontSize: "var(--text-body)", color: "var(--color-warm-gray)" }}>Tipo</span>
          <select
            value={tipo}
            onChange={(ev) => setTipo(ev.target.value as "entrada" | "salida")}
            style={selectFieldStyle}
          >
            <option value="entrada">Entrada</option>
            <option value="salida">Salida</option>
          </select>
        </label>
        <Input
          label="Cantidad"
          value={cantidad}
          onChange={(ev) => setCantidad(ev.target.value)}
          inputMode="numeric"
          required
        />
        <Input label="Motivo" value={motivo} onChange={(ev) => setMotivo(ev.target.value)} required />
        {error ? (
          <p style={{ margin: 0, color: "var(--color-error-text)", fontSize: "var(--text-body)" }}>{error}</p>
        ) : null}
        <div style={{ display: "flex", gap: "var(--spacing-8)", justifyContent: "flex-end" }}>
          <Button type="button" variant="ghost" onClick={onClose}>
            Cancelar
          </Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? "Aplicando…" : "Aplicar ajuste"}
          </Button>
        </div>
      </form>
    </ModalShell>
  );
}

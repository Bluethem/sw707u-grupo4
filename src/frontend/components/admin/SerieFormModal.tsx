"use client";

import { FormEvent, useEffect, useState } from "react";
import { Button, Input } from "@/components/ui";
import { ApiError, apiFetch } from "@/lib/api";
import type { SerieItem, TipoComprobante } from "@/lib/caja-types";
import { ModalShell } from "./ModalShell";
import { selectFieldStyle } from "./table-styles";

type Mode = "create" | "edit";

type Props = {
  open: boolean;
  mode: Mode;
  serieRow: SerieItem | null;
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

export function SerieFormModal({ open, mode, serieRow, onClose, onSaved }: Props) {
  const [tipo, setTipo] = useState<TipoComprobante>("boleta");
  const [serie, setSerie] = useState("");
  const [siguiente, setSiguiente] = useState("1");
  const [activa, setActiva] = useState(true);
  const [error, setError] = useState<string | undefined>();
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!open) return;
    setError(undefined);
    if (mode === "create") {
      setTipo("boleta");
      setSerie("");
      setSiguiente("1");
      setActiva(true);
      return;
    }
    if (!serieRow) return;
    setTipo(serieRow.tipo);
    setSerie(serieRow.serie);
    setSiguiente(String(serieRow.siguiente));
    setActiva(serieRow.activa);
  }, [open, mode, serieRow]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(undefined);
    const nextNum = Number.parseInt(siguiente, 10);
    if (!Number.isFinite(nextNum) || nextNum < 1) {
      setError("El correlativo debe ser un entero ≥ 1");
      return;
    }
    setSubmitting(true);
    try {
      if (mode === "create") {
        const trimmed = serie.trim();
        if (!trimmed) {
          setError("Indica la serie (ej. B001 o F001)");
          return;
        }
        await apiFetch<SerieItem>("/api/v1/series", {
          method: "POST",
          body: { tipo, serie: trimmed, siguiente: nextNum },
        });
      } else if (serieRow) {
        await apiFetch<SerieItem>("/api/v1/series", {
          method: "PATCH",
          body: {
            tipo: serieRow.tipo,
            serie: serieRow.serie,
            activa,
            siguiente: nextNum,
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
      title={mode === "create" ? "Nueva serie" : "Editar serie"}
      onClose={onClose}
    >
      <form onSubmit={(ev) => void onSubmit(ev)} style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-12)" }}>
        {mode === "create" ? (
          <>
            <label style={labelStyle}>
              <span style={labelTextStyle}>Tipo</span>
              <select
                value={tipo}
                onChange={(ev) => setTipo(ev.target.value as TipoComprobante)}
                style={selectFieldStyle}
              >
                <option value="boleta">Boleta</option>
                <option value="factura">Factura</option>
              </select>
            </label>
            <Input
              label="Serie"
              placeholder="B001 o F001"
              value={serie}
              onChange={(ev) => setSerie(ev.target.value)}
              autoComplete="off"
            />
          </>
        ) : (
          <p style={{ margin: 0, fontSize: "var(--text-body)", color: "var(--color-warm-gray)" }}>
            {serieRow?.tipo} · {serieRow?.serie}
          </p>
        )}
        <Input
          label="Siguiente correlativo"
          inputMode="numeric"
          value={siguiente}
          onChange={(ev) => setSiguiente(ev.target.value.replace(/\D/g, ""))}
        />
        {mode === "edit" ? (
          <label style={{ display: "flex", alignItems: "center", gap: "var(--spacing-8)", cursor: "pointer" }}>
            <input type="checkbox" checked={activa} onChange={(ev) => setActiva(ev.target.checked)} />
            <span style={{ fontSize: "var(--text-body)" }}>Activa</span>
          </label>
        ) : null}
        {error ? (
          <p style={{ margin: 0, color: "var(--color-error-text)", fontSize: "var(--text-body)" }}>{error}</p>
        ) : null}
        <div style={{ display: "flex", justifyContent: "flex-end", gap: "var(--spacing-8)", marginTop: "var(--spacing-4)" }}>
          <Button type="button" variant="ghost" onClick={onClose} disabled={submitting}>
            Cancelar
          </Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? "Guardando…" : "Guardar"}
          </Button>
        </div>
      </form>
    </ModalShell>
  );
}

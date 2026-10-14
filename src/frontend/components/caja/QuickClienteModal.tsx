"use client";

import { FormEvent, useState } from "react";
import { Button, Input } from "@/components/ui";
import { ApiError, apiFetch } from "@/lib/api";
import type { ClienteItem } from "@/lib/caja-types";

type Props = {
  open: boolean;
  onClose: () => void;
  onCreated: (cliente: ClienteItem) => void;
};

export function QuickClienteModal({ open, onClose, onCreated }: Props) {
  const [tipoDoc, setTipoDoc] = useState<"DNI" | "RUC">("DNI");
  const [numero, setNumero] = useState("");
  const [razonSocial, setRazonSocial] = useState("");
  const [direccion, setDireccion] = useState("");
  const [error, setError] = useState<string | undefined>();
  const [submitting, setSubmitting] = useState(false);

  if (!open) return null;

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(undefined);
    setSubmitting(true);
    try {
      const cliente = await apiFetch<ClienteItem>("/api/v1/clientes", {
        method: "POST",
        body: {
          tipo_doc: tipoDoc,
          numero: numero.trim(),
          razon_social: razonSocial.trim(),
          direccion: direccion.trim(),
          correo: "",
        },
      });
      onCreated(cliente);
      setNumero("");
      setRazonSocial("");
      setDireccion("");
      onClose();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "No se pudo crear el cliente");
    } finally {
      setSubmitting(false);
    }
  }

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
          maxWidth: "420px",
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
          Cliente rápido
        </h2>
        <form
          onSubmit={(e) => void onSubmit(e)}
          style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-12)" }}
        >
          <label style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-4)" }}>
            <span style={{ fontSize: "var(--text-body)", color: "var(--color-warm-gray)" }}>
              Tipo documento
            </span>
            <select
              value={tipoDoc}
              onChange={(ev) => setTipoDoc(ev.target.value as "DNI" | "RUC")}
              style={{
                padding: "4px 12px",
                borderRadius: "var(--radius-inputs)",
                border: "1px solid var(--color-stone-muted)",
                fontFamily: "var(--font-family-body)",
                fontSize: "var(--text-body)",
              }}
            >
              <option value="DNI">DNI</option>
              <option value="RUC">RUC</option>
            </select>
          </label>
          <Input label="Número" value={numero} onChange={(ev) => setNumero(ev.target.value)} required />
          <Input
            label="Razón social / nombre"
            value={razonSocial}
            onChange={(ev) => setRazonSocial(ev.target.value)}
            required
          />
          <Input label="Dirección" value={direccion} onChange={(ev) => setDireccion(ev.target.value)} />
          {error ? (
            <p style={{ margin: 0, color: "var(--color-error-text)", fontSize: "var(--text-body)" }}>
              {error}
            </p>
          ) : null}
          <div style={{ display: "flex", gap: "var(--spacing-8)", justifyContent: "flex-end" }}>
            <Button type="button" variant="ghost" onClick={onClose}>
              Cancelar
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? "Guardando…" : "Crear"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

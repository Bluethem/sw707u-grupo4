"use client";

import { FormEvent, useEffect, useState } from "react";
import { Button, Input } from "@/components/ui";
import { ApiError, apiFetch } from "@/lib/api";
import type { ClienteItem } from "@/lib/caja-types";
import { ModalShell } from "./ModalShell";
import { selectFieldStyle } from "./table-styles";

type Mode = "create" | "edit";

type Props = {
  open: boolean;
  mode: Mode;
  clienteId: string | null;
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

export function ClienteFormModal({ open, mode, clienteId, onClose, onSaved }: Props) {
  const [tipoDoc, setTipoDoc] = useState<"DNI" | "RUC">("DNI");
  const [numero, setNumero] = useState("");
  const [razonSocial, setRazonSocial] = useState("");
  const [direccion, setDireccion] = useState("");
  const [correo, setCorreo] = useState("");
  const [activo, setActivo] = useState(true);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!open) return;
    setError(undefined);
    if (mode === "create") {
      setTipoDoc("DNI");
      setNumero("");
      setRazonSocial("");
      setDireccion("");
      setCorreo("");
      setActivo(true);
      return;
    }
    if (!clienteId) return;
    setLoadingDetail(true);
    void apiFetch<ClienteItem>(`/api/v1/clientes/${clienteId}`)
      .then((c) => {
        setTipoDoc(c.tipo_doc === "RUC" ? "RUC" : "DNI");
        setNumero(c.numero);
        setRazonSocial(c.razon_social);
        setDireccion(c.direccion);
        setCorreo(c.correo);
        setActivo(c.activo);
      })
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : "No se pudo cargar el cliente");
      })
      .finally(() => setLoadingDetail(false));
  }, [open, mode, clienteId]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(undefined);
    setSubmitting(true);
    try {
      if (mode === "create") {
        await apiFetch<ClienteItem>("/api/v1/clientes", {
          method: "POST",
          body: {
            tipo_doc: tipoDoc,
            numero: numero.trim(),
            razon_social: razonSocial.trim(),
            direccion: direccion.trim(),
            correo: correo.trim(),
          },
        });
      } else if (clienteId) {
        await apiFetch<ClienteItem>(`/api/v1/clientes/${clienteId}`, {
          method: "PATCH",
          body: {
            razon_social: razonSocial.trim(),
            direccion: direccion.trim(),
            correo: correo.trim(),
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
      title={mode === "create" ? "Nuevo cliente" : "Editar cliente"}
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
            <label style={labelStyle}>
              <span style={labelTextStyle}>Tipo documento</span>
              <select
                value={tipoDoc}
                onChange={(ev) => setTipoDoc(ev.target.value as "DNI" | "RUC")}
                style={selectFieldStyle}
              >
                <option value="DNI">DNI</option>
                <option value="RUC">RUC</option>
              </select>
            </label>
          ) : (
            <Input label="Tipo documento" value={tipoDoc} readOnly disabled />
          )}
          <Input
            label="Número"
            value={numero}
            onChange={(ev) => setNumero(ev.target.value)}
            readOnly={mode === "edit"}
            disabled={mode === "edit"}
            required={mode === "create"}
          />
          <Input
            label="Razón social / nombre"
            value={razonSocial}
            onChange={(ev) => setRazonSocial(ev.target.value)}
            required
          />
          <Input label="Dirección" value={direccion} onChange={(ev) => setDireccion(ev.target.value)} />
          <Input
            label="Correo"
            type="email"
            value={correo}
            onChange={(ev) => setCorreo(ev.target.value)}
          />
          {mode === "edit" ? (
            <label style={{ ...labelStyle, flexDirection: "row", alignItems: "center", gap: "var(--spacing-8)" }}>
              <input type="checkbox" checked={activo} onChange={(ev) => setActivo(ev.target.checked)} />
              <span style={labelTextStyle}>Activo</span>
            </label>
          ) : null}
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

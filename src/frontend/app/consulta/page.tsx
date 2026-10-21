"use client";

import { useSearchParams } from "next/navigation";
import { FormEvent, Suspense, useMemo, useState } from "react";
import { Button, Card, Input } from "@/components/ui";
import { ApiError, apiFetch } from "@/lib/api";

type ConsultaResponse = {
  id: string;
  tipo: string;
  serie: string;
  numero: number;
  fecha: string;
  total: string;
  estado: string;
  razon_social_emisor: string;
  token: string;
};

function ConsultaForm() {
  const searchParams = useSearchParams();
  const defaults = useMemo(
    () => ({
      ruc: searchParams.get("ruc") ?? "",
      serie: searchParams.get("serie") ?? "",
      numeroComprobante: searchParams.get("numero") ?? "",
    }),
    [searchParams],
  );

  const [ruc, setRuc] = useState(defaults.ruc);
  const [tipoDoc, setTipoDoc] = useState("DNI");
  const [numero, setNumero] = useState("");
  const [serie, setSerie] = useState(defaults.serie);
  const [numeroComprobante, setNumeroComprobante] = useState(defaults.numeroComprobante);
  const [fecha, setFecha] = useState("");
  const [monto, setMonto] = useState("");
  const [formError, setFormError] = useState<string | undefined>();
  const [resultado, setResultado] = useState<ConsultaResponse | undefined>();
  const [submitting, setSubmitting] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setFormError(undefined);
    setResultado(undefined);
    setSubmitting(true);
    try {
      const data = await apiFetch<ConsultaResponse>("/api/v1/publico/consulta", {
        method: "POST",
        token: null,
        body: {
          ruc,
          tipo_doc: tipoDoc,
          numero,
          serie,
          numero_comprobante: Number(numeroComprobante),
          fecha,
          monto,
        },
      });
      setResultado(data);
    } catch (err) {
      if (err instanceof ApiError) {
        setFormError(err.message);
      } else {
        setFormError("No se pudo consultar el comprobante");
      }
    } finally {
      setSubmitting(false);
    }
  }

  const apiBase = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

  return (
    <Card style={{ width: "100%", maxWidth: "480px" }}>
      <h1
        style={{
          fontSize: "var(--text-heading-sm)",
          lineHeight: "var(--leading-heading-sm)",
          letterSpacing: "var(--tracking-heading-sm)",
          marginBottom: "var(--spacing-8)",
          textAlign: "center",
        }}
      >
        Consulta tu <span className="highlight-span">comprobante</span>
      </h1>
      <p
        style={{
          textAlign: "center",
          marginBottom: "var(--spacing-24)",
          fontSize: "var(--text-body)",
        }}
      >
        FlowFacture — portal del cliente
      </p>
      <form
        onSubmit={(e) => void onSubmit(e)}
        style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-16)" }}
      >
        <Input label="RUC emisor" name="ruc" value={ruc} onChange={(ev) => setRuc(ev.target.value)} required />
        <Input
          label="Tipo documento (DNI / RUC)"
          name="tipo_doc"
          value={tipoDoc}
          onChange={(ev) => setTipoDoc(ev.target.value)}
          required
        />
        <Input
          label="Número documento"
          name="numero"
          value={numero}
          onChange={(ev) => setNumero(ev.target.value)}
          required
        />
        <Input label="Serie" name="serie" value={serie} onChange={(ev) => setSerie(ev.target.value)} required />
        <Input
          label="Número comprobante"
          name="numero_comprobante"
          inputMode="numeric"
          value={numeroComprobante}
          onChange={(ev) => setNumeroComprobante(ev.target.value)}
          required
        />
        <Input label="Fecha emisión (AAAA-MM-DD)" name="fecha" type="date" value={fecha} onChange={(ev) => setFecha(ev.target.value)} required />
        <Input label="Monto total" name="monto" value={monto} onChange={(ev) => setMonto(ev.target.value)} required />
        {formError ? (
          <p style={{ margin: 0, fontSize: "var(--text-body)", color: "var(--color-error-text)" }}>{formError}</p>
        ) : null}
        <Button type="submit" disabled={submitting} style={{ width: "100%" }}>
          {submitting ? "Consultando…" : "Consultar"}
        </Button>
      </form>
      {resultado ? (
        <div
          style={{
            marginTop: "var(--spacing-24)",
            display: "flex",
            flexDirection: "column",
            gap: "var(--spacing-12)",
            fontSize: "var(--text-body)",
          }}
        >
          <p style={{ margin: 0, color: "var(--color-ink-black)" }}>
            <strong>{resultado.razon_social_emisor}</strong> — {resultado.tipo} {resultado.serie}-
            {resultado.numero}
          </p>
          <p style={{ margin: 0 }}>Estado: {resultado.estado}</p>
          <div style={{ display: "flex", gap: "var(--spacing-8)", flexWrap: "wrap" }}>
            <a
              href={`${apiBase}/api/v1/publico/comprobantes/${resultado.id}/pdf?token=${encodeURIComponent(resultado.token)}`}
              style={{ color: "var(--color-cyan-edge)" }}
            >
              Descargar PDF
            </a>
            <a
              href={`${apiBase}/api/v1/publico/comprobantes/${resultado.id}/xml?token=${encodeURIComponent(resultado.token)}`}
              style={{ color: "var(--color-cyan-edge)" }}
            >
              Descargar XML
            </a>
          </div>
        </div>
      ) : null}
    </Card>
  );
}

export default function ConsultaPage() {
  return (
    <main>
      <Suspense fallback={null}>
        <ConsultaForm />
      </Suspense>
    </main>
  );
}

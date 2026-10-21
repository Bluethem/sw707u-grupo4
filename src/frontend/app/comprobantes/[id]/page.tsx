"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { AppNav } from "@/components/shell/AppNav";
import { RequireAuth } from "@/components/shell/RequireAuth";
import { Button } from "@/components/ui";
import { Badge, isBadgeStatus } from "@/components/ui/Badge";
import { ApiError, apiFetch, getApiBaseUrl } from "@/lib/api";
import { getToken } from "@/lib/auth";
import type { ComprobanteItem } from "@/lib/caja-types";
import { formatMoneyDecimal } from "@/lib/money";

type MeResponse = {
  rol: string;
};

function canReenviar(estado: string): boolean {
  return estado === "PENDIENTE" || estado === "ENVIADO";
}

function canSolicitarBaja(tipo: string, estado: string): boolean {
  if (estado === "PENDIENTE_BAJA" || estado === "ANULADA") return false;
  if (tipo === "factura") return estado === "ACEPTADO";
  if (tipo === "boleta") {
    return estado === "ACEPTADO" || estado === "PENDIENTE" || estado === "ENVIADO";
  }
  return false;
}

const cardSurface: React.CSSProperties = {
  background: "var(--color-pure-white)",
  borderRadius: "var(--radius-cards)",
  border: "1px solid var(--color-stone-border)",
  padding: "var(--spacing-16)",
  boxShadow: "none",
};

const thStyle: React.CSSProperties = {
  textAlign: "left",
  padding: "8px 12px",
  fontSize: "12px",
  fontWeight: "var(--font-weight-medium)",
  color: "var(--color-warm-gray)",
  borderBottom: "1px solid var(--color-stone-border)",
};

const tdStyle: React.CSSProperties = {
  padding: "10px 12px",
  fontSize: "var(--text-body)",
  borderBottom: "1px solid var(--color-stone-border)",
};

function shouldPoll(estado: string): boolean {
  return estado === "PENDIENTE" || estado === "ENVIADO";
}

export default function ComprobanteDetallePage() {
  const params = useParams();
  const router = useRouter();
  const id = typeof params.id === "string" ? params.id : "";

  const [comp, setComp] = useState<ComprobanteItem | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>();
  const [polling, setPolling] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);
  const [actionBusy, setActionBusy] = useState<"reenviar" | "baja" | null>(null);
  const [downloadBusy, setDownloadBusy] = useState<"pdf" | "xml" | "cdr" | null>(null);
  const [cdrBlocked, setCdrBlocked] = useState(false);
  const [toast, setToast] = useState<{ ok: boolean; text: string } | null>(null);

  useEffect(() => {
    let cancelled = false;
    apiFetch<MeResponse>("/api/v1/auth/me")
      .then((me) => {
        if (!cancelled) setIsAdmin(me.rol === "administrador");
      })
      .catch(() => {
        if (!cancelled) setIsAdmin(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!id) return;

    let cancelled = false;
    let timer: ReturnType<typeof setInterval> | null = null;
    let pollCount = 0;

    async function refresh(): Promise<ComprobanteItem | null> {
      try {
        const data = await apiFetch<ComprobanteItem>(`/api/v1/comprobantes/${encodeURIComponent(id)}`);
        if (!cancelled) {
          setComp(data);
          setError(undefined);
        }
        return data;
      } catch (e) {
        if (e instanceof ApiError && e.status === 401) {
          router.replace("/login");
          return null;
        }
        if (e instanceof ApiError && e.status === 404) {
          if (!cancelled) {
            setError("Comprobante no encontrado");
            setComp(null);
          }
          return null;
        }
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "Error al cargar");
        }
        return null;
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    void (async () => {
      setLoading(true);
      const data = await refresh();
      if (cancelled || !data || !shouldPoll(data.estado)) {
        setPolling(false);
        return;
      }

      setPolling(true);
      timer = setInterval(() => {
        pollCount += 1;
        if (pollCount >= 10) {
          if (timer) clearInterval(timer);
          setPolling(false);
          return;
        }
        void refresh().then((d) => {
          if (d && !shouldPoll(d.estado)) {
            if (timer) clearInterval(timer);
            setPolling(false);
          }
        });
      }, 3000);
    })();

    return () => {
      cancelled = true;
      if (timer) clearInterval(timer);
    };
  }, [id, router]);

  const cliente = comp?.cliente;

  async function reloadComprobante() {
    if (!id) return;
    try {
      const data = await apiFetch<ComprobanteItem>(`/api/v1/comprobantes/${encodeURIComponent(id)}`);
      setComp(data);
      setError(undefined);
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        router.replace("/login");
        return;
      }
      setToast({ ok: false, text: e instanceof ApiError ? e.message : "Error al actualizar" });
    }
  }

  async function handleReenviar() {
    if (!comp) return;
    setActionBusy("reenviar");
    setToast(null);
    try {
      const updated = await apiFetch<ComprobanteItem>(
        `/api/v1/comprobantes/${encodeURIComponent(comp.id)}/reenviar`,
        { method: "POST" },
      );
      setComp(updated);
      setToast({ ok: true, text: "Reenvío encolado." });
      if (shouldPoll(updated.estado)) setPolling(true);
    } catch (e) {
      setToast({ ok: false, text: e instanceof ApiError ? e.message : "No se pudo reenviar" });
    } finally {
      setActionBusy(null);
    }
  }

  async function downloadDocumento(kind: "pdf" | "xml" | "cdr") {
    if (!comp) return;
    setDownloadBusy(kind);
    setToast(null);
    if (kind === "cdr") {
      setCdrBlocked(false);
    }
    const token = getToken();
    try {
      const res = await fetch(
        `${getApiBaseUrl()}/api/v1/comprobantes/${encodeURIComponent(comp.id)}/${kind}`,
        { headers: token ? { Authorization: `Bearer ${token}` } : {} },
      );
      if (!res.ok) {
        let message = `Error HTTP ${res.status}`;
        const text = await res.text();
        try {
          const parsed = JSON.parse(text) as { error?: string };
          if (typeof parsed.error === "string") {
            message = parsed.error;
          }
        } catch {
          if (text) message = text;
        }
        if (res.status === 409 && kind === "cdr") {
          setCdrBlocked(true);
        }
        throw new ApiError(message, res.status);
      }
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `${comp.serie}-${comp.numero}.${kind === "cdr" ? "cdr" : kind}`;
      a.click();
      URL.revokeObjectURL(url);
      if (kind === "cdr") {
        setCdrBlocked(false);
      }
    } catch (e) {
      setToast({
        ok: false,
        text: e instanceof ApiError ? e.message : "No se pudo descargar",
      });
    } finally {
      setDownloadBusy(null);
    }
  }

  async function handleBaja() {
    if (!comp) return;
    const label = comp.tipo === "factura" ? "Dar de baja" : "Solicitar baja";
    const motivo = window.prompt(`${label}: indica el motivo`);
    if (motivo === null) return;
    const trimmed = motivo.trim();
    if (!trimmed) {
      setToast({ ok: false, text: "El motivo es obligatorio." });
      return;
    }
    setActionBusy("baja");
    setToast(null);
    try {
      const updated = await apiFetch<ComprobanteItem>(
        `/api/v1/comprobantes/${encodeURIComponent(comp.id)}/baja`,
        { method: "POST", body: { motivo: trimmed } },
      );
      setComp(updated);
      setToast({
        ok: true,
        text: comp.tipo === "factura" ? "Baja solicitada." : "Baja pendiente de resumen diario.",
      });
    } catch (e) {
      setToast({ ok: false, text: e instanceof ApiError ? e.message : "No se pudo solicitar la baja" });
    } finally {
      setActionBusy(null);
    }
  }

  const showReenviar = isAdmin && comp ? canReenviar(comp.estado) : false;
  const showBaja = isAdmin && comp ? canSolicitarBaja(comp.tipo, comp.estado) : false;

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
        <p style={{ margin: "0 0 var(--spacing-12)" }}>
          <Link href="/comprobantes" style={{ color: "var(--color-cyan-edge)", fontSize: "var(--text-body)" }}>
            ← Comprobantes
          </Link>
        </p>

        {loading && !comp ? (
          <p style={{ color: "var(--color-warm-gray)" }}>Cargando…</p>
        ) : error && !comp ? (
          <p style={{ color: "var(--color-error-text)" }}>{error}</p>
        ) : comp ? (
          <>
            <div
              style={{
                display: "flex",
                flexWrap: "wrap",
                alignItems: "center",
                gap: "var(--spacing-12)",
                marginBottom: "var(--spacing-16)",
              }}
            >
              <h1
                style={{
                  margin: 0,
                  fontSize: "var(--text-subheading)",
                  fontFamily: "var(--font-family-display)",
                  fontWeight: "var(--font-weight-regular)",
                  textTransform: "capitalize",
                }}
              >
                {comp.tipo} {comp.serie}-{comp.numero}
              </h1>
              {isBadgeStatus(comp.estado) ? (
                <Badge status={comp.estado} />
              ) : (
                <span style={{ fontSize: "12px" }}>{comp.estado}</span>
              )}
              {polling ? (
                <span style={{ fontSize: "12px", color: "var(--color-warm-gray)" }}>Actualizando estado…</span>
              ) : null}
            </div>

            {toast ? (
              <p
                style={{
                  margin: "0 0 var(--spacing-12)",
                  fontSize: "var(--text-body)",
                  color: toast.ok ? "var(--color-cyan-edge)" : "var(--color-error-text)",
                }}
              >
                {toast.text}
              </p>
            ) : null}

            <div
              style={{
                display: "flex",
                flexWrap: "wrap",
                gap: "var(--spacing-8)",
                marginBottom: "var(--spacing-16)",
              }}
            >
              <Button
                type="button"
                variant="ghost"
                disabled={downloadBusy !== null}
                onClick={() => void downloadDocumento("pdf")}
              >
                {downloadBusy === "pdf" ? "Descargando PDF…" : "PDF"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                disabled={downloadBusy !== null}
                onClick={() => void downloadDocumento("xml")}
              >
                {downloadBusy === "xml" ? "Descargando XML…" : "XML"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                disabled={downloadBusy !== null || cdrBlocked}
                onClick={() => void downloadDocumento("cdr")}
              >
                {downloadBusy === "cdr" ? "Descargando CDR…" : "CDR"}
              </Button>
              {cdrBlocked ? (
                <Button
                  type="button"
                  variant="ghost"
                  disabled={downloadBusy !== null}
                  onClick={() => void downloadDocumento("cdr")}
                >
                  Reintentar CDR
                </Button>
              ) : null}
            </div>

            {showReenviar || showBaja ? (
              <div
                style={{
                  display: "flex",
                  flexWrap: "wrap",
                  gap: "var(--spacing-8)",
                  marginBottom: "var(--spacing-16)",
                }}
              >
                {showReenviar ? (
                  <Button
                    type="button"
                    variant="ghost"
                    disabled={actionBusy !== null}
                    onClick={() => void handleReenviar()}
                  >
                    {actionBusy === "reenviar" ? "Reenviando…" : "Reenviar"}
                  </Button>
                ) : null}
                {showBaja ? (
                  <Button
                    type="button"
                    variant="ghost"
                    disabled={actionBusy !== null}
                    onClick={() => void handleBaja()}
                    style={{ color: "var(--color-error-text)" }}
                  >
                    {actionBusy === "baja"
                      ? "Procesando…"
                      : comp.tipo === "factura"
                        ? "Dar de baja"
                        : "Solicitar baja"}
                  </Button>
                ) : null}
                <Button type="button" variant="ghost" disabled={actionBusy !== null} onClick={() => void reloadComprobante()}>
                  Actualizar
                </Button>
              </div>
            ) : null}

            <div style={{ ...cardSurface, marginBottom: "var(--spacing-16)" }}>
              <dl
                style={{
                  margin: 0,
                  display: "grid",
                  gridTemplateColumns: "repeat(auto-fill, minmax(200px, 1fr))",
                  gap: "var(--spacing-12)",
                  fontSize: "var(--text-body)",
                }}
              >
                <div>
                  <dt style={{ margin: 0, fontSize: "12px", color: "var(--color-warm-gray)" }}>Cliente</dt>
                  <dd style={{ margin: "4px 0 0" }}>
                    {cliente ? (
                      <>
                        {cliente.razon_social || "—"}
                        <br />
                        <span style={{ color: "var(--color-warm-gray)", fontSize: "12px" }}>
                          {cliente.tipo_doc} {cliente.numero}
                        </span>
                      </>
                    ) : (
                      "—"
                    )}
                  </dd>
                </div>
                <div>
                  <dt style={{ margin: 0, fontSize: "12px", color: "var(--color-warm-gray)" }}>Valor venta</dt>
                  <dd style={{ margin: "4px 0 0" }}>{formatMoneyDecimal(comp.valor_venta)}</dd>
                </div>
                <div>
                  <dt style={{ margin: 0, fontSize: "12px", color: "var(--color-warm-gray)" }}>IGV</dt>
                  <dd style={{ margin: "4px 0 0" }}>{formatMoneyDecimal(comp.igv)}</dd>
                </div>
                <div>
                  <dt style={{ margin: 0, fontSize: "12px", color: "var(--color-warm-gray)" }}>Total</dt>
                  <dd style={{ margin: "4px 0 0", fontWeight: "var(--font-weight-semibold)" }}>
                    {formatMoneyDecimal(comp.total)}
                  </dd>
                </div>
                {comp.baja_motivo ? (
                  <div style={{ gridColumn: "1 / -1" }}>
                    <dt style={{ margin: 0, fontSize: "12px", color: "var(--color-warm-gray)" }}>Motivo de baja</dt>
                    <dd style={{ margin: "4px 0 0" }}>{comp.baja_motivo}</dd>
                  </div>
                ) : null}
              </dl>
            </div>

            <div style={{ ...cardSurface, overflowX: "auto" }}>
              <h2
                style={{
                  margin: "0 0 var(--spacing-12)",
                  fontSize: "var(--text-body-lg)",
                  fontWeight: "var(--font-weight-medium)",
                }}
              >
                Líneas
              </h2>
              {comp.lineas.length === 0 ? (
                <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Sin líneas.</p>
              ) : (
                <table style={{ width: "100%", borderCollapse: "collapse", minWidth: 560 }}>
                  <thead>
                    <tr>
                      <th style={thStyle}>Producto</th>
                      <th style={thStyle}>Cant.</th>
                      <th style={{ ...thStyle, textAlign: "right" }}>P. unit.</th>
                      <th style={thStyle}>Afect.</th>
                      <th style={{ ...thStyle, textAlign: "right" }}>Total</th>
                    </tr>
                  </thead>
                  <tbody>
                    {comp.lineas.map((l, i) => (
                      <tr key={`${l.producto_id}-${i}`}>
                        <td style={tdStyle}>
                          <span style={{ fontFamily: "monospace", fontSize: "12px" }}>{l.producto_id}</span>
                        </td>
                        <td style={tdStyle}>{l.cantidad}</td>
                        <td style={{ ...tdStyle, textAlign: "right" }}>{formatMoneyDecimal(l.precio_unitario)}</td>
                        <td style={tdStyle}>{l.afectacion}</td>
                        <td style={{ ...tdStyle, textAlign: "right" }}>{formatMoneyDecimal(l.total)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </>
        ) : null}
      </main>
    </RequireAuth>
  );
}

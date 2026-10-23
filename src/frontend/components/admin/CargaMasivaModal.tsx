"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui";
import { ApiError, apiFetch, getApiBaseUrl } from "@/lib/api";
import { getToken } from "@/lib/auth";
import { ModalShell } from "./ModalShell";

// ────────────────────────────────────────────────
// Types
// ────────────────────────────────────────────────

type ConfigCatalogo = {
  precio_minimo: string;
  precio_maximo: string | null;
  carga_max_filas: number;
  carga_max_mb: number;
  unidades_permitidas: string[];
  afectaciones_permitidas: string[];
};

type ErrorFila = {
  fila: number;
  columna: string;
  mensaje: string;
};

// Fila editable en la previsualización
type FilaEditable = {
  codigo: string;
  descripcion: string;
  unidad: string;
  precio: string;
  afectacion: string;
  controla_stock: boolean;
  stock_inicial: string; // string para edición; "" si no aplica
};

type Estado =
  | "idle"
  | "cargando_config"
  | "listo"
  | "procesando_preview"
  | "previsualizando"
  | "confirmando"
  | "exito"
  | "error_validacion"
  | "error_fatal";

// ────────────────────────────────────────────────
// Props
// ────────────────────────────────────────────────

type Props = {
  open: boolean;
  onClose: () => void;
  onSaved: () => void;
};

// ────────────────────────────────────────────────
// Helpers de estilo
// ────────────────────────────────────────────────

const inputCell: React.CSSProperties = {
  width: "100%",
  padding: "3px 6px",
  fontSize: "13px",
  border: "1px solid var(--color-stone-muted)",
  borderRadius: "4px",
  fontFamily: "var(--font-family-body)",
  background: "var(--color-pure-white)",
  boxSizing: "border-box",
};

const thPreview: React.CSSProperties = {
  padding: "6px 8px",
  fontSize: "12px",
  fontWeight: "600",
  color: "var(--color-warm-gray)",
  borderBottom: "1px solid var(--color-stone-border)",
  whiteSpace: "nowrap",
  background: "var(--color-pure-white)",
  position: "sticky",
  top: 0,
};

const tdPreview: React.CSSProperties = {
  padding: "4px 6px",
  borderBottom: "1px solid var(--color-stone-border)",
  verticalAlign: "middle",
};

// ────────────────────────────────────────────────
// Component principal
// ────────────────────────────────────────────────

export function CargaMasivaModal({ open, onClose, onSaved }: Props) {
  const router = useRouter();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [estado, setEstado] = useState<Estado>("idle");
  const [config, setConfig] = useState<ConfigCatalogo | null>(null);
  const [archivo, setArchivo] = useState<File | null>(null);
  const [drag, setDrag] = useState(false);

  // Resultado del preview
  const [filasEditables, setFilasEditables] = useState<FilaEditable[]>([]);

  // Errores de validación (preview o confirmar)
  const [erroresFila, setErroresFila] = useState<ErrorFila[]>([]);
  const [errorFatal, setErrorFatal] = useState<string | undefined>();
  const [cantidadCargada, setCantidadCargada] = useState<number | null>(null);

  // ── Cargar config al abrir ─────────────────────────────────────────

  const cargarConfig = useCallback(async () => {
    setEstado("cargando_config");
    setErrorFatal(undefined);
    try {
      const data = await apiFetch<ConfigCatalogo>("/api/v1/configuracion/catalogo");
      setConfig(data);
      setEstado("listo");
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) { router.replace("/login"); return; }
      setErrorFatal(e instanceof Error ? e.message : "No se pudo cargar la configuración");
      setEstado("error_fatal");
    }
  }, [router]);

  useEffect(() => {
    if (open) {
      setArchivo(null);
      setFilasEditables([]);
      setErroresFila([]);
      setErrorFatal(undefined);
      setCantidadCargada(null);
      void cargarConfig();
    }
  }, [open, cargarConfig]);

  // ── Selección de archivo ──────────────────────────────────────────

  function seleccionarArchivo(f: File | null) {
    setErroresFila([]);
    setErrorFatal(undefined);
    setCantidadCargada(null);
    setFilasEditables([]);
    if (!f) { setArchivo(null); return; }
    const ext = f.name.split(".").pop()?.toLowerCase();
    if (ext !== "xlsx" && ext !== "xls") {
      setErrorFatal("El archivo debe tener extensión .xlsx o .xls");
      setArchivo(null);
      return;
    }
    const maxMB = config?.carga_max_mb ?? 20;
    if (f.size > maxMB * 1024 * 1024) {
      setErrorFatal(`El archivo supera el límite de ${maxMB} MB`);
      setArchivo(null);
      return;
    }
    setArchivo(f);
    if (fileInputRef.current) fileInputRef.current.value = "";
  }

  function handleDragOver(e: React.DragEvent) { e.preventDefault(); setDrag(true); }
  function handleDragLeave() { setDrag(false); }
  function handleDrop(e: React.DragEvent) {
    e.preventDefault(); setDrag(false);
    seleccionarArchivo(e.dataTransfer.files[0] ?? null);
  }

  // ── Paso 1: Enviar a preview ───────────────────────────────────────

  async function handlePreview() {
    if (!archivo) return;
    setEstado("procesando_preview");
    setErroresFila([]);
    setErrorFatal(undefined);

    const form = new FormData();
    form.append("archivo", archivo);

    try {
      const token = getToken();
      const resp = await fetch(`${getApiBaseUrl()}/api/v1/productos/carga-masiva/preview`, {
        method: "POST",
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        body: form,
      });

      if (resp.status === 401) { router.replace("/login"); return; }

      const json = await resp.json() as Record<string, unknown>;

      if (!resp.ok) {
        const errores = json.errores as ErrorFila[] | undefined;
        if (errores && errores.length > 0) {
          setErroresFila(errores);
          setEstado("error_validacion");
        } else {
          setErrorFatal((json.error as string) ?? `Error ${resp.status}`);
          setEstado("error_fatal");
        }
        return;
      }

      // Éxito: convertir filas a editables
      const filas = (json.filas as Array<{
        codigo: string; descripcion: string; unidad: string; precio: string;
        afectacion: string; controla_stock: boolean; stock_inicial: number | null;
      }>) ?? [];

      setFilasEditables(filas.map((f) => ({
        codigo: f.codigo,
        descripcion: f.descripcion,
        unidad: f.unidad,
        precio: f.precio,
        afectacion: f.afectacion,
        controla_stock: f.controla_stock,
        stock_inicial: f.stock_inicial != null ? String(f.stock_inicial) : "",
      })));
      setEstado("previsualizando");
    } catch (e) {
      setErrorFatal(e instanceof Error ? e.message : "Error de red");
      setEstado("error_fatal");
    }
  }

  // ── Paso 2: Confirmar la carga con las filas editadas ─────────────

  async function handleConfirmar() {
    setEstado("confirmando");
    setErroresFila([]);
    setErrorFatal(undefined);

    // Convertir filas editables al formato que espera el backend
    const filas = filasEditables.map((f) => ({
      codigo: f.codigo,
      descripcion: f.descripcion,
      unidad: f.unidad,
      precio: f.precio,
      afectacion: f.afectacion,
      controla_stock: f.controla_stock,
      stock_inicial: f.controla_stock && f.stock_inicial !== ""
        ? parseInt(f.stock_inicial, 10)
        : null,
    }));

    try {
      const json = await apiFetch<{ cantidad?: number; errores?: ErrorFila[] }>(
        "/api/v1/productos/carga-masiva/confirmar",
        { method: "POST", body: { filas } }
      );

      if (json.errores && json.errores.length > 0) {
        setErroresFila(json.errores);
        setEstado("previsualizando"); // volver a la tabla para corregir
        return;
      }

      setCantidadCargada(json.cantidad ?? filasEditables.length);
      setEstado("exito");
      onSaved();
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) { router.replace("/login"); return; }
      if (e instanceof ApiError) {
        // Puede ser 422 con errores de validación
        const body = e.body as { errores?: ErrorFila[] } | undefined;
        if (body?.errores && body.errores.length > 0) {
          setErroresFila(body.errores);
          setEstado("previsualizando");
          return;
        }
      }
      setErrorFatal(e instanceof Error ? e.message : "Error de red");
      setEstado("error_fatal");
    }
  }

  function volverASeleccion() {
    setFilasEditables([]);
    setErroresFila([]);
    setErrorFatal(undefined);
    setEstado("listo");
  }

  // ── Render ────────────────────────────────────────────────────────

  const enProgreso = estado === "procesando_preview" || estado === "confirmando";

  // Si estamos en la pantalla de previsualización, mostramos el modal más ancho
  const maxWidth = estado === "previsualizando" ? 900 : 640;

  return (
    <ModalShell open={open} title="Carga masiva de productos" onClose={onClose} maxWidth={maxWidth}>
      <div style={{ display: "flex", flexDirection: "column", gap: "var(--spacing-20)" }}>

        {/* ── Pantalla: Selección de archivo ── */}
        {(estado === "listo" || estado === "procesando_preview" ||
          estado === "error_validacion" || estado === "error_fatal") && (
          <>
            {config && <ConfigResumen config={config} />}

            <div>
              <div
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
                onClick={() => fileInputRef.current?.click()}
                style={{
                  border: `2px dashed ${drag ? "var(--color-cyan-signal)" : "var(--color-stone-border)"}`,
                  borderRadius: "var(--radius-cards)",
                  padding: "var(--spacing-24) var(--spacing-16)",
                  textAlign: "center",
                  cursor: "pointer",
                  background: drag ? "var(--color-sky-wash)" : "var(--color-stone-canvas)",
                  transition: "border-color 0.15s, background 0.15s",
                }}
              >
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".xlsx,.xls"
                  style={{ display: "none" }}
                  onChange={(e) => seleccionarArchivo(e.target.files?.[0] ?? null)}
                />
                <p style={{ margin: "0 0 var(--spacing-8)", fontSize: "var(--text-body)", color: "var(--color-ink-black)" }}>
                  {archivo
                    ? <><strong>{archivo.name}</strong> ({(archivo.size / 1024).toFixed(1)} KB)</>
                    : "Arrastra un archivo Excel aquí, o haz clic para seleccionar"
                  }
                </p>
                <p style={{ margin: 0, fontSize: "13px", color: "var(--color-warm-gray)" }}>
                  Formatos: .xlsx, .xls{config ? ` · Máx. ${config.carga_max_mb} MB` : ""}
                </p>
              </div>
              <ColsEsperadas />
            </div>

            {errorFatal && <MsgError>{errorFatal}</MsgError>}

            {estado === "error_validacion" && erroresFila.length > 0 && (
              <TablaErrores errores={erroresFila} />
            )}

            <Acciones>
              <Button type="button" variant="ghost" onClick={onClose} disabled={enProgreso}>
                Cancelar
              </Button>
              <Button
                type="button"
                disabled={!archivo || enProgreso}
                onClick={() => void handlePreview()}
              >
                {estado === "procesando_preview" ? "Procesando…" : "Previsualizar"}
              </Button>
            </Acciones>
          </>
        )}

        {/* ── Pantalla: Cargando config ── */}
        {estado === "cargando_config" && (
          <p style={{ margin: 0, color: "var(--color-warm-gray)" }}>Cargando configuración…</p>
        )}

        {/* ── Pantalla: Previsualización editable ── */}
        {(estado === "previsualizando" || estado === "confirmando") && (
          <>
            <div style={{ display: "flex", alignItems: "center", gap: "var(--spacing-12)" }}>
              <span style={{ fontSize: "var(--text-body)", color: "var(--color-warm-gray)" }}>
                {filasEditables.length} producto{filasEditables.length === 1 ? "" : "s"} listos para cargar.
                Puedes editar los valores antes de confirmar.
              </span>
            </div>

            {erroresFila.length > 0 && <TablaErrores errores={erroresFila} />}

            <TablaPreview
              filas={filasEditables}
              onChange={setFilasEditables}
              unidadesPermitidas={config?.unidades_permitidas ?? []}
              afectacionesPermitidas={config?.afectaciones_permitidas ?? []}
            />

            <Acciones>
              <Button type="button" variant="ghost" onClick={volverASeleccion} disabled={enProgreso}>
                ← Volver
              </Button>
              <Button
                type="button"
                disabled={filasEditables.length === 0 || enProgreso}
                onClick={() => void handleConfirmar()}
              >
                {estado === "confirmando" ? "Cargando…" : `Confirmar ${filasEditables.length} producto${filasEditables.length === 1 ? "" : "s"}`}
              </Button>
            </Acciones>
          </>
        )}

        {/* ── Pantalla: Éxito ── */}
        {estado === "exito" && cantidadCargada !== null && (
          <>
            <MsgSuccess>
              Se cargaron correctamente <strong>{cantidadCargada}</strong>{" "}
              producto{cantidadCargada === 1 ? "" : "s"}.
            </MsgSuccess>
            <Acciones>
              <Button type="button" onClick={onClose}>Cerrar</Button>
            </Acciones>
          </>
        )}

      </div>
    </ModalShell>
  );
}

// ────────────────────────────────────────────────
// TablaPreview — tabla editable de filas
// ────────────────────────────────────────────────

type TablaPreviewProps = {
  filas: FilaEditable[];
  onChange: (filas: FilaEditable[]) => void;
  unidadesPermitidas: string[];
  afectacionesPermitidas: string[];
};

function TablaPreview({ filas, onChange, unidadesPermitidas, afectacionesPermitidas }: TablaPreviewProps) {
  function updateFila(idx: number, patch: Partial<FilaEditable>) {
    onChange(filas.map((f, i) => (i === idx ? { ...f, ...patch } : f)));
  }

  function eliminarFila(idx: number) {
    onChange(filas.filter((_, i) => i !== idx));
  }

  return (
    <div style={{
      maxHeight: 380,
      overflowY: "auto",
      overflowX: "auto",
      border: "1px solid var(--color-stone-border)",
      borderRadius: "var(--radius-cards)",
    }}>
      <table style={{ width: "100%", borderCollapse: "collapse", minWidth: 760 }}>
        <thead>
          <tr>
            <th style={thPreview}>#</th>
            <th style={thPreview}>Código</th>
            <th style={{ ...thPreview, minWidth: 180 }}>Descripción</th>
            <th style={thPreview}>Unidad</th>
            <th style={thPreview}>Precio</th>
            <th style={thPreview}>Afectación</th>
            <th style={thPreview}>Stock</th>
            <th style={thPreview}>Stk. Inicial</th>
            <th style={thPreview}></th>
          </tr>
        </thead>
        <tbody>
          {filas.map((fila, idx) => (
            <tr key={idx} style={{ background: idx % 2 === 0 ? "var(--color-pure-white)" : "var(--color-stone-canvas)" }}>
              <td style={{ ...tdPreview, color: "var(--color-warm-gray)", fontSize: "12px", textAlign: "center", width: 28 }}>
                {idx + 1}
              </td>
              <td style={{ ...tdPreview, width: 100 }}>
                <input
                  style={inputCell}
                  value={fila.codigo}
                  onChange={(e) => updateFila(idx, { codigo: e.target.value })}
                />
              </td>
              <td style={{ ...tdPreview, minWidth: 180 }}>
                <input
                  style={inputCell}
                  value={fila.descripcion}
                  onChange={(e) => updateFila(idx, { descripcion: e.target.value })}
                />
              </td>
              <td style={{ ...tdPreview, width: 80 }}>
                <select
                  style={{ ...inputCell, padding: "3px 4px" }}
                  value={fila.unidad}
                  onChange={(e) => updateFila(idx, { unidad: e.target.value })}
                >
                  {unidadesPermitidas.map((u) => (
                    <option key={u} value={u}>{u}</option>
                  ))}
                </select>
              </td>
              <td style={{ ...tdPreview, width: 80 }}>
                <input
                  style={inputCell}
                  value={fila.precio}
                  onChange={(e) => updateFila(idx, { precio: e.target.value })}
                  inputMode="decimal"
                />
              </td>
              <td style={{ ...tdPreview, width: 100 }}>
                <select
                  style={{ ...inputCell, padding: "3px 4px" }}
                  value={fila.afectacion}
                  onChange={(e) => updateFila(idx, { afectacion: e.target.value })}
                >
                  {afectacionesPermitidas.map((a) => (
                    <option key={a} value={a}>{a.charAt(0).toUpperCase() + a.slice(1)}</option>
                  ))}
                </select>
              </td>
              <td style={{ ...tdPreview, width: 60, textAlign: "center" }}>
                <input
                  type="checkbox"
                  checked={fila.controla_stock}
                  onChange={(e) => updateFila(idx, {
                    controla_stock: e.target.checked,
                    stock_inicial: e.target.checked ? (fila.stock_inicial || "0") : "",
                  })}
                />
              </td>
              <td style={{ ...tdPreview, width: 80 }}>
                {fila.controla_stock ? (
                  <input
                    style={inputCell}
                    value={fila.stock_inicial}
                    onChange={(e) => updateFila(idx, { stock_inicial: e.target.value })}
                    inputMode="numeric"
                  />
                ) : (
                  <span style={{ color: "var(--color-warm-gray)", fontSize: "12px" }}>—</span>
                )}
              </td>
              <td style={{ ...tdPreview, width: 32, textAlign: "center" }}>
                <button
                  type="button"
                  title="Eliminar fila"
                  onClick={() => eliminarFila(idx)}
                  style={{
                    background: "none",
                    border: "none",
                    cursor: "pointer",
                    color: "var(--color-warm-gray)",
                    fontSize: "16px",
                    padding: "0 4px",
                    lineHeight: 1,
                  }}
                >
                  ×
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ────────────────────────────────────────────────
// Sub-componentes auxiliares
// ────────────────────────────────────────────────

function ConfigResumen({ config }: { config: ConfigCatalogo }) {
  return (
    <div style={{
      padding: "var(--spacing-12)",
      borderRadius: "var(--radius-cards)",
      border: "1px solid var(--color-stone-border)",
      background: "var(--color-stone-canvas)",
      fontSize: "13px",
      color: "var(--color-warm-gray)",
    }}>
      <strong style={{ color: "var(--color-ink-black)", display: "block", marginBottom: "var(--spacing-8)" }}>
        Reglas de validación activas
      </strong>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "4px 16px" }}>
        <span>Precio mínimo: <strong style={{ color: "var(--color-ink-black)" }}>S/ {config.precio_minimo}</strong></span>
        <span>Precio máximo: <strong style={{ color: "var(--color-ink-black)" }}>
          {config.precio_maximo ? `S/ ${config.precio_maximo}` : "Sin límite"}
        </strong></span>
        <span>Filas máx.: <strong style={{ color: "var(--color-ink-black)" }}>{config.carga_max_filas}</strong></span>
        <span>Tamaño máx.: <strong style={{ color: "var(--color-ink-black)" }}>{config.carga_max_mb} MB</strong></span>
        <span>Unidades: <strong style={{ color: "var(--color-ink-black)" }}>{config.unidades_permitidas.join(", ")}</strong></span>
        <span>Afectaciones: <strong style={{ color: "var(--color-ink-black)" }}>
          {config.afectaciones_permitidas.map((a) => a.charAt(0).toUpperCase() + a.slice(1)).join(", ")}
        </strong></span>
      </div>
    </div>
  );
}

function ColsEsperadas() {
  const cols = ["codigo", "descripcion", "unidad", "precio", "afectacion", "controla_stock", "stock_inicial"];
  return (
    <div style={{ marginTop: "var(--spacing-8)" }}>
      <p style={{ margin: "0 0 4px", fontSize: "13px", color: "var(--color-warm-gray)" }}>
        Columnas requeridas en la primera hoja:
      </p>
      <div style={{ display: "flex", flexWrap: "wrap", gap: "4px" }}>
        {cols.map((c) => (
          <code key={c} style={{
            fontSize: "12px", padding: "2px 6px", borderRadius: "4px",
            background: "var(--color-stone-canvas)", border: "1px solid var(--color-stone-border)",
            color: "var(--color-ink-black)",
          }}>{c}</code>
        ))}
      </div>
    </div>
  );
}

function TablaErrores({ errores }: { errores: ErrorFila[] }) {
  const thS: React.CSSProperties = {
    padding: "8px 12px", textAlign: "left", fontSize: "13px",
    fontWeight: "var(--font-weight-medium)", color: "var(--color-warm-gray)",
    borderBottom: "1px solid var(--color-stone-border)", whiteSpace: "nowrap",
  };
  const tdS: React.CSSProperties = {
    padding: "8px 12px", fontSize: "var(--text-body)", color: "var(--color-ink-black)",
    borderBottom: "1px solid var(--color-stone-border)", verticalAlign: "top",
  };
  return (
    <div>
      <p style={{ margin: "0 0 var(--spacing-8)", fontSize: "var(--text-body)", color: "var(--color-error-text)", fontWeight: "var(--font-weight-medium)" }}>
        {errores.length} error{errores.length === 1 ? "" : "es"} encontrado{errores.length === 1 ? "" : "s"}. Corrige y vuelve a intentarlo.
      </p>
      <div style={{ maxHeight: 200, overflowY: "auto", border: "1px solid var(--color-stone-border)", borderRadius: "var(--radius-cards)" }}>
        <table style={{ width: "100%", borderCollapse: "collapse" }}>
          <thead style={{ position: "sticky", top: 0, background: "var(--color-pure-white)" }}>
            <tr>
              <th style={thS}>Fila</th>
              <th style={thS}>Columna</th>
              <th style={{ ...thS, width: "100%" }}>Mensaje</th>
            </tr>
          </thead>
          <tbody>
            {errores.map((err, i) => (
              <tr key={i}>
                <td style={{ ...tdS, fontVariantNumeric: "tabular-nums" }}>{err.fila}</td>
                <td style={{ ...tdS, fontFamily: "monospace", fontSize: "13px" }}>{err.columna}</td>
                <td style={tdS}>{err.mensaje}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function MsgError({ children }: { children: React.ReactNode }) {
  return (
    <p style={{
      margin: 0, padding: "var(--spacing-12)", borderRadius: "var(--radius-cards)",
      background: "var(--color-error-bg, #fff0f0)", color: "var(--color-error-text)",
      fontSize: "var(--text-body)",
    }}>{children}</p>
  );
}

function MsgSuccess({ children }: { children: React.ReactNode }) {
  return (
    <p style={{
      margin: 0, padding: "var(--spacing-12)", borderRadius: "var(--radius-cards)",
      background: "var(--color-success-bg, #f0fff4)", color: "var(--color-success-text, #166534)",
      fontSize: "var(--text-body)",
    }}>{children}</p>
  );
}

function Acciones({ children }: { children: React.ReactNode }) {
  return (
    <div style={{ display: "flex", gap: "var(--spacing-12)", justifyContent: "flex-end" }}>
      {children}
    </div>
  );
}

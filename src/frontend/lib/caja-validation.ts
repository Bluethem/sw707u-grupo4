import type { ClienteItem, TipoComprobante } from "./caja-types";

const BOLETA_UMBRAL = 700;

export function clienteRequerido(
  tipo: TipoComprobante,
  totalPreview: number,
): boolean {
  if (tipo === "factura") return true;
  return totalPreview > BOLETA_UMBRAL;
}

export function validarClienteParaEmitir(
  tipo: TipoComprobante,
  totalPreview: number,
  cliente: ClienteItem | null,
): string | null {
  if (!clienteRequerido(tipo, totalPreview)) return null;
  if (!cliente) {
    return tipo === "factura"
      ? "La factura requiere un cliente con RUC"
      : "Boleta mayor a S/ 700 requiere cliente (DNI o RUC)";
  }
  if (tipo === "factura") {
    if (cliente.tipo_doc !== "RUC") {
      return "La factura requiere un cliente con RUC";
    }
  } else if (totalPreview > BOLETA_UMBRAL) {
    if (cliente.tipo_doc !== "DNI" && cliente.tipo_doc !== "RUC") {
      return "Indique cliente con DNI o RUC";
    }
  }
  return null;
}

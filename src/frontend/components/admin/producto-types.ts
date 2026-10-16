import type { Paginated, ProductoItem } from "@/lib/caja-types";

export type { Paginated };

export type ProductoDetalle = ProductoItem & {
  unidad: string;
  empresa_id?: string;
  creado_en?: string;
  actualizado_en?: string;
};

export type StockMovimiento = {
  id: string;
  empresa_id: string;
  producto_id: string;
  tipo: string;
  cantidad: number;
  motivo: string;
  creado_en: string;
  usuario_id?: string;
};

export const AFECTACIONES = ["gravado", "exonerado", "inafecto"] as const;
export type Afectacion = (typeof AFECTACIONES)[number];

export type ConfigCatalogo = {
  empresa_id?: string;
  precio_minimo: string;
  precio_maximo: string | null;
  carga_max_filas: number;
  carga_max_mb: number;
  unidades_permitidas: string[];
  afectaciones_permitidas: string[];
  actualizado_en?: string;
};

export const UNIDADES_SUNAT = ["NIU", "ZZ", "KGM", "LTR", "MTR", "BX"] as const;
export const AFECTACIONES_IGV = ["gravado", "exonerado", "inafecto"] as const;

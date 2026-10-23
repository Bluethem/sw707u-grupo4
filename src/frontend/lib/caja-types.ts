export type TipoComprobante = "boleta" | "factura";

export type ProductoItem = {
  id: string;
  codigo: string;
  descripcion: string;
  precio: string;
  afectacion: string;
  controla_stock: boolean;
  activo: boolean;
  stock?: number;
};

export type ClienteItem = {
  id: string;
  tipo_doc: string;
  numero: string;
  razon_social: string;
  direccion: string;
  correo: string;
  activo: boolean;
};

export type SerieItem = {
  empresa_id: string;
  tipo: TipoComprobante;
  serie: string;
  siguiente: number;
  activa: boolean;
};

export type CartLine = {
  productoId: string;
  codigo: string;
  descripcion: string;
  precioUnitario: string;
  afectacion: string;
  cantidad: number;
};

export type ComprobanteEmitido = {
  id: string;
  tipo: TipoComprobante;
  serie: string;
  numero: number;
  estado: string;
  total: string;
};

export type ComprobanteCliente = {
  tipo_doc: string;
  numero: string;
  razon_social: string;
};

export type ComprobanteLinea = {
  producto_id: string;
  cantidad: number;
  precio_unitario: string;
  afectacion: string;
  valor_venta: string;
  igv: string;
  total: string;
};

/** GET /api/v1/comprobantes and GET /api/v1/comprobantes/{id} */
export type ComprobanteItem = {
  id: string;
  tipo: TipoComprobante;
  serie: string;
  numero: number;
  estado: string;
  cliente: ComprobanteCliente | null;
  valor_venta: string;
  igv: string;
  total: string;
  idempotency_key: string;
  baja_motivo?: string;
  lineas: ComprobanteLinea[];
};

/** GET/POST /api/v1/resumenes */
export type ResumenComprobanteItem = {
  comprobante_id: string;
  anulado: boolean;
  tipo?: TipoComprobante | string;
  serie?: string;
  numero?: number;
  estado?: string;
};

export type ResumenDiarioItem = {
  id: string;
  empresa_id: string;
  fecha: string;
  correlativo: number;
  estado: string;
  ticket: string;
  creado_en: string;
  comprobantes: ResumenComprobanteItem[];
};

export type Paginated<T> = {
  items: T[];
  total: number;
};

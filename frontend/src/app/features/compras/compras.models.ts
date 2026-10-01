export interface CatalogOption {
  id: string;
  nombre: string;
}
export interface PurchaseProduct {
  id: string;
  nombre: string;
  sku: string | null;
  codigo_barras: string | null;
}
export interface PackagingUnit extends CatalogOption {
  tipo: string;
  factor_a_base: number;
}
export interface PurchaseDetail {
  producto_negocio_id: string;
  unidad_medida_id: string;
  cantidad_empaques: number;
  costo_empaque: number;
  descuento: number;
  piezas_danadas: number;
  piezas_faltantes: number;
  numero_lote: string | null;
  fecha_caducidad: string | null;
  producto_nuevo: boolean;
  nombre_producto_nuevo: string;
  codigo_barras_nuevo: string;
}
export interface CreatePurchaseRequest {
  proveedor_id: string;
  sucursal_id: string;
  tipo_documento: 'TICKET' | 'FACTURA';
  folio_documento: string;
  numero_cuenta: string | null;
  fecha_documento: string;
  fecha_recepcion: string;
  forma_pago: string;
  incluye_ieps: boolean;
  ieps: number;
  incluye_iva: boolean;
  observaciones: string | null;
  detalles: PurchaseDetail[];
}
export interface PurchaseRow {
  id: string;
  proveedor_id: string;
  sucursal_id: string;
  tipo_documento: string;
  folio_documento: string;
  fecha_recepcion: string;
  forma_pago: string;
  importe_pagado: number;
  total: number;
}
export interface PurchaseListResponse {
  items: PurchaseRow[];
  total: number;
}

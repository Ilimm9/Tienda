export interface CatalogOption {
  id: string;
  nombre: string;
}
export interface PurchaseProduct {
  id: string;
  nombre: string;
  sku: string | null;
  codigo_barras: string | null;
  del_proveedor: boolean;
  /** Tasas recordadas del producto (fracción); null si nunca se capturaron. */
  tasa_ieps?: number | null;
  tasa_iva?: number | null;
}
export interface PurchaseProductGroup {
  label: string;
  items: PurchaseProduct[];
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
  /** Fracción: 0.08 = 8 %. */
  tasa_ieps: number;
  tasa_iva: number;
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
  /** Importe de IEPS del ticket; la diferencia se reparte entre los productos con IEPS. */
  ieps_ajustado: number | null;
  observaciones: string | null;
  detalles: PurchaseDetail[];
}
export interface PurchaseRow {
  id: string;
  proveedor_id: string;
  sucursal_id: string;
  tipo_documento: string;
  folio_documento: string;
  fecha_documento?: string;
  fecha_recepcion: string;
  forma_pago: string;
  importe_pagado: number;
  total: number;
  subtotal?: number;
  ieps?: number;
  iva?: number;
  numero_cuenta?: string | null;
  observaciones?: string | null;
  /** ACTIVA | SUSPENDIDA. */
  estado?: string;
  motivo_suspension?: string | null;
  suspendido_en?: string | null;
  proveedor?: string;
  productos?: number;
  piezas?: number;
  costos_pendientes?: number;
  costos_validados?: number;
  costos_cancelados?: number;
}
/** Filtro del Excel: compras elegidas o proveedor y fechas de recepción. */
export interface ExportFilter {
  sucursal_id: string;
  compra_id?: string[];
  proveedor_id?: string;
  desde?: string;
  hasta?: string;
  comprobacion: boolean;
  incluir_suspendidas: boolean;
}
export interface PurchaseListResponse {
  items: PurchaseRow[];
  total: number;
}

/** Renglón del detalle de una compra. */
export interface PurchaseViewRow {
  id: string;
  producto_negocio_id: string;
  unidad_medida_id: string;
  producto: string;
  unidad: string;
  cantidad_empaques: number;
  piezas_por_empaque: number;
  total_piezas: number;
  piezas_buenas: number;
  piezas_danadas: number;
  piezas_faltantes: number;
  costo_empaque: number;
  descuento: number;
  subtotal: number;
  ieps: number;
  iva: number;
  tasa_ieps: number;
  tasa_iva: number;
  impuestos_por_renglon: boolean;
  numero_lote: string | null;
  fecha_caducidad: string | null;
  costo_final: number;
  /** PENDIENTE | AUTORIZADA | CANCELADA | ''. */
  estado_costo: string;
}
export interface PurchaseView extends PurchaseRow {
  fecha_documento: string;
  subtotal: number;
  ieps: number;
  iva: number;
  renglones: PurchaseViewRow[];
}
export interface EditPurchaseDetail {
  id?: string;
  quitar?: boolean;
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
  tasa_ieps: number;
  tasa_iva: number;
}
export interface EditPurchaseRequest {
  tipo_documento: string;
  folio_documento: string;
  numero_cuenta: string | null;
  fecha_documento: string;
  fecha_recepcion: string;
  forma_pago: string;
  observaciones: string | null;
  motivo: string | null;
  detalles?: EditPurchaseDetail[];
}

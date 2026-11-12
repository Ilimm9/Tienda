import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface PriceProposal {
  id: string;
  producto_negocio_id: string;
  producto: string;
  piezas_buenas: number | string;
  costo_capturado: number | string;
  costo_anterior: number | string | null;
  precio_anterior: number | string;
  margen_sugerido: number | string;
  precio_sugerido: number | string;
  compra_id: string;
  tipo_documento: string;
  folio_documento: string;
  proveedor_id: string;
  proveedor: string;
  fecha_documento: string;
  fecha_recepcion: string;
  total_compra: number | string;
  codigo_barras: string | null;
  nota: string;
  impuestos_por_renglon: boolean;
  tasa_ieps: number | string;
  tasa_iva: number | string;
  producto_nuevo: boolean;
}
export interface PriceHistoryItem { id: string; origen: 'ALTA_INICIAL' | 'COMPRA_AUTORIZADA' | 'AJUSTE_PRECIO_MANUAL' | 'CARGA_MASIVA'; costo_anterior: number | null; costo_nuevo: number | null; precio_anterior: number | null; precio_nuevo: number | null; motivo: string | null; creado_en: string; }
export interface FiltroPendientes { q?: string; proveedor_id?: string; desde?: string; hasta?: string; }
export interface Autorizacion { costo: number; margen: number; precio_venta: number; }

@Injectable({ providedIn: 'root' })
export class PreciosService {
  private readonly http = inject(HttpClient);
  private base(businessId: string) { return `${environment.apiUrl}/negocios/${businessId}/precios`; }
  pending(businessId: string, branchId: string, filtro: FiltroPendientes = {}): Observable<{ items: PriceProposal[] }> {
    let params = new HttpParams().set('sucursal_id', branchId);
    for (const [key, value] of Object.entries(filtro)) if (value) params = params.set(key, value);
    return this.http.get<{ items: PriceProposal[] }>(`${this.base(businessId)}/propuestas`, { params });
  }
  authorize(businessId: string, id: string, value: Autorizacion): Observable<unknown> { return this.http.patch(`${this.base(businessId)}/propuestas/${id}/autorizar`, value); }
  /** Valida varios renglones de un ticket en una sola transacción. */
  authorizePurchase(businessId: string, purchaseId: string, renglones: (Autorizacion & { propuesta_id: string })[]): Observable<unknown> {
    return this.http.patch(`${this.base(businessId)}/compras/${purchaseId}/autorizar`, { renglones });
  }
  history(businessId: string, branchId: string, productId: string): Observable<{items: PriceHistoryItem[]}> { return this.http.get<{items: PriceHistoryItem[]}>(`${this.base(businessId)}/historial`, { params: { sucursal_id: branchId, producto_negocio_id: productId } }); }
  adjust(businessId: string, productId: string, value: {sucursal_id: string; precio_venta: number; motivo: string}): Observable<PriceHistoryItem> { return this.http.patch<PriceHistoryItem>(`${this.base(businessId)}/productos/${productId}`, value); }
}

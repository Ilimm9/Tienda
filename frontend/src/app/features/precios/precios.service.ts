import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface PriceProposal { id: string; producto: string; piezas_buenas: number | string; costo_capturado: number | string; costo_anterior: number | string | null; precio_anterior: number | string; margen_sugerido: number | string; precio_sugerido: number | string; ganancia?: number; }
@Injectable({ providedIn: 'root' })
export class PreciosService {
  private readonly http = inject(HttpClient);
  private base(businessId: string) { return `${environment.apiUrl}/negocios/${businessId}/precios`; }
  pending(businessId: string, branchId: string): Observable<{items: PriceProposal[]}> { return this.http.get<{items: PriceProposal[]}>(`${this.base(businessId)}/propuestas`, { params: { sucursal_id: branchId } }); }
  authorize(businessId: string, id: string, value: {costo:number; margen:number; precio_venta:number}): Observable<unknown> { return this.http.patch(`${this.base(businessId)}/propuestas/${id}/autorizar`, value); }
}

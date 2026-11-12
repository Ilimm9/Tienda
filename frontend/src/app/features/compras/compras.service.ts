import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import {
  CatalogOption,
  CreatePurchaseRequest,
  EditPurchaseRequest,
  ExportFilter,
  PurchaseView,
  PackagingUnit,
  PurchaseListResponse,
  PurchaseProduct,
} from './compras.models';

@Injectable({ providedIn: 'root' })
export class ComprasService {
  private readonly http = inject(HttpClient);
  private base(id: string): string {
    return `${environment.apiUrl}/negocios/${id}`;
  }
  list(id: string, branchId: string): Observable<PurchaseListResponse> {
    return this.http.get<PurchaseListResponse>(`${this.base(id)}/compras`, {
      params: { sucursal_id: branchId },
    });
  }
  providers(id: string): Observable<CatalogOption[]> {
    return this.http.get<CatalogOption[]>(`${this.base(id)}/catalogo/proveedores`);
  }
  createProvider(id: string, request: { nombre: string }): Observable<void> {
    return this.http.post<void>(`${this.base(id)}/catalogo/proveedores`, request);
  }
  branches(id: string): Observable<CatalogOption[]> {
    return this.http.get<CatalogOption[]>(`${this.base(id)}/sucursales`);
  }
  units(id: string): Observable<PackagingUnit[]> {
    return this.http.get<PackagingUnit[]>(`${this.base(id)}/catalogo/unidades-medida`);
  }
  products(id: string, providerId: string): Observable<PurchaseProduct[]> {
    return this.http.get<PurchaseProduct[]>(`${this.base(id)}/compras/productos`, {
      params: { proveedor_id: providerId },
    });
  }
  create(id: string, request: CreatePurchaseRequest): Observable<unknown> {
    return this.http.post(`${this.base(id)}/compras`, request);
  }
  createUnit(
    id: string,
    request: {
      nombre: string;
      simbolo: string;
      tipo: string;
      factor_a_base: number;
      decimales: number;
    },
  ): Observable<void> {
    return this.http.post<void>(`${this.base(id)}/compras/unidades`, request);
  }
  view(id: string, purchaseId: string): Observable<PurchaseView> {
    return this.http.get<PurchaseView>(`${this.base(id)}/compras/${purchaseId}`);
  }
  edit(id: string, purchaseId: string, request: EditPurchaseRequest): Observable<unknown> {
    return this.http.patch(`${this.base(id)}/compras/${purchaseId}`, request);
  }
  suspend(id: string, purchaseId: string, motivo: string): Observable<unknown> {
    return this.http.post(`${this.base(id)}/compras/${purchaseId}/suspender`, { motivo });
  }
  reactivate(id: string, purchaseId: string, motivo: string): Observable<unknown> {
    return this.http.post(`${this.base(id)}/compras/${purchaseId}/reactivar`, { motivo });
  }
  /** Descarga el .xlsx de compras (hoja 1 y COMPROBACIÓN). */
  exportar(id: string, filtro: ExportFilter): Observable<Blob> {
    let params = new HttpParams().set('sucursal_id', filtro.sucursal_id).set('comprobacion', filtro.comprobacion).set('incluir_suspendidas', filtro.incluir_suspendidas);
    for (const compra of filtro.compra_id ?? []) params = params.append('compra_id', compra);
    for (const key of ['proveedor_id', 'desde', 'hasta'] as const) {
      const value = filtro[key];
      if (value) params = params.set(key, value);
    }
    return this.http.get(`${this.base(id)}/compras/exportar`, { params, responseType: 'blob' });
  }
}

/** Guarda un archivo descargado con un enlace temporal. */
export function guardarArchivo(archivo: Blob, nombre: string): void {
  const url = URL.createObjectURL(archivo);
  const enlace = document.createElement('a');
  enlace.href = url;
  enlace.download = nombre;
  enlace.click();
  URL.revokeObjectURL(url);
}

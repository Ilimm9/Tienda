import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import {
  CatalogOption,
  CreatePurchaseRequest,
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
  list(id: string): Observable<PurchaseListResponse> {
    return this.http.get<PurchaseListResponse>(`${this.base(id)}/compras`);
  }
  providers(id: string): Observable<CatalogOption[]> {
    return this.http.get<CatalogOption[]>(`${this.base(id)}/catalogo/proveedores`);
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
      codigo: string;
      nombre: string;
      simbolo: string;
      tipo: string;
      factor_a_base: number;
      decimales: number;
    },
  ): Observable<void> {
    return this.http.post<void>(`${this.base(id)}/compras/unidades`, request);
  }
}

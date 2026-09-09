import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { CatalogOption, CreateProductRequest, ProductListResponse, ProductLookup } from './product.models';

@Injectable({ providedIn: 'root' })
export class ProductosService {
  private readonly http = inject(HttpClient);

  listByBusiness(businessId: string): Observable<ProductListResponse> {
    return this.http.get<ProductListResponse>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos`,
    );
  }

  listCategories(businessId: string): Observable<CatalogOption[]> {
    return this.http.get<CatalogOption[]>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/categorias`,
    );
  }

  listBrands(businessId: string): Observable<CatalogOption[]> {
    return this.http.get<CatalogOption[]>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/marcas`,
    );
  }

  listBranches(businessId: string): Observable<CatalogOption[]> {
    return this.http.get<CatalogOption[]>(`${environment.apiUrl}/negocios/${businessId}/sucursales`);
  }

  lookupProduct(businessId: string, barcode: string): Observable<ProductLookup> {
    return this.http.get<ProductLookup>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos/consulta-codigo/${encodeURIComponent(barcode)}`,
    );
  }

  create(businessId: string, payload: CreateProductRequest): Observable<void> {
    return this.http.post<void>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos`,
      payload,
    );
  }
}

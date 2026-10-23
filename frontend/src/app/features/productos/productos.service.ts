import { HttpClient } from '@angular/common/http';
import { effect, inject, Injectable } from '@angular/core';
import { Observable, tap } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { environment } from '../../../environments/environment';
import {
  CatalogOption,
  CreateUnitRequest,
  CreateProductRequest,
  ProductImportJob,
  ProductImportPreview,
  ProductListResponse,
  ProductLookup,
  ProductRow,
  UnitOption,
  UpdateProductRequest,
  AdjustInventoryRequest,
  AddProductToBranchRequest,
  BranchCatalogProduct,
} from './product.models';

@Injectable({ providedIn: 'root' })
export class ProductosService {
  private readonly http = inject(HttpClient);
  private readonly contexto = inject(ContextoService);
  // Última lista recibida por negocio y sucursal; permite mostrarla al volver a la vista
  // mientras se refresca en segundo plano. Solo vive en memoria durante la sesión.
  private readonly listas = new Map<string, ProductRow[]>();

  constructor() {
    effect(() => {
      if (!this.contexto.negocio()) this.olvidarListas();
    });
  }

  listByBusiness(businessId: string, branchId?: string): Observable<ProductListResponse> {
    return this.http.get<ProductListResponse>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos`,
      branchId ? { params: { sucursal_id: branchId } } : undefined,
    ).pipe(tap(({ items }) => this.listas.set(this.claveLista(businessId, branchId), items ?? [])));
  }

  listaGuardada(businessId: string, branchId?: string): ProductRow[] | null {
    return this.listas.get(this.claveLista(businessId, branchId)) ?? null;
  }

  olvidarListas(): void {
    this.listas.clear();
  }

  private claveLista(businessId: string, branchId?: string): string {
    return `${businessId}:${branchId ?? ''}`;
  }

  private olvidarAlTerminar<T>(request: Observable<T>): Observable<T> {
    return request.pipe(tap(() => this.olvidarListas()));
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

  createBrand(businessId: string, nombre: string): Observable<void> {
    return this.http.post<void>(`${environment.apiUrl}/negocios/${businessId}/catalogo/marcas`, {
      nombre,
    });
  }

  createCategory(businessId: string, nombre: string): Observable<void> {
    return this.http.post<void>(`${environment.apiUrl}/negocios/${businessId}/catalogo/categorias`, { nombre });
  }

  createUnit(businessId: string, payload: CreateUnitRequest): Observable<void> {
    return this.http.post<void>(`${environment.apiUrl}/negocios/${businessId}/catalogo/unidades-medida`, payload);
  }

  listBranches(businessId: string): Observable<CatalogOption[]> {
    return this.http.get<CatalogOption[]>(
      `${environment.apiUrl}/negocios/${businessId}/sucursales`,
    );
  }
  listUnits(businessId: string): Observable<UnitOption[]> {
    return this.http.get<UnitOption[]>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/unidades-medida`,
    );
  }
  listProviders(businessId: string): Observable<CatalogOption[]> {
    return this.http.get<CatalogOption[]>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/proveedores`,
    );
  }

  lookupProduct(businessId: string, barcode: string): Observable<ProductLookup> {
    return this.http.get<ProductLookup>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos/consulta-codigo/${encodeURIComponent(barcode)}`,
    );
  }

  create(businessId: string, payload: CreateProductRequest): Observable<void> {
    return this.olvidarAlTerminar(this.http.post<void>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos`,
      payload,
    ));
  }

  update(businessId: string, productId: string, payload: UpdateProductRequest): Observable<void> {
    return this.olvidarAlTerminar(this.http.patch<void>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos/${productId}`,
      payload,
    ));
  }

  adjustInventory(businessId: string, productId: string, payload: AdjustInventoryRequest): Observable<void> {
    return this.olvidarAlTerminar(this.http.patch<void>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos/${productId}/inventario`,
      payload,
    ));
  }

  /** Productos del negocio que aún no están en la sucursal (solo propietario). */
  listNotInBranch(businessId: string, branchId: string): Observable<BranchCatalogProduct[]> {
    return this.http.get<BranchCatalogProduct[]>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos/fuera-de-sucursal`,
      { params: { sucursal_id: branchId } },
    );
  }

  /** Agrega productos del catálogo a la sucursal, sin precio ni stock. */
  addToBranch(businessId: string, branchId: string, productIds: string[] | 'todos'): Observable<{ agregados: number }> {
    const body = productIds === 'todos'
      ? { sucursal_id: branchId, todos: true }
      : { sucursal_id: branchId, producto_ids: productIds };
    return this.olvidarAlTerminar(this.http.post<{ agregados: number }>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos/agregar-a-sucursal`,
      body,
    ));
  }

  /** Agrega un producto simple existente con el precio, costo y stock del formulario. */
  addProductToBranch(businessId: string, productId: string, payload: AddProductToBranchRequest): Observable<void> {
    return this.olvidarAlTerminar(this.http.post<void>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos/${productId}/agregar-a-sucursal`,
      payload,
    ));
  }

  deactivate(businessId: string, productId: string): Observable<void> {
    return this.olvidarAlTerminar(this.http.delete<void>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos/${productId}`,
    ));
  }

  productImportTemplateUrl(businessId: string): string {
    return `${environment.apiUrl}/negocios/${businessId}/catalogo/productos/importacion/plantilla`;
  }

  importProducts(
    businessId: string,
    branchId: string,
    file: File,
    authorizeNewCatalogs = false,
  ): Observable<ProductImportJob> {
    const data = new FormData();
    data.append('archivo', file);
    data.append('sucursal_id', branchId);
    if (authorizeNewCatalogs) data.append('autorizar_catalogos_nuevos', 'true');
    return this.http.post<ProductImportJob>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos/importar`,
      data,
    );
  }

  productImportStatus(businessId: string, importId: string): Observable<ProductImportJob> {
    return this.http.get<ProductImportJob>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos/importaciones/${importId}`,
    ).pipe(tap((job) => { if (job.estado === 'completada') this.olvidarListas(); }));
  }

  previewProductImport(
    businessId: string,
    branchId: string,
    file: File,
  ): Observable<ProductImportPreview> {
    const data = new FormData();
    data.append('archivo', file);
    data.append('sucursal_id', branchId);
    return this.http.post<ProductImportPreview>(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos/validar-importacion`,
      data,
    );
  }

  downloadProductImportErrors(businessId: string, branchId: string, file: File): Observable<Blob> {
    const data = new FormData();
    data.append('archivo', file);
    data.append('sucursal_id', branchId);
    return this.http.post(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos/importacion/errores`,
      data,
      { responseType: 'blob' },
    );
  }
}

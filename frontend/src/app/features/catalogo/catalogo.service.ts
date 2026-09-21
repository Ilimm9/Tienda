import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { CatalogImportResult, Categoria, Marca, Proveedor, UnidadMedida } from './catalogo.models';

@Injectable({ providedIn: 'root' })
export class CatalogoService {
  private readonly http = inject(HttpClient);
  private readonly api = environment.apiUrl;
  marcas(negocioId: string): Observable<Marca[]> {
    return this.http.get<Marca[]>(`${this.api}/negocios/${negocioId}/catalogo/marcas`);
  }
  categorias(negocioId: string): Observable<Categoria[]> {
    return this.http.get<Categoria[]>(`${this.api}/negocios/${negocioId}/catalogo/categorias`);
  }
  proveedores(id: string): Observable<Proveedor[]> {
    return this.http.get<Proveedor[]>(`${this.api}/negocios/${id}/catalogo/proveedores`);
  }
  unidades(negocioId: string): Observable<UnidadMedida[]> {
    return this.http.get<UnidadMedida[]>(`${this.api}/negocios/${negocioId}/catalogo/unidades-medida`);
  }
  crear(negocioId: string, path: string, payload: object): Observable<void> {
    return this.http.post<void>(`${this.api}/negocios/${negocioId}/catalogo/${path}`, payload);
  }
  actualizar(negocioId: string, path: string, id: string, payload: object): Observable<void> {
    return this.http.patch<void>(`${this.api}/negocios/${negocioId}/catalogo/${path}/${id}`, payload);
  }
  importar(negocioId: string, section: 'marcas' | 'categorias' | 'unidades', file: File): Observable<CatalogImportResult> {
    const data = new FormData();
    data.append('archivo', file);
    return this.http.post<CatalogImportResult>(`${this.api}/negocios/${negocioId}/catalogo/${section}/importar`, data);
  }
  plantillaUrl(negocioId: string, section: 'marcas' | 'categorias' | 'unidades'): string {
    if (section === 'unidades') return `${this.api}/negocios/${negocioId}/catalogo/unidades/importacion/plantilla`;
    return `${this.api}/negocios/${negocioId}/catalogo/${section}/importacion/plantilla`;
  }
}

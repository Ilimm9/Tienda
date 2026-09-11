import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { CatalogImportResult, Categoria, Marca, Proveedor, UnidadMedida } from './catalogo.models';

@Injectable({ providedIn: 'root' })
export class CatalogoService {
  private readonly http = inject(HttpClient);
  private readonly api = environment.apiUrl;
  marcas(): Observable<Marca[]> {
    return this.http.get<Marca[]>(`${this.api}/catalogo/marcas`);
  }
  categorias(): Observable<Categoria[]> {
    return this.http.get<Categoria[]>(`${this.api}/catalogo/categorias`);
  }
  proveedores(id: string): Observable<Proveedor[]> {
    return this.http.get<Proveedor[]>(`${this.api}/negocios/${id}/catalogo/proveedores`);
  }
  unidades(): Observable<UnidadMedida[]> {
    return this.http.get<UnidadMedida[]>(`${this.api}/catalogo/unidades-medida`);
  }
  crear(path: string, payload: object): Observable<void> {
    return this.http.post<void>(`${this.api}/${path}`, payload);
  }
  actualizar(path: string, id: string, payload: object): Observable<void> {
    return this.http.patch<void>(`${this.api}/${path}/${id}`, payload);
  }
  importar(section: 'marcas' | 'categorias' | 'unidades', file: File): Observable<CatalogImportResult> {
    const data = new FormData();
    data.append('archivo', file);
    return this.http.post<CatalogImportResult>(`${this.api}/catalogo/${section}/importar`, data);
  }
  plantillaUrl(section: 'marcas' | 'categorias' | 'unidades'): string {
    if (section === 'unidades') return `${this.api}/catalogo/unidades/importacion/plantilla`;
    return `${this.api}/catalogo/${section}/importacion/plantilla`;
  }
}

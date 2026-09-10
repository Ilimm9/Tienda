import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Categoria, Marca, Proveedor } from './catalogo.models';

@Injectable({ providedIn: 'root' })
export class CatalogoService {
  private readonly http = inject(HttpClient);
  private readonly api = environment.apiUrl;
  marcas(): Observable<Marca[]> { return this.http.get<Marca[]>(`${this.api}/catalogo/marcas`); }
  categorias(): Observable<Categoria[]> { return this.http.get<Categoria[]>(`${this.api}/catalogo/categorias`); }
  proveedores(id: string): Observable<Proveedor[]> { return this.http.get<Proveedor[]>(`${this.api}/negocios/${id}/catalogo/proveedores`); }
  crear(path: string, payload: object): Observable<void> { return this.http.post<void>(`${this.api}/${path}`, payload); }
  actualizar(path: string, id: string, payload: object): Observable<void> { return this.http.patch<void>(`${this.api}/${path}/${id}`, payload); }
}

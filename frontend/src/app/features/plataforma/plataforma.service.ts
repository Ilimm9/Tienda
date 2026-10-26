import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  NegocioPlataformaDetalle,
  PaginaCuentas,
  PaginaNegocios,
  PlataformaResumen,
  UsuariosResponse,
} from './plataforma.models';

@Injectable({ providedIn: 'root' })
export class PlataformaService {
  private readonly http = inject(HttpClient);
  private readonly url = `${environment.apiUrl}/plataforma`;

  resumen(): Observable<PlataformaResumen> {
    return this.http.get<PlataformaResumen>(`${this.url}/resumen`);
  }

  negocios(q = '', estado = '', page = 1, pageSize = 20): Observable<PaginaNegocios> {
    let params = new HttpParams().set('page', page).set('page_size', pageSize);
    if (q.trim()) params = params.set('q', q.trim());
    if (estado) params = params.set('estado', estado);
    return this.http.get<PaginaNegocios>(`${this.url}/negocios`, { params });
  }

  cuentas(q = '', estado = '', page = 1, pageSize = 20): Observable<PaginaCuentas> {
    let params = new HttpParams().set('page', page).set('page_size', pageSize);
    if (q.trim()) params = params.set('q', q.trim());
    if (estado) params = params.set('estado', estado);
    return this.http.get<PaginaCuentas>(`${this.url}/cuentas`, { params });
  }

  negocio(id: string): Observable<NegocioPlataformaDetalle> {
    return this.http.get<NegocioPlataformaDetalle>(`${this.url}/negocios/${id}`);
  }

  usuarios(id: string): Observable<UsuariosResponse> {
    return this.http.get<UsuariosResponse>(`${this.url}/negocios/${id}/usuarios`);
  }
}

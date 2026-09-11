import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  ActualizarNegocioPayload,
  CrearNegocioPayload,
  NegocioDetalle,
  NegocioListResponse,
} from './negocio.models';

@Injectable({ providedIn: 'root' })
export class NegocioService {
  private readonly http = inject(HttpClient);
  private readonly url = `${environment.apiUrl}/negocios`;

  listar(estado: 'activo' | 'archivado' = 'activo'): Observable<NegocioListResponse> {
    return this.http.get<NegocioListResponse>(this.url, { params: { estado } });
  }

  crear(payload: CrearNegocioPayload): Observable<NegocioDetalle> {
    return this.http.post<NegocioDetalle>(this.url, payload);
  }

  obtener(id: string): Observable<NegocioDetalle> {
    return this.http.get<NegocioDetalle>(`${this.url}/${id}`);
  }

  actualizar(id: string, payload: ActualizarNegocioPayload): Observable<NegocioDetalle> {
    return this.http.patch<NegocioDetalle>(`${this.url}/${id}`, payload);
  }

  archivar(id: string): Observable<void> {
    return this.http.delete<void>(`${this.url}/${id}`);
  }

  restaurar(id: string): Observable<NegocioDetalle> {
    return this.http.post<NegocioDetalle>(`${this.url}/${id}/restaurar`, {});
  }
}

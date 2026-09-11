import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  ActualizarSucursalPayload,
  CrearSucursalPayload,
  SucursalDetalle,
  SucursalListResponse,
} from './sucursal.models';

@Injectable({ providedIn: 'root' })
export class SucursalService {
  private readonly http = inject(HttpClient);

  listar(
    negocioId: string,
    estado: 'activo' | 'archivado' = 'activo',
    buscar = '',
  ): Observable<SucursalListResponse> {
    let params = new HttpParams().set('estado', estado);
    if (buscar.trim()) params = params.set('buscar', buscar.trim());
    return this.http.get<SucursalListResponse>(this.baseUrl(negocioId), { params });
  }

  obtener(negocioId: string, sucursalId: string): Observable<SucursalDetalle> {
    return this.http.get<SucursalDetalle>(`${this.baseUrl(negocioId)}/${sucursalId}`);
  }

  crear(negocioId: string, payload: CrearSucursalPayload): Observable<SucursalDetalle> {
    return this.http.post<SucursalDetalle>(this.baseUrl(negocioId), payload);
  }

  actualizar(
    negocioId: string,
    sucursalId: string,
    payload: ActualizarSucursalPayload,
  ): Observable<SucursalDetalle> {
    return this.http.patch<SucursalDetalle>(`${this.baseUrl(negocioId)}/${sucursalId}`, payload);
  }

  archivar(negocioId: string, sucursalId: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl(negocioId)}/${sucursalId}`);
  }

  restaurar(negocioId: string, sucursalId: string): Observable<SucursalDetalle> {
    return this.http.post<SucursalDetalle>(
      `${this.baseUrl(negocioId)}/${sucursalId}/restaurar`,
      {},
    );
  }

  private baseUrl(negocioId: string): string {
    return `${environment.apiUrl}/negocios/${negocioId}/administracion/sucursales`;
  }
}

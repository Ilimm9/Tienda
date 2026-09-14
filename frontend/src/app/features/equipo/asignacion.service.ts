import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { AsignacionListResponse, AsignarSucursalPayload } from './asignacion.models';

@Injectable({ providedIn: 'root' })
export class AsignacionService {
  private readonly http = inject(HttpClient);

  listar(negocioId: string, empleadoId: string, incluirFinalizadas = false): Observable<AsignacionListResponse> {
    const params = incluirFinalizadas ? new HttpParams().set('estado', 'todos') : undefined;
    return this.http.get<AsignacionListResponse>(this.baseUrl(negocioId, empleadoId), { params });
  }

  asignar(
    negocioId: string,
    empleadoId: string,
    payload: AsignarSucursalPayload,
  ): Observable<AsignacionListResponse> {
    return this.http.post<AsignacionListResponse>(this.baseUrl(negocioId, empleadoId), payload);
  }

  establecerPrincipal(
    negocioId: string,
    empleadoId: string,
    asignacionId: string,
  ): Observable<AsignacionListResponse> {
    return this.http.post<AsignacionListResponse>(
      `${this.baseUrl(negocioId, empleadoId)}/${asignacionId}/principal`,
      {},
    );
  }

  finalizar(negocioId: string, empleadoId: string, asignacionId: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl(negocioId, empleadoId)}/${asignacionId}`);
  }

  private baseUrl(negocioId: string, empleadoId: string): string {
    return `${environment.apiUrl}/negocios/${negocioId}/administracion/empleados/${empleadoId}/sucursales`;
  }
}

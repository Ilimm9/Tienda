import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  ActualizarEmpleadoPayload,
  CrearEmpleadoPayload,
  EmpleadoDetalle,
  EmpleadoListResponse,
} from './empleado.models';

@Injectable({ providedIn: 'root' })
export class EmpleadoService {
  private readonly http = inject(HttpClient);

  listar(negocioId: string, estado = '', buscar = ''): Observable<EmpleadoListResponse> {
    let params = new HttpParams();
    if (estado) params = params.set('estado', estado);
    if (buscar.trim()) params = params.set('buscar', buscar.trim());
    return this.http.get<EmpleadoListResponse>(this.baseUrl(negocioId), { params });
  }

  obtener(negocioId: string, empleadoId: string): Observable<EmpleadoDetalle> {
    return this.http.get<EmpleadoDetalle>(`${this.baseUrl(negocioId)}/${empleadoId}`);
  }

  crear(negocioId: string, payload: CrearEmpleadoPayload): Observable<EmpleadoDetalle> {
    return this.http.post<EmpleadoDetalle>(this.baseUrl(negocioId), payload);
  }

  actualizar(
    negocioId: string,
    empleadoId: string,
    payload: ActualizarEmpleadoPayload,
  ): Observable<EmpleadoDetalle> {
    return this.http.patch<EmpleadoDetalle>(`${this.baseUrl(negocioId)}/${empleadoId}`, payload);
  }

  private baseUrl(negocioId: string): string {
    return `${environment.apiUrl}/negocios/${negocioId}/administracion/empleados`;
  }
}

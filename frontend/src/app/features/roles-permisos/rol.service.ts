import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  ActualizarRolPayload,
  CrearRolPayload,
  ListaRespuesta,
  MiembroRoles,
  Permiso,
  RolDetalle,
  RolResumen,
} from './rol.models';

@Injectable({ providedIn: 'root' })
export class RolService {
  private readonly http = inject(HttpClient);

  permisos(negocioId: string): Observable<ListaRespuesta<Permiso>> {
    return this.http.get<ListaRespuesta<Permiso>>(`${this.baseUrl(negocioId)}/permisos`);
  }

  misPermisos(negocioId: string): Observable<ListaRespuesta<string>> {
    return this.http.get<ListaRespuesta<string>>(`${this.baseUrl(negocioId)}/mis-permisos`);
  }

  listar(negocioId: string, incluirInactivos = false): Observable<ListaRespuesta<RolResumen>> {
    const params = incluirInactivos ? new HttpParams().set('estado', 'todos') : undefined;
    return this.http.get<ListaRespuesta<RolResumen>>(`${this.baseUrl(negocioId)}/roles`, { params });
  }

  obtener(negocioId: string, rolId: string): Observable<RolDetalle> {
    return this.http.get<RolDetalle>(`${this.baseUrl(negocioId)}/roles/${rolId}`);
  }

  crear(negocioId: string, payload: CrearRolPayload): Observable<RolDetalle> {
    return this.http.post<RolDetalle>(`${this.baseUrl(negocioId)}/roles`, payload);
  }

  actualizar(negocioId: string, rolId: string, payload: ActualizarRolPayload): Observable<RolDetalle> {
    return this.http.patch<RolDetalle>(`${this.baseUrl(negocioId)}/roles/${rolId}`, payload);
  }

  eliminar(negocioId: string, rolId: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl(negocioId)}/roles/${rolId}`);
  }

  miembros(negocioId: string): Observable<ListaRespuesta<MiembroRoles>> {
    return this.http.get<ListaRespuesta<MiembroRoles>>(`${this.baseUrl(negocioId)}/miembros`);
  }

  asignarRoles(negocioId: string, membresiaId: string, roles: string[]): Observable<void> {
    return this.http.put<void>(`${this.baseUrl(negocioId)}/miembros/${membresiaId}/roles`, { roles });
  }

  private baseUrl(negocioId: string): string {
    return `${environment.apiUrl}/negocios/${negocioId}/administracion`;
  }
}

import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  CrearInvitacionPayload,
  InvitacionCreada,
  InvitacionListResponse,
  InvitacionPublica,
} from './invitacion.models';

@Injectable({ providedIn: 'root' })
export class InvitacionService {
  private readonly http = inject(HttpClient);

  listar(negocioId: string, estado = ''): Observable<InvitacionListResponse> {
    const params = estado ? new HttpParams().set('estado', estado) : undefined;
    return this.http.get<InvitacionListResponse>(this.baseUrl(negocioId), { params });
  }

  crear(negocioId: string, payload: CrearInvitacionPayload): Observable<InvitacionCreada> {
    return this.http.post<InvitacionCreada>(this.baseUrl(negocioId), payload);
  }

  cancelar(negocioId: string, invitacionId: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl(negocioId)}/${invitacionId}`);
  }

  /** Consulta pública del enlace: quien lo abre puede no tener cuenta todavía. */
  consultar(token: string): Observable<InvitacionPublica> {
    return this.http.get<InvitacionPublica>(`${environment.apiUrl}/invitaciones/${token}`);
  }

  aceptar(token: string): Observable<void> {
    return this.http.post<void>(`${environment.apiUrl}/invitaciones/${token}/aceptar`, {});
  }

  /** Construye el enlace copiable que se entrega al invitado. */
  enlaceDeToken(token: string): string {
    return `${window.location.origin}/invitacion/${token}`;
  }

  private baseUrl(negocioId: string): string {
    return `${environment.apiUrl}/negocios/${negocioId}/administracion/invitaciones`;
  }
}

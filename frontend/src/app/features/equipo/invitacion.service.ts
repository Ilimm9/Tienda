import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  CrearInvitacionPayload,
  DesafioRegistro,
  FiltroInvitaciones,
  InvitacionAceptada,
  InvitacionCreada,
  InvitacionListResponse,
  InvitacionPublica,
  RegistroInvitacionPayload,
} from './invitacion.models';

@Injectable({ providedIn: 'root' })
export class InvitacionService {
  private readonly http = inject(HttpClient);

  listar(negocioId: string, filtro: FiltroInvitaciones = {}): Observable<InvitacionListResponse> {
    let params = new HttpParams();
    if (filtro.sinAceptar) params = params.set('sin_aceptar', 'true');
    else if (filtro.estado) params = params.set('estado', filtro.estado);
    if (filtro.sucursalId) params = params.set('sucursal_id', filtro.sucursalId);
    return this.http.get<InvitacionListResponse>(this.baseUrl(negocioId), { params });
  }

  crear(negocioId: string, payload: CrearInvitacionPayload): Observable<InvitacionCreada> {
    return this.http.post<InvitacionCreada>(this.baseUrl(negocioId), payload);
  }

  /** Emite un enlace nuevo con 72 horas; con `correo` corrige además al destinatario. */
  reenviar(negocioId: string, invitacionId: string, correo?: string): Observable<InvitacionCreada> {
    return this.http.post<InvitacionCreada>(
      `${this.baseUrl(negocioId)}/${invitacionId}/reenviar`,
      correo ? { correo } : {},
    );
  }

  cancelar(negocioId: string, invitacionId: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl(negocioId)}/${invitacionId}`);
  }

  /** Consulta pública del enlace: quien lo abre puede no tener cuenta todavía. */
  consultar(token: string): Observable<InvitacionPublica> {
    return this.http.get<InvitacionPublica>(this.urlPublica(token));
  }

  /** Crea la cuenta pendiente del invitado; no consume la invitación. */
  registrar(token: string, payload: RegistroInvitacionPayload): Observable<DesafioRegistro> {
    return this.http.post<DesafioRegistro>(`${this.urlPublica(token)}/registro`, payload);
  }

  aceptar(token: string): Observable<InvitacionAceptada> {
    return this.http.post<InvitacionAceptada>(`${this.urlPublica(token)}/aceptar`, {});
  }

  /** Construye el enlace copiable que se entrega al invitado. */
  enlaceDeToken(token: string): string {
    return `${window.location.origin}/invitacion/${token}`;
  }

  private urlPublica(token: string): string {
    return `${environment.apiUrl}/invitaciones/${encodeURIComponent(token)}`;
  }

  private baseUrl(negocioId: string): string {
    return `${environment.apiUrl}/negocios/${negocioId}/administracion/invitaciones`;
  }
}

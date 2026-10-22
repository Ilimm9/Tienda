import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { ActualizarPuestoPayload, CrearPuestoPayload, Puesto, PuestoListResponse } from './puesto.models';

@Injectable({ providedIn: 'root' })
export class PuestoService {
  private readonly http = inject(HttpClient);

  listar(negocioId: string): Observable<PuestoListResponse> {
    return this.http.get<PuestoListResponse>(this.baseUrl(negocioId));
  }

  crear(negocioId: string, payload: CrearPuestoPayload): Observable<Puesto> {
    return this.http.post<Puesto>(this.baseUrl(negocioId), payload);
  }

  actualizar(negocioId: string, puestoId: string, payload: ActualizarPuestoPayload): Observable<Puesto> {
    return this.http.patch<Puesto>(`${this.baseUrl(negocioId)}/${puestoId}`, payload);
  }

  private baseUrl(negocioId: string): string {
    return `${environment.apiUrl}/negocios/${negocioId}/administracion/puestos`;
  }
}

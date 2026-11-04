import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject, Injector } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';

import { environment } from '../../../environments/environment';
import { AccesoDenegadoService } from '../../contexto/acceso-denegado.service';
import { AuthService } from './auth.service';

// Respuestas 401 que cada pantalla maneja por su cuenta: login, consulta de sesión e invitación pública.
const rutasCon401Propio = [`${environment.apiUrl}/auth/`, `${environment.apiUrl}/invitaciones/`];

export const authInterceptor: HttpInterceptorFn = (request, next) => {
  // Los servicios dependen de HttpClient: se resuelven de forma perezosa para no crear un ciclo.
  const injector = inject(Injector);
  return next(request.clone({ withCredentials: true })).pipe(
    catchError((error: unknown) => {
      if (error instanceof HttpErrorResponse && error.status === 403 && error.error?.codigo === 'ACCESO_DENEGADO') {
        injector.get(AccesoDenegadoService).atender();
      }
      if (
        error instanceof HttpErrorResponse &&
        error.status === 401 &&
        !rutasCon401Propio.some((ruta) => request.url.startsWith(ruta))
      ) {
        // La sesión se guarda en memoria entre navegaciones; si venció en el servidor, se vuelve al login.
        const auth = injector.get(AuthService);
        if (auth.currentUser()) {
          auth.sesionTerminada();
          void injector.get(Router).navigate(['/login']);
        }
      }
      return throwError(() => error);
    }),
  );
};

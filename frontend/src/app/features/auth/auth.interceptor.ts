import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject, Injector } from '@angular/core';
import { catchError, throwError } from 'rxjs';

import { AccesoDenegadoService } from '../../contexto/acceso-denegado.service';

export const authInterceptor: HttpInterceptorFn = (request, next) => {
  // El servicio depende de HttpClient: se resuelve de forma perezosa para no crear un ciclo.
  const injector = inject(Injector);
  return next(request.clone({ withCredentials: true })).pipe(
    catchError((error: unknown) => {
      if (error instanceof HttpErrorResponse && error.status === 403 && error.error?.codigo === 'ACCESO_DENEGADO') {
        injector.get(AccesoDenegadoService).atender();
      }
      return throwError(() => error);
    }),
  );
};

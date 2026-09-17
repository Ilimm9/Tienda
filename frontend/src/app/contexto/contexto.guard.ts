import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { map } from 'rxjs';

import { ContextoService } from './contexto.service';

export const contextoGuard: CanActivateFn = () => {
  const contexto = inject(ContextoService);
  const router = inject(Router);
  return contexto.asegurarInicializado().pipe(
    map((estado) => estado === 'listo' || estado === 'sin_sucursal'
      ? true
      : router.createUrlTree([estado === 'error' ? '/negocios' : '/seleccionar-negocio'])),
  );
};

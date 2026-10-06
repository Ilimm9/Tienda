import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { map } from 'rxjs';

import { ContextoService } from './contexto.service';

const RUTA_ASISTENTE = '/configuracion-inicial';

/**
 * Guard de Inicio. Ofrece el asistente a quien le falta empresa o sucursal, salvo que lo haya
 * pospuesto en esta sesión; entonces Inicio se muestra aunque todavía no exista contexto.
 */
export const altaInicialGuard: CanActivateFn = () => {
  const contexto = inject(ContextoService);
  const router = inject(Router);
  return contexto.asegurarInicializado().pipe(
    map((estado) => {
      const alta = contexto.estadoAlta();
      const pendiente = alta === 'requiere_empresa' || alta === 'requiere_sucursal';
      if (pendiente && !contexto.altaPospuesta()) return router.createUrlTree([RUTA_ASISTENTE]);
      if (alta === 'requiere_empresa') return true;
      if (estado === 'listo' || estado === 'sin_sucursal') return true;
      return router.createUrlTree([estado === 'error' ? '/negocios' : '/seleccionar-negocio']);
    }),
  );
};

/** El asistente solo sirve a quien puede completar el alta; los demás vuelven a Inicio. */
export const asistenteGuard: CanActivateFn = () => {
  const contexto = inject(ContextoService);
  const router = inject(Router);
  return contexto.asegurarInicializado().pipe(
    map(() => {
      const alta = contexto.estadoAlta();
      return alta === 'requiere_empresa' || alta === 'requiere_sucursal' ? true : router.createUrlTree(['/inicio']);
    }),
  );
};

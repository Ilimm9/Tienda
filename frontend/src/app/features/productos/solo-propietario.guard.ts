import { inject } from '@angular/core';
import { CanActivateFn, RedirectCommand, Router } from '@angular/router';
import { map } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { nombreDeSeccion } from '../../contexto/permiso.guard';

/**
 * Crear e importar productos es solo del propietario del negocio activo (capability 31).
 * Es una ayuda de navegación: la API responde `SOLO_PROPIETARIOS` a los demás.
 */
export const soloPropietarioGuard: CanActivateFn = (route) => {
  const contexto = inject(ContextoService);
  const router = inject(Router);
  return contexto.asegurarInicializado().pipe(
    map((estado) => {
      if (estado === 'error') return true;
      if (contexto.negocio()?.tipo_miembro === 'propietario') return true;
      return new RedirectCommand(
        router.createUrlTree(['/sin-acceso'], { queryParams: { seccion: nombreDeSeccion(route) } }),
        { skipLocationChange: true },
      );
    }),
  );
};

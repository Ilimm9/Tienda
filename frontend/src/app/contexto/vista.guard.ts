import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { map, switchMap } from 'rxjs';

import { ContextoService } from './contexto.service';
import { SeccionOcultable, VistasService } from './vistas.service';

/**
 * Una sección oculta en Configuración > Vistas (`data.vista`) no se abre: lleva a Inicio.
 * Es una ayuda de navegación; la API sigue respondiendo.
 */
export const vistaGuard: CanActivateFn = (route) => {
  const contexto = inject(ContextoService);
  const vistas = inject(VistasService);
  const router = inject(Router);
  const seccion = route.data?.['vista'] as SeccionOcultable | undefined;
  return contexto.asegurarInicializado().pipe(
    switchMap(() => vistas.asegurar()),
    map(() => (vistas.visible(seccion) ? true : router.createUrlTree(['/inicio']))),
  );
};

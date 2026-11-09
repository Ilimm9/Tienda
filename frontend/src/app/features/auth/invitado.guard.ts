import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { catchError, map, of } from 'rxjs';

import { AuthService, destinoTrasSesion } from './auth.service';

/**
 * Pantallas solo para visitantes (login, registro, recuperar contraseña): con sesión activa,
 * en esta u otra pestaña, se vuelve a la app hasta cerrar sesión.
 */
export const invitadoGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  return auth.asegurarSesion().pipe(
    map((usuario) => router.createUrlTree([destinoTrasSesion(usuario)])),
    catchError(() => of(true)),
  );
};

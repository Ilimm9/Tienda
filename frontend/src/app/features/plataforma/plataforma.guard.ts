import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { catchError, map, of } from 'rxjs';

import { AuthService } from '../auth/auth.service';

export const plataformaGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);
  return auth.me().pipe(
    map(({ usuario }) =>
      usuario.roles_plataforma?.includes('ADMIN_LECTURA')
        ? true
        : router.createUrlTree(['/sin-acceso']),
    ),
    catchError(() => of(router.createUrlTree(['/login']))),
  );
};

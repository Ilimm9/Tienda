import { inject } from '@angular/core';
import { ActivatedRouteSnapshot, CanActivateChildFn, CanActivateFn, RedirectCommand, Router } from '@angular/router';
import { map } from 'rxjs';

import { ContextoService } from './contexto.service';
import { PermisoRequerido } from './permisos';

/**
 * Exige los permisos declarados en `data.permiso` de la ruta o de sus ancestros.
 * Sin permiso muestra «Sin acceso» conservando la URL. Es una ayuda de navegación: la API sigue autorizando.
 */
export const permisoGuard: CanActivateFn = (route) => {
  const contexto = inject(ContextoService);
  const router = inject(Router);
  return contexto.asegurarInicializado().pipe(
    map((estado) => {
      const requeridos = permisosDeRuta(route);
      if (!requeridos.length || estado === 'error') return true;
      // Las pantallas de un negocio concreto (`/negocios/:negocioId/...`) se evalúan contra ese negocio.
      const negocioId = negocioDeRuta(route);
      if (!negocioId && estado !== 'listo' && estado !== 'sin_sucursal') return true;
      const permitido = requeridos.every((permiso) =>
        negocioId ? contexto.puedeEn(negocioId, permiso, true) : contexto.puede(permiso),
      );
      if (permitido) return true;
      return new RedirectCommand(
        router.createUrlTree(['/sin-acceso'], { queryParams: { seccion: nombreDeSeccion(route) } }),
        { skipLocationChange: true },
      );
    }),
  );
};

/** Misma regla para `canActivateChild`, que recibe la ruta hija. */
export const permisoChildGuard: CanActivateChildFn = (child, state) => permisoGuard(child, state);

export function negocioDeRuta(route: ActivatedRouteSnapshot): string | null {
  for (let actual: ActivatedRouteSnapshot | null = route; actual; actual = actual.parent) {
    const id = actual.paramMap.get('negocioId');
    if (id) return id;
  }
  return null;
}

/** Permisos declarados en la ruta y en todos sus ancestros. */
export function permisosDeRuta(route: ActivatedRouteSnapshot): PermisoRequerido[] {
  const permisos: PermisoRequerido[] = [];
  for (let actual: ActivatedRouteSnapshot | null = route; actual; actual = actual.parent) {
    // `data` se hereda en rutas sin componente: se toma solo lo declarado en cada nivel.
    const propio = actual.routeConfig?.data?.['permiso'] as PermisoRequerido | undefined;
    if (propio) permisos.push(propio);
  }
  return permisos;
}

export function nombreDeSeccion(route: ActivatedRouteSnapshot): string | null {
  for (let actual: ActivatedRouteSnapshot | null = route; actual; actual = actual.parent) {
    const nombre = actual.routeConfig?.data?.['breadcrumb'];
    if (typeof nombre === 'string' && nombre) return nombre;
  }
  return null;
}

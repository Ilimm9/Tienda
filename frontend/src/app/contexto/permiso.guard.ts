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
      const sinAcceso = () =>
        new RedirectCommand(
          router.createUrlTree(['/sin-acceso'], { queryParams: { seccion: nombreDeSeccion(route) } }),
          { skipLocationChange: true },
        );
      if (estado === 'error') return true;
      // Reglas de cuenta, no de un negocio: ver la sección Negocios y registrar uno nuevo.
      const acceso = accesosDeRuta(route);
      if (acceso.includes('crear-negocio') && !contexto.puedeCrearNegocio()) return sinAcceso();
      if (acceso.includes('ver-negocios') && !contexto.puedeVerNegocios()) return sinAcceso();
      const requeridos = permisosDeRuta(route);
      if (!requeridos.length) return true;
      // Las pantallas de un negocio concreto (`/negocios/:negocioId/...`) se evalúan contra ese negocio.
      const negocioId = negocioDeRuta(route);
      if (!negocioId && estado !== 'listo' && estado !== 'sin_sucursal') return true;
      const permitido = requeridos.every((permiso) =>
        negocioId ? contexto.puedeEn(negocioId, permiso, true) : contexto.puede(permiso),
      );
      return permitido ? true : sinAcceso();
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

export type AccesoDeCuenta = 'ver-negocios' | 'crear-negocio';

/** Reglas `data.acceso` declaradas en la ruta y en sus ancestros. */
export function accesosDeRuta(route: ActivatedRouteSnapshot): AccesoDeCuenta[] {
  const accesos: AccesoDeCuenta[] = [];
  for (let actual: ActivatedRouteSnapshot | null = route; actual; actual = actual.parent) {
    const propio = actual.routeConfig?.data?.['acceso'] as AccesoDeCuenta | undefined;
    if (propio) accesos.push(propio);
  }
  return accesos;
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

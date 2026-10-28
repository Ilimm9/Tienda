import { inject } from '@angular/core';
import { CanActivateFn, RedirectCommand, Router, Routes } from '@angular/router';
import { map } from 'rxjs';
import { ContextoService } from '../../contexto/contexto.service';
import { permisoGuard } from '../../contexto/permiso.guard';
import { PERMISOS } from '../../contexto/permisos';

/** El historial y los tickets se abren con `ventas.ver` o con `ventas.registrar` (solo ventas propias). */
export const ventasLecturaGuard: CanActivateFn = () => {
  const contexto = inject(ContextoService);
  const router = inject(Router);
  return contexto.asegurarInicializado().pipe(
    map((estado) => {
      if (estado !== 'listo' || contexto.puedeAlguno(PERMISOS.ventaVer, PERMISOS.ventaRegistrar))
        return true;
      return new RedirectCommand(
        router.createUrlTree(['/sin-acceso'], { queryParams: { seccion: 'Ventas' } }),
        {
          skipLocationChange: true,
        },
      );
    }),
  );
};

export const VENTAS_ROUTES: Routes = [
  {
    path: '',
    canActivate: [ventasLecturaGuard],
    loadComponent: () => import('./ventas.component').then((module) => module.VentasComponent),
  },
  {
    path: 'nueva',
    canActivate: [permisoGuard],
    data: { breadcrumb: 'Nueva venta', permiso: PERMISOS.ventaRegistrar },
    loadComponent: () =>
      import('./nueva-venta.component').then((module) => module.NuevaVentaComponent),
  },
  {
    path: ':ventaId',
    canActivate: [ventasLecturaGuard],
    data: { breadcrumb: 'Ticket' },
    loadComponent: () =>
      import('./ticket-venta.component').then((module) => module.TicketVentaComponent),
  },
];

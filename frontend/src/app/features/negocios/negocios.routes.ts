import { Routes } from '@angular/router';

import { permisoChildGuard } from '../../contexto/permiso.guard';
import { PERMISOS } from '../../contexto/permisos';

export const NEGOCIOS_ROUTES: Routes = [
  {
    path: '',
    canActivateChild: [permisoChildGuard],
    data: { breadcrumb: 'Negocios' },
    children: [
      {
        path: '',
        pathMatch: 'full',
        data: { title: 'Mis negocios', acceso: 'ver-negocios' },
        loadComponent: () =>
          import('./negocios.component').then((module) => module.NegociosComponent),
      },
      {
        path: 'nuevo',
        data: { breadcrumb: 'Registrar negocio', title: 'Registrar nuevo negocio', acceso: 'crear-negocio' },
        loadComponent: () =>
          import('./negocio-form.component').then((module) => module.NegocioFormComponent),
      },
      {
        path: ':negocioId/editar',
        data: { breadcrumb: 'Editar', title: 'Editar negocio', permiso: PERMISOS.negocioEditar },
        loadComponent: () =>
          import('./negocio-form.component').then((module) => module.NegocioFormComponent),
      },
      {
        path: ':negocioId/sucursales',
        data: { breadcrumb: 'Sucursales', title: 'Sucursales', permiso: PERMISOS.sucursalVer },
        children: [
          {
            path: '',
            pathMatch: 'full',
            loadComponent: () =>
              import('../sucursales/sucursales.component').then(
                (module) => module.SucursalesComponent,
              ),
          },
          {
            path: 'nueva',
            data: { breadcrumb: 'Nueva', title: 'Nueva sucursal', permiso: PERMISOS.sucursalCrear },
            loadComponent: () =>
              import('../sucursales/sucursal-form.component').then(
                (module) => module.SucursalFormComponent,
              ),
          },
          {
            path: ':sucursalId/editar',
            data: { breadcrumb: 'Editar', title: 'Editar sucursal', permiso: PERMISOS.sucursalEditar },
            loadComponent: () =>
              import('../sucursales/sucursal-form.component').then(
                (module) => module.SucursalFormComponent,
              ),
          },
          {
            path: ':sucursalId',
            data: { breadcrumb: 'Detalle', title: 'Detalle de sucursal' },
            loadComponent: () =>
              import('../sucursales/sucursal-detalle.component').then(
                (module) => module.SucursalDetalleComponent,
              ),
          },
        ],
      },
      {
        path: ':negocioId',
        data: { breadcrumb: 'Detalle', title: 'Datos del negocio', permiso: PERMISOS.negocioVer },
        loadComponent: () =>
          import('./negocio-detalle.component').then((module) => module.NegocioDetalleComponent),
      },
    ],
  },
];

import { Routes } from '@angular/router';

import { contextoGuard } from '../../contexto/contexto.guard';
import { permisoChildGuard } from '../../contexto/permiso.guard';
import { PERMISOS } from '../../contexto/permisos';

export const ROLES_PERMISOS_ROUTES: Routes = [
  {
    path: '',
    canActivateChild: [contextoGuard, permisoChildGuard],
    data: { breadcrumb: 'Roles y permisos', permiso: PERMISOS.rolVer },
    children: [
      {
        path: '',
        data: { breadcrumb: 'Roles y permisos' },
        loadComponent: () =>
          import('./roles.component').then((module) => module.RolesComponent),
      },
      {
        path: 'nuevo',
        data: { breadcrumb: 'Nuevo rol', permiso: PERMISOS.rolGestionar },
        loadComponent: () =>
          import('./rol-form.component').then((module) => module.RolFormComponent),
      },
      {
        path: 'miembros/:membresiaId',
        // Abre la ventana «Asignar roles» encima de la lista.
        data: { breadcrumb: 'Asignar roles', permiso: PERMISOS.rolAsignar },
        loadComponent: () =>
          import('./roles.component').then((module) => module.RolesComponent),
      },
      {
        path: ':rolId',
        data: { breadcrumb: 'Editar rol', permiso: PERMISOS.rolGestionar },
        loadComponent: () =>
          import('./rol-form.component').then((module) => module.RolFormComponent),
      },
    ],
  },
];

import { Routes } from '@angular/router';

import { contextoGuard } from '../../contexto/contexto.guard';

export const ROLES_PERMISOS_ROUTES: Routes = [
  {
    path: '',
    canActivateChild: [contextoGuard],
    data: { breadcrumb: 'Roles y permisos' },
    children: [
      {
        path: '',
        data: { breadcrumb: 'Roles y permisos' },
        loadComponent: () =>
          import('./roles.component').then((module) => module.RolesComponent),
      },
      {
        path: 'nuevo',
        data: { breadcrumb: 'Nuevo rol' },
        loadComponent: () =>
          import('./rol-form.component').then((module) => module.RolFormComponent),
      },
      {
        path: 'miembros/:membresiaId',
        data: { breadcrumb: 'Asignar roles' },
        loadComponent: () =>
          import('./miembro-roles.component').then((module) => module.MiembroRolesComponent),
      },
      {
        path: ':rolId',
        data: { breadcrumb: 'Editar rol' },
        loadComponent: () =>
          import('./rol-form.component').then((module) => module.RolFormComponent),
      },
    ],
  },
];

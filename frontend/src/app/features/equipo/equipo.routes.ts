import { Routes } from '@angular/router';

import { contextoGuard } from '../../contexto/contexto.guard';

export const EQUIPO_ROUTES: Routes = [
  {
    path: '',
    canActivateChild: [contextoGuard],
    data: { breadcrumb: 'Equipo' },
    children: [
      { path: '', redirectTo: 'empleados', pathMatch: 'full' },
      {
        path: 'empleados',
        data: { breadcrumb: 'Empleados' },
        children: [
          {
            path: '',
            loadComponent: () =>
              import('./empleados.component').then((module) => module.EmpleadosComponent),
          },
          {
            path: 'nuevo',
            data: { breadcrumb: 'Nuevo empleado' },
            loadComponent: () =>
              import('./empleado-form.component').then((module) => module.EmpleadoFormComponent),
          },
          {
            path: ':empleadoId',
            data: { breadcrumb: 'Detalle' },
            loadComponent: () =>
              import('./empleado-detalle.component').then(
                (module) => module.EmpleadoDetalleComponent,
              ),
          },
          {
            path: ':empleadoId/sucursales',
            data: { breadcrumb: 'Sucursales asignadas' },
            loadComponent: () =>
              import('./empleado-sucursales.component').then(
                (module) => module.EmpleadoSucursalesComponent,
              ),
          },
          {
            path: ':empleadoId/editar',
            data: { breadcrumb: 'Editar empleado' },
            loadComponent: () =>
              import('./empleado-form.component').then((module) => module.EmpleadoFormComponent),
          },
        ],
      },
      {
        path: 'invitaciones',
        data: { breadcrumb: 'Invitaciones' },
        loadComponent: () =>
          import('./invitaciones.component').then((module) => module.InvitacionesComponent),
      },
    ],
  },
];

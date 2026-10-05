import { Routes } from '@angular/router';

import { contextoGuard } from '../../contexto/contexto.guard';
import { permisoChildGuard } from '../../contexto/permiso.guard';
import { PERMISOS } from '../../contexto/permisos';

export const EQUIPO_ROUTES: Routes = [
  {
    path: '',
    canActivateChild: [contextoGuard, permisoChildGuard],
    data: { breadcrumb: 'Equipo' },
    children: [
      { path: '', redirectTo: 'empleados', pathMatch: 'full' },
      {
        path: 'empleados',
        data: { breadcrumb: 'Empleados', permiso: PERMISOS.empleadoVer },
        children: [
          {
            path: '',
            loadComponent: () =>
              import('./empleados.component').then((module) => module.EmpleadosComponent),
          },
          {
            path: 'nuevo',
            data: { breadcrumb: 'Nuevo empleado', permiso: PERMISOS.empleadoGestionar },
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
            data: { breadcrumb: 'Sucursales asignadas', permiso: PERMISOS.asignacionVer },
            loadComponent: () =>
              import('./empleado-sucursales.component').then(
                (module) => module.EmpleadoSucursalesComponent,
              ),
          },
          {
            path: ':empleadoId/editar',
            data: { breadcrumb: 'Editar empleado', permiso: PERMISOS.empleadoGestionar },
            loadComponent: () =>
              import('./empleado-form.component').then((module) => module.EmpleadoFormComponent),
          },
        ],
      },
      {
        path: 'invitaciones',
        data: { breadcrumb: 'Invitaciones', permiso: PERMISOS.invitacionVer },
        children: [
          {
            path: '',
            loadComponent: () =>
              import('./invitaciones/invitaciones').then((module) => module.Invitaciones),
          },
          {
            path: 'nueva',
            data: {
              breadcrumb: 'Nueva invitación',
              permiso: [PERMISOS.invitacionEnviar, PERMISOS.empleadoVer, PERMISOS.rolVer],
            },
            loadComponent: () =>
              import('./invitaciones/invitacion-nueva').then((module) => module.InvitacionNueva),
          },
          {
            path: ':invitacionId/corregir-correo',
            data: {
              breadcrumb: 'Corregir correo',
              permiso: [PERMISOS.invitacionEnviar, PERMISOS.empleadoGestionar],
            },
            loadComponent: () =>
              import('./invitaciones/corregir-correo').then((module) => module.CorregirCorreo),
          },
        ],
      },
    ],
  },
];

import { Routes } from '@angular/router';

import { authGuard } from './features/auth/auth.guard';
import { contextoGuard } from './contexto/contexto.guard';

export const routes: Routes = [
  { path: '', redirectTo: 'inicio', pathMatch: 'full' },
  {
    path: 'login',
    loadComponent: () =>
      import('./features/auth/login/login.component').then((module) => module.LoginComponent),
  },
  {
    path: 'registro',
    loadComponent: () =>
      import('./features/auth/register/register.component').then(
        (module) => module.RegisterComponent,
      ),
  },
  {
    path: 'verificar-correo',
    loadComponent: () =>
      import('./features/auth/verification/verification.component').then(
        (module) => module.VerificationComponent,
      ),
  },
  {
    // La aceptación de invitación vive fuera del shell autenticado: el invitado puede no tener cuenta.
    path: 'invitacion/:token',
    loadComponent: () =>
      import('./features/invitacion/aceptar-invitacion.component').then(
        (module) => module.AceptarInvitacionComponent,
      ),
  },
  {
    path: '',
    canActivateChild: [authGuard],
    loadComponent: () =>
      import('./layout/shell/app-shell.component').then((module) => module.AppShellComponent),
    children: [
      {
        path: 'inicio',
        canActivate: [contextoGuard],
        data: { breadcrumb: 'Inicio' },
        loadComponent: () =>
          import('./features/home/home.component').then((module) => module.HomeComponent),
      },
      {
        path: 'negocios',
        loadChildren: () =>
          import('./features/negocios/negocios.routes').then((module) => module.NEGOCIOS_ROUTES),
      },
      {
        path: 'seleccionar-negocio',
        loadComponent: () =>
          import('./contexto/seleccionar-negocio.component').then(
            (module) => module.SeleccionarNegocioComponent,
          ),
      },
      {
        path: 'sucursales',
        canActivate: [contextoGuard],
        loadChildren: () =>
          import('./features/sucursales/sucursales.routes').then(
            (module) => module.SUCURSALES_ROUTES,
          ),
      },
      {
        path: 'ventas',
        loadChildren: () =>
          import('./features/ventas/ventas.routes').then((module) => module.VENTAS_ROUTES),
      },
      {
        path: 'productos',
        redirectTo: 'catalogo/productos',
        pathMatch: 'full',
      },
      {
        path: 'catalogo',
        loadChildren: () =>
          import('./features/catalogo/catalogo.routes').then((module) => module.CATALOGO_ROUTES),
      },
      {
        path: 'equipo',
        loadChildren: () =>
          import('./features/equipo/equipo.routes').then((module) => module.EQUIPO_ROUTES),
      },
      {
        path: 'roles-permisos',
        loadChildren: () =>
          import('./features/roles-permisos/roles-permisos.routes').then(
            (module) => module.ROLES_PERMISOS_ROUTES,
          ),
      },
    ],
  },
  { path: '**', redirectTo: 'inicio' },
];

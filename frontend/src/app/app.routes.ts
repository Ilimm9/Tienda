import { Routes } from '@angular/router';

import { authGuard } from './features/auth/auth.guard';

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
    path: '',
    canActivateChild: [authGuard],
    loadComponent: () =>
      import('./layout/shell/app-shell.component').then((module) => module.AppShellComponent),
    children: [
      {
        path: 'inicio',
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
        path: 'sucursales',
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

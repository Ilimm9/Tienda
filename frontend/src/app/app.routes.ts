import { Routes } from '@angular/router';

import { authGuard } from './features/auth/auth.guard';

export const routes: Routes = [
  { path: '', redirectTo: 'login', pathMatch: 'full' },
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
    path: 'inicio',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./features/home/home.component').then((module) => module.HomeComponent),
  },
  { path: '**', redirectTo: 'login' },
];

import { Routes } from '@angular/router';

export const PLATAFORMA_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./plataforma-shell.component').then((module) => module.PlataformaShellComponent),
    children: [
      { path: '', loadComponent: () => import('./plataforma-dashboard.component').then((module) => module.PlataformaDashboardComponent) },
      { path: 'negocios', loadComponent: () => import('./plataforma-negocios.component').then((module) => module.PlataformaNegociosComponent) },
      { path: 'cuentas', loadComponent: () => import('./plataforma-cuentas.component').then((module) => module.PlataformaCuentasComponent) },
      { path: 'negocios/:negocioId', loadComponent: () => import('./plataforma-negocio-detalle.component').then((module) => module.PlataformaNegocioDetalleComponent) },
    ],
  },
];

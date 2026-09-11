import { Routes } from '@angular/router';

export const NEGOCIOS_ROUTES: Routes = [
  {
    path: '',
    data: { breadcrumb: 'Negocios' },
    children: [
      {
        path: '',
        pathMatch: 'full',
        data: { title: 'Mis negocios' },
        loadComponent: () =>
          import('./negocios.component').then((module) => module.NegociosComponent),
      },
      {
        path: 'nuevo',
        data: { breadcrumb: 'Registrar negocio', title: 'Registrar nuevo negocio' },
        loadComponent: () =>
          import('./negocio-form.component').then((module) => module.NegocioFormComponent),
      },
      {
        path: ':negocioId/editar',
        data: { breadcrumb: 'Editar', title: 'Editar negocio' },
        loadComponent: () =>
          import('./negocio-form.component').then((module) => module.NegocioFormComponent),
      },
      {
        path: ':negocioId',
        data: { breadcrumb: 'Detalle', title: 'Datos del negocio' },
        loadComponent: () =>
          import('./negocio-detalle.component').then((module) => module.NegocioDetalleComponent),
      },
    ],
  },
];

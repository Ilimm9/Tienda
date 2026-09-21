import { Routes } from '@angular/router';

export const PRODUCTOS_ROUTES: Routes = [
  {
    path: '',
    pathMatch: 'full',
    data: {
      breadcrumb: 'Productos',
      title: 'Productos',
      description: 'Administra el catálogo de productos de tus negocios.',
      icon: 'pi pi-tags',
    },
    loadComponent: () =>
      import('./productos.component').then((module) => module.ProductosComponent),
  },
  {
    path: 'nuevo',
    data: { breadcrumb: 'Agregar producto', title: 'Agregar producto', mode: 'create' },
    loadComponent: () => import('./productos.component').then((module) => module.ProductosComponent),
  },
  {
    path: 'importar',
    data: { breadcrumb: 'Carga masiva', title: 'Carga masiva', mode: 'import' },
    loadComponent: () => import('./productos.component').then((module) => module.ProductosComponent),
  },
  {
    path: ':productoId/editar',
    data: { breadcrumb: 'Editar producto', title: 'Editar producto', mode: 'edit' },
    loadComponent: () => import('./productos.component').then((module) => module.ProductosComponent),
  },
];

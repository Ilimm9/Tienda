import { Routes } from '@angular/router';

export const PRODUCTOS_ROUTES: Routes = [
  {
    path: '',
    data: {
      breadcrumb: 'Productos',
      title: 'Productos',
      description: 'Administra el catálogo de productos de tus negocios.',
      icon: 'pi pi-tags',
    },
    loadComponent: () =>
      import('./productos.component').then((module) => module.ProductosComponent),
  },
];

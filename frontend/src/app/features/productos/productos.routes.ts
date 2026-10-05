import { Routes } from '@angular/router';

import { permisoGuard } from '../../contexto/permiso.guard';
import { PERMISOS } from '../../contexto/permisos';

export const PRODUCTOS_ROUTES: Routes = [
  {
    path: '',
    pathMatch: 'full',
    canActivate: [permisoGuard],
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
    data: { breadcrumb: 'Agregar producto', title: 'Agregar producto', mode: 'create', permiso: PERMISOS.catalogoGestionar },
    canActivate: [permisoGuard],
    loadComponent: () => import('./productos.component').then((module) => module.ProductosComponent),
  },
  {
    path: 'importar',
    data: { breadcrumb: 'Carga masiva', title: 'Carga masiva', mode: 'import', permiso: PERMISOS.catalogoGestionar },
    canActivate: [permisoGuard],
    loadComponent: () => import('./productos.component').then((module) => module.ProductosComponent),
  },
  {
    path: ':productoId/editar',
    data: { breadcrumb: 'Editar producto', title: 'Editar producto', mode: 'edit', permiso: PERMISOS.catalogoGestionar },
    canActivate: [permisoGuard],
    loadComponent: () => import('./productos.component').then((module) => module.ProductosComponent),
  },
];

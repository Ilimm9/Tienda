import { Routes } from '@angular/router';

const productos = () => import('../productos/productos.component').then((module) => module.ProductosComponent);
const catalogo = () => import('./catalogo.component').then((module) => module.CatalogoComponent);

export const CATALOGO_ROUTES: Routes = [
  { path: '', redirectTo: 'productos', pathMatch: 'full' },
  { path: 'productos', data: { breadcrumb: 'Productos' }, loadComponent: productos },
  { path: 'marcas', data: { breadcrumb: 'Marcas', section: 'marcas' }, loadComponent: catalogo },
  { path: 'categorias', data: { breadcrumb: 'Categorías', section: 'categorias' }, loadComponent: catalogo },
  { path: 'proveedores', data: { breadcrumb: 'Proveedores', section: 'proveedores' }, loadComponent: catalogo },
];

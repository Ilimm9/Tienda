import { Routes } from '@angular/router';
import { contextoGuard } from '../../contexto/contexto.guard';

const productos = () => import('../productos/productos.component').then((module) => module.ProductosComponent);
const catalogo = () => import('./catalogo.component').then((module) => module.CatalogoComponent);

export const CATALOGO_ROUTES: Routes = [
  { path: '', redirectTo: 'productos', pathMatch: 'full' },
  { path: 'productos', canActivate: [contextoGuard], data: { breadcrumb: 'Productos' }, loadComponent: productos },
  { path: 'marcas', data: { breadcrumb: 'Marcas', section: 'marcas' }, loadComponent: catalogo },
  { path: 'categorias', data: { breadcrumb: 'Categorías', section: 'categorias' }, loadComponent: catalogo },
  { path: 'proveedores', canActivate: [contextoGuard], data: { breadcrumb: 'Proveedores', section: 'proveedores' }, loadComponent: catalogo },
  { path: 'unidades-medida', data: { breadcrumb: 'Unidades de medida', section: 'unidades' }, loadComponent: catalogo },
];

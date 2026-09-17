import { Routes } from '@angular/router';
import { contextoGuard } from '../../contexto/contexto.guard';

const catalogo = () => import('./catalogo.component').then((module) => module.CatalogoComponent);

export const CATALOGO_ROUTES: Routes = [
  { path: '', redirectTo: 'productos', pathMatch: 'full' },
  { path: 'productos', canActivate: [contextoGuard], data: { breadcrumb: 'Productos' }, loadChildren: () => import('../productos/productos.routes').then((module) => module.PRODUCTOS_ROUTES) },
  { path: 'marcas/nuevo', data: { breadcrumb: 'Agregar marca', section: 'marcas', mode: 'create' }, loadComponent: catalogo },
  { path: 'marcas/importar', data: { breadcrumb: 'Importar marcas', section: 'marcas', mode: 'import' }, loadComponent: catalogo },
  { path: 'marcas/:id/editar', data: { breadcrumb: 'Editar marca', section: 'marcas', mode: 'edit' }, loadComponent: catalogo },
  { path: 'marcas', data: { breadcrumb: 'Marcas', section: 'marcas' }, loadComponent: catalogo },
  { path: 'categorias/nuevo', data: { breadcrumb: 'Agregar categoría', section: 'categorias', mode: 'create' }, loadComponent: catalogo },
  { path: 'categorias/importar', data: { breadcrumb: 'Importar categorías', section: 'categorias', mode: 'import' }, loadComponent: catalogo },
  { path: 'categorias/:id/editar', data: { breadcrumb: 'Editar categoría', section: 'categorias', mode: 'edit' }, loadComponent: catalogo },
  { path: 'categorias', data: { breadcrumb: 'Categorías', section: 'categorias' }, loadComponent: catalogo },
  { path: 'proveedores', canActivate: [contextoGuard], data: { breadcrumb: 'Proveedores', section: 'proveedores' }, loadComponent: catalogo },
  { path: 'unidades-medida/nuevo', data: { breadcrumb: 'Agregar unidad', section: 'unidades', mode: 'create' }, loadComponent: catalogo },
  { path: 'unidades-medida/importar', data: { breadcrumb: 'Importar unidades', section: 'unidades', mode: 'import' }, loadComponent: catalogo },
  { path: 'unidades-medida/:id/editar', data: { breadcrumb: 'Editar unidad', section: 'unidades', mode: 'edit' }, loadComponent: catalogo },
  { path: 'unidades-medida', data: { breadcrumb: 'Unidades de medida', section: 'unidades' }, loadComponent: catalogo },
];

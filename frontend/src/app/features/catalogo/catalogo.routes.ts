import { Routes } from '@angular/router';
import { contextoGuard } from '../../contexto/contexto.guard';
import { permisoGuard } from '../../contexto/permiso.guard';
import { PERMISOS } from '../../contexto/permisos';

const catalogo = () => import('./catalogo.component').then((module) => module.CatalogoComponent);

export const CATALOGO_ROUTES: Routes = [
  { path: '', redirectTo: 'productos', pathMatch: 'full' },
  { path: 'productos', canActivate: [contextoGuard], data: { breadcrumb: 'Productos', permiso: PERMISOS.catalogoVer }, loadChildren: () => import('../productos/productos.routes').then((module) => module.PRODUCTOS_ROUTES) },
  { canActivate: [contextoGuard, permisoGuard], path: 'marcas/nuevo', data: { breadcrumb: 'Agregar marca', section: 'marcas', mode: 'create', permiso: PERMISOS.catalogoGestionar }, loadComponent: catalogo },
  { canActivate: [contextoGuard, permisoGuard], path: 'marcas/importar', data: { breadcrumb: 'Importar marcas', section: 'marcas', mode: 'import', permiso: PERMISOS.catalogoGestionar }, loadComponent: catalogo },
  { canActivate: [contextoGuard, permisoGuard], path: 'marcas/:id/editar', data: { breadcrumb: 'Editar marca', section: 'marcas', mode: 'edit', permiso: PERMISOS.catalogoGestionar }, loadComponent: catalogo },
  { canActivate: [contextoGuard, permisoGuard], path: 'marcas', data: { breadcrumb: 'Marcas', section: 'marcas', permiso: PERMISOS.catalogoVer }, loadComponent: catalogo },
  { canActivate: [contextoGuard, permisoGuard], path: 'categorias/nuevo', data: { breadcrumb: 'Agregar categoría', section: 'categorias', mode: 'create', permiso: PERMISOS.catalogoGestionar }, loadComponent: catalogo },
  { canActivate: [contextoGuard, permisoGuard], path: 'categorias/importar', data: { breadcrumb: 'Importar categorías', section: 'categorias', mode: 'import', permiso: PERMISOS.catalogoGestionar }, loadComponent: catalogo },
  { canActivate: [contextoGuard, permisoGuard], path: 'categorias/:id/editar', data: { breadcrumb: 'Editar categoría', section: 'categorias', mode: 'edit', permiso: PERMISOS.catalogoGestionar }, loadComponent: catalogo },
  { canActivate: [contextoGuard, permisoGuard], path: 'categorias', data: { breadcrumb: 'Categorías', section: 'categorias', permiso: PERMISOS.catalogoVer }, loadComponent: catalogo },
  { canActivate: [contextoGuard, permisoGuard], path: 'unidades-medida/nuevo', data: { breadcrumb: 'Agregar unidad', section: 'unidades', mode: 'create', permiso: PERMISOS.catalogoGestionar }, loadComponent: catalogo },
  { canActivate: [contextoGuard, permisoGuard], path: 'unidades-medida/importar', data: { breadcrumb: 'Importar unidades', section: 'unidades', mode: 'import', permiso: PERMISOS.catalogoGestionar }, loadComponent: catalogo },
  { canActivate: [contextoGuard, permisoGuard], path: 'unidades-medida/:id/editar', data: { breadcrumb: 'Editar unidad', section: 'unidades', mode: 'edit', permiso: PERMISOS.catalogoGestionar }, loadComponent: catalogo },
  { canActivate: [contextoGuard, permisoGuard], path: 'unidades-medida', data: { breadcrumb: 'Unidades de medida', section: 'unidades', permiso: PERMISOS.catalogoVer }, loadComponent: catalogo },
];

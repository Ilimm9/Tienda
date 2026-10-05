import { PermisoRequerido, PERMISOS } from '../contexto/permisos';

export interface NavigationItem {
  readonly label: string;
  readonly icon: string;
  readonly route?: string;
  /** Permiso de lectura de la sección; sin él, el elemento no aparece en el menú. */
  readonly permiso?: PermisoRequerido;
  readonly children?: readonly NavigationItem[];
}

export const NAVIGATION_ITEMS: readonly NavigationItem[] = [
  { label: 'Inicio', icon: 'pi pi-home', route: '/inicio' },
  { label: 'Negocios', icon: 'pi pi-briefcase', route: '/negocios' },
  { label: 'Sucursales', icon: 'pi pi-map-marker', route: '/sucursales', permiso: PERMISOS.sucursalVer },
  { label: 'Ventas', icon: 'pi pi-shopping-cart', route: '/ventas' },
  {
    label: 'Catálogo',
    icon: 'pi pi-tags',
    children: [
      { label: 'Productos', icon: 'pi pi-tags', route: '/catalogo/productos', permiso: PERMISOS.catalogoVer },
      { label: 'Marcas', icon: 'pi pi-bookmark', route: '/catalogo/marcas', permiso: PERMISOS.catalogoVer },
      { label: 'Categorías', icon: 'pi pi-list', route: '/catalogo/categorias', permiso: PERMISOS.catalogoVer },
      { label: 'Unidades de medida', icon: 'pi pi-sliders-h', route: '/catalogo/unidades-medida', permiso: PERMISOS.catalogoVer },
    ],
  },
  { label: 'Proveedores', icon: 'pi pi-truck', route: '/proveedores', permiso: PERMISOS.catalogoVer },
  { label: 'Compras', icon: 'pi pi-shopping-bag', route: '/compras', permiso: PERMISOS.compraVer },
  { label: 'Validar costos y precios', icon: 'pi pi-verified', route: '/precios/validaciones', permiso: PERMISOS.precioVer },
  {
    label: 'Equipo',
    icon: 'pi pi-users',
    children: [
      { label: 'Empleados', icon: 'pi pi-user', route: '/equipo/empleados', permiso: PERMISOS.empleadoVer },
      { label: 'Invitaciones', icon: 'pi pi-send', route: '/equipo/invitaciones', permiso: PERMISOS.invitacionVer },
    ],
  },
  { label: 'Roles y permisos', icon: 'pi pi-lock', route: '/roles-permisos', permiso: PERMISOS.rolVer },
];

/** Deja solo lo que `puede` permite; un grupo sin hijos visibles desaparece. */
export function filtrarNavegacion(
  items: readonly NavigationItem[],
  puede: (permiso: PermisoRequerido | undefined) => boolean,
): NavigationItem[] {
  return items.flatMap((item) => {
    if (!puede(item.permiso)) return [];
    if (!item.children) return [item];
    const children = item.children.filter((child) => puede(child.permiso));
    return children.length ? [{ ...item, children }] : [];
  });
}

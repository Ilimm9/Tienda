export interface NavigationItem {
  readonly label: string;
  readonly icon: string;
  readonly route?: string;
  readonly children?: readonly NavigationItem[];
}

export const NAVIGATION_ITEMS: readonly NavigationItem[] = [
  { label: 'Inicio', icon: 'pi pi-home', route: '/inicio' },
  { label: 'Negocios', icon: 'pi pi-briefcase', route: '/negocios' },
  { label: 'Sucursales', icon: 'pi pi-map-marker', route: '/sucursales' },
  { label: 'Ventas', icon: 'pi pi-shopping-cart', route: '/ventas' },
  { label: 'Productos', icon: 'pi pi-tags', route: '/productos' },
  {
    label: 'Equipo',
    icon: 'pi pi-users',
    children: [
      { label: 'Empleados', icon: 'pi pi-user', route: '/equipo/empleados' },
      { label: 'Invitaciones', icon: 'pi pi-send', route: '/equipo/invitaciones' },
    ],
  },
  { label: 'Roles y permisos', icon: 'pi pi-lock', route: '/roles-permisos' },
];

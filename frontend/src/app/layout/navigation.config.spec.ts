import { NAVIGATION_ITEMS } from './navigation.config';

describe('NAVIGATION_ITEMS', () => {
  it('contains every administrative destination', () => {
    const routes = NAVIGATION_ITEMS.flatMap((item) => [
      ...(item.route ? [item.route] : []),
      ...(item.children?.flatMap((child) => (child.route ? [child.route] : [])) ?? []),
    ]);

    expect(routes).toEqual([
      '/inicio',
      '/negocios',
      '/sucursales',
      '/ventas',
      '/catalogo/productos',
      '/catalogo/marcas',
      '/catalogo/categorias',
      '/catalogo/unidades-medida',
      '/proveedores',
      '/equipo/empleados',
      '/equipo/invitaciones',
      '/roles-permisos',
    ]);
  });

  it('shows Proveedores as a root entry instead of a Catálogo child', () => {
    const catalogo = NAVIGATION_ITEMS.find((item) => item.label === 'Catálogo');
    const proveedores = NAVIGATION_ITEMS.find((item) => item.label === 'Proveedores');

    expect(catalogo?.children?.some((item) => item.label === 'Proveedores')).toBe(false);
    expect(proveedores).toMatchObject({ route: '/proveedores', icon: 'pi pi-truck' });
  });
});

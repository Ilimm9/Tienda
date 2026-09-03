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
      '/equipo/empleados',
      '/equipo/invitaciones',
      '/roles-permisos',
    ]);
  });
});

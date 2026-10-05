import { filtrarNavegacion, NAVIGATION_ITEMS } from './navigation.config';

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
      '/compras',
      '/precios/validaciones',
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

  it('oculta lo que el rol no permite y retira los grupos vacíos', () => {
    const otorgados = ['catalogo.ver', 'equipo.invitaciones.ver'];
    const visibles = filtrarNavegacion(NAVIGATION_ITEMS, (permiso) =>
      [permiso ?? []].flat().every((codigo) => otorgados.includes(codigo)),
    );

    expect(visibles.map((item) => item.label)).toEqual([
      'Inicio', 'Negocios', 'Ventas', 'Catálogo', 'Proveedores', 'Equipo',
    ]);
    expect(visibles.find((item) => item.label === 'Equipo')?.children?.map((item) => item.label)).toEqual(['Invitaciones']);
  });

  it('muestra todo el menú con todos los permisos', () => {
    expect(filtrarNavegacion(NAVIGATION_ITEMS, () => true)).toHaveLength(NAVIGATION_ITEMS.length);
  });
});

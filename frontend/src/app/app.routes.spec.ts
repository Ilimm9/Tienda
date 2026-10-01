import { contextoGuard } from './contexto/contexto.guard';
import { CATALOGO_ROUTES } from './features/catalogo/catalogo.routes';
import { routes } from './app.routes';

describe('rutas principales', () => {
  it('agrupa login y registro bajo el shell de acceso persistente', () => {
    const authShell = routes.find(
      (route) => route.path === '' && route.children?.some((child) => child.path === 'login'),
    );

    expect(authShell?.loadComponent).toBeDefined();
    expect(authShell?.children?.map((route) => route.path)).toEqual(['login', 'registro']);
    expect(authShell?.children?.every((route) => route.loadComponent !== undefined)).toBe(true);
  });

  it('declara Proveedores como sección independiente protegida por contexto', () => {
    const shell = routes.find((route) => route.path === '' && route.canActivateChild !== undefined);
    const proveedores = shell?.children?.find((route) => route.path === 'proveedores');

    expect(proveedores).toMatchObject({
      path: 'proveedores',
      canActivate: [contextoGuard],
      data: { breadcrumb: 'Proveedores', section: 'proveedores' },
    });
    expect(proveedores?.loadComponent).toBeDefined();
  });

  it('elimina Proveedores de las rutas de Catálogo', () => {
    expect(CATALOGO_ROUTES.some((route) => route.path === 'proveedores')).toBe(false);
  });

  it('declara páginas protegidas para crear, editar e importar proveedores', () => {
    const shell = routes.find((route) => route.path === '' && route.canActivateChild !== undefined);
    const providerPaths = shell?.children
      ?.filter((route) => route.path?.startsWith('proveedores'))
      .map((route) => route.path);

    expect(providerPaths).toEqual([
      'proveedores',
      'proveedores/nuevo',
      'proveedores/importar',
      'proveedores/:id/editar',
    ]);
  });
});

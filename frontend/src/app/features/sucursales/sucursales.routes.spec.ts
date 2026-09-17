import { NEGOCIOS_ROUTES } from '../negocios/negocios.routes';
import { SUCURSALES_ROUTES } from './sucursales.routes';

describe('rutas de sucursales', () => {
  it('redirige ruta global al negocio activo', () => {
    expect(SUCURSALES_ROUTES[0]).toMatchObject({
      path: '',
      pathMatch: 'full',
    });
    expect(SUCURSALES_ROUTES[0].loadComponent).toBeDefined();
  });

  it('declara lista, alta, detalle y edición dentro del negocio', () => {
    const branches = NEGOCIOS_ROUTES[0].children?.find(
      (route) => route.path === ':negocioId/sucursales',
    );
    expect(branches?.children?.map((route) => route.path)).toEqual([
      '',
      'nueva',
      ':sucursalId/editar',
      ':sucursalId',
    ]);
  });
});

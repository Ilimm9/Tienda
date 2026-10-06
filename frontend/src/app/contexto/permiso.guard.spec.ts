import { TestBed } from '@angular/core/testing';
import { ActivatedRouteSnapshot, convertToParamMap, GuardResult, RedirectCommand } from '@angular/router';
import { firstValueFrom, isObservable, of } from 'rxjs';

import { ContextoService } from './contexto.service';
import { EstadoContexto } from './contexto.models';
import { permisoGuard } from './permiso.guard';

function ruta(data: Record<string, unknown>, params: Record<string, string> = {}, parent: unknown = null): ActivatedRouteSnapshot {
  return { routeConfig: { data }, paramMap: convertToParamMap(params), parent } as unknown as ActivatedRouteSnapshot;
}

function resolver(
  route: ActivatedRouteSnapshot,
  otorgados: string[],
  estado: EstadoContexto = 'listo',
  negocios: Record<string, string[]> = {},
): Promise<GuardResult> {
  TestBed.configureTestingModule({
    providers: [
      {
        provide: ContextoService,
        useValue: {
          asegurarInicializado: () => of(estado),
          puede: (requerido: string | string[]) => [requerido].flat().every((codigo) => otorgados.includes(codigo)),
          puedeEn: (id: string, requerido: string | string[], siDesconocido = false) =>
            id in negocios ? [requerido].flat().every((codigo) => negocios[id].includes(codigo)) : siDesconocido,
        },
      },
    ],
  });
  const resultado = TestBed.runInInjectionContext(() => permisoGuard(route, {} as never));
  return isObservable(resultado) ? firstValueFrom(resultado) : Promise.resolve(resultado as GuardResult);
}

describe('permisoGuard', () => {
  afterEach(() => TestBed.resetTestingModule());

  it('permite rutas sin permiso declarado', async () => {
    expect(await resolver(ruta({ breadcrumb: 'Inicio' }), [])).toBe(true);
  });

  it('permite cuando el negocio activo otorga el permiso', async () => {
    expect(await resolver(ruta({ permiso: 'roles.ver' }), ['roles.ver'])).toBe(true);
  });

  it('muestra «Sin acceso» sin cambiar la URL cuando falta el permiso', async () => {
    const resultado = await resolver(ruta({ permiso: 'roles.ver', breadcrumb: 'Roles y permisos' }), []);

    expect(resultado).toBeInstanceOf(RedirectCommand);
    const redireccion = resultado as RedirectCommand;
    expect(redireccion.redirectTo.toString()).toBe('/sin-acceso?seccion=Roles%20y%20permisos');
    expect(redireccion.navigationBehaviorOptions?.skipLocationChange).toBe(true);
  });

  it('exige también los permisos de las rutas ancestras', async () => {
    const padre = ruta({ permiso: 'roles.ver' });
    const hija = ruta({ permiso: 'roles.gestionar' }, {}, padre);

    expect(await resolver(hija, ['roles.gestionar'])).toBeInstanceOf(RedirectCommand);
    TestBed.resetTestingModule();
    expect(await resolver(hija, ['roles.gestionar', 'roles.ver'])).toBe(true);
  });

  it('exige todos los permisos de una lista', async () => {
    const route = ruta({ permiso: ['equipo.invitaciones.enviar', 'roles.ver'] });

    expect(await resolver(route, ['equipo.invitaciones.enviar'])).toBeInstanceOf(RedirectCommand);
  });

  it('evalúa las pantallas de un negocio concreto contra ese negocio', async () => {
    const route = ruta({ permiso: 'sucursales.crear' }, { negocioId: 'otro' });

    expect(await resolver(route, ['sucursales.crear'], 'listo', { otro: [] })).toBeInstanceOf(RedirectCommand);
    TestBed.resetTestingModule();
    expect(await resolver(route, [], 'listo', { otro: ['sucursales.crear'] })).toBe(true);
  });

  it('deja pasar un negocio que el contexto no conoce; la API decide', async () => {
    expect(await resolver(ruta({ permiso: 'negocios.ver' }, { negocioId: 'archivado' }), [])).toBe(true);
  });

  it('no decide mientras falta elegir negocio', async () => {
    expect(await resolver(ruta({ permiso: 'roles.ver' }), [], 'requiere_negocio')).toBe(true);
  });
});

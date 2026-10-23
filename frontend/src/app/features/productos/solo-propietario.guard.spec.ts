import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { ActivatedRouteSnapshot, GuardResult, RedirectCommand } from '@angular/router';
import { firstValueFrom, isObservable, of } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { soloPropietarioGuard } from './solo-propietario.guard';

function resolver(tipoMiembro: 'propietario' | 'miembro' | null): Promise<GuardResult> {
  TestBed.configureTestingModule({
    providers: [
      {
        provide: ContextoService,
        useValue: {
          asegurarInicializado: () => of('listo'),
          negocio: signal(tipoMiembro ? { id: 'negocio-1', tipo_miembro: tipoMiembro } : null),
        },
      },
    ],
  });
  const route = { routeConfig: { data: { breadcrumb: 'Agregar producto' } }, parent: null } as unknown as ActivatedRouteSnapshot;
  const resultado = TestBed.runInInjectionContext(() => soloPropietarioGuard(route, {} as never));
  return isObservable(resultado) ? firstValueFrom(resultado) : Promise.resolve(resultado as GuardResult);
}

describe('soloPropietarioGuard', () => {
  afterEach(() => TestBed.resetTestingModule());

  it('deja pasar al propietario', async () => {
    expect(await resolver('propietario')).toBe(true);
  });

  it('lleva a «Sin acceso» a un miembro aunque tenga permisos de catálogo', async () => {
    expect(await resolver('miembro')).toBeInstanceOf(RedirectCommand);
  });

  it('lleva a «Sin acceso» sin negocio activo', async () => {
    expect(await resolver(null)).toBeInstanceOf(RedirectCommand);
  });
});

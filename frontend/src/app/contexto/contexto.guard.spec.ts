import { TestBed } from '@angular/core/testing';
import { GuardResult, Router } from '@angular/router';
import { of } from 'rxjs';
import { firstValueFrom, isObservable } from 'rxjs';

import { contextoGuard } from './contexto.guard';
import { ContextoService } from './contexto.service';
import { EstadoContexto } from './contexto.models';

function resolver(estado: EstadoContexto): Promise<GuardResult> {
  TestBed.configureTestingModule({
    providers: [
      {
        provide: ContextoService,
        useValue: { asegurarInicializado: () => of(estado) },
      },
    ],
  });
  const resultado = TestBed.runInInjectionContext(() =>
    contextoGuard({} as never, {} as never),
  );
  return isObservable(resultado) ? firstValueFrom(resultado) : Promise.resolve(resultado as GuardResult);
}

describe('contextoGuard', () => {
  afterEach(() => TestBed.resetTestingModule());

  it('permite entrar cuando el contexto quedó listo', async () => {
    expect(await resolver('listo')).toBe(true);
  });

  it('permite entrar cuando el negocio no tiene sucursales activas', async () => {
    expect(await resolver('sin_sucursal')).toBe(true);
  });

  it('envía a elegir negocio cuando falta seleccionarlo', async () => {
    const resultado = await resolver('requiere_negocio');
    const router = TestBed.inject(Router);

    expect(resultado).toEqual(router.createUrlTree(['/seleccionar-negocio']));
  });

  it('envía a negocios cuando el contexto no pudo resolverse', async () => {
    const resultado = await resolver('error');
    const router = TestBed.inject(Router);

    expect(resultado).toEqual(router.createUrlTree(['/negocios']));
  });
});

import { TestBed } from '@angular/core/testing';
import { GuardResult, Router } from '@angular/router';
import { firstValueFrom, isObservable, of } from 'rxjs';

import { altaInicialGuard, asistenteGuard } from './alta-inicial.guard';
import { ContextoService } from './contexto.service';
import { EstadoAlta, EstadoContexto } from './contexto.models';

function resolver(
  guard: typeof altaInicialGuard,
  alta: EstadoAlta,
  estado: EstadoContexto = 'listo',
  pospuesta = false,
): Promise<GuardResult> {
  TestBed.configureTestingModule({
    providers: [
      {
        provide: ContextoService,
        useValue: { asegurarInicializado: () => of(estado), estadoAlta: () => alta, altaPospuesta: () => pospuesta },
      },
    ],
  });
  const resultado = TestBed.runInInjectionContext(() => guard({} as never, {} as never));
  return isObservable(resultado) ? firstValueFrom(resultado) : Promise.resolve(resultado as GuardResult);
}

describe('altaInicialGuard', () => {
  afterEach(() => TestBed.resetTestingModule());

  it('lleva al asistente cuando falta la empresa', async () => {
    const resultado = await resolver(altaInicialGuard, 'requiere_empresa', 'requiere_negocio');
    expect(resultado).toEqual(TestBed.inject(Router).createUrlTree(['/configuracion-inicial']));
  });

  it('lleva al asistente cuando falta la sucursal', async () => {
    const resultado = await resolver(altaInicialGuard, 'requiere_sucursal', 'sin_sucursal');
    expect(resultado).toEqual(TestBed.inject(Router).createUrlTree(['/configuracion-inicial']));
  });

  it('respeta «hacerlo después» y abre Inicio aunque no haya empresa', async () => {
    expect(await resolver(altaInicialGuard, 'requiere_empresa', 'requiere_negocio', true)).toBe(true);
    TestBed.resetTestingModule();
    expect(await resolver(altaInicialGuard, 'requiere_sucursal', 'sin_sucursal', true)).toBe(true);
  });

  it('no manda al asistente a un invitado sin sucursal', async () => {
    expect(await resolver(altaInicialGuard, 'sin_asignacion', 'sin_sucursal')).toBe(true);
  });

  it('con varios negocios sin elegir conserva la selección de negocio', async () => {
    const resultado = await resolver(altaInicialGuard, 'listo', 'requiere_negocio');
    expect(resultado).toEqual(TestBed.inject(Router).createUrlTree(['/seleccionar-negocio']));
  });

  it('con contexto listo abre Inicio', async () => {
    expect(await resolver(altaInicialGuard, 'listo')).toBe(true);
  });
});

describe('asistenteGuard', () => {
  afterEach(() => TestBed.resetTestingModule());

  it('abre el asistente a quien puede completar el alta', async () => {
    expect(await resolver(asistenteGuard, 'requiere_empresa', 'requiere_negocio')).toBe(true);
    TestBed.resetTestingModule();
    expect(await resolver(asistenteGuard, 'requiere_sucursal', 'sin_sucursal')).toBe(true);
  });

  it('devuelve a Inicio a quien ya terminó o no puede completarla', async () => {
    const inicio = () => TestBed.inject(Router).createUrlTree(['/inicio']);
    expect(await resolver(asistenteGuard, 'listo')).toEqual(inicio());
    TestBed.resetTestingModule();
    expect(await resolver(asistenteGuard, 'sin_asignacion', 'sin_sucursal')).toEqual(inicio());
  });
});

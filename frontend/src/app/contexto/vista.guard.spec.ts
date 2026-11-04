import { TestBed } from '@angular/core/testing';
import { ActivatedRouteSnapshot, Router, RouterStateSnapshot, UrlTree, provideRouter } from '@angular/router';
import { firstValueFrom, Observable, of } from 'rxjs';

import { ContextoService } from './contexto.service';
import { vistaGuard } from './vista.guard';
import { VistasService } from './vistas.service';

describe('vistaGuard', () => {
  function ejecutar(vista: string | undefined, ocultas: string[]): Promise<boolean | UrlTree> {
    TestBed.configureTestingModule({
      providers: [
        provideRouter([]),
        { provide: ContextoService, useValue: { asegurarInicializado: () => of('listo') } },
        { provide: VistasService, useValue: { asegurar: () => of(undefined), visible: (seccion?: string) => !seccion || !ocultas.includes(seccion) } },
      ],
    });
    const route = { data: vista ? { vista } : {} } as unknown as ActivatedRouteSnapshot;
    const resultado = TestBed.runInInjectionContext(() => vistaGuard(route, {} as RouterStateSnapshot));
    return firstValueFrom(resultado as Observable<boolean | UrlTree>);
  }

  afterEach(() => TestBed.resetTestingModule());

  it('deja entrar a una sección visible o sin vista', async () => {
    expect(await ejecutar('categorias', ['marcas'])).toBe(true);
    TestBed.resetTestingModule();
    expect(await ejecutar(undefined, ['categorias'])).toBe(true);
  });

  it('lleva a Inicio cuando la sección está oculta', async () => {
    const resultado = await ejecutar('categorias', ['categorias']);
    expect(resultado instanceof UrlTree).toBe(true);
    expect(TestBed.inject(Router).serializeUrl(resultado as UrlTree)).toBe('/inicio');
  });
});

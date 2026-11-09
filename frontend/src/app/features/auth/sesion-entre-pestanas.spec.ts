import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { of, Subject, throwError } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { AuthService, AvisoSesion } from './auth.service';
import { sincronizarSesionEntrePestanas } from './sesion-entre-pestanas';

describe('sincronizarSesionEntrePestanas', () => {
  let avisos: Subject<AvisoSesion>;
  const auth = { sesionEnOtraPestana: null as unknown as Subject<AvisoSesion>, asegurarSesion: vi.fn() };
  const router = { url: '/inicio', navigate: vi.fn(), navigateByUrl: vi.fn() };
  const contexto = { limpiar: vi.fn() };

  beforeEach(() => {
    vi.clearAllMocks();
    avisos = new Subject<AvisoSesion>();
    auth.sesionEnOtraPestana = avisos;
    TestBed.configureTestingModule({
      providers: [
        { provide: AuthService, useValue: auth },
        { provide: Router, useValue: router },
        { provide: ContextoService, useValue: contexto },
      ],
    });
    TestBed.runInInjectionContext(sincronizarSesionEntrePestanas);
  });

  it('un cierre en otra pestaña lleva al login y limpia el contexto', () => {
    router.url = '/catalogo/productos?pagina=2';
    avisos.next('cierre');

    expect(contexto.limpiar).toHaveBeenCalled();
    expect(router.navigate).toHaveBeenCalledWith(['/login'], { replaceUrl: true });
  });

  it('un cierre no interrumpe pantallas públicas', () => {
    router.url = '/restablecer-contrasena?token=abc';
    avisos.next('cierre');

    expect(router.navigate).not.toHaveBeenCalled();
  });

  it('un inicio en otra pestaña saca del login hacia la app', () => {
    router.url = '/login';
    auth.asegurarSesion.mockReturnValue(of({ id: '1', correo: 'persona@ejemplo.com' }));
    avisos.next('inicio');

    expect(router.navigateByUrl).toHaveBeenCalledWith('/inicio', { replaceUrl: true });
  });

  it('un inicio no mueve pantallas que no son de visitante', () => {
    router.url = '/invitacion/token-1';
    avisos.next('inicio');

    expect(auth.asegurarSesion).not.toHaveBeenCalled();
  });

  it('sin sesión válida el login sigue disponible', () => {
    router.url = '/registro';
    auth.asegurarSesion.mockReturnValue(throwError(() => new Error('unauthorized')));
    avisos.next('inicio');

    expect(router.navigateByUrl).not.toHaveBeenCalled();
  });
});

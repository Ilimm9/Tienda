import { TestBed } from '@angular/core/testing';
import { ActivatedRouteSnapshot, Router, RouterStateSnapshot, UrlTree } from '@angular/router';
import { firstValueFrom, Observable, of, throwError } from 'rxjs';

import { AuthService } from './auth.service';
import { invitadoGuard } from './invitado.guard';

describe('invitadoGuard', () => {
  const auth = { asegurarSesion: vi.fn() };
  const router = { createUrlTree: vi.fn(() => new UrlTree()) };

  beforeEach(() => {
    vi.clearAllMocks();
    TestBed.configureTestingModule({
      providers: [
        { provide: AuthService, useValue: auth },
        { provide: Router, useValue: router },
      ],
    });
  });

  function runGuard(): Observable<boolean | UrlTree> {
    return TestBed.runInInjectionContext(
      () =>
        invitadoGuard({} as ActivatedRouteSnapshot, {} as RouterStateSnapshot) as Observable<
          boolean | UrlTree
        >,
    );
  }

  it('deja ver la pantalla de acceso sin sesión', async () => {
    auth.asegurarSesion.mockReturnValue(throwError(() => new Error('unauthorized')));

    await expect(firstValueFrom(runGuard())).resolves.toBe(true);
    expect(router.createUrlTree).not.toHaveBeenCalled();
  });

  it('con sesión activa regresa a inicio', async () => {
    auth.asegurarSesion.mockReturnValue(of({ id: '1', correo: 'persona@ejemplo.com' }));

    await expect(firstValueFrom(runGuard())).resolves.toBeInstanceOf(UrlTree);
    expect(router.createUrlTree).toHaveBeenCalledWith(['/inicio']);
  });

  it('con sesión de plataforma regresa al panel de plataforma', async () => {
    auth.asegurarSesion.mockReturnValue(
      of({ id: '1', correo: 'admin@ejemplo.com', roles_plataforma: ['ADMIN_LECTURA'] }),
    );

    await expect(firstValueFrom(runGuard())).resolves.toBeInstanceOf(UrlTree);
    expect(router.createUrlTree).toHaveBeenCalledWith(['/administracion-plataforma']);
  });
});

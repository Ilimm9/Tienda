import { TestBed } from '@angular/core/testing';
import { ActivatedRouteSnapshot, Router, RouterStateSnapshot, UrlTree } from '@angular/router';
import { firstValueFrom, Observable, of, throwError } from 'rxjs';

import { AuthService } from '../auth/auth.service';
import { fueraDePlataformaGuard, plataformaGuard } from './plataforma.guard';

describe('plataformaGuard', () => {
  const auth = { me: vi.fn() };
  const router = { createUrlTree: vi.fn(() => new UrlTree()) };

  beforeEach(() => {
    vi.clearAllMocks();
    TestBed.configureTestingModule({ providers: [{ provide: AuthService, useValue: auth }, { provide: Router, useValue: router }] });
  });

  function run(): Observable<boolean | UrlTree> {
    return TestBed.runInInjectionContext(() => plataformaGuard({} as ActivatedRouteSnapshot, {} as RouterStateSnapshot) as Observable<boolean | UrlTree>);
  }

  it('acepta ADMIN_LECTURA', async () => {
    auth.me.mockReturnValue(of({ autenticado: true, usuario: { id: '1', correo: 'admin@ejemplo.com', roles_plataforma: ['ADMIN_LECTURA'] } }));
    await expect(firstValueFrom(run())).resolves.toBe(true);
  });

  it('rechaza roles de negocio', async () => {
    auth.me.mockReturnValue(of({ autenticado: true, usuario: { id: '1', correo: 'propietario@ejemplo.com', roles_plataforma: [] } }));
    await expect(firstValueFrom(run())).resolves.toBeInstanceOf(UrlTree);
    expect(router.createUrlTree).toHaveBeenCalledWith(['/sin-acceso']);
  });
});

describe('fueraDePlataformaGuard', () => {
  const auth = { me: vi.fn() };
  const router = { createUrlTree: vi.fn(() => new UrlTree()) };

  beforeEach(() => {
    vi.clearAllMocks();
    TestBed.configureTestingModule({ providers: [{ provide: AuthService, useValue: auth }, { provide: Router, useValue: router }] });
  });

  afterEach(() => TestBed.resetTestingModule());

  function run(): Observable<boolean | UrlTree> {
    return TestBed.runInInjectionContext(() => fueraDePlataformaGuard({} as ActivatedRouteSnapshot, {} as RouterStateSnapshot) as Observable<boolean | UrlTree>);
  }

  it('devuelve ADMIN_LECTURA al panel de plataforma', async () => {
    auth.me.mockReturnValue(of({ autenticado: true, usuario: { id: '1', correo: 'admin@ejemplo.com', roles_plataforma: ['ADMIN_LECTURA'] } }));
    await expect(firstValueFrom(run())).resolves.toBeInstanceOf(UrlTree);
    expect(router.createUrlTree).toHaveBeenCalledWith(['/administracion-plataforma']);
  });

  it('deja pasar cuentas de negocio', async () => {
    auth.me.mockReturnValue(of({ autenticado: true, usuario: { id: '1', correo: 'propietario@ejemplo.com', roles_plataforma: [] } }));
    await expect(firstValueFrom(run())).resolves.toBe(true);
  });

  it('sin sesión deja decidir a authGuard', async () => {
    auth.me.mockReturnValue(throwError(() => new Error('401')));
    await expect(firstValueFrom(run())).resolves.toBe(true);
  });
});

import { TestBed } from '@angular/core/testing';
import { ActivatedRouteSnapshot, Router, RouterStateSnapshot, UrlTree } from '@angular/router';
import { firstValueFrom, Observable, of, throwError } from 'rxjs';

import { authGuard } from './auth.guard';
import { AuthService } from './auth.service';

describe('authGuard', () => {
  const auth = { me: vi.fn() };
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
        authGuard({} as ActivatedRouteSnapshot, {} as RouterStateSnapshot) as Observable<
          boolean | UrlTree
        >,
    );
  }

  it('allows access when the session is valid', async () => {
    auth.me.mockReturnValue(of({ usuario: { id: '1', correo: 'persona@ejemplo.com' } }));

    await expect(firstValueFrom(runGuard())).resolves.toBe(true);
  });

  it('redirects to login when the session is invalid', async () => {
    auth.me.mockReturnValue(throwError(() => new Error('unauthorized')));

    await expect(firstValueFrom(runGuard())).resolves.toBeInstanceOf(UrlTree);
    expect(router.createUrlTree).toHaveBeenCalledWith(['/login']);
  });
});

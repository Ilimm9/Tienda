import { TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';

import { FeedbackService } from '../../../shared/feedback/feedback.service';
import { AuthService } from '../auth.service';
import { ResetPasswordComponent } from './reset-password.component';

describe('ResetPasswordComponent', () => {
  const auth = { resetPassword: vi.fn() };
  const feedback = { success: vi.fn() };
  let router: Router;

  beforeEach(async () => {
    vi.clearAllMocks();
    await TestBed.configureTestingModule({
      imports: [ResetPasswordComponent],
      providers: [
        provideRouter([]),
        { provide: AuthService, useValue: auth },
        { provide: FeedbackService, useValue: feedback },
      ],
    }).compileComponents();
    router = TestBed.inject(Router);
    vi.spyOn(router, 'navigate').mockResolvedValue(true);
  });

  afterEach(() => history.replaceState(null, '', '/'));

  function create(hash: string) {
    history.replaceState(null, '', `/restablecer-contrasena${hash}`);
    return TestBed.createComponent(ResetPasswordComponent);
  }

  it('reads the token from the fragment, removes it from the URL and resets the password', () => {
    auth.resetPassword.mockReturnValue(of({ mensaje: 'Tu contraseña se actualizó.' }));
    const fixture = create('#desafio=challenge-1&token=abc_DEF-123');
    const component = fixture.componentInstance;

    expect(component.linkValid).toBe(true);
    expect(window.location.hash).toBe('');

    component.form.setValue({ contrasena: 'nueva-contraseña', confirmarContrasena: 'otra-cosa' });
    component.submit();
    expect(auth.resetPassword).not.toHaveBeenCalled();

    component.form.setValue({ contrasena: 'nueva-contraseña', confirmarContrasena: 'nueva-contraseña' });
    component.submit();
    expect(auth.resetPassword).toHaveBeenCalledWith({
      desafio_id: 'challenge-1',
      token: 'abc_DEF-123',
      contrasena: 'nueva-contraseña',
    });
    expect(feedback.success).toHaveBeenCalledWith('Contraseña actualizada', 'Tu contraseña se actualizó.');
    expect(router.navigate).toHaveBeenCalledWith(['/login']);
  });

  it('rejects passwords longer than 72 bytes', () => {
    const fixture = create('#desafio=challenge-1&token=abc');
    const component = fixture.componentInstance;
    const long = 'ñ'.repeat(37);

    component.form.setValue({ contrasena: long, confirmarContrasena: long });
    component.submit();
    expect(component.form.controls.contrasena.hasError('maxBytes')).toBe(true);
    expect(auth.resetPassword).not.toHaveBeenCalled();
  });

  it('marks an incomplete link as invalid and shows API errors', () => {
    const incomplete = create('#desafio=challenge-1');
    incomplete.detectChanges();
    expect(incomplete.componentInstance.linkValid).toBe(false);
    expect(incomplete.nativeElement.querySelector('[role="alert"]').textContent).toContain('no es válido');

    auth.resetPassword.mockReturnValue(throwError(() => ({ error: { mensaje: 'el enlace no es válido o expiró' } })));
    const component = create('#desafio=challenge-1&token=abc').componentInstance;
    component.form.setValue({ contrasena: 'nueva-contraseña', confirmarContrasena: 'nueva-contraseña' });
    component.submit();
    expect(component.error()).toBe('el enlace no es válido o expiró');
    expect(router.navigate).not.toHaveBeenCalled();
  });
});

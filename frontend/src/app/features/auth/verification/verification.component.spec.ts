import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router } from '@angular/router';
import { of, throwError } from 'rxjs';

import { FeedbackService } from '../../../shared/feedback/feedback.service';
import { AuthService } from '../auth.service';
import { VerificationComponent } from './verification.component';

describe('VerificationComponent', () => {
  const auth = { verifyEmail: vi.fn(), resendVerification: vi.fn() };
  const router = { navigate: vi.fn(() => Promise.resolve(true)) };
  const feedback = { success: vi.fn() };

  beforeEach(async () => {
    vi.useFakeTimers();
    vi.clearAllMocks();
    sessionStorage.clear();
    await TestBed.configureTestingModule({
      imports: [VerificationComponent],
      providers: [
        { provide: AuthService, useValue: auth },
        { provide: Router, useValue: router },
        { provide: FeedbackService, useValue: feedback },
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              queryParamMap: { get: (key: string) => (key === 'desafio' ? 'challenge-1' : null) },
            },
          },
        },
      ],
    }).compileComponents();
  });

  afterEach(() => vi.useRealTimers());

  it('validates six digits and starts a session after verification', () => {
    auth.verifyEmail.mockReturnValue(of({
      mensaje: 'Correo verificado',
      usuario: { id: '1', correo: 'persona@ejemplo.com' },
    }));
    const fixture = TestBed.createComponent(VerificationComponent);
    const component = fixture.componentInstance;

    component.form.setValue({ codigo: '12ab' });
    component.submit();
    expect(auth.verifyEmail).not.toHaveBeenCalled();

    component.form.setValue({ codigo: '123456' });
    component.submit();
    expect(auth.verifyEmail).toHaveBeenCalledWith({ desafio_id: 'challenge-1', codigo: '123456' });
    expect(feedback.success).toHaveBeenCalledWith('Correo verificado', 'Tu cuenta ya está activa.');
    expect(router.navigate).toHaveBeenCalledWith(['/inicio']);
    fixture.destroy();
  });

  it('allows resend only after the countdown and replaces the challenge', () => {
    auth.resendVerification.mockReturnValue(of({
      mensaje: 'Código reenviado',
      desafio_id: 'challenge-2',
      correo_enmascarado: 'p***@ejemplo.com',
      reenviar_en_segundos: 60,
    }));
    const fixture = TestBed.createComponent(VerificationComponent);
    const component = fixture.componentInstance;

    component.resend();
    expect(auth.resendVerification).not.toHaveBeenCalled();
    vi.advanceTimersByTime(60_000);
    component.resend();
    expect(auth.resendVerification).toHaveBeenCalledWith({ desafio_id: 'challenge-1' });
    expect(sessionStorage.getItem('tienda.verification.challenge')).toBe('challenge-2');
    fixture.destroy();
  });

  it('shows the generic API error for an invalid code', () => {
    auth.verifyEmail.mockReturnValue(throwError(() => ({ error: { mensaje: 'el código no es válido o expiró' } })));
    const fixture = TestBed.createComponent(VerificationComponent);
    const component = fixture.componentInstance;
    component.form.setValue({ codigo: '123456' });
    component.submit();
    expect(component.error()).toBe('el código no es válido o expiró');
    fixture.destroy();
  });
});

import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router } from '@angular/router';
import { of, throwError } from 'rxjs';

import { FeedbackService } from '../../../shared/feedback/feedback.service';
import { AuthService } from '../auth.service';
import { RegisterComponent } from './register.component';

describe('RegisterComponent', () => {
  const auth = { register: vi.fn() };
  const router = { navigate: vi.fn(() => Promise.resolve(true)) };
  const feedback = { success: vi.fn() };

  beforeEach(async () => {
    vi.clearAllMocks();
    sessionStorage.clear();
    await TestBed.configureTestingModule({
      imports: [RegisterComponent],
      providers: [
        { provide: AuthService, useValue: auth },
        { provide: Router, useValue: router },
        { provide: FeedbackService, useValue: feedback },
        { provide: ActivatedRoute, useValue: { snapshot: {} } },
      ],
    }).compileComponents();
  });

  it('rejects mismatched passwords', () => {
    const component = TestBed.createComponent(RegisterComponent).componentInstance;
    component.form.patchValue({ contrasena: '12345678', confirmarContrasena: '87654321' });

    expect(component.form.hasError('passwordsMismatch')).toBe(true);
  });

  it('maps the form to the existing API contract', () => {
    auth.register.mockReturnValue(of({
      mensaje: 'Código enviado',
      desafio_id: 'challenge-1',
      correo_enmascarado: 'j***@ejemplo.com',
      reenviar_en_segundos: 60,
    }));
    const component = TestBed.createComponent(RegisterComponent).componentInstance;
    component.form.setValue({
      nombreCompleto: 'Juan Pérez',
      correo: 'juan@ejemplo.com',
      telefono: '',
      contrasena: '12345678',
      confirmarContrasena: '12345678',
    });

    component.submit();

    expect(auth.register).toHaveBeenCalledWith({
      nombre_completo: 'Juan Pérez',
      correo: 'juan@ejemplo.com',
      telefono: '',
      contrasena: '12345678',
    });
    expect(feedback.success).toHaveBeenCalledWith(
      'Código enviado',
      'Revisa tu correo para activar la cuenta.',
    );
    expect(sessionStorage.getItem('tienda.verification.challenge')).toBe('challenge-1');
    expect(router.navigate).toHaveBeenCalledWith(['/verificar-correo'], {
      queryParams: { desafio: 'challenge-1' },
    });
    expect(component.loading()).toBe(false);
  });

  it('keeps API errors and both password toggles', () => {
    auth.register.mockReturnValue(
      throwError(() => ({ error: { mensaje: 'el correo electrónico ya está registrado' } })),
    );
    const component = TestBed.createComponent(RegisterComponent).componentInstance;
    component.form.setValue({
      nombreCompleto: 'Juan Pérez',
      correo: 'juan@ejemplo.com',
      telefono: '55 1234 5678',
      contrasena: '12345678',
      confirmarContrasena: '12345678',
    });

    component.togglePassword();
    component.toggleConfirmation();
    component.submit();

    expect(component.passwordVisible()).toBe(true);
    expect(component.confirmationVisible()).toBe(true);
    expect(component.error()).toBe('el correo electrónico ya está registrado');
  });

  it('preserves a recoverable challenge when email delivery fails', () => {
    auth.register.mockReturnValue(
      throwError(() => ({
        status: 503,
        error: { mensaje: 'No fue posible enviar', desafio_id: 'challenge-retry' },
      })),
    );
    const component = TestBed.createComponent(RegisterComponent).componentInstance;
    component.form.setValue({
      nombreCompleto: 'Juan Pérez',
      correo: 'juan@ejemplo.com',
      telefono: '',
      contrasena: '12345678',
      confirmarContrasena: '12345678',
    });

    component.submit();

    expect(sessionStorage.getItem('tienda.verification.challenge')).toBe('challenge-retry');
    expect(router.navigate).toHaveBeenCalledWith(['/verificar-correo'], {
      queryParams: { desafio: 'challenge-retry', envio: 'pendiente' },
    });
  });
});

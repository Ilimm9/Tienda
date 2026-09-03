import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router } from '@angular/router';
import { of, throwError } from 'rxjs';

import { AuthService } from '../auth.service';
import { RegisterComponent } from './register.component';

describe('RegisterComponent', () => {
  const auth = { register: vi.fn() };
  const router = { navigate: vi.fn(() => Promise.resolve(true)) };

  beforeEach(async () => {
    vi.clearAllMocks();
    await TestBed.configureTestingModule({
      imports: [RegisterComponent],
      providers: [
        { provide: AuthService, useValue: auth },
        { provide: Router, useValue: router },
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
    auth.register.mockReturnValue(of({ mensaje: 'Cuenta creada correctamente' }));
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
    expect(router.navigate).toHaveBeenCalledWith(['/login'], {
      queryParams: { registrado: '1' },
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
});

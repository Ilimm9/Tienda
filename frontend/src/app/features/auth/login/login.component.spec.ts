import { ActivatedRoute, Router } from '@angular/router';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';

import { AuthService } from '../auth.service';
import { LoginComponent } from './login.component';

describe('LoginComponent', () => {
  const auth = { login: vi.fn() };
  const router = { navigate: vi.fn(() => Promise.resolve(true)) };

  beforeEach(async () => {
    vi.clearAllMocks();
    await TestBed.configureTestingModule({
      imports: [LoginComponent],
      providers: [
        { provide: AuthService, useValue: auth },
        { provide: Router, useValue: router },
        { provide: ActivatedRoute, useValue: { snapshot: {} } },
      ],
    }).compileComponents();
  });

  it('validates and submits the same login payload', () => {
    auth.login.mockReturnValue(of({ usuario: { id: '1', correo: 'persona@ejemplo.com' } }));
    const fixture = TestBed.createComponent(LoginComponent);
    const component = fixture.componentInstance;

    component.submit();
    expect(auth.login).not.toHaveBeenCalled();
    expect(component.form.controls.correo.touched).toBe(true);

    component.form.setValue({
      correo: 'persona@ejemplo.com',
      contrasena: 'secreto',
      recordarme: false,
    });
    component.submit();

    expect(auth.login).toHaveBeenCalledWith(component.form.getRawValue());
    expect(router.navigate).toHaveBeenCalledWith(['/inicio']);
    expect(component.loading()).toBe(false);
  });

  it('keeps the API error message and password toggle', () => {
    auth.login.mockReturnValue(
      throwError(() => ({ error: { mensaje: 'credenciales inválidas' } })),
    );
    const component = TestBed.createComponent(LoginComponent).componentInstance;
    component.form.setValue({
      correo: 'persona@ejemplo.com',
      contrasena: 'incorrecta',
      recordarme: false,
    });

    component.togglePassword();
    component.submit();

    expect(component.passwordVisible()).toBe(true);
    expect(component.error()).toBe('credenciales inválidas');
  });
});

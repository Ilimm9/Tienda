import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';

import { AuthService } from '../auth.service';
import { RecoverPasswordComponent } from './recover-password.component';

describe('RecoverPasswordComponent', () => {
  const auth = { requestPasswordReset: vi.fn() };

  beforeEach(async () => {
    vi.clearAllMocks();
    await TestBed.configureTestingModule({
      imports: [RecoverPasswordComponent],
      providers: [provideRouter([]), { provide: AuthService, useValue: auth }],
    }).compileComponents();
  });

  it('validates the email and shows the uniform confirmation', () => {
    auth.requestPasswordReset.mockReturnValue(of({ mensaje: 'Si existe una cuenta con ese correo, enviaremos un enlace.' }));
    const fixture = TestBed.createComponent(RecoverPasswordComponent);
    const component = fixture.componentInstance;

    component.form.setValue({ correo: 'no-es-correo' });
    component.submit();
    expect(auth.requestPasswordReset).not.toHaveBeenCalled();

    component.form.setValue({ correo: 'persona@ejemplo.com' });
    component.submit();
    fixture.detectChanges();
    expect(auth.requestPasswordReset).toHaveBeenCalledWith({ correo: 'persona@ejemplo.com' });
    expect(fixture.nativeElement.querySelector('[role="status"]').textContent).toContain('Si existe una cuenta');
    expect(fixture.nativeElement.querySelector('form')).toBeNull();
  });

  it('shows API errors without leaving the form', () => {
    auth.requestPasswordReset.mockReturnValue(throwError(() => ({ error: { mensaje: 'No fue posible procesar la solicitud' } })));
    const fixture = TestBed.createComponent(RecoverPasswordComponent);
    const component = fixture.componentInstance;

    component.form.setValue({ correo: 'persona@ejemplo.com' });
    component.submit();
    expect(component.error()).toBe('No fue posible procesar la solicitud');
    expect(component.sentMessage()).toBe('');
  });
});

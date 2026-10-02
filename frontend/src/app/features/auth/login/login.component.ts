import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { NonNullableFormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { finalize } from 'rxjs';

import { consumirRetornoInvitacion } from '../../invitacion/retorno-invitacion';
import { AuthService } from '../auth.service';
import { AuthTransitionService } from '../auth-transition.service';

@Component({
  selector: 'app-login',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './login.component.html',
  styleUrl: './login.component.css',
})
export class LoginComponent {
  private readonly formBuilder = inject(NonNullableFormBuilder);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  private readonly transition = inject(AuthTransitionService);

  readonly loading = signal(false);
  readonly error = signal('');
  readonly passwordVisible = signal(false);
  readonly form = this.formBuilder.group({
    correo: ['', [Validators.required, Validators.email]],
    contrasena: ['', Validators.required],
    recordarme: false,
  });

  togglePassword(): void {
    this.passwordVisible.update((visible) => !visible);
  }

  switchToRegister(): void {
    this.transition.request('registro');
  }

  submit(): void {
    this.error.set('');
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.loading.set(true);
    this.auth
      .login(this.form.getRawValue())
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        // Quien llegó desde un enlace de invitación vuelve a él para confirmar su acceso.
        next: () => {
          const retorno = consumirRetornoInvitacion();
          void (retorno ? this.router.navigateByUrl(retorno) : this.router.navigate(['/inicio']));
        },
        error: (response: HttpErrorResponse) =>
          this.error.set(response.error?.mensaje || 'No fue posible iniciar sesión.'),
      });
  }
}

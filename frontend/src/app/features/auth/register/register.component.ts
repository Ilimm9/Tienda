import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import {
  AbstractControl,
  NonNullableFormBuilder,
  ReactiveFormsModule,
  ValidationErrors,
  Validators,
} from '@angular/forms';
import { Router } from '@angular/router';
import { finalize } from 'rxjs';

import { FeedbackService } from '../../../shared/feedback/feedback.service';
import { AuthService } from '../auth.service';
import { AuthTransitionService } from '../auth-transition.service';

function matchingPasswords(control: AbstractControl): ValidationErrors | null {
  const password = control.get('contrasena');
  const confirmation = control.get('confirmarContrasena');

  if (!password || !confirmation || !confirmation.value) {
    return null;
  }

  return password.value === confirmation.value ? null : { passwordsMismatch: true };
}

@Component({
  selector: 'app-register',
  imports: [ReactiveFormsModule],
  templateUrl: './register.component.html',
  styleUrl: './register.component.css',
})
export class RegisterComponent {
  private readonly formBuilder = inject(NonNullableFormBuilder);
  private readonly auth = inject(AuthService);
  private readonly feedback = inject(FeedbackService);
  private readonly router = inject(Router);
  private readonly transition = inject(AuthTransitionService);

  readonly loading = signal(false);
  readonly error = signal('');
  /** Correo que el servidor rechazó por pertenecer ya a una cuenta; se limpia al cambiarlo. */
  readonly correoRegistrado = signal('');
  readonly passwordVisible = signal(false);
  readonly confirmationVisible = signal(false);
  readonly form = this.formBuilder.group(
    {
      nombres: ['', [Validators.required, Validators.pattern(/\S/), Validators.maxLength(120)]],
      primerApellido: ['', [Validators.required, Validators.pattern(/\S/), Validators.maxLength(120)]],
      segundoApellido: ['', Validators.maxLength(120)],
      correo: ['', [Validators.required, Validators.email, Validators.maxLength(254)]],
      telefono: ['', Validators.maxLength(30)],
      contrasena: ['', [Validators.required, Validators.minLength(8), Validators.maxLength(72)]],
      confirmarContrasena: ['', Validators.required],
    },
    { validators: matchingPasswords },
  );

  togglePassword(): void {
    this.passwordVisible.update((visible) => !visible);
  }

  toggleConfirmation(): void {
    this.confirmationVisible.update((visible) => !visible);
  }

  /** El aviso de correo repetido solo aplica mientras el campo conserva ese mismo correo. */
  correoYaRegistrado(): boolean {
    const rechazado = this.correoRegistrado();
    return !!rechazado && this.form.controls.correo.value.trim().toLowerCase() === rechazado;
  }

  recoverPassword(): void {
    void this.router.navigate(['/recuperar-contrasena']);
  }

  switchToLogin(): void {
    this.transition.request('login');
  }

  submit(): void {
    this.error.set('');

    if (this.form.invalid || this.correoYaRegistrado()) {
      this.form.markAllAsTouched();
      return;
    }

    const value = this.form.getRawValue();
    this.loading.set(true);
    this.auth
      .register({
        nombres: value.nombres.trim(),
        primer_apellido: value.primerApellido.trim(),
        segundo_apellido: value.segundoApellido.trim(),
        correo: value.correo.trim(),
        telefono: value.telefono.trim(),
        contrasena: value.contrasena,
      })
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (response) => {
          sessionStorage.setItem('tienda.verification.challenge', response.desafio_id);
          this.feedback.success('Código enviado', 'Revisa tu correo para activar la cuenta.');
          void this.router.navigate(['/verificar-correo'], {
            queryParams: { desafio: response.desafio_id },
          });
        },
        error: (response: HttpErrorResponse) => {
          const challengeID = response.error?.desafio_id;
          if (response.status === 503 && challengeID) {
            sessionStorage.setItem('tienda.verification.challenge', challengeID);
            void this.router.navigate(['/verificar-correo'], {
              queryParams: { desafio: challengeID, envio: 'pendiente' },
            });
            return;
          }
          if (response.status === 409 && response.error?.codigo === 'CORREO_YA_REGISTRADO') {
            this.correoRegistrado.set(value.correo.trim().toLowerCase());
            this.form.controls.correo.markAsTouched();
            return;
          }
          this.error.set(response.error?.mensaje || 'No fue posible crear la cuenta.');
        },
      });
  }
}

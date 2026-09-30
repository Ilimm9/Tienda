import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import {
  AbstractControl,
  NonNullableFormBuilder,
  ReactiveFormsModule,
  ValidationErrors,
  Validators,
} from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { finalize } from 'rxjs';

import { FeedbackService } from '../../../shared/feedback/feedback.service';
import { AuthService } from '../auth.service';

function matchingPasswords(control: AbstractControl): ValidationErrors | null {
  const password = control.get('contrasena');
  const confirmation = control.get('confirmarContrasena');
  if (!password || !confirmation || !confirmation.value) {
    return null;
  }
  return password.value === confirmation.value ? null : { passwordsMismatch: true };
}

// bcrypt solo considera los primeros 72 bytes; el backend rechaza contraseñas más largas.
function maxUtf8Bytes(limit: number) {
  return (control: AbstractControl): ValidationErrors | null =>
    new TextEncoder().encode(control.value ?? '').length > limit ? { maxBytes: true } : null;
}

@Component({
  selector: 'app-reset-password',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './reset-password.component.html',
  styleUrl: './reset-password.component.css',
})
export class ResetPasswordComponent {
  private readonly formBuilder = inject(NonNullableFormBuilder);
  private readonly auth = inject(AuthService);
  private readonly feedback = inject(FeedbackService);
  private readonly router = inject(Router);

  readonly loading = signal(false);
  readonly error = signal('');
  readonly passwordVisible = signal(false);
  readonly form = this.formBuilder.group(
    {
      contrasena: ['', [Validators.required, Validators.minLength(8), maxUtf8Bytes(72)]],
      confirmarContrasena: ['', Validators.required],
    },
    { validators: matchingPasswords },
  );

  private readonly challengeID: string;
  private readonly token: string;
  readonly linkValid: boolean;

  constructor() {
    // El token viaja en el fragmento (#), que el navegador no envía al servidor ni en Referer.
    const params = new URLSearchParams(window.location.hash.replace(/^#/, ''));
    this.challengeID = params.get('desafio') ?? '';
    this.token = params.get('token') ?? '';
    this.linkValid = this.challengeID !== '' && this.token !== '';
    // Se retira el token de la barra de direcciones y del historial.
    history.replaceState(history.state, '', window.location.pathname);
  }

  togglePassword(): void {
    this.passwordVisible.update((visible) => !visible);
  }

  submit(): void {
    this.error.set('');
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.loading.set(true);
    this.auth
      .resetPassword({
        desafio_id: this.challengeID,
        token: this.token,
        contrasena: this.form.controls.contrasena.value,
      })
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (response) => {
          this.form.reset();
          this.feedback.success('Contraseña actualizada', response.mensaje);
          void this.router.navigate(['/login']);
        },
        error: (response: HttpErrorResponse) =>
          this.error.set(response.error?.mensaje || 'No fue posible actualizar la contraseña.'),
      });
  }
}

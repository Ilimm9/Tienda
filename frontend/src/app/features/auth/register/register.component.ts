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

import { environment } from '../../../../environments/environment';
import { AuthService } from '../auth.service';

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
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './register.component.html',
  styleUrl: './register.component.css',
})
export class RegisterComponent {
  private readonly formBuilder = inject(NonNullableFormBuilder);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  readonly loading = signal(false);
  readonly error = signal('');
  readonly passwordVisible = signal(false);
  readonly confirmationVisible = signal(false);
  readonly imageUrl = environment.loginHeaderImageUrl;
  readonly form = this.formBuilder.group(
    {
      nombreCompleto: ['', [Validators.required, Validators.minLength(3)]],
      correo: ['', [Validators.required, Validators.email]],
      telefono: '',
      contrasena: ['', [Validators.required, Validators.minLength(8)]],
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

  submit(): void {
    this.error.set('');

    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const value = this.form.getRawValue();
    this.loading.set(true);
    this.auth
      .register({
        nombre_completo: value.nombreCompleto,
        correo: value.correo,
        telefono: value.telefono || '',
        contrasena: value.contrasena,
      })
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: () => void this.router.navigate(['/login'], { queryParams: { registrado: '1' } }),
        error: (response: HttpErrorResponse) =>
          this.error.set(response.error?.mensaje || 'No fue posible crear la cuenta.'),
      });
  }
}

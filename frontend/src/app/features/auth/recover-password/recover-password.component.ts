import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { NonNullableFormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { finalize } from 'rxjs';

import { AuthService } from '../auth.service';

@Component({
  selector: 'app-recover-password',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './recover-password.component.html',
  styleUrl: './recover-password.component.css',
})
export class RecoverPasswordComponent {
  private readonly formBuilder = inject(NonNullableFormBuilder);
  private readonly auth = inject(AuthService);

  readonly loading = signal(false);
  readonly error = signal('');
  // La respuesta es uniforme: nunca confirma si el correo tiene cuenta.
  readonly sentMessage = signal('');
  readonly form = this.formBuilder.group({
    correo: ['', [Validators.required, Validators.email]],
  });

  submit(): void {
    this.error.set('');
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.loading.set(true);
    this.auth
      .requestPasswordReset({ correo: this.form.controls.correo.value })
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (response) => this.sentMessage.set(response.mensaje),
        error: (response: HttpErrorResponse) =>
          this.error.set(response.error?.mensaje || 'No fue posible procesar la solicitud.'),
      });
  }
}

import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnDestroy, inject, signal } from '@angular/core';
import { NonNullableFormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { finalize } from 'rxjs';

import { FeedbackService } from '../../../shared/feedback/feedback.service';
import { AuthService } from '../auth.service';

const CHALLENGE_STORAGE_KEY = 'tienda.verification.challenge';

@Component({
  selector: 'app-verification',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './verification.component.html',
  styleUrl: './verification.component.css',
})
export class VerificationComponent implements OnDestroy {
  private readonly formBuilder = inject(NonNullableFormBuilder);
  private readonly auth = inject(AuthService);
  private readonly feedback = inject(FeedbackService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private timer?: ReturnType<typeof setInterval>;

  readonly loading = signal(false);
  readonly resending = signal(false);
  readonly error = signal(
    this.route.snapshot.queryParamMap.get('envio') === 'pendiente'
      ? 'No pudimos entregar el primer código. Espera el contador y solicita uno nuevo.'
      : '',
  );
  readonly secondsUntilResend = signal(60);
  readonly form = this.formBuilder.group({
    codigo: ['', [Validators.required, Validators.pattern(/^\d{6}$/)]],
  });

  private challengeID =
    this.route.snapshot.queryParamMap.get('desafio') || sessionStorage.getItem(CHALLENGE_STORAGE_KEY) || '';

  constructor() {
    if (this.challengeID) {
      sessionStorage.setItem(CHALLENGE_STORAGE_KEY, this.challengeID);
    }
    this.startCountdown(60);
  }

  ngOnDestroy(): void {
    if (this.timer) clearInterval(this.timer);
  }

  submit(): void {
    this.error.set('');
    if (!this.challengeID) {
      this.error.set('No encontramos una verificación pendiente. Registra tu cuenta nuevamente.');
      return;
    }
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.loading.set(true);
    this.auth
      .verifyEmail({ desafio_id: this.challengeID, codigo: this.form.controls.codigo.value })
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: () => {
          sessionStorage.removeItem(CHALLENGE_STORAGE_KEY);
          this.feedback.success('Correo verificado', 'Tu cuenta ya está activa.');
          void this.router.navigate(['/inicio']);
        },
        error: (response: HttpErrorResponse) =>
          this.error.set(response.error?.mensaje || 'No fue posible verificar el código.'),
      });
  }

  resend(): void {
    if (!this.challengeID || this.secondsUntilResend() > 0 || this.resending()) return;
    this.error.set('');
    this.resending.set(true);
    this.auth
      .resendVerification({ desafio_id: this.challengeID })
      .pipe(finalize(() => this.resending.set(false)))
      .subscribe({
        next: (response) => {
          this.challengeID = response.desafio_id;
          sessionStorage.setItem(CHALLENGE_STORAGE_KEY, response.desafio_id);
          this.form.reset();
          this.startCountdown(response.reenviar_en_segundos);
          this.feedback.success('Código reenviado', 'Revisa nuevamente tu correo.');
        },
        error: (response: HttpErrorResponse) => {
          this.error.set(response.error?.mensaje || 'No fue posible reenviar el código.');
          if (response.status === 429) this.startCountdown(60);
        },
      });
  }

  private startCountdown(seconds: number): void {
    if (this.timer) clearInterval(this.timer);
    this.secondsUntilResend.set(Math.max(0, seconds));
    this.timer = setInterval(() => {
      this.secondsUntilResend.update((current) => Math.max(0, current - 1));
      if (this.secondsUntilResend() === 0 && this.timer) {
        clearInterval(this.timer);
        this.timer = undefined;
      }
    }, 1000);
  }
}

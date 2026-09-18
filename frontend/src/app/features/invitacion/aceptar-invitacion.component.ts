import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';

import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { InvitacionPublica } from '../equipo/invitacion.models';
import { InvitacionService } from '../equipo/invitacion.service';

@Component({
  selector: 'app-aceptar-invitacion',
  imports: [CommonModule, RouterLink],
  templateUrl: './aceptar-invitacion.component.html',
  styleUrl: './aceptar-invitacion.component.css',
})
export class AceptarInvitacionComponent {
  private readonly invitacionService = inject(InvitacionService);
  private readonly contexto = inject(ContextoService);
  private readonly feedback = inject(FeedbackService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  readonly token = this.route.snapshot.paramMap.get('token') ?? '';
  readonly invitacion = signal<InvitacionPublica | null>(null);
  readonly loading = signal(true);
  readonly processing = signal(false);
  readonly error = signal<string | null>(null);

  constructor() {
    this.load();
  }

  /** Conserva el token para volver aquí después de iniciar sesión o registrarse. */
  get retorno(): string {
    return `/invitacion/${this.token}`;
  }

  aceptar(): void {
    if (this.processing()) return;
    this.processing.set(true);
    this.invitacionService.aceptar(this.token).subscribe({
      next: () => {
        this.processing.set(false);
        this.feedback.success('Invitación aceptada', 'Ya formas parte del negocio.');
        // El nuevo negocio debe aparecer en el selector de contexto de inmediato.
        this.contexto.recargar().subscribe({
          next: () => void this.router.navigate(['/inicio']),
          error: () => void this.router.navigate(['/inicio']),
        });
      },
      error: (response: HttpErrorResponse) => {
        this.processing.set(false);
        if (response.status === 401) {
          this.error.set('Inicia sesión con el correo invitado para aceptar.');
          return;
        }
        this.error.set(
          (response.error as { mensaje?: string } | null)?.mensaje ??
            'No fue posible aceptar la invitación.',
        );
      },
    });
  }

  private load(): void {
    if (!this.token) {
      this.loading.set(false);
      this.error.set('El enlace de invitación no es válido.');
      return;
    }
    this.invitacionService.consultar(this.token).subscribe({
      next: (invitacion) => {
        this.invitacion.set(invitacion);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(
          response.status === 404
            ? 'Este enlace de invitación no existe.'
            : (response.error as { mensaje?: string } | null)?.mensaje ??
              'Este enlace ya no está vigente.',
        );
      },
    });
  }
}

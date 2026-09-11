import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';

import { FeedbackService } from '../../shared/feedback/feedback.service';
import { ApiErrorResponse, NegocioDetalle } from './negocio.models';
import { NegocioService } from './negocio.service';

@Component({
  selector: 'app-negocio-detalle',
  imports: [CommonModule, RouterLink],
  templateUrl: './negocio-detalle.component.html',
  styleUrl: './negocio-detalle.component.css',
})
export class NegocioDetalleComponent {
  private readonly negocioService = inject(NegocioService);
  private readonly feedback = inject(FeedbackService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  readonly negocioId = this.route.snapshot.paramMap.get('negocioId') ?? '';
  readonly business = signal<NegocioDetalle | null>(null);
  readonly loading = signal(true);
  readonly processing = signal(false);
  readonly error = signal<string | null>(null);

  constructor() {
    this.load();
  }

  async archive(): Promise<void> {
    const business = this.business();
    if (!business) return;

    const confirmed = await this.feedback.confirmDanger({
      titulo: `Archivar ${business.nombre_comercial}`,
      descripcion:
        'El negocio dejará de estar disponible para operar, pero sus datos se conservarán.',
      textoConfirmar: 'Sí, archivar',
    });
    if (!confirmed) return;

    this.processing.set(true);
    this.error.set(null);
    this.negocioService.archivar(business.id).subscribe({
      next: () => {
        this.processing.set(false);
        this.feedback.success('Negocio archivado');
        void this.router.navigate(['/negocios']);
      },
      error: (response: HttpErrorResponse) => {
        this.processing.set(false);
        this.feedback.error(
          'No fue posible archivar el negocio',
          this.errorMessage(response, 'Intenta nuevamente.'),
        );
      },
    });
  }

  restore(): void {
    const business = this.business();
    if (!business) return;
    this.processing.set(true);
    this.negocioService.restaurar(business.id).subscribe({
      next: (restored) => {
        this.business.set(restored);
        this.processing.set(false);
        this.feedback.success('Negocio restaurado');
      },
      error: (response: HttpErrorResponse) => {
        this.processing.set(false);
        this.feedback.error(
          'No fue posible restaurar el negocio',
          this.errorMessage(response, 'Intenta nuevamente.'),
        );
      },
    });
  }

  private load(): void {
    this.negocioService.obtener(this.negocioId).subscribe({
      next: (business) => {
        this.business.set(business);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.handleError(response, 'No fue posible cargar el negocio.');
      },
    });
  }

  private handleError(response: HttpErrorResponse, fallback: string): void {
    this.processing.set(false);
    this.error.set(this.errorMessage(response, fallback));
  }

  private errorMessage(response: HttpErrorResponse, fallback: string): string {
    return (response.error as ApiErrorResponse | null)?.mensaje ?? fallback;
  }
}

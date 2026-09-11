import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';

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

  archive(): void {
    const business = this.business();
    if (
      !business ||
      !window.confirm(`¿Archivar ${business.nombre_comercial}? Sus datos se conservarán.`)
    )
      return;
    this.processing.set(true);
    this.negocioService.archivar(business.id).subscribe({
      next: () => {
        this.processing.set(false);
        void this.router.navigate(['/negocios']);
      },
      error: (response: HttpErrorResponse) =>
        this.handleError(response, 'No fue posible archivar el negocio.'),
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
      },
      error: (response: HttpErrorResponse) =>
        this.handleError(response, 'No fue posible restaurar el negocio.'),
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
    this.error.set((response.error as ApiErrorResponse | null)?.mensaje ?? fallback);
  }
}

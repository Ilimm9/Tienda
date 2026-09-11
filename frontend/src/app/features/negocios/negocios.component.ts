import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';

import { ApiErrorResponse, NegocioResumen } from './negocio.models';
import { NegocioService } from './negocio.service';

@Component({
  selector: 'app-negocios',
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './negocios.component.html',
  styleUrl: './negocios.component.css',
})
export class NegociosComponent {
  private readonly negocioService = inject(NegocioService);

  readonly negocios = signal<NegocioResumen[]>([]);
  readonly loading = signal(true);
  readonly processingId = signal<string | null>(null);
  readonly error = signal<string | null>(null);
  readonly estado = signal<'activo' | 'archivado'>('activo');
  readonly search = signal('');
  readonly filteredBusinesses = computed(() => {
    const query = this.search().trim().toLocaleLowerCase('es-MX');
    if (!query) return this.negocios();
    return this.negocios().filter(
      (business) =>
        business.nombre_comercial.toLocaleLowerCase('es-MX').includes(query) ||
        business.rfc?.toLocaleLowerCase('es-MX').includes(query),
    );
  });

  constructor() {
    this.load();
  }

  setStatus(status: 'activo' | 'archivado'): void {
    if (status === this.estado()) return;
    this.estado.set(status);
    this.load();
  }

  updateSearch(value: string): void {
    this.search.set(value);
  }

  archive(business: NegocioResumen): void {
    if (!window.confirm(`¿Archivar ${business.nombre_comercial}? Sus datos se conservarán.`))
      return;
    this.processingId.set(business.id);
    this.error.set(null);
    this.negocioService.archivar(business.id).subscribe({
      next: () => {
        this.processingId.set(null);
        this.load();
      },
      error: (response: HttpErrorResponse) => {
        this.processingId.set(null);
        this.error.set(this.errorMessage(response, 'No fue posible archivar el negocio.'));
      },
    });
  }

  restore(business: NegocioResumen): void {
    this.processingId.set(business.id);
    this.error.set(null);
    this.negocioService.restaurar(business.id).subscribe({
      next: () => {
        this.processingId.set(null);
        this.load();
      },
      error: (response: HttpErrorResponse) => {
        this.processingId.set(null);
        this.error.set(this.errorMessage(response, 'No fue posible restaurar el negocio.'));
      },
    });
  }

  private load(): void {
    this.loading.set(true);
    this.error.set(null);
    this.negocioService.listar(this.estado()).subscribe({
      next: ({ items }) => {
        this.negocios.set(items ?? []);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(this.errorMessage(response, 'No fue posible cargar tus negocios.'));
      },
    });
  }

  private errorMessage(response: HttpErrorResponse, fallback: string): string {
    return (response.error as ApiErrorResponse | null)?.mensaje ?? fallback;
  }
}

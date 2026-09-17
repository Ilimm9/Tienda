import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { NegocioDetalle } from '../negocios/negocio.models';
import { NegocioService } from '../negocios/negocio.service';
import { SucursalApiError, SucursalResumen } from './sucursal.models';
import { SucursalService } from './sucursal.service';

@Component({
  selector: 'app-sucursales',
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './sucursales.component.html',
  styleUrl: './sucursales.component.css',
})
export class SucursalesComponent {
  private readonly sucursalService = inject(SucursalService);
  private readonly negocioService = inject(NegocioService);
  private readonly feedback = inject(FeedbackService);
  private readonly contexto = inject(ContextoService);
  private readonly route = inject(ActivatedRoute);

  readonly negocioId = this.route.snapshot.paramMap.get('negocioId') ?? '';
  readonly business = signal<NegocioDetalle | null>(null);
  readonly sucursales = signal<SucursalResumen[]>([]);
  readonly estado = signal<'activo' | 'archivado'>(
    this.route.snapshot.queryParamMap.get('estado') === 'archivado' ? 'archivado' : 'activo',
  );
  readonly buscar = signal('');
  readonly loading = signal(true);
  readonly processingId = signal<string | null>(null);
  readonly error = signal<string | null>(null);

  constructor() {
    this.load();
  }

  setStatus(status: 'activo' | 'archivado'): void {
    if (status === this.estado()) return;
    this.estado.set(status);
    this.load();
  }

  updateSearch(value: string): void {
    this.buscar.set(value);
  }

  search(): void {
    this.load();
  }

  clearSearch(): void {
    if (!this.buscar()) return;
    this.buscar.set('');
    this.load();
  }

  async archive(branch: SucursalResumen): Promise<void> {
    const confirmed = await this.feedback.confirmDanger({
      titulo: `Archivar ${branch.nombre}`,
      descripcion:
        'La sucursal dejará de estar disponible para operar, pero sus datos se conservarán.',
      textoConfirmar: 'Sí, archivar',
    });
    if (!confirmed) return;

    this.processingId.set(branch.id);
    this.sucursalService.archivar(this.negocioId, branch.id).subscribe({
      next: () => {
        this.processingId.set(null);
        this.feedback.success('Sucursal archivada');
        this.contexto.recargar().subscribe();
        this.load();
      },
      error: (response: HttpErrorResponse) => {
        this.processingId.set(null);
        this.feedback.error(
          'No fue posible archivar la sucursal',
          this.errorMessage(response, 'Intenta nuevamente.'),
        );
      },
    });
  }

  restore(branch: SucursalResumen): void {
    this.processingId.set(branch.id);
    this.sucursalService.restaurar(this.negocioId, branch.id).subscribe({
      next: () => {
        this.processingId.set(null);
        this.feedback.success('Sucursal restaurada');
        this.contexto.recargar().subscribe();
        this.load();
      },
      error: (response: HttpErrorResponse) => {
        this.processingId.set(null);
        this.feedback.error(
          'No fue posible restaurar la sucursal',
          this.errorMessage(response, 'Intenta nuevamente.'),
        );
      },
    });
  }

  private load(): void {
    this.loading.set(true);
    this.error.set(null);
    forkJoin({
      business: this.negocioService.obtener(this.negocioId),
      branches: this.sucursalService.listar(this.negocioId, this.estado(), this.buscar()),
    }).subscribe({
      next: ({ business, branches }) => {
        this.business.set(business);
        this.sucursales.set(branches.items ?? []);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(this.errorMessage(response, 'No fue posible cargar las sucursales.'));
      },
    });
  }

  private errorMessage(response: HttpErrorResponse, fallback: string): string {
    return (response.error as SucursalApiError | null)?.mensaje ?? fallback;
  }
}

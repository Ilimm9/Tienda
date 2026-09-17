import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { NegocioDetalle } from '../negocios/negocio.models';
import { NegocioService } from '../negocios/negocio.service';
import { SucursalApiError, SucursalDetalle } from './sucursal.models';
import { SucursalService } from './sucursal.service';

@Component({
  selector: 'app-sucursal-detalle',
  imports: [CommonModule, RouterLink],
  templateUrl: './sucursal-detalle.component.html',
  styleUrl: './sucursal-detalle.component.css',
})
export class SucursalDetalleComponent {
  private readonly sucursalService = inject(SucursalService);
  private readonly negocioService = inject(NegocioService);
  private readonly feedback = inject(FeedbackService);
  private readonly contexto = inject(ContextoService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  readonly negocioId = this.route.snapshot.paramMap.get('negocioId') ?? '';
  readonly sucursalId = this.route.snapshot.paramMap.get('sucursalId') ?? '';
  readonly business = signal<NegocioDetalle | null>(null);
  readonly branch = signal<SucursalDetalle | null>(null);
  readonly loading = signal(true);
  readonly processing = signal(false);
  readonly error = signal<string | null>(null);

  constructor() {
    this.load();
  }

  async archive(): Promise<void> {
    const branch = this.branch();
    if (!branch || this.processing()) return;
    const confirmed = await this.feedback.confirmDanger({
      titulo: `Archivar ${branch.nombre}`,
      descripcion:
        'La sucursal dejará de estar disponible para operar, pero sus datos se conservarán.',
      textoConfirmar: 'Sí, archivar',
    });
    if (!confirmed) return;

    this.processing.set(true);
    this.sucursalService.archivar(this.negocioId, this.sucursalId).subscribe({
      next: () => {
        this.processing.set(false);
        this.feedback.success('Sucursal archivada');
        this.contexto.recargar().subscribe();
        void this.router.navigate(['/negocios', this.negocioId, 'sucursales'], {
          queryParams: { estado: 'archivado' },
        });
      },
      error: (response: HttpErrorResponse) => {
        this.processing.set(false);
        this.feedback.error(
          'No fue posible archivar la sucursal',
          this.errorMessage(response, 'Intenta nuevamente.'),
        );
      },
    });
  }

  restore(): void {
    if (this.processing()) return;
    this.processing.set(true);
    this.sucursalService.restaurar(this.negocioId, this.sucursalId).subscribe({
      next: (branch) => {
        this.processing.set(false);
        this.branch.set(branch);
        this.feedback.success('Sucursal restaurada');
        this.contexto.recargar().subscribe();
      },
      error: (response: HttpErrorResponse) => {
        this.processing.set(false);
        this.feedback.error(
          'No fue posible restaurar la sucursal',
          this.errorMessage(response, 'Intenta nuevamente.'),
        );
      },
    });
  }

  fullAddress(branch: SucursalDetalle): string {
    const address = branch.direccion;
    if (!address) return 'Sin dirección registrada';
    return [
      [address.calle, address.numero_exterior].filter(Boolean).join(' '),
      address.numero_interior ? `Int. ${address.numero_interior}` : '',
      address.colonia,
      address.codigo_postal,
      address.ciudad,
      address.municipio,
      address.estado,
      address.codigo_pais,
    ]
      .filter(Boolean)
      .join(', ');
  }

  private load(): void {
    forkJoin({
      business: this.negocioService.obtener(this.negocioId),
      branch: this.sucursalService.obtener(this.negocioId, this.sucursalId),
    }).subscribe({
      next: ({ business, branch }) => {
        this.business.set(business);
        this.branch.set(branch);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(this.errorMessage(response, 'No fue posible cargar la sucursal.'));
      },
    });
  }

  private errorMessage(response: HttpErrorResponse, fallback: string): string {
    return (response.error as SucursalApiError | null)?.mensaje ?? fallback;
  }
}

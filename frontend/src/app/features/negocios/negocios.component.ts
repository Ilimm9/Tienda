import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';

import { FeedbackService } from '../../shared/feedback/feedback.service';
import { FechaMexicoPipe } from '../../shared/fecha-mexico.pipe';
import { ApiErrorResponse, NegocioResumen } from './negocio.models';
import { NegocioService } from './negocio.service';
import { ContextoService } from '../../contexto/contexto.service';
import { CodigoPermiso, PERMISOS } from '../../contexto/permisos';

@Component({
  selector: 'app-negocios',
  imports: [CommonModule, FechaMexicoPipe, FormsModule, RouterLink],
  templateUrl: './negocios.component.html',
  styleUrl: './negocios.component.css',
})
export class NegociosComponent {
  private readonly negocioService = inject(NegocioService);
  private readonly feedback = inject(FeedbackService);
  private readonly router = inject(Router);

  readonly negocios = signal<NegocioResumen[]>([]);
  readonly loading = signal(true);
  readonly processingId = signal<string | null>(null);
  readonly error = signal<string | null>(null);
  readonly estado = signal<'activo' | 'archivado'>('activo');
  readonly search = signal('');
  readonly estados = [
    { valor: 'activo', etiqueta: 'Activos' },
    { valor: 'archivado', etiqueta: 'Archivados' },
  ] as const;
  readonly filteredBusinesses = computed(() => {
    const query = this.search().trim().toLocaleLowerCase('es-MX');
    if (!query) return this.negocios();
    return this.negocios().filter(
      (business) =>
        business.nombre_comercial.toLocaleLowerCase('es-MX').includes(query) ||
        business.rfc?.toLocaleLowerCase('es-MX').includes(query),
    );
  });

  private readonly contexto = inject(ContextoService);
  readonly P = PERMISOS;
  readonly puedeCrear = this.contexto.puedeCrearNegocio;

  /** Un negocio archivado no está en el contexto: ahí decide ser propietario. */
  puedeEn(business: NegocioResumen, codigo: CodigoPermiso): boolean {
    return this.contexto.puedeEn(business.id, codigo, business.tipo_miembro === 'propietario');
  }

  constructor() {
    this.load();
  }

  /** Negocio que el contexto tiene abierto para operar. */
  esActual(business: NegocioResumen): boolean {
    return this.contexto.negocio()?.id === business.id;
  }

  /** Nombre de la sucursal principal según el contexto; los archivados no aparecen ahí. */
  principal(business: NegocioResumen): string {
    const negocio = this.contexto.negocios().find((item) => item.id === business.id);
    return negocio?.sucursales.find((item) => item.es_principal)?.nombre ?? '—';
  }

  iniciales(nombre: string): string {
    return (
      nombre
        .trim()
        .split(/\s+/)
        .map((palabra) => palabra[0])
        .join('')
        .slice(0, 2)
        .toUpperCase() || 'N'
    );
  }

  /** Cambia el contexto al negocio y abre su inicio, igual que el selector de la barra superior. */
  entrar(business: NegocioResumen): void {
    this.contexto.seleccionarNegocio(business.id);
    if (this.contexto.negocio()?.id !== business.id) return;
    this.feedback.info('Negocio activo', business.nombre_comercial);
    void this.router.navigate(['/inicio']);
  }

  setStatus(status: 'activo' | 'archivado'): void {
    if (status === this.estado()) return;
    this.estado.set(status);
    this.load();
  }

  updateSearch(value: string): void {
    this.search.set(value);
  }

  async archive(business: NegocioResumen): Promise<void> {
    const confirmed = await this.feedback.confirmDanger({
      titulo: `Archivar ${business.nombre_comercial}`,
      descripcion:
        'El negocio dejará de estar disponible para operar, pero sus datos se conservarán.',
      textoConfirmar: 'Sí, archivar',
    });
    if (!confirmed) return;

    this.processingId.set(business.id);
    this.error.set(null);
    this.negocioService.archivar(business.id).subscribe({
      next: () => {
        this.processingId.set(null);
        this.contexto.recargar().subscribe();
        this.feedback.success('Negocio archivado');
        this.load();
      },
      error: (response: HttpErrorResponse) => {
        this.processingId.set(null);
        this.feedback.error(
          'No fue posible archivar el negocio',
          this.errorMessage(response, 'Intenta nuevamente.'),
        );
      },
    });
  }

  restore(business: NegocioResumen): void {
    this.processingId.set(business.id);
    this.error.set(null);
    this.negocioService.restaurar(business.id).subscribe({
      next: () => {
        this.processingId.set(null);
        this.contexto.recargar().subscribe();
        this.feedback.success('Negocio restaurado');
        this.load();
      },
      error: (response: HttpErrorResponse) => {
        this.processingId.set(null);
        this.feedback.error(
          'No fue posible restaurar el negocio',
          this.errorMessage(response, 'Intenta nuevamente.'),
        );
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

import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { FechaMexicoPipe } from '../../shared/fecha-mexico.pipe';
import { AsignacionResumen } from './asignacion.models';
import { AsignacionService } from './asignacion.service';
import { EmpleadoDetalle } from './empleado.models';
import { EmpleadoService } from './empleado.service';

@Component({
  selector: 'app-empleado-sucursales',
  imports: [CommonModule, FechaMexicoPipe, FormsModule, RouterLink],
  templateUrl: './empleado-sucursales.component.html',
})
export class EmpleadoSucursalesComponent {
  private readonly asignacionService = inject(AsignacionService);
  private readonly empleadoService = inject(EmpleadoService);
  private readonly feedback = inject(FeedbackService);
  readonly contexto = inject(ContextoService);
  /** Permisos del negocio activo, resueltos una sola vez en el contexto. */
  readonly misPermisos = computed(() => [...this.contexto.permisos()]);
  private readonly route = inject(ActivatedRoute);

  readonly empleadoId = this.route.snapshot.paramMap.get('empleadoId') ?? '';
  readonly empleado = signal<EmpleadoDetalle | null>(null);
  readonly asignaciones = signal<AsignacionResumen[]>([]);
  readonly loading = signal(true);
  readonly processingId = signal<string | null>(null);
  readonly error = signal<string | null>(null);

  sucursalSeleccionada = '';
  comoPrincipal = false;

  readonly negocioId = computed(() => this.contexto.negocio()?.id ?? '');
  readonly puedeEditar = computed(() => this.misPermisos().includes('equipo.asignaciones.editar'));

  /** Solo ofrece sucursales activas del negocio que aún no estén asignadas. */
  readonly disponibles = computed(() => {
    const asignadas = new Set(this.asignaciones().map((item) => item.sucursal_id));
    return (this.contexto.negocio()?.sucursales ?? []).filter((sucursal) => !asignadas.has(sucursal.id));
  });

  constructor() {
    this.load();
  }

  asignar(): void {
    if (!this.sucursalSeleccionada || this.processingId()) return;
    this.processingId.set('nueva');
    this.asignacionService
      .asignar(this.negocioId(), this.empleadoId, {
        sucursal_id: this.sucursalSeleccionada,
        es_principal: this.comoPrincipal,
      })
      .subscribe({
        next: (respuesta) => {
          this.processingId.set(null);
          this.asignaciones.set(respuesta.items);
          this.sucursalSeleccionada = '';
          this.comoPrincipal = false;
          this.feedback.success('Sucursal asignada');
        },
        error: (response: HttpErrorResponse) => {
          this.processingId.set(null);
          this.feedback.error('No fue posible asignar la sucursal', this.mensaje(response));
        },
      });
  }

  marcarPrincipal(asignacion: AsignacionResumen): void {
    if (asignacion.es_principal || this.processingId()) return;
    this.processingId.set(asignacion.id);
    this.asignacionService
      .establecerPrincipal(this.negocioId(), this.empleadoId, asignacion.id)
      .subscribe({
        next: (respuesta) => {
          this.processingId.set(null);
          this.asignaciones.set(respuesta.items);
          this.feedback.success('Sucursal principal actualizada');
        },
        error: (response: HttpErrorResponse) => {
          this.processingId.set(null);
          this.feedback.error('No fue posible cambiar la principal', this.mensaje(response));
        },
      });
  }

  async retirar(asignacion: AsignacionResumen): Promise<void> {
    const confirmado = await this.feedback.confirmDanger({
      titulo: `Retirar ${asignacion.nombre_sucursal}`,
      descripcion: 'El empleado dejará de estar asignado, pero se conserva el historial.',
      textoConfirmar: 'Sí, retirar',
    });
    if (!confirmado) return;

    this.processingId.set(asignacion.id);
    this.asignacionService.finalizar(this.negocioId(), this.empleadoId, asignacion.id).subscribe({
      next: () => {
        this.processingId.set(null);
        this.feedback.success('Asignación retirada');
        this.load();
      },
      error: (response: HttpErrorResponse) => {
        this.processingId.set(null);
        this.feedback.error('No fue posible retirar la asignación', this.mensaje(response));
      },
    });
  }

  private load(): void {
    const negocioId = this.negocioId();
    if (!negocioId) {
      this.loading.set(false);
      this.error.set('Selecciona un negocio para administrar su equipo.');
      return;
    }
    this.loading.set(true);
    this.error.set(null);
    forkJoin({
      empleado: this.empleadoService.obtener(negocioId, this.empleadoId),
      asignaciones: this.asignacionService.listar(negocioId, this.empleadoId),
    }).subscribe({
      next: ({ empleado, asignaciones }) => {
        this.empleado.set(empleado);
        this.asignaciones.set(asignaciones.items);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(
          response.status === 404
            ? 'No fue posible encontrar el empleado solicitado.'
            : response.status === 403
              ? 'No tienes permiso para ver las asignaciones de este negocio.'
              : 'No fue posible cargar las asignaciones.',
        );
      },
    });
  }

  private mensaje(response: HttpErrorResponse): string {
    return (response.error as { mensaje?: string } | null)?.mensaje ?? 'Intenta nuevamente.';
  }
}

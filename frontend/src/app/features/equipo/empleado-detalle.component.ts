import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { FechaMexicoPipe } from '../../shared/fecha-mexico.pipe';
import { EmpleadoApiError, EmpleadoDetalle, EstadoEmpleado } from './empleado.models';
import { EmpleadoService } from './empleado.service';
import { avatarDeEstado, ETIQUETA_ESTADO_EMPLEADO, iniciales, PASTILLA_ESTADO_EMPLEADO } from './presentacion';

@Component({
  selector: 'app-empleado-detalle',
  imports: [CommonModule, FechaMexicoPipe, RouterLink],
  templateUrl: './empleado-detalle.component.html',
})
export class EmpleadoDetalleComponent {
  private readonly empleadoService = inject(EmpleadoService);
  private readonly feedback = inject(FeedbackService);
  private readonly contexto = inject(ContextoService);
  /** Permisos del negocio activo, resueltos una sola vez en el contexto. */
  readonly misPermisos = computed(() => [...this.contexto.permisos()]);
  private readonly route = inject(ActivatedRoute);

  readonly empleadoId = this.route.snapshot.paramMap.get('empleadoId') ?? '';
  readonly empleado = signal<EmpleadoDetalle | null>(null);
  readonly loading = signal(true);
  readonly processing = signal(false);
  readonly error = signal<string | null>(null);

  readonly iniciales = iniciales;
  readonly avatarDeEstado = avatarDeEstado;
  readonly pastillaEstado = PASTILLA_ESTADO_EMPLEADO;
  readonly etiquetaEstado = ETIQUETA_ESTADO_EMPLEADO;

  readonly puedeGestionar = computed(() => this.misPermisos().includes('equipo.empleados.gestionar'));
  /** Misma exigencia que la ruta de nueva invitación: emitir y leer empleados y roles. */
  readonly puedeInvitar = computed(() =>
    this.contexto.puede([PERMISOS.invitacionEnviar, PERMISOS.empleadoVer, PERMISOS.rolVer]),
  );
  readonly puedeVerAsignaciones = computed(() => this.contexto.puede(PERMISOS.asignacionVer));

  constructor() {
    this.load();
  }

  get negocioId(): string {
    return this.contexto.negocio()?.id ?? '';
  }

  cambiarEstado(estado: EstadoEmpleado): void {
    if (this.processing()) return;
    this.processing.set(true);
    this.empleadoService.actualizar(this.negocioId, this.empleadoId, { estado }).subscribe({
      next: (empleado) => {
        this.processing.set(false);
        this.empleado.set(empleado);
        this.feedback.success('Estado actualizado');
      },
      error: (response: HttpErrorResponse) => {
        this.processing.set(false);
        const apiError = response.error as EmpleadoApiError | null;
        this.feedback.error('No fue posible cambiar el estado', apiError?.mensaje ?? 'Intenta nuevamente.');
      },
    });
  }

  private load(): void {
    const negocioId = this.negocioId;
    if (!negocioId) {
      this.loading.set(false);
      this.error.set('Selecciona un negocio para administrar su equipo.');
      return;
    }
    forkJoin({
      empleado: this.empleadoService.obtener(negocioId, this.empleadoId),
    }).subscribe({
      next: ({ empleado }) => {
        this.empleado.set(empleado);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(
          response.status === 404
            ? 'No fue posible encontrar el empleado solicitado.'
            : response.status === 403
              ? 'No tienes permiso para ver el equipo de este negocio.'
              : 'No fue posible cargar el empleado.',
        );
      },
    });
  }
}

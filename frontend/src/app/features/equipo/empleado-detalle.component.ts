import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { RolService } from '../roles-permisos/rol.service';
import { EmpleadoApiError, EmpleadoDetalle, EstadoEmpleado } from './empleado.models';
import { EmpleadoService } from './empleado.service';

@Component({
  selector: 'app-empleado-detalle',
  imports: [CommonModule, RouterLink],
  templateUrl: './empleado-detalle.component.html',
  styleUrl: './empleado-detalle.component.css',
})
export class EmpleadoDetalleComponent {
  private readonly empleadoService = inject(EmpleadoService);
  private readonly rolService = inject(RolService);
  private readonly feedback = inject(FeedbackService);
  private readonly contexto = inject(ContextoService);
  private readonly route = inject(ActivatedRoute);

  readonly empleadoId = this.route.snapshot.paramMap.get('empleadoId') ?? '';
  readonly empleado = signal<EmpleadoDetalle | null>(null);
  readonly misPermisos = signal<string[]>([]);
  readonly loading = signal(true);
  readonly processing = signal(false);
  readonly error = signal<string | null>(null);

  readonly puedeGestionar = computed(() => this.misPermisos().includes('equipo.empleados.gestionar'));
  readonly puedeInvitar = computed(() => this.misPermisos().includes('equipo.invitaciones.enviar'));

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
      permisos: this.rolService.misPermisos(negocioId),
    }).subscribe({
      next: ({ empleado, permisos }) => {
        this.empleado.set(empleado);
        this.misPermisos.set(permisos.items);
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

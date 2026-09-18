import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TableModule } from 'primeng/table';
import { forkJoin } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { RolResumen } from '../roles-permisos/rol.models';
import { RolService } from '../roles-permisos/rol.service';
import { EmpleadoResumen } from './empleado.models';
import { EmpleadoService } from './empleado.service';
import { EstadoInvitacion, InvitacionResumen } from './invitacion.models';
import { InvitacionService } from './invitacion.service';

@Component({
  selector: 'app-invitaciones',
  imports: [CommonModule, FormsModule, TableModule],
  templateUrl: './invitaciones.component.html',
  styleUrl: './invitaciones.component.css',
})
export class InvitacionesComponent {
  private readonly invitacionService = inject(InvitacionService);
  private readonly empleadoService = inject(EmpleadoService);
  private readonly rolService = inject(RolService);
  private readonly feedback = inject(FeedbackService);
  readonly contexto = inject(ContextoService);

  readonly invitaciones = signal<InvitacionResumen[]>([]);
  readonly invitables = signal<EmpleadoResumen[]>([]);
  readonly roles = signal<RolResumen[]>([]);
  readonly misPermisos = signal<string[]>([]);
  readonly estado = signal<EstadoInvitacion | ''>('');
  readonly loading = signal(true);
  readonly processingId = signal<string | null>(null);
  readonly error = signal<string | null>(null);
  readonly enlaceGenerado = signal<string | null>(null);

  empleadoSeleccionado = '';
  rolSeleccionado = '';

  readonly negocioId = computed(() => this.contexto.negocio()?.id ?? '');
  readonly puedeEnviar = computed(() => this.misPermisos().includes('equipo.invitaciones.enviar'));

  readonly estados: { valor: EstadoInvitacion | ''; etiqueta: string }[] = [
    { valor: '', etiqueta: 'Todas' },
    { valor: 'pendiente', etiqueta: 'Pendientes' },
    { valor: 'aceptada', etiqueta: 'Aceptadas' },
    { valor: 'expirada', etiqueta: 'Expiradas' },
    { valor: 'cancelada', etiqueta: 'Canceladas' },
  ];

  constructor() {
    this.load();
  }

  setEstado(valor: EstadoInvitacion | ''): void {
    if (valor === this.estado()) return;
    this.estado.set(valor);
    this.load();
  }

  invitar(): void {
    if (!this.empleadoSeleccionado || this.processingId()) return;
    this.processingId.set('nueva');
    this.enlaceGenerado.set(null);
    this.invitacionService
      .crear(this.negocioId(), {
        empleado_id: this.empleadoSeleccionado,
        rol_predeterminado_id: this.rolSeleccionado || null,
      })
      .subscribe({
        next: (creada) => {
          this.processingId.set(null);
          this.enlaceGenerado.set(this.invitacionService.enlaceDeToken(creada.token));
          this.empleadoSeleccionado = '';
          this.rolSeleccionado = '';
          this.feedback.success('Invitación creada', 'Copia el enlace y compártelo con la persona.');
          this.load();
        },
        error: (response: HttpErrorResponse) => {
          this.processingId.set(null);
          this.feedback.error('No fue posible crear la invitación', this.mensaje(response));
        },
      });
  }

  nombreDeRol(rolId: string | null): string {
    if (!rolId) return 'Sin rol';
    return this.roles().find((rol) => rol.id === rolId)?.nombre ?? 'Rol retirado';
  }

  async copiar(enlace: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(enlace);
      this.feedback.success('Enlace copiado');
    } catch {
      this.feedback.warning('Copia el enlace manualmente');
    }
  }

  async cancelar(invitacion: InvitacionResumen): Promise<void> {
    const confirmado = await this.feedback.confirmDanger({
      titulo: `Cancelar invitación de ${invitacion.correo}`,
      descripcion: 'El enlace dejará de funcionar de inmediato.',
      textoConfirmar: 'Sí, cancelar',
      textoCancelar: 'Conservar',
    });
    if (!confirmado) return;

    this.processingId.set(invitacion.id);
    this.invitacionService.cancelar(this.negocioId(), invitacion.id).subscribe({
      next: () => {
        this.processingId.set(null);
        this.feedback.success('Invitación cancelada');
        this.load();
      },
      error: (response: HttpErrorResponse) => {
        this.processingId.set(null);
        this.feedback.error('No fue posible cancelar la invitación', this.mensaje(response));
      },
    });
  }

  private load(): void {
    const negocioId = this.negocioId();
    if (!negocioId) {
      this.loading.set(false);
      this.error.set('Selecciona un negocio para administrar sus invitaciones.');
      return;
    }
    this.loading.set(true);
    this.error.set(null);
    forkJoin({
      invitaciones: this.invitacionService.listar(negocioId, this.estado()),
      empleados: this.empleadoService.listar(negocioId),
      roles: this.rolService.listar(negocioId),
      permisos: this.rolService.misPermisos(negocioId),
    }).subscribe({
      next: ({ invitaciones, empleados, roles, permisos }) => {
        this.invitaciones.set(invitaciones.items);
        // Solo tiene sentido invitar a quien ya está registrado, tiene correo y aún no tiene cuenta.
        this.invitables.set(
          empleados.items.filter((empleado) => !empleado.tiene_cuenta && Boolean(empleado.correo)),
        );
        this.roles.set(roles.items);
        this.misPermisos.set(permisos.items);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(
          response.status === 403
            ? 'No tienes permiso para ver las invitaciones de este negocio.'
            : 'No fue posible cargar las invitaciones.',
        );
      },
    });
  }

  private mensaje(response: HttpErrorResponse): string {
    return (response.error as { mensaje?: string } | null)?.mensaje ?? 'Intenta nuevamente.';
  }
}

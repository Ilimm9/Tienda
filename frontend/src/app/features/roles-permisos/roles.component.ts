import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TableModule } from 'primeng/table';
import { forkJoin } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { MiembroRoles, RolApiError, RolResumen } from './rol.models';
import { RolService } from './rol.service';

@Component({
  selector: 'app-roles',
  imports: [CommonModule, RouterLink, TableModule],
  templateUrl: './roles.component.html',
  styleUrl: './roles.component.css',
})
export class RolesComponent {
  private readonly rolService = inject(RolService);
  private readonly feedback = inject(FeedbackService);
  readonly contexto = inject(ContextoService);

  readonly roles = signal<RolResumen[]>([]);
  readonly miembros = signal<MiembroRoles[]>([]);
  readonly misPermisos = signal<string[]>([]);
  readonly loading = signal(true);
  readonly processingId = signal<string | null>(null);
  readonly error = signal<string | null>(null);

  readonly negocioId = computed(() => this.contexto.negocio()?.id ?? '');
  readonly puedeGestionar = computed(() => this.misPermisos().includes('roles.gestionar'));
  readonly puedeAsignar = computed(() => this.misPermisos().includes('roles.asignar'));

  constructor() {
    this.load();
  }

  nombreDeRol(rolId: string): string {
    return this.roles().find((rol) => rol.id === rolId)?.nombre ?? 'Rol retirado';
  }

  async eliminar(rol: RolResumen): Promise<void> {
    if (this.processingId()) return;
    const confirmado = await this.feedback.confirmDanger({
      titulo: `Eliminar ${rol.nombre}`,
      descripcion: 'El rol dejará de estar disponible para asignarse a nuevos miembros.',
      textoConfirmar: 'Sí, eliminar',
    });
    if (!confirmado) return;

    this.processingId.set(rol.id);
    this.rolService.eliminar(this.negocioId(), rol.id).subscribe({
      next: () => {
        this.processingId.set(null);
        this.feedback.success('Rol eliminado');
        this.load();
      },
      error: (response: HttpErrorResponse) => {
        this.processingId.set(null);
        this.feedback.error('No fue posible eliminar el rol', this.mensajeError(response));
      },
    });
  }

  private load(): void {
    const negocioId = this.negocioId();
    if (!negocioId) {
      this.loading.set(false);
      this.error.set('Selecciona un negocio para administrar sus roles.');
      return;
    }
    this.loading.set(true);
    this.error.set(null);
    forkJoin({
      roles: this.rolService.listar(negocioId, true),
      miembros: this.rolService.miembros(negocioId),
      permisos: this.rolService.misPermisos(negocioId),
    }).subscribe({
      next: ({ roles, miembros, permisos }) => {
        this.roles.set(roles.items);
        this.miembros.set(miembros.items);
        this.misPermisos.set(permisos.items);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(
          response.status === 403
            ? 'No tienes permiso para ver los roles de este negocio.'
            : 'No fue posible cargar los roles.',
        );
      },
    });
  }

  private mensajeError(response: HttpErrorResponse): string {
    return (response.error as RolApiError | null)?.mensaje ?? 'Intenta nuevamente.';
  }
}

import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { MiembroRoles, RolApiError, RolResumen } from './rol.models';
import { RolService } from './rol.service';

@Component({
  selector: 'app-miembro-roles',
  imports: [CommonModule, RouterLink],
  templateUrl: './miembro-roles.component.html',
  styleUrl: './miembro-roles.component.css',
})
export class MiembroRolesComponent {
  private readonly rolService = inject(RolService);
  private readonly feedback = inject(FeedbackService);
  private readonly contexto = inject(ContextoService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  readonly membresiaId = this.route.snapshot.paramMap.get('membresiaId') ?? '';
  readonly roles = signal<RolResumen[]>([]);
  readonly miembro = signal<MiembroRoles | null>(null);
  readonly seleccionados = signal<Set<string>>(new Set());
  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);

  constructor() {
    this.load();
  }

  get negocioId(): string {
    return this.contexto.negocio()?.id ?? '';
  }

  estaSeleccionado(rolId: string): boolean {
    return this.seleccionados().has(rolId);
  }

  alternar(rolId: string): void {
    const copia = new Set(this.seleccionados());
    if (copia.has(rolId)) copia.delete(rolId);
    else copia.add(rolId);
    this.seleccionados.set(copia);
  }

  guardar(): void {
    if (this.saving()) return;
    this.saving.set(true);
    this.error.set(null);
    this.rolService.asignarRoles(this.negocioId, this.membresiaId, [...this.seleccionados()]).subscribe({
      next: () => {
        this.saving.set(false);
        this.feedback.success('Roles actualizados');
        void this.router.navigate(['/roles-permisos']);
      },
      error: (response: HttpErrorResponse) => {
        this.saving.set(false);
        const apiError = response.error as RolApiError | null;
        this.error.set(apiError?.mensaje ?? 'No fue posible actualizar los roles.');
      },
    });
  }

  private load(): void {
    const negocioId = this.negocioId;
    if (!negocioId) {
      this.loading.set(false);
      this.error.set('Selecciona un negocio para administrar sus roles.');
      return;
    }
    forkJoin({
      roles: this.rolService.listar(negocioId, false),
      miembros: this.rolService.miembros(negocioId),
    }).subscribe({
      next: ({ roles, miembros }) => {
        this.roles.set(roles.items);
        const actual = miembros.items.find((item) => item.membresia_id === this.membresiaId) ?? null;
        this.miembro.set(actual);
        this.seleccionados.set(new Set(actual?.roles ?? []));
        this.loading.set(false);
        if (!actual) this.error.set('No fue posible encontrar la membresía solicitada.');
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(
          response.status === 403
            ? 'No tienes permiso para asignar roles en este negocio.'
            : 'No fue posible cargar la membresía.',
        );
      },
    });
  }
}

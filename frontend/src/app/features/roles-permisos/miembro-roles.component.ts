import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, ElementRef, inject, input, OnInit, output, signal, viewChild } from '@angular/core';
import { RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { enfocarPrimerInvalido } from '../../shared/formularios/formulario';
import { ModalComponent } from '../../shared/ui/modal/modal.component';
import { iniciales } from '../equipo/presentacion';
import { MiembroRoles, RolApiError, RolResumen } from './rol.models';
import { RolService } from './rol.service';

/** Ventana para elegir los roles de un miembro; se abre encima de Roles y permisos. */
@Component({
  selector: 'app-miembro-roles',
  imports: [CommonModule, RouterLink, ModalComponent],
  templateUrl: './miembro-roles.component.html',
})
export class MiembroRolesComponent implements OnInit {
  private readonly rolService = inject(RolService);
  private readonly feedback = inject(FeedbackService);
  private readonly contexto = inject(ContextoService);
  private readonly host = inject<ElementRef<HTMLElement>>(ElementRef);
  private readonly modal = viewChild(ModalComponent);

  readonly membresiaId = input.required<string>();
  readonly cerrar = output<void>();
  readonly guardado = output<void>();

  readonly roles = signal<RolResumen[]>([]);
  readonly miembro = signal<MiembroRoles | null>(null);
  readonly seleccionados = signal<Set<string>>(new Set());
  private readonly originales = signal<Set<string>>(new Set());
  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly intentado = signal(false);
  readonly sinRoles = computed(() => this.intentado() && this.seleccionados().size === 0);
  readonly cambiado = computed(() => {
    const actual = this.seleccionados();
    const antes = this.originales();
    return actual.size !== antes.size || [...actual].some((id) => !antes.has(id));
  });
  readonly iniciales = iniciales;

  ngOnInit(): void {
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

  cancelar(): void {
    const modal = this.modal();
    if (modal) void modal.solicitarCierre();
    else this.cerrar.emit();
  }

  guardar(): void {
    this.intentado.set(true);
    if (this.seleccionados().size === 0) {
      enfocarPrimerInvalido(this.host.nativeElement);
      return;
    }
    if (this.saving()) return;
    this.saving.set(true);
    this.error.set(null);
    this.rolService.asignarRoles(this.negocioId, this.membresiaId(), [...this.seleccionados()]).subscribe({
      next: () => {
        this.saving.set(false);
        this.contexto.recargar().subscribe();
        const persona = this.miembro();
        this.feedback.success('Roles actualizados', persona ? `${persona.nombre || persona.correo} ya tiene sus nuevos permisos.` : undefined);
        this.guardado.emit();
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
        const actual = miembros.items.find((item) => item.membresia_id === this.membresiaId()) ?? null;
        this.miembro.set(actual);
        this.seleccionados.set(new Set(actual?.roles ?? []));
        this.originales.set(new Set(actual?.roles ?? []));
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

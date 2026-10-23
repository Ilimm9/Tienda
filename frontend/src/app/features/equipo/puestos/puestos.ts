import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';

import { ContextoService } from '../../../contexto/contexto.service';
import { PERMISOS } from '../../../contexto/permisos';
import { FeedbackService } from '../../../shared/feedback/feedback.service';
import { CREAR_NUEVO } from '../presentacion';
import { RolResumen } from '../../roles-permisos/rol.models';
import { RolService } from '../../roles-permisos/rol.service';
import { ActualizarPuestoPayload, Puesto } from '../puesto.models';
import { PuestoService } from '../puesto.service';

@Component({
  selector: 'app-puestos',
  imports: [FormsModule],
  templateUrl: './puestos.html',
})
export class Puestos {
  private readonly puestoService = inject(PuestoService);
  private readonly rolService = inject(RolService);
  private readonly feedback = inject(FeedbackService);
  readonly contexto = inject(ContextoService);

  readonly puestos = signal<Puesto[]>([]);
  readonly roles = signal<RolResumen[]>([]);
  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly nuevo = signal('');
  readonly nuevoRol = signal('');
  readonly errorNuevo = signal<string | null>(null);

  readonly negocioId = computed(() => this.contexto.negocio()?.id ?? '');
  readonly puedeGestionar = computed(() => this.contexto.puede(PERMISOS.empleadoGestionar));
  readonly puedeVerRoles = computed(() => this.contexto.puede(PERMISOS.rolVer));
  readonly puedeCrearRol = computed(() => this.contexto.puede(PERMISOS.rolGestionar));
  readonly CREAR_NUEVO = CREAR_NUEVO;
  private readonly router = inject(Router);

  /** «+ Crear nuevo rol…» lleva a Roles y permisos. */
  elegirRol(valor: string): void {
    if (valor === CREAR_NUEVO) void this.router.navigate(['/roles-permisos/nuevo']);
    else this.nuevoRol.set(valor);
  }
  /** Aviso inmediato; el servidor vuelve a comprobarlo sin distinguir mayúsculas. */
  readonly duplicado = computed(() => {
    const nombre = this.nuevo().trim().toLowerCase();
    return Boolean(nombre) && this.puestos().some((puesto) => puesto.nombre.toLowerCase() === nombre);
  });

  constructor() {
    this.load();
  }

  nombreRol(id: string | null): string {
    return this.roles().find((rol) => rol.id === id)?.nombre ?? 'Sin rol sugerido';
  }

  agregar(): void {
    const nombre = this.nuevo().trim();
    this.errorNuevo.set(null);
    if (nombre.length < 2) {
      this.errorNuevo.set('Usa entre 2 y 120 caracteres.');
      return;
    }
    if (this.duplicado() || this.saving()) return;
    this.saving.set(true);
    this.puestoService.crear(this.negocioId(), { nombre, rol_sugerido_id: this.nuevoRol() || null }).subscribe({
      next: (puesto) => {
        this.saving.set(false);
        this.puestos.update((items) => [...items, puesto].sort((a, b) => a.nombre.localeCompare(b.nombre)));
        this.nuevo.set('');
        this.nuevoRol.set('');
        this.feedback.success('Puesto agregado');
      },
      error: (response: HttpErrorResponse) => {
        this.saving.set(false);
        this.errorNuevo.set(this.mensaje(response));
      },
    });
  }

  actualizar(puesto: Puesto, cambios: ActualizarPuestoPayload): void {
    if (cambios.nombre !== undefined) {
      cambios.nombre = cambios.nombre.trim();
      if (cambios.nombre === puesto.nombre) return;
    }
    this.puestoService.actualizar(this.negocioId(), puesto.id, cambios).subscribe({
      next: (actualizado) => {
        this.puestos.update((items) => items.map((item) => (item.id === actualizado.id ? actualizado : item)));
        this.feedback.success('Puesto actualizado');
      },
      error: (response: HttpErrorResponse) => {
        // Se recarga para descartar lo escrito en la fila rechazada.
        this.feedback.error('No fue posible actualizar el puesto', this.mensaje(response));
        this.load();
      },
    });
  }

  private load(): void {
    const negocioId = this.negocioId();
    if (!negocioId) {
      this.loading.set(false);
      this.error.set('Selecciona un negocio para administrar sus puestos.');
      return;
    }
    this.puestoService.listar(negocioId).subscribe({
      next: ({ items }) => {
        this.puestos.set(items);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(
          response.status === 403 ? 'No tienes permiso para ver los puestos.' : 'No fue posible cargar los puestos.',
        );
      },
    });
    if (this.puedeVerRoles()) {
      this.rolService.listar(negocioId).subscribe({
        next: ({ items }) => this.roles.set(items.filter((rol) => rol.activo && rol.codigo.toUpperCase() !== 'PROPIETARIO')),
      });
    }
  }

  private mensaje(response: HttpErrorResponse): string {
    const cuerpo = response.error as { mensaje?: string; campos?: Record<string, string> } | null;
    return cuerpo?.campos?.['nombre'] ? `El nombre ${cuerpo.campos['nombre']}.` : (cuerpo?.mensaje ?? 'Intenta nuevamente.');
  }
}

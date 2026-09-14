import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { TableModule } from 'primeng/table';
import { forkJoin } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { RolService } from '../roles-permisos/rol.service';
import { EmpleadoResumen, EstadoEmpleado } from './empleado.models';
import { EmpleadoService } from './empleado.service';

@Component({
  selector: 'app-empleados',
  imports: [CommonModule, FormsModule, RouterLink, TableModule],
  templateUrl: './empleados.component.html',
  styleUrl: './empleados.component.css',
})
export class EmpleadosComponent {
  private readonly empleadoService = inject(EmpleadoService);
  private readonly rolService = inject(RolService);
  readonly contexto = inject(ContextoService);

  readonly empleados = signal<EmpleadoResumen[]>([]);
  readonly misPermisos = signal<string[]>([]);
  readonly estado = signal<EstadoEmpleado | ''>('');
  readonly buscar = signal('');
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);

  readonly negocioId = computed(() => this.contexto.negocio()?.id ?? '');
  readonly puedeGestionar = computed(() => this.misPermisos().includes('equipo.empleados.gestionar'));

  readonly estados: { valor: EstadoEmpleado | ''; etiqueta: string }[] = [
    { valor: '', etiqueta: 'Todos' },
    { valor: 'pendiente', etiqueta: 'Pendientes' },
    { valor: 'activo', etiqueta: 'Activos' },
    { valor: 'suspendido', etiqueta: 'Suspendidos' },
    { valor: 'terminado', etiqueta: 'Terminados' },
  ];

  constructor() {
    this.load();
  }

  setEstado(valor: EstadoEmpleado | ''): void {
    if (valor === this.estado()) return;
    this.estado.set(valor);
    this.load();
  }

  updateSearch(valor: string): void {
    this.buscar.set(valor);
  }

  search(): void {
    this.load();
  }

  clearSearch(): void {
    if (!this.buscar()) return;
    this.buscar.set('');
    this.load();
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
      empleados: this.empleadoService.listar(negocioId, this.estado(), this.buscar()),
      permisos: this.rolService.misPermisos(negocioId),
    }).subscribe({
      next: ({ empleados, permisos }) => {
        this.empleados.set(empleados.items);
        this.misPermisos.set(permisos.items);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(
          response.status === 403
            ? 'No tienes permiso para ver el equipo de este negocio.'
            : 'No fue posible cargar los empleados.',
        );
      },
    });
  }
}

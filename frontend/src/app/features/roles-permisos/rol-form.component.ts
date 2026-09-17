import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, DestroyRef, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { forkJoin, of } from 'rxjs';

import { CambiosPendientesService } from '../../contexto/cambios-pendientes.service';
import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { Permiso, RolApiError, RolDetalle } from './rol.models';
import { RolService } from './rol.service';

interface ModuloPermisos {
  codigo: string;
  nombre: string;
  permisos: Permiso[];
}

const nombresDeModulo: Record<string, string> = {
  negocios: 'Negocios',
  sucursales: 'Sucursales',
  equipo: 'Equipo',
  roles: 'Roles y permisos',
  catalogo: 'Catálogo',
  ventas: 'Ventas',
};

@Component({
  selector: 'app-rol-form',
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './rol-form.component.html',
  styleUrl: './rol-form.component.css',
})
export class RolFormComponent {
  private readonly rolService = inject(RolService);
  private readonly feedback = inject(FeedbackService);
  private readonly contexto = inject(ContextoService);
  private readonly cambiosPendientes = inject(CambiosPendientesService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly fb = inject(FormBuilder);

  readonly rolId = this.route.snapshot.paramMap.get('rolId') ?? '';
  readonly editing = Boolean(this.rolId);
  readonly modulos = signal<ModuloPermisos[]>([]);
  readonly seleccionados = signal<Set<string>>(new Set());
  readonly rol = signal<RolDetalle | null>(null);
  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly fieldErrors = signal<Record<string, string>>({});

  readonly form = this.fb.nonNullable.group({
    codigo: ['', [Validators.required, Validators.minLength(2), Validators.maxLength(60)]],
    nombre: ['', [Validators.required, Validators.minLength(2), Validators.maxLength(120)]],
    descripcion: [''],
    activo: [true],
  });

  constructor() {
    this.load();
    const baja = this.cambiosPendientes.registrar(() => this.form.dirty && !this.saving());
    inject(DestroyRef).onDestroy(baja);
  }

  get negocioId(): string {
    return this.contexto.negocio()?.id ?? '';
  }

  nombreModulo(codigo: string): string {
    return nombresDeModulo[codigo] ?? codigo;
  }

  estaSeleccionado(permisoId: string): boolean {
    return this.seleccionados().has(permisoId);
  }

  alternarPermiso(permisoId: string): void {
    const copia = new Set(this.seleccionados());
    if (copia.has(permisoId)) copia.delete(permisoId);
    else copia.add(permisoId);
    this.seleccionados.set(copia);
    this.form.markAsDirty();
  }

  alternarModulo(modulo: ModuloPermisos): void {
    const copia = new Set(this.seleccionados());
    const todos = modulo.permisos.every((permiso) => copia.has(permiso.id));
    for (const permiso of modulo.permisos) {
      if (todos) copia.delete(permiso.id);
      else copia.add(permiso.id);
    }
    this.seleccionados.set(copia);
    this.form.markAsDirty();
  }

  moduloCompleto(modulo: ModuloPermisos): boolean {
    return modulo.permisos.length > 0 && modulo.permisos.every((permiso) => this.estaSeleccionado(permiso.id));
  }

  submit(): void {
    this.form.markAllAsTouched();
    this.error.set(null);
    this.fieldErrors.set({});
    if (this.form.invalid || this.saving()) return;

    const value = this.form.getRawValue();
    const permisos = [...this.seleccionados()];
    this.saving.set(true);
    const request = this.editing
      ? this.rolService.actualizar(this.negocioId, this.rolId, {
          nombre: value.nombre.trim(),
          descripcion: value.descripcion.trim() || null,
          activo: value.activo,
          permisos,
        })
      : this.rolService.crear(this.negocioId, {
          codigo: value.codigo.trim().toUpperCase(),
          nombre: value.nombre.trim(),
          descripcion: value.descripcion.trim() || null,
          permisos,
        });

    request.subscribe({
      next: () => {
        this.saving.set(false);
        this.form.markAsPristine();
        this.feedback.success(this.editing ? 'Rol actualizado' : 'Rol registrado');
        void this.router.navigate(['/roles-permisos']);
      },
      error: (response: HttpErrorResponse) => {
        this.saving.set(false);
        const apiError = response.error as RolApiError | null;
        this.error.set(apiError?.mensaje ?? 'No fue posible guardar el rol.');
        this.fieldErrors.set(apiError?.campos ?? {});
      },
    });
  }

  hasError(control: 'codigo' | 'nombre'): boolean {
    const field = this.form.controls[control];
    return (field.touched && field.invalid) || Boolean(this.fieldErrors()[control]);
  }

  private load(): void {
    const negocioId = this.negocioId;
    if (!negocioId) {
      this.loading.set(false);
      this.error.set('Selecciona un negocio para administrar sus roles.');
      return;
    }
    forkJoin({
      permisos: this.rolService.permisos(negocioId),
      rol: this.editing ? this.rolService.obtener(negocioId, this.rolId) : of(null),
    }).subscribe({
      next: ({ permisos, rol }) => {
        this.modulos.set(this.agruparPorModulo(permisos.items));
        if (rol) {
          this.rol.set(rol);
          this.form.patchValue({
            codigo: rol.codigo,
            nombre: rol.nombre,
            descripcion: rol.descripcion ?? '',
            activo: rol.activo,
          });
          this.form.controls.codigo.disable();
          this.seleccionados.set(new Set(rol.permisos));
        }
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(
          response.status === 403
            ? 'No tienes permiso para administrar roles en este negocio.'
            : 'No fue posible cargar el rol.',
        );
      },
    });
  }

  private agruparPorModulo(permisos: Permiso[]): ModuloPermisos[] {
    const modulos = new Map<string, ModuloPermisos>();
    for (const permiso of permisos) {
      const actual = modulos.get(permiso.codigo_modulo) ?? {
        codigo: permiso.codigo_modulo,
        nombre: this.nombreModulo(permiso.codigo_modulo),
        permisos: [],
      };
      actual.permisos.push(permiso);
      modulos.set(permiso.codigo_modulo, actual);
    }
    return [...modulos.values()];
  }
}

import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, DestroyRef, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';

import { CambiosPendientesService } from '../../contexto/cambios-pendientes.service';
import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { EmpleadoApiError, EmpleadoDetalle } from './empleado.models';
import { EmpleadoService } from './empleado.service';

@Component({
  selector: 'app-empleado-form',
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './empleado-form.component.html',
  styleUrl: './empleado-form.component.css',
})
export class EmpleadoFormComponent {
  private readonly empleadoService = inject(EmpleadoService);
  private readonly feedback = inject(FeedbackService);
  private readonly contexto = inject(ContextoService);
  private readonly cambiosPendientes = inject(CambiosPendientesService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly fb = inject(FormBuilder);

  readonly empleadoId = this.route.snapshot.paramMap.get('empleadoId') ?? '';
  readonly editing = Boolean(this.empleadoId);
  readonly empleado = signal<EmpleadoDetalle | null>(null);
  readonly loading = signal(this.editing);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly fieldErrors = signal<Record<string, string>>({});

  readonly form = this.fb.nonNullable.group({
    numero_empleado: [''],
    nombre: ['', [Validators.required, Validators.minLength(2), Validators.maxLength(100)]],
    segundo_nombre: [''],
    primer_apellido: ['', [Validators.required, Validators.minLength(2), Validators.maxLength(100)]],
    segundo_apellido: [''],
    correo: [''],
    telefono: [''],
    puesto: [''],
    contratado_en: [''],
  });

  constructor() {
    if (this.editing) this.load();
    const baja = this.cambiosPendientes.registrar(() => this.form.dirty && !this.saving());
    inject(DestroyRef).onDestroy(baja);
  }

  get negocioId(): string {
    return this.contexto.negocio()?.id ?? '';
  }

  hasError(control: 'nombre' | 'primer_apellido' | 'numero_empleado' | 'correo'): boolean {
    const field = this.form.controls[control];
    return (field.touched && field.invalid) || Boolean(this.fieldErrors()[control]);
  }

  submit(): void {
    this.form.markAllAsTouched();
    this.error.set(null);
    this.fieldErrors.set({});
    if (this.form.invalid || this.saving()) return;

    const value = this.form.getRawValue();
    this.saving.set(true);
    const request = this.editing
      ? this.empleadoService.actualizar(this.negocioId, this.empleadoId, {
          numero_empleado: this.opcional(value.numero_empleado),
          nombre: value.nombre.trim(),
          segundo_nombre: this.opcional(value.segundo_nombre),
          primer_apellido: value.primer_apellido.trim(),
          segundo_apellido: this.opcional(value.segundo_apellido),
          correo: this.opcional(value.correo),
          telefono: this.opcional(value.telefono),
          puesto: this.opcional(value.puesto),
          contratado_en: this.opcional(value.contratado_en),
        })
      : this.empleadoService.crear(this.negocioId, {
          numero_empleado: this.opcional(value.numero_empleado),
          nombre: value.nombre.trim(),
          segundo_nombre: this.opcional(value.segundo_nombre),
          primer_apellido: value.primer_apellido.trim(),
          segundo_apellido: this.opcional(value.segundo_apellido),
          correo: this.opcional(value.correo),
          telefono: this.opcional(value.telefono),
          puesto: this.opcional(value.puesto),
          contratado_en: this.opcional(value.contratado_en),
        });

    request.subscribe({
      next: (empleado) => {
        this.saving.set(false);
        this.form.markAsPristine();
        this.feedback.success(this.editing ? 'Empleado actualizado' : 'Empleado registrado');
        void this.router.navigate(['/equipo/empleados', empleado.id]);
      },
      error: (response: HttpErrorResponse) => {
        this.saving.set(false);
        const apiError = response.error as EmpleadoApiError | null;
        this.error.set(apiError?.mensaje ?? 'No fue posible guardar el empleado.');
        this.fieldErrors.set(apiError?.campos ?? {});
      },
    });
  }

  private opcional(valor: string): string | null {
    const limpio = valor.trim();
    return limpio === '' ? null : limpio;
  }

  private load(): void {
    const negocioId = this.negocioId;
    if (!negocioId) {
      this.loading.set(false);
      this.error.set('Selecciona un negocio para administrar su equipo.');
      return;
    }
    this.empleadoService.obtener(negocioId, this.empleadoId).subscribe({
      next: (empleado) => {
        this.empleado.set(empleado);
        this.form.patchValue({
          numero_empleado: empleado.numero_empleado ?? '',
          nombre: empleado.nombre,
          segundo_nombre: empleado.segundo_nombre ?? '',
          primer_apellido: empleado.primer_apellido,
          segundo_apellido: empleado.segundo_apellido ?? '',
          correo: empleado.correo ?? '',
          telefono: empleado.telefono ?? '',
          puesto: empleado.puesto ?? '',
          contratado_en: empleado.contratado_en?.slice(0, 10) ?? '',
        });
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(
          response.status === 403
            ? 'No tienes permiso para administrar el equipo de este negocio.'
            : 'No fue posible cargar el empleado.',
        );
      },
    });
  }
}

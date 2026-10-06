import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, DestroyRef, ElementRef, computed, inject, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';

import { CambiosPendientesService } from '../../contexto/cambios-pendientes.service';
import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { enfocarPrimerInvalido, fechaNoFutura, sinSoloEspacios } from '../../shared/formularios/formulario';
import { EmpleadoApiError, EmpleadoDetalle } from './empleado.models';
import { EmpleadoService } from './empleado.service';

type CampoEmpleado =
  | 'nombre' | 'segundo_nombre' | 'primer_apellido' | 'segundo_apellido'
  | 'correo' | 'telefono' | 'puesto' | 'numero_empleado' | 'contratado_en';

const MENSAJES: Record<CampoEmpleado, string> = {
  nombre: 'Usa entre 2 y 100 caracteres.',
  segundo_nombre: 'Usa hasta 100 caracteres.',
  primer_apellido: 'Usa entre 2 y 100 caracteres.',
  segundo_apellido: 'Usa hasta 100 caracteres.',
  correo: 'Escribe un correo válido, como nombre@correo.mx.',
  telefono: 'Usa de 7 a 30 dígitos; se permiten espacios, guiones, paréntesis y +.',
  puesto: 'Usa hasta 120 caracteres.',
  numero_empleado: 'Usa hasta 40 caracteres: letras, números, guion o guion bajo.',
  contratado_en: 'La fecha no puede ser posterior a hoy.',
};

@Component({
  selector: 'app-empleado-form',
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './empleado-form.component.html',
})
export class EmpleadoFormComponent {
  private readonly empleadoService = inject(EmpleadoService);
  private readonly feedback = inject(FeedbackService);
  private readonly contexto = inject(ContextoService);
  private readonly cambiosPendientes = inject(CambiosPendientesService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly fb = inject(FormBuilder);
  private readonly host = inject<ElementRef<HTMLElement>>(ElementRef);

  readonly empleadoId = this.route.snapshot.paramMap.get('empleadoId') ?? '';
  readonly editing = Boolean(this.empleadoId);
  readonly empleado = signal<EmpleadoDetalle | null>(null);
  readonly loading = signal(this.editing);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly fieldErrors = signal<Record<string, string>>({});

  readonly form = this.fb.nonNullable.group({
    // Los límites repiten los del backend (`empleado_service.go`) para avisar antes de enviar.
    numero_empleado: ['', Validators.pattern(/^\s*[A-Za-z0-9][A-Za-z0-9_-]{0,39}\s*$/)],
    nombre: ['', [Validators.required, sinSoloEspacios, Validators.minLength(2), Validators.maxLength(100)]],
    segundo_nombre: ['', Validators.maxLength(100)],
    primer_apellido: ['', [Validators.required, sinSoloEspacios, Validators.minLength(2), Validators.maxLength(100)]],
    segundo_apellido: ['', Validators.maxLength(100)],
    correo: ['', [Validators.email, Validators.maxLength(254)]],
    telefono: ['', Validators.pattern(/^\s*[0-9+()\s-]{7,30}\s*$/)],
    puesto: ['', Validators.maxLength(120)],
    contratado_en: ['', fechaNoFutura],
  });

  /** Valores en vivo para la vista previa de la tarjeta lateral. */
  private readonly valores = toSignal(this.form.valueChanges, { initialValue: this.form.getRawValue() });
  readonly nombreVista = computed(() => {
    const v = this.valores();
    const partes = [v.nombre, v.segundo_nombre, v.primer_apellido, v.segundo_apellido]
      .map((parte) => (parte ?? '').trim())
      .filter(Boolean);
    return partes.join(' ');
  });
  readonly inicialesVista = computed(() => {
    const v = this.valores();
    const letras = [v.nombre, v.primer_apellido].map((parte) => (parte ?? '').trim().charAt(0)).join('');
    return letras.toUpperCase() || '?';
  });
  readonly puestoVista = computed(() => (this.valores().puesto ?? '').trim());
  readonly correoVista = computed(() => (this.valores().correo ?? '').trim());
  readonly telefonoVista = computed(() => (this.valores().telefono ?? '').trim());
  readonly numeroVista = computed(() => (this.valores().numero_empleado ?? '').trim());

  constructor() {
    if (this.editing) this.load();
    const baja = this.cambiosPendientes.registrar(() => this.form.dirty && !this.saving());
    inject(DestroyRef).onDestroy(baja);
  }

  get negocioId(): string {
    return this.contexto.negocio()?.id ?? '';
  }

  hasError(control: CampoEmpleado): boolean {
    const field = this.form.controls[control];
    return (field.touched && field.invalid) || Boolean(this.fieldErrors()[control]);
  }

  /** El mensaje del servidor tiene prioridad: describe un rechazo que el formulario no puede anticipar. */
  mensaje(control: CampoEmpleado): string {
    return this.fieldErrors()[control] || MENSAJES[control];
  }

  submit(): void {
    this.form.markAllAsTouched();
    this.error.set(null);
    this.fieldErrors.set({});
    if (this.form.invalid) {
      enfocarPrimerInvalido(this.host.nativeElement);
      return;
    }
    if (this.saving()) return;

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
        enfocarPrimerInvalido(this.host.nativeElement);
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

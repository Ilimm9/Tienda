import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { NonNullableFormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';

import { ContextoService } from '../../../contexto/contexto.service';
import { FeedbackService } from '../../../shared/feedback/feedback.service';
import { InvitacionResumen } from '../invitacion.models';
import { InvitacionService } from '../invitacion.service';
import { iniciales } from '../presentacion';

interface EnlaceEmitido {
  enlace: string;
  correo: string;
  correoEnviado: boolean;
}

@Component({
  selector: 'app-corregir-correo',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './corregir-correo.html',
})
export class CorregirCorreo {
  private readonly invitacionService = inject(InvitacionService);
  private readonly contexto = inject(ContextoService);
  private readonly feedback = inject(FeedbackService);
  private readonly route = inject(ActivatedRoute);
  private readonly formBuilder = inject(NonNullableFormBuilder);

  private readonly invitacionId = this.route.snapshot.paramMap.get('invitacionId') ?? '';
  readonly negocioId = computed(() => this.contexto.negocio()?.id ?? '');

  readonly invitacion = signal<InvitacionResumen | null>(null);
  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly enlaceEmitido = signal<EnlaceEmitido | null>(null);

  readonly form = this.formBuilder.group({
    correo: ['', [Validators.required, Validators.email, Validators.maxLength(254)]],
    confirmacion: ['', Validators.required],
  });

  constructor() {
    this.load();
  }

  get noCoincide(): boolean {
    const { correo, confirmacion } = this.form.getRawValue();
    return this.normalizar(correo) !== this.normalizar(confirmacion);
  }

  get correoConError(): boolean {
    const control = this.form.controls.correo;
    return control.touched && (control.invalid || this.sinCambio);
  }

  get confirmacionConError(): boolean {
    return this.form.controls.confirmacion.touched && this.noCoincide;
  }

  /** Iniciales del empleado invitado para la cabecera del diálogo. */
  iniciales(): string {
    const datos = this.invitacion();
    return iniciales(datos?.nombre_empleado, datos?.correo ?? '');
  }

  get sinCambio(): boolean {
    return this.normalizar(this.form.controls.correo.value) === this.invitacion()?.correo;
  }

  guardar(): void {
    this.error.set(null);
    if (this.form.invalid || this.noCoincide || this.sinCambio) {
      this.form.markAllAsTouched();
      return;
    }
    if (this.saving()) return;
    this.saving.set(true);
    this.invitacionService
      .reenviar(this.negocioId(), this.invitacionId, this.normalizar(this.form.controls.correo.value))
      .subscribe({
        next: (creada) => {
          this.saving.set(false);
          this.enlaceEmitido.set({
            enlace: this.invitacionService.enlaceDeToken(creada.token),
            correo: creada.invitacion.correo,
            correoEnviado: creada.correo_enviado,
          });
          this.feedback.success('Correo corregido', 'El enlace anterior ya no funciona.');
        },
        error: (response: HttpErrorResponse) => {
          this.saving.set(false);
          // El formulario se conserva: un reintento no vuelve a pedir los datos.
          this.error.set(
            (response.error as { mensaje?: string } | null)?.mensaje ??
              'No fue posible corregir el correo. Intenta nuevamente.',
          );
        },
      });
  }

  async copiar(enlace: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(enlace);
      this.feedback.success('Enlace copiado');
    } catch {
      this.feedback.warning('Copia el enlace manualmente');
    }
  }

  private normalizar(correo: string): string {
    return correo.trim().toLowerCase();
  }

  private load(): void {
    const negocioId = this.negocioId();
    if (!negocioId || !this.invitacionId) {
      this.loading.set(false);
      this.error.set('No encontramos la invitación.');
      return;
    }
    this.invitacionService.listar(negocioId).subscribe({
      next: ({ items }) => {
        const invitacion = items.find((item) => item.id === this.invitacionId) ?? null;
        this.invitacion.set(invitacion);
        this.loading.set(false);
        if (!invitacion) this.error.set('No encontramos la invitación.');
        else if (invitacion.estado === 'aceptada')
          this.error.set('La invitación ya fue aceptada; el correo ya no puede corregirse aquí.');
      },
      error: () => {
        this.loading.set(false);
        this.error.set('No fue posible cargar la invitación.');
      },
    });
  }
}

import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';

import {
  ActualizarNegocioPayload,
  ApiErrorResponse,
  CrearNegocioPayload,
  DireccionNegocio,
  NegocioDetalle,
} from './negocio.models';
import { NegocioService } from './negocio.service';

@Component({
  selector: 'app-negocio-form',
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './negocio-form.component.html',
  styleUrl: './negocio-form.component.css',
})
export class NegocioFormComponent {
  private readonly formBuilder = inject(FormBuilder);
  private readonly negocioService = inject(NegocioService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  readonly negocioId = this.route.snapshot.paramMap.get('negocioId');
  readonly editing = this.negocioId !== null;
  readonly loading = signal(this.editing);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly fieldErrors = signal<Record<string, string>>({});
  readonly slug = signal<string | null>(null);

  readonly form = this.formBuilder.nonNullable.group({
    nombre_comercial: [
      '',
      [Validators.required, Validators.minLength(2), Validators.maxLength(180)],
    ],
    razon_social: ['', Validators.maxLength(220)],
    rfc: ['', Validators.pattern(/^[A-Za-zÑñ&]{3,4}[0-9]{6}[A-Za-z0-9]{3}$/)],
    telefono: ['', Validators.maxLength(30)],
    correo: ['', [Validators.email, Validators.maxLength(254)]],
    codigo_moneda: ['MXN', [Validators.required, Validators.pattern(/^[A-Z]{3}$/)]],
    zona_horaria: ['America/Mexico_City', Validators.required],
    codigo_pais: ['MX', Validators.pattern(/^[A-Za-z]{2}$/)],
    estado: ['', Validators.maxLength(120)],
    municipio: ['', Validators.maxLength(120)],
    ciudad: ['', Validators.maxLength(120)],
    colonia: ['', Validators.maxLength(150)],
    codigo_postal: ['', Validators.maxLength(12)],
    calle: ['', Validators.maxLength(180)],
    numero_exterior: ['', Validators.maxLength(30)],
    numero_interior: ['', Validators.maxLength(30)],
    referencias: [''],
  });

  constructor() {
    if (this.negocioId) this.load(this.negocioId);
  }

  submit(): void {
    this.form.markAllAsTouched();
    this.error.set(null);
    this.fieldErrors.set({});
    if (this.form.invalid || this.saving()) return;

    const value = this.form.getRawValue();
    const address = this.addressPayload(value);
    this.saving.set(true);

    const request = this.negocioId
      ? this.negocioService.actualizar(this.negocioId, {
          nombre_comercial: value.nombre_comercial.trim(),
          razon_social: this.nullable(value.razon_social),
          rfc: this.nullable(value.rfc)?.toUpperCase() ?? null,
          telefono: this.nullable(value.telefono),
          correo: this.nullable(value.correo)?.toLowerCase() ?? null,
          codigo_moneda: value.codigo_moneda,
          zona_horaria: value.zona_horaria,
          direccion: address,
        } satisfies ActualizarNegocioPayload)
      : this.negocioService.crear({
          nombre_comercial: value.nombre_comercial.trim(),
          ...this.optionalProperty('razon_social', value.razon_social),
          ...this.optionalProperty('rfc', value.rfc.toUpperCase()),
          ...this.optionalProperty('telefono', value.telefono),
          ...this.optionalProperty('correo', value.correo.toLowerCase()),
          codigo_moneda: value.codigo_moneda,
          zona_horaria: value.zona_horaria,
          ...(address ? { direccion: address } : {}),
        } satisfies CrearNegocioPayload);

    request.subscribe({
      next: (business: NegocioDetalle) => {
        this.saving.set(false);
        void this.router.navigate(['/negocios', business.id]);
      },
      error: (response: HttpErrorResponse) => {
        this.saving.set(false);
        const apiError = response.error as ApiErrorResponse | null;
        this.error.set(apiError?.mensaje ?? 'No fue posible guardar el negocio.');
        this.fieldErrors.set(apiError?.campos ?? {});
      },
    });
  }

  hasError(control: keyof typeof this.form.controls): boolean {
    const field = this.form.controls[control];
    return field.invalid && (field.dirty || field.touched);
  }

  cancelLink(): readonly string[] {
    return this.negocioId ? ['/negocios', this.negocioId] : ['/negocios'];
  }

  private load(id: string): void {
    this.negocioService.obtener(id).subscribe({
      next: (business) => {
        const address = business.direccion;
        this.slug.set(business.slug);
        this.form.patchValue({
          nombre_comercial: business.nombre_comercial,
          razon_social: business.razon_social ?? '',
          rfc: business.rfc ?? '',
          telefono: business.telefono ?? '',
          correo: business.correo ?? '',
          codigo_moneda: business.codigo_moneda,
          zona_horaria: business.zona_horaria,
          codigo_pais: address?.codigo_pais ?? 'MX',
          estado: address?.estado ?? '',
          municipio: address?.municipio ?? '',
          ciudad: address?.ciudad ?? '',
          colonia: address?.colonia ?? '',
          codigo_postal: address?.codigo_postal ?? '',
          calle: address?.calle ?? '',
          numero_exterior: address?.numero_exterior ?? '',
          numero_interior: address?.numero_interior ?? '',
          referencias: address?.referencias ?? '',
        });
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(
          (response.error as ApiErrorResponse | null)?.mensaje ??
            'No fue posible cargar el negocio.',
        );
      },
    });
  }

  private addressPayload(
    value: ReturnType<typeof this.form.getRawValue>,
  ): Omit<DireccionNegocio, 'id'> | null {
    const optionalValues = [
      value.estado,
      value.municipio,
      value.ciudad,
      value.colonia,
      value.codigo_postal,
      value.calle,
      value.numero_exterior,
      value.numero_interior,
      value.referencias,
    ];
    if (!optionalValues.some((item) => item.trim())) return null;
    return {
      codigo_pais: value.codigo_pais.trim().toUpperCase() || 'MX',
      estado: this.nullable(value.estado),
      municipio: this.nullable(value.municipio),
      ciudad: this.nullable(value.ciudad),
      colonia: this.nullable(value.colonia),
      codigo_postal: this.nullable(value.codigo_postal),
      calle: this.nullable(value.calle),
      numero_exterior: this.nullable(value.numero_exterior),
      numero_interior: this.nullable(value.numero_interior),
      referencias: this.nullable(value.referencias),
    };
  }

  private nullable(value: string): string | null {
    return value.trim() || null;
  }

  private optionalProperty<K extends string>(key: K, value: string): Partial<Record<K, string>> {
    const cleaned = value.trim();
    return cleaned ? ({ [key]: cleaned } as Record<K, string>) : {};
  }
}

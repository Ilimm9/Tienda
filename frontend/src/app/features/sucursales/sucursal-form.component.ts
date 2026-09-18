import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, DestroyRef, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { CambiosPendientesService } from '../../contexto/cambios-pendientes.service';
import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { NegocioDetalle } from '../negocios/negocio.models';
import { NegocioService } from '../negocios/negocio.service';
import {
  ActualizarSucursalPayload,
  CrearSucursalPayload,
  DireccionSucursal,
  SucursalApiError,
  SucursalDetalle,
} from './sucursal.models';
import { SucursalService } from './sucursal.service';

@Component({
  selector: 'app-sucursal-form',
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './sucursal-form.component.html',
  styleUrl: './sucursal-form.component.css',
})
export class SucursalFormComponent {
  private readonly formBuilder = inject(FormBuilder);
  private readonly sucursalService = inject(SucursalService);
  private readonly negocioService = inject(NegocioService);
  private readonly feedback = inject(FeedbackService);
  private readonly contexto = inject(ContextoService);
  private readonly cambiosPendientes = inject(CambiosPendientesService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  readonly negocioId = this.route.snapshot.paramMap.get('negocioId') ?? '';
  readonly sucursalId = this.route.snapshot.paramMap.get('sucursalId');
  readonly editing = this.sucursalId !== null;
  readonly business = signal<NegocioDetalle | null>(null);
  readonly branch = signal<SucursalDetalle | null>(null);
  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly fieldErrors = signal<Record<string, string>>({});

  readonly form = this.formBuilder.nonNullable.group({
    codigo: [
      '',
      [
        Validators.required,
        Validators.minLength(2),
        Validators.maxLength(40),
        Validators.pattern(/^[A-Za-z0-9][A-Za-z0-9_-]{1,39}$/),
      ],
    ],
    nombre: ['', [Validators.required, Validators.minLength(2), Validators.maxLength(160)]],
    telefono: ['', Validators.maxLength(30)],
    es_principal: false,
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
    this.load();
    // Declara trabajo sin guardar para que un cambio de contexto pida confirmación.
    const baja = this.cambiosPendientes.registrar(() => this.form.dirty && !this.saving());
    inject(DestroyRef).onDestroy(baja);
  }

  submit(): void {
    this.form.markAllAsTouched();
    this.error.set(null);
    this.fieldErrors.set({});
    if (this.form.invalid || this.saving()) return;
    if (this.business()?.tipo_miembro !== 'propietario') {
      this.error.set('No tienes permiso para modificar sucursales.');
      return;
    }

    const value = this.form.getRawValue();
    const address = this.addressPayload(value);
    this.saving.set(true);
    const request = this.sucursalId
      ? this.sucursalService.actualizar(this.negocioId, this.sucursalId, {
          nombre: value.nombre.trim(),
          telefono: this.nullable(value.telefono),
          es_principal: value.es_principal,
          direccion: address,
        } satisfies ActualizarSucursalPayload)
      : this.sucursalService.crear(this.negocioId, {
          codigo: value.codigo.trim().toUpperCase(),
          nombre: value.nombre.trim(),
          ...this.optionalProperty('telefono', value.telefono),
          es_principal: value.es_principal,
          ...(address ? { direccion: address } : {}),
        } satisfies CrearSucursalPayload);

    request.subscribe({
      next: (branch) => {
        this.saving.set(false);
        const promoted = this.editing && !this.branch()?.es_principal && branch.es_principal;
        this.feedback.success(
          promoted
            ? 'Sucursal principal actualizada'
            : this.editing
              ? 'Sucursal actualizada'
              : 'Sucursal registrada',
        );
        this.contexto.recargar().subscribe();
        void this.router.navigate(['/negocios', this.negocioId, 'sucursales', branch.id]);
      },
      error: (response: HttpErrorResponse) => {
        this.saving.set(false);
        const apiError = response.error as SucursalApiError | null;
        this.error.set(apiError?.mensaje ?? 'No fue posible guardar la sucursal.');
        this.fieldErrors.set(apiError?.campos ?? {});
      },
    });
  }

  hasError(control: keyof typeof this.form.controls): boolean {
    const field = this.form.controls[control];
    return field.invalid && (field.dirty || field.touched);
  }

  cancelLink(): readonly string[] {
    return this.sucursalId
      ? ['/negocios', this.negocioId, 'sucursales', this.sucursalId]
      : ['/negocios', this.negocioId, 'sucursales'];
  }

  previewLocation(): string {
    const value = this.form.getRawValue();
    return [value.ciudad.trim(), value.estado.trim()].filter(Boolean).join(', ') || 'Sin dirección';
  }

  private load(): void {
    const businessRequest = this.negocioService.obtener(this.negocioId);
    if (this.sucursalId) {
      forkJoin({
        business: businessRequest,
        branch: this.sucursalService.obtener(this.negocioId, this.sucursalId),
      }).subscribe({
        next: ({ business, branch }) => {
          this.business.set(business);
          this.branch.set(branch);
          this.patchBranch(branch);
          this.loading.set(false);
        },
        error: (response: HttpErrorResponse) => this.handleLoadError(response),
      });
      return;
    }

    businessRequest.subscribe({
      next: (business) => {
        this.business.set(business);
        if (!business.tiene_sucursales) {
          this.form.controls.es_principal.setValue(true);
          this.form.controls.es_principal.disable({ emitEvent: false });
        }
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => this.handleLoadError(response),
    });
  }

  private handleLoadError(response: HttpErrorResponse): void {
    this.loading.set(false);
    this.error.set(
      (response.error as SucursalApiError | null)?.mensaje ??
        'No fue posible cargar la información.',
    );
  }

  private patchBranch(branch: SucursalDetalle): void {
    const address = branch.direccion;
    this.form.patchValue({
      codigo: branch.codigo,
      nombre: branch.nombre,
      telefono: branch.telefono ?? '',
      es_principal: branch.es_principal,
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
    if (branch.es_principal) {
      this.form.controls.es_principal.disable({ emitEvent: false });
    }
  }

  private addressPayload(
    value: ReturnType<typeof this.form.getRawValue>,
  ): Omit<DireccionSucursal, 'id'> | null {
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

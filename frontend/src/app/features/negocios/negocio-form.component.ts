import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, ElementRef, inject, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';

import { FeedbackService } from '../../shared/feedback/feedback.service';
import {
  ActualizarNegocioPayload,
  ApiErrorResponse,
  CrearNegocioPayload,
  DireccionNegocio,
  NegocioDetalle,
} from './negocio.models';
import { NegocioService } from './negocio.service';
import { ContextoService } from '../../contexto/contexto.service';
import { enfocarPrimerInvalido } from '../../shared/formularios/formulario';

const CAMPOS_DOMICILIO = [
  'calle',
  'numero_exterior',
  'numero_interior',
  'colonia',
  'codigo_postal',
  'municipio',
  'ciudad',
  'estado',
  'referencias',
] as const;

const MONEDAS: Record<string, string> = {
  MXN: 'MXN · Peso mexicano',
  USD: 'USD · Dólar estadounidense',
};

const ZONAS: Record<string, string> = {
  'America/Mexico_City': 'Centro de México (UTC−6)',
  'America/Cancun': 'Quintana Roo (UTC−5)',
  'America/Chihuahua': 'Chihuahua (UTC−6)',
  'America/Tijuana': 'Baja California (UTC−8)',
};

/** Aproximación del identificador que generará la API; al editar se muestra el real. */
function slugDe(nombre: string): string {
  return nombre
    .toLowerCase()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');
}

function inicialesDe(nombre: string): string {
  return (
    nombre
      .split(/\s+/)
      .filter(Boolean)
      .map((palabra) => palabra[0])
      .join('')
      .slice(0, 2)
      .toUpperCase() || 'N'
  );
}

@Component({
  selector: 'app-negocio-form',
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './negocio-form.component.html',
})
export class NegocioFormComponent {
  private readonly formBuilder = inject(FormBuilder);
  private readonly host = inject<ElementRef<HTMLElement>>(ElementRef);
  private readonly negocioService = inject(NegocioService);
  private readonly feedback = inject(FeedbackService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly contexto = inject(ContextoService);

  readonly negocioId = this.route.snapshot.paramMap.get('negocioId');
  readonly editing = this.negocioId !== null;
  readonly loading = signal(this.editing);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly fieldErrors = signal<Record<string, string>>({});
  readonly slug = signal<string | null>(null);
  readonly nombreGuardado = signal('');
  readonly domicilioAbierto = signal(false);
  readonly regionalAbierta = signal(false);

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

  /** Valores vivos para la vista previa y los indicadores de cada sección. */
  private readonly valores = toSignal(this.form.valueChanges, {
    initialValue: this.form.getRawValue(),
  });

  readonly vistaPrevia = computed(() => {
    const valores = this.valores();
    const nombre = valores.nombre_comercial?.trim() ?? '';
    return {
      nombre: nombre || 'Nombre comercial',
      razon: valores.razon_social?.trim() || 'Sin razón social',
      rfc: valores.rfc?.trim().toUpperCase() || '—',
      slug: this.slug() ?? (slugDe(nombre) || '—'),
      moneda: valores.codigo_moneda ?? 'MXN',
      iniciales: inicialesDe(nombre),
    };
  });

  readonly nombreListo = computed(
    () => (this.valores().nombre_comercial?.trim().length ?? 0) >= 2,
  );

  /** Longitud y tipo de persona que corresponde al RFC capturado. */
  readonly rfcAyuda = computed(() => {
    const largo = this.valores().rfc?.trim().length ?? 0;
    if (!largo) return { largo, valido: false, texto: '12 caracteres para persona moral, 13 para persona física.' };
    if (largo === 12) return { largo, valido: true, texto: 'Persona moral' };
    if (largo === 13) return { largo, valido: true, texto: 'Persona física' };
    return { largo, valido: false, texto: `Faltan ${12 - Math.min(largo, 12)} caracteres para persona moral.` };
  });

  readonly monedaTexto = computed(() => MONEDAS[this.valores().codigo_moneda ?? ''] ?? this.valores().codigo_moneda);
  readonly zonaTexto = computed(() => ZONAS[this.valores().zona_horaria ?? ''] ?? this.valores().zona_horaria);

  constructor() {
    if (this.negocioId) this.load(this.negocioId);
  }

  irA(id: string): void {
    this.host.nativeElement.querySelector(`#${id}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }

  submit(): void {
    this.form.markAllAsTouched();
    this.error.set(null);
    this.fieldErrors.set({});
    if (this.form.invalid) {
      // Una sección plegada con errores se abre para que el foco llegue al campo.
      if (CAMPOS_DOMICILIO.some((campo) => this.form.controls[campo].invalid)) this.domicilioAbierto.set(true);
      if (this.form.controls.codigo_moneda.invalid || this.form.controls.zona_horaria.invalid) this.regionalAbierta.set(true);
      enfocarPrimerInvalido(this.host.nativeElement);
      return;
    }
    if (this.saving()) return;

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
        this.feedback.success(this.editing ? 'Cambios guardados' : 'Negocio registrado');
        // El negocio nuevo trae el rol propietario: se recarga el contexto para que menú y permisos lo reflejen.
        const abrir = () => void this.router.navigate(['/negocios', business.id]);
        this.contexto.recargar().subscribe({ next: abrir, error: abrir });
      },
      error: (response: HttpErrorResponse) => {
        this.saving.set(false);
        const apiError = response.error as ApiErrorResponse | null;
        this.error.set(apiError?.mensaje ?? 'No fue posible guardar el negocio.');
        const campos = apiError?.campos ?? {};
        this.fieldErrors.set(campos);
        if (CAMPOS_DOMICILIO.some((campo) => campos[campo])) this.domicilioAbierto.set(true);
        if (campos['codigo_moneda'] || campos['zona_horaria']) this.regionalAbierta.set(true);
      },
    });
  }

  /** Inválido por validación local o porque el servidor rechazó el campo. */
  invalido(control: keyof typeof this.form.controls): boolean {
    return this.hasError(control) || Boolean(this.fieldErrors()[control]);
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
        this.nombreGuardado.set(business.nombre_comercial);
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
        this.domicilioAbierto.set(CAMPOS_DOMICILIO.some((campo) => Boolean(address?.[campo]?.trim())));
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

import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, DestroyRef, ElementRef, computed, inject, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { Observable, concat, last } from 'rxjs';

import { CambiosPendientesService } from '../../contexto/cambios-pendientes.service';
import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { enfocarPrimerInvalido, fechaNoFutura, sinSoloEspacios } from '../../shared/formularios/formulario';
import { RolResumen } from '../roles-permisos/rol.models';
import { RolService } from '../roles-permisos/rol.service';
import { AsignacionResumen } from './asignacion.models';
import { AsignacionService } from './asignacion.service';
import { EmpleadoApiError, EmpleadoDetalle } from './empleado.models';
import { EmpleadoService } from './empleado.service';
import { InvitacionService } from './invitacion.service';
import { digitosTelefono, formatoTelefono, siguienteNumero } from './numero-empleado';
import { Puesto } from './puesto.models';
import { PuestoService } from './puesto.service';

type CampoEmpleado =
  | 'nombre' | 'segundo_nombre' | 'primer_apellido' | 'segundo_apellido'
  | 'correo' | 'telefono' | 'puesto_id' | 'numero_empleado' | 'contratado_en';

const MENSAJES: Record<CampoEmpleado, string> = {
  nombre: 'Usa entre 2 y 100 caracteres.',
  segundo_nombre: 'Usa hasta 100 caracteres.',
  primer_apellido: 'Usa entre 2 y 100 caracteres.',
  segundo_apellido: 'Usa hasta 100 caracteres.',
  correo: 'Escribe un correo válido, como nombre@correo.mx.',
  telefono: 'Escribe 10 dígitos, sin lada.',
  puesto_id: 'Elige un puesto activo del catálogo.',
  numero_empleado: 'Usa hasta 40 caracteres: letras, números, guion o guion bajo.',
  contratado_en: 'La fecha no puede ser posterior a hoy.',
};

// El rol propietario nunca se delega por invitación; el servidor también lo rechaza.
const CODIGO_ROL_PROPIETARIO = 'PROPIETARIO';

/** Alta en 3 pasos: datos, sucursal e invitación; el 4 es el resumen final. */
type Paso = 1 | 2 | 3 | 4;

interface InvitacionEnviada {
  correo: string;
  correoEnviado: boolean;
  enlace: string;
}

@Component({
  selector: 'app-empleado-form',
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './empleado-form.component.html',
})
export class EmpleadoFormComponent {
  private readonly empleadoService = inject(EmpleadoService);
  private readonly puestoService = inject(PuestoService);
  private readonly asignacionService = inject(AsignacionService);
  private readonly invitacionService = inject(InvitacionService);
  private readonly rolService = inject(RolService);
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

  readonly paso = signal<Paso>(1);
  /** Empleado ya guardado por el paso 1; volver a ese paso lo actualiza en vez de crear otro. */
  readonly creado = signal<EmpleadoDetalle | null>(null);
  readonly puestos = signal<Puesto[]>([]);
  readonly numerosUsados = signal<(string | null)[]>([]);
  readonly asignaciones = signal<AsignacionResumen[]>([]);
  readonly seleccion = signal<ReadonlySet<string>>(new Set());
  readonly principalId = signal('');
  readonly roles = signal<RolResumen[]>([]);
  readonly rolId = signal('');
  readonly invitacion = signal<InvitacionEnviada | null>(null);

  readonly sucursales = computed(() => this.contexto.negocio()?.sucursales ?? []);
  readonly puedeAsignar = computed(() => this.contexto.puede(PERMISOS.asignacionEditar));
  readonly puedeInvitar = computed(() => this.contexto.puede([PERMISOS.invitacionEnviar, PERMISOS.rolVer]));
  readonly asignadas = computed(() => new Set(this.asignaciones().map((item) => item.sucursal_id)));
  readonly principalAsignada = computed(() => this.asignaciones().find((item) => item.es_principal) ?? this.asignaciones()[0] ?? null);
  /** Puestos activos más el que ya tiene el empleado, aunque se haya desactivado. */
  readonly opcionesPuesto = computed(() => {
    const actual = this.empleado()?.puesto_id ?? this.creado()?.puesto_id;
    return this.puestos().filter((puesto) => puesto.activo || puesto.id === actual);
  });
  readonly rolSugerido = computed(() => {
    const puesto = this.puestos().find((item) => item.id === this.creado()?.puesto_id);
    return this.roles().find((rol) => rol.id === puesto?.rol_sugerido_id) ?? null;
  });
  /** Motivo por el que el paso 3 no puede enviar la invitación; null cuando sí puede. */
  readonly bloqueoInvitacion = computed<'correo' | 'sucursal' | 'permiso' | null>(() => {
    if (!this.puedeInvitar()) return 'permiso';
    if (!this.creado()?.correo) return 'correo';
    if (!this.principalAsignada()) return 'sucursal';
    return null;
  });

  readonly form = this.fb.nonNullable.group({
    // Los límites repiten los del backend (`empleado_service.go`) para avisar antes de enviar.
    numero_empleado: ['', Validators.pattern(/^\s*[A-Za-z0-9][A-Za-z0-9_-]{0,39}\s*$/)],
    nombre: ['', [Validators.required, sinSoloEspacios, Validators.minLength(2), Validators.maxLength(100)]],
    segundo_nombre: ['', Validators.maxLength(100)],
    primer_apellido: ['', [Validators.required, sinSoloEspacios, Validators.minLength(2), Validators.maxLength(100)]],
    segundo_apellido: ['', Validators.maxLength(100)],
    correo: ['', [Validators.email, Validators.maxLength(254)]],
    telefono: ['', Validators.pattern(/^\d{2} \d{4} \d{4}$/)],
    puesto_id: [''],
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
  readonly puestoVista = computed(() => this.puestos().find((puesto) => puesto.id === this.valores().puesto_id)?.nombre ?? '');
  readonly correoVista = computed(() => (this.valores().correo ?? '').trim());
  readonly telefonoVista = computed(() => (this.valores().telefono ?? '').trim());
  readonly numeroVista = computed(() => (this.valores().numero_empleado ?? '').trim());

  constructor() {
    if (this.editing) this.load();
    this.cargarApoyos();
    const baja = this.cambiosPendientes.registrar(() => this.form.dirty && !this.saving());
    inject(DestroyRef).onDestroy(baja);
  }

  /** Mantiene el teléfono como `55 1234 5678` mientras se escribe. */
  formatearTelefono(): void {
    const control = this.form.controls.telefono;
    control.setValue(formatoTelefono(control.value));
  }

  /** Propone el siguiente número con las iniciales de la sucursal elegida (o la del contexto). */
  generarNumero(): void {
    const sucursalId = this.principalId() || this.contexto.sucursal()?.id;
    const sucursal = this.sucursales().find((item) => item.id === sucursalId) ?? this.contexto.sucursal();
    const control = this.form.controls.numero_empleado;
    control.setValue(siguienteNumero(sucursal?.nombre ?? '', this.numerosUsados()));
    control.markAsDirty();
  }

  /** Se puede volver al paso 1 o 2 en cualquier momento; el 3 requiere que el empleado exista. */
  irPaso(paso: Paso): void {
    if (paso === 3 && this.creado()) this.irAInvitacion();
    else if (paso === 1 || (this.creado() && paso === 2)) this.paso.set(paso);
  }

  /** El rol se propone desde el puesto la primera vez; luego se respeta lo elegido. */
  private irAInvitacion(): void {
    if (!this.rolId()) this.rolId.set(this.rolSugerido()?.id ?? '');
    this.paso.set(3);
  }

  alternarSucursal(id: string, marcada: boolean): void {
    if (this.asignadas().has(id)) return;
    const seleccion = new Set(this.seleccion());
    if (marcada) seleccion.add(id);
    else seleccion.delete(id);
    this.seleccion.set(seleccion);
    if (!marcada && this.principalId() === id) this.principalId.set([...this.asignadas(), ...seleccion][0] ?? '');
    if (marcada && !this.principalId()) this.principalId.set(id);
  }

  /** Asigna las sucursales nuevas y, si cambió, la principal; después pasa a la invitación. */
  guardarSucursales(): void {
    const empleado = this.creado();
    if (!empleado || this.saving()) return;
    const nuevas = [...this.seleccion()].filter((id) => !this.asignadas().has(id));
    const principal = this.principalId();
    const cambiarPrincipal = this.asignaciones().find((item) => item.sucursal_id === principal && !item.es_principal);
    const pasos: Observable<unknown>[] = nuevas.map((id) =>
      this.asignacionService.asignar(this.negocioId, empleado.id, { sucursal_id: id, es_principal: id === principal }),
    );
    if (cambiarPrincipal) pasos.push(this.asignacionService.establecerPrincipal(this.negocioId, empleado.id, cambiarPrincipal.id));
    if (!pasos.length) {
      this.irAInvitacion();
      return;
    }
    this.saving.set(true);
    this.error.set(null);
    concat(...pasos, this.asignacionService.listar(this.negocioId, empleado.id)).pipe(last()).subscribe({
      next: (lista) => {
        this.saving.set(false);
        this.asignaciones.set((lista as { items: AsignacionResumen[] }).items);
        this.seleccion.set(new Set());
        this.irAInvitacion();
      },
      error: (response: HttpErrorResponse) => {
        this.saving.set(false);
        this.error.set((response.error as EmpleadoApiError | null)?.mensaje ?? 'No fue posible asignar las sucursales.');
        this.recargarAsignaciones();
      },
    });
  }

  enviarInvitacion(): void {
    const empleado = this.creado();
    const sucursal = this.principalAsignada();
    if (!empleado || !sucursal || this.bloqueoInvitacion() || this.saving()) return;
    if (!this.rolId()) {
      this.error.set('Elige el rol que tendrá.');
      return;
    }
    this.saving.set(true);
    this.error.set(null);
    this.invitacionService
      .crear(this.negocioId, { empleado_id: empleado.id, sucursal_id: sucursal.sucursal_id, rol_predeterminado_id: this.rolId() })
      .subscribe({
        next: (creada) => {
          this.saving.set(false);
          this.invitacion.set({
            correo: creada.invitacion.correo,
            correoEnviado: creada.correo_enviado,
            enlace: this.invitacionService.enlaceDeToken(creada.token),
          });
          this.paso.set(4);
        },
        error: (response: HttpErrorResponse) => {
          this.saving.set(false);
          this.error.set((response.error as EmpleadoApiError | null)?.mensaje ?? 'No fue posible enviar la invitación.');
        },
      });
  }

  terminarSinInvitar(): void {
    this.invitacion.set(null);
    this.paso.set(4);
  }

  /** Reinicia el asistente para dar de alta a otra persona. */
  otraAlta(): void {
    this.form.reset();
    this.creado.set(null);
    this.asignaciones.set([]);
    this.invitacion.set(null);
    this.rolId.set('');
    this.error.set(null);
    this.prepararSucursales();
    this.cargarApoyos();
    this.paso.set(1);
  }

  async copiar(enlace: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(enlace);
      this.feedback.success('Enlace copiado');
    } catch {
      this.feedback.warning('Copia el enlace manualmente');
    }
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
    const creado = this.creado();
    this.saving.set(true);
    const request = this.editing || creado
      ? this.empleadoService.actualizar(this.negocioId, creado?.id ?? this.empleadoId, {
          numero_empleado: this.opcional(value.numero_empleado),
          nombre: value.nombre.trim(),
          segundo_nombre: this.opcional(value.segundo_nombre),
          primer_apellido: value.primer_apellido.trim(),
          segundo_apellido: this.opcional(value.segundo_apellido),
          correo: this.opcional(value.correo),
          telefono: this.opcional(digitosTelefono(value.telefono)),
          puesto_id: value.puesto_id || null,
          contratado_en: this.opcional(value.contratado_en),
        })
      : this.empleadoService.crear(this.negocioId, {
          numero_empleado: this.opcional(value.numero_empleado),
          nombre: value.nombre.trim(),
          segundo_nombre: this.opcional(value.segundo_nombre),
          primer_apellido: value.primer_apellido.trim(),
          segundo_apellido: this.opcional(value.segundo_apellido),
          correo: this.opcional(value.correo),
          telefono: this.opcional(digitosTelefono(value.telefono)),
          puesto_id: value.puesto_id || null,
          contratado_en: this.opcional(value.contratado_en),
        });

    request.subscribe({
      next: (empleado) => {
        this.saving.set(false);
        this.form.markAsPristine();
        if (this.editing) {
          this.feedback.success('Empleado actualizado');
          void this.router.navigate(['/equipo/empleados', empleado.id]);
          return;
        }
        if (!creado) {
          this.feedback.success('Empleado registrado');
          this.numerosUsados.update((numeros) => [...numeros, empleado.numero_empleado]);
          this.prepararSucursales();
        }
        this.creado.set(empleado);
        this.paso.set(2);
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

  /** Catálogo de puestos, números usados y roles para invitar; ninguno bloquea el formulario. */
  private cargarApoyos(): void {
    const negocioId = this.negocioId;
    if (!negocioId) return;
    this.puestoService.listar(negocioId).subscribe({ next: ({ items }) => this.puestos.set(items), error: () => undefined });
    if (this.editing) return;
    this.empleadoService.listar(negocioId).subscribe({
      next: ({ items }) => this.numerosUsados.set(items.map((item) => item.numero_empleado)),
      error: () => undefined,
    });
    if (this.puedeInvitar()) {
      this.rolService.listar(negocioId, false, true).subscribe({
        next: ({ items }) =>
          this.roles.set(items.filter((rol) => rol.activo && rol.codigo.toUpperCase() !== CODIGO_ROL_PROPIETARIO)),
        error: () => undefined,
      });
    }
    this.prepararSucursales();
  }

  /** Por defecto se propone la sucursal del contexto como principal. */
  private prepararSucursales(): void {
    const id = this.contexto.sucursal()?.id ?? '';
    this.seleccion.set(new Set(id ? [id] : []));
    this.principalId.set(id);
  }

  private recargarAsignaciones(): void {
    const empleado = this.creado();
    if (!empleado) return;
    this.asignacionService.listar(this.negocioId, empleado.id).subscribe({
      next: ({ items }) => this.asignaciones.set(items),
      error: () => undefined,
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
          telefono: formatoTelefono(empleado.telefono),
          puesto_id: empleado.puesto_id ?? '',
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

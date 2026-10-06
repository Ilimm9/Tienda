import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, ElementRef, inject, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';

import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { enfocarPrimerInvalido, sinSoloEspacios } from '../../shared/formularios/formulario';
import { ApiErrorResponse, NegocioResumen } from '../negocios/negocio.models';
import { NegocioService } from '../negocios/negocio.service';
import { SucursalService } from '../sucursales/sucursal.service';

type Paso = 'empresa' | 'sucursal' | 'listo';

const MONEDAS: Record<string, string> = {
  MXN: 'MXN · Peso mexicano',
  USD: 'USD · Dólar estadounidense',
};

const ZONAS: Record<string, string> = {
  'America/Mexico_City': 'Ciudad de México (UTC−6)',
  'America/Cancun': 'Quintana Roo (UTC−5)',
  'America/Chihuahua': 'Chihuahua (UTC−6)',
  'America/Tijuana': 'Baja California (UTC−8)',
};

/**
 * Asistente de alta inicial: empresa, primera sucursal y cierre.
 * Usa los mismos servicios y validaciones que los formularios completos; aquí solo se pide lo mínimo.
 */
@Component({
  selector: 'app-alta-inicial',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './alta-inicial.html',
})
export class AltaInicial {
  private readonly fb = inject(FormBuilder);
  private readonly negocios = inject(NegocioService);
  private readonly sucursales = inject(SucursalService);
  private readonly contexto = inject(ContextoService);
  private readonly router = inject(Router);
  private readonly host = inject<ElementRef<HTMLElement>>(ElementRef);

  readonly monedas = Object.entries(MONEDAS);
  readonly zonas = Object.entries(ZONAS);
  /** Para invitar primero hay que registrar al empleado: el cierre lleva a ese formulario. */
  readonly puedeRegistrarEquipo = computed(() => this.contexto.puede(PERMISOS.empleadoGestionar));

  // Con la empresa ya registrada se retoma en la sucursal: nunca se duplica.
  readonly paso = signal<Paso>(this.contexto.estadoAlta() === 'requiere_sucursal' ? 'sucursal' : 'empresa');
  readonly guardando = signal(false);
  readonly error = signal<string | null>(null);
  readonly campos = signal<Record<string, string>>({});
  readonly masDatosEmpresa = signal(false);
  readonly masDatosSucursal = signal(false);
  /** Empresas propias archivadas: se ofrece restaurarlas antes de crear otra. */
  readonly archivadas = signal<NegocioResumen[]>([]);
  readonly empresa = signal<{ id: string; nombre: string } | null>(this.empresaDelContexto());
  readonly sucursalCreada = signal<{ nombre: string; codigo: string } | null>(null);

  // Mismos límites que `negocio-form` y `sucursal-form`.
  readonly empresaForm = this.fb.nonNullable.group({
    nombre_comercial: ['', [Validators.required, sinSoloEspacios, Validators.minLength(2), Validators.maxLength(180)]],
    codigo_moneda: ['MXN', [Validators.required, Validators.pattern(/^[A-Z]{3}$/)]],
    zona_horaria: ['America/Mexico_City', Validators.required],
    razon_social: ['', Validators.maxLength(220)],
    rfc: ['', Validators.pattern(/^[A-Za-zÑñ&]{3,4}[0-9]{6}[A-Za-z0-9]{3}$/)],
    telefono: ['', Validators.maxLength(30)],
    correo: ['', [Validators.email, Validators.maxLength(254)]],
  });
  readonly sucursalForm = this.fb.nonNullable.group({
    nombre: ['', [Validators.required, sinSoloEspacios, Validators.minLength(2), Validators.maxLength(160)]],
    codigo: [
      'SUC-001',
      [Validators.required, Validators.pattern(/^[A-Za-z0-9][A-Za-z0-9_-]{1,39}$/)],
    ],
    telefono: ['', Validators.maxLength(30)],
    calle: ['', Validators.maxLength(180)],
    numero_exterior: ['', Validators.maxLength(30)],
    colonia: ['', Validators.maxLength(150)],
    ciudad: ['', Validators.maxLength(120)],
    estado: ['', Validators.maxLength(120)],
    codigo_postal: ['', Validators.maxLength(12)],
  });

  private readonly valoresEmpresa = toSignal(this.empresaForm.valueChanges, {
    initialValue: this.empresaForm.getRawValue(),
  });
  private readonly valoresSucursal = toSignal(this.sucursalForm.valueChanges, {
    initialValue: this.sucursalForm.getRawValue(),
  });
  // Resumen en vivo del mosaico.
  readonly nombreEmpresaVista = computed(
    () => this.empresa()?.nombre || (this.valoresEmpresa().nombre_comercial ?? '').trim() || 'Tu empresa',
  );
  readonly monedaVista = computed(() => this.valoresEmpresa().codigo_moneda ?? 'MXN');
  readonly zonaVista = computed(
    () => (ZONAS[this.valoresEmpresa().zona_horaria ?? ''] ?? '').replace(/ \(.*\)$/, ''),
  );
  readonly nombreSucursalVista = computed(() => (this.valoresSucursal().nombre ?? '').trim() || 'Tu sucursal');
  readonly codigoSucursalVista = computed(() => {
    const control = this.sucursalForm.controls.codigo;
    this.valoresSucursal();
    return control.valid ? control.value.trim().toUpperCase() : 'Código por confirmar';
  });

  constructor() {
    if (this.paso() === 'empresa') this.buscarArchivadas();
  }

  invalidoEmpresa(control: keyof typeof this.empresaForm.controls): boolean {
    const campo = this.empresaForm.controls[control];
    return (campo.invalid && (campo.dirty || campo.touched)) || Boolean(this.campos()[control]);
  }

  invalidoSucursal(control: keyof typeof this.sucursalForm.controls): boolean {
    const campo = this.sucursalForm.controls[control];
    return (campo.invalid && (campo.dirty || campo.touched)) || Boolean(this.campos()[control]);
  }

  /** «Correcto» solo después de escribir algo válido; un campo intacto no se felicita. */
  correctoSucursal(control: 'nombre' | 'codigo'): boolean {
    const campo = this.sucursalForm.controls[control];
    return campo.valid && campo.dirty && !this.campos()[control];
  }

  guardarEmpresa(): void {
    this.empresaForm.markAllAsTouched();
    this.error.set(null);
    this.campos.set({});
    if (this.empresaForm.invalid) {
      // Un dato opcional inválido puede estar plegado: se abre para que el error se vea.
      if (!this.empresaForm.controls.nombre_comercial.invalid) this.masDatosEmpresa.set(true);
      enfocarPrimerInvalido(this.host.nativeElement);
      return;
    }
    if (this.guardando()) return;
    this.guardando.set(true);
    const valor = this.empresaForm.getRawValue();
    this.negocios
      .crear({
        nombre_comercial: valor.nombre_comercial.trim(),
        codigo_moneda: valor.codigo_moneda,
        zona_horaria: valor.zona_horaria,
        ...this.opcional('razon_social', valor.razon_social),
        ...this.opcional('rfc', valor.rfc.toUpperCase()),
        ...this.opcional('telefono', valor.telefono),
        ...this.opcional('correo', valor.correo.toLowerCase()),
      })
      .subscribe({
        next: (negocio) => this.empresaLista(negocio.id, negocio.nombre_comercial),
        error: (respuesta: HttpErrorResponse) => this.fallo(respuesta, 'No fue posible registrar la empresa.'),
      });
  }

  restaurar(negocio: NegocioResumen): void {
    if (this.guardando()) return;
    this.guardando.set(true);
    this.error.set(null);
    this.negocios.restaurar(negocio.id).subscribe({
      next: (restaurado) => this.empresaLista(restaurado.id, restaurado.nombre_comercial),
      error: (respuesta: HttpErrorResponse) => this.fallo(respuesta, 'No fue posible restaurar la empresa.'),
    });
  }

  crearSucursal(): void {
    this.sucursalForm.markAllAsTouched();
    this.error.set(null);
    this.campos.set({});
    const empresa = this.empresa();
    if (this.sucursalForm.invalid || !empresa) {
      const basicos = this.sucursalForm.controls;
      if (!basicos.nombre.invalid && !basicos.codigo.invalid) this.masDatosSucursal.set(true);
      enfocarPrimerInvalido(this.host.nativeElement);
      return;
    }
    if (this.guardando()) return;
    this.guardando.set(true);
    const valor = this.sucursalForm.getRawValue();
    const direccion = this.direccion(valor);
    const codigo = valor.codigo.trim().toUpperCase();
    this.sucursales
      .crear(empresa.id, {
        codigo,
        nombre: valor.nombre.trim(),
        es_principal: true,
        ...this.opcional('telefono', valor.telefono),
        ...(direccion ? { direccion } : {}),
      })
      .subscribe({
        next: (sucursal) => {
          this.sucursalCreada.set({ nombre: sucursal.nombre, codigo: sucursal.codigo });
          // El contexto recargado deja seleccionadas la empresa y su sucursal principal.
          const cerrar = () => {
            this.contexto.seleccionarNegocio(empresa.id);
            this.contexto.seleccionarSucursal(sucursal.id);
            this.guardando.set(false);
            this.paso.set('listo');
          };
          this.contexto.recargar().subscribe({ next: cerrar, error: cerrar });
        },
        error: (respuesta: HttpErrorResponse) => this.fallo(respuesta, 'No fue posible crear la sucursal.'),
      });
  }

  /** «Hacerlo después»: Inicio se muestra con el aviso y el asistente no vuelve a abrirse en esta sesión. */
  posponer(): void {
    this.contexto.posponerAlta();
    void this.router.navigate(['/inicio']);
  }

  private empresaLista(id: string, nombre: string): void {
    const seguir = () => {
      this.contexto.seleccionarNegocio(id);
      this.empresa.set({ id, nombre });
      this.guardando.set(false);
      // Una empresa restaurada puede traer ya sus sucursales.
      if (this.contexto.estadoAlta() === 'listo') void this.router.navigate(['/inicio']);
      else this.paso.set('sucursal');
    };
    this.contexto.recargar().subscribe({ next: seguir, error: seguir });
  }

  private fallo(respuesta: HttpErrorResponse, respaldo: string): void {
    this.guardando.set(false);
    const cuerpo = respuesta.error as ApiErrorResponse | null;
    this.campos.set(cuerpo?.campos ?? {});
    this.error.set(cuerpo?.mensaje ?? respaldo);
    enfocarPrimerInvalido(this.host.nativeElement);
  }

  private buscarArchivadas(): void {
    this.negocios.listar('archivado').subscribe({
      next: ({ items }) => this.archivadas.set(items.filter((item) => item.tipo_miembro === 'propietario')),
      // Sin la lista el asistente sigue sirviendo: solo no ofrece restaurar.
      error: () => this.archivadas.set([]),
    });
  }

  private empresaDelContexto(): { id: string; nombre: string } | null {
    const negocio = this.contexto.negocio();
    return negocio ? { id: negocio.id, nombre: negocio.nombre_comercial } : null;
  }

  private direccion(valor: ReturnType<typeof this.sucursalForm.getRawValue>) {
    const partes = [valor.calle, valor.numero_exterior, valor.colonia, valor.ciudad, valor.estado, valor.codigo_postal];
    if (!partes.some((parte) => parte.trim())) return null;
    const limpio = (texto: string) => texto.trim() || null;
    return {
      codigo_pais: 'MX',
      estado: limpio(valor.estado),
      municipio: null,
      ciudad: limpio(valor.ciudad),
      colonia: limpio(valor.colonia),
      codigo_postal: limpio(valor.codigo_postal),
      calle: limpio(valor.calle),
      numero_exterior: limpio(valor.numero_exterior),
      numero_interior: null,
      referencias: null,
    };
  }

  private opcional<K extends string>(clave: K, valor: string): Partial<Record<K, string>> {
    const limpio = valor.trim();
    return limpio ? ({ [clave]: limpio } as Record<K, string>) : {};
  }
}

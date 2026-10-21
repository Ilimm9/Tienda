import { HttpErrorResponse } from '@angular/common/http';
import {Component, DestroyRef,ElementRef,Injector,ViewChild,afterNextRender, computed, inject,signal,} from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { gsap } from 'gsap';
import { SelectModule } from 'primeng/select';

import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { enfocarPrimerInvalido, sinSoloEspacios } from '../../shared/formularios/formulario';
import { ApiErrorResponse, NegocioResumen } from '../negocios/negocio.models';
import { NegocioService } from '../negocios/negocio.service';
import { SucursalService } from '../sucursales/sucursal.service';

type Paso = 'empresa' | 'sucursal' | 'listo';
type PosicionOrbital = 'superior' | 'superiorIntermedia' | 'activa' | 'intermedia' | 'inferior';
type PosicionRelativa = -2 | -1 | 0 | 1 | 2;

interface CardAntes {
  rect: DOMRect;
  posicion: PosicionOrbital;
  angulo: number;
}

interface GeometriaOrbital {
  x: number;
  y: number;
  radio: number;
}

const ORBITA = {
  radioHalo: 0.90, //Fracción del radio visual del halo ocupada por la trayectoria de las cards. 
  separacionAngular: 50, //Separación entre posiciones vecinas alrededor del extremo activo (180°).
  duracion: 0.72,
  escalaActiva: 1,
  escalaCercana: 0.97,
  escalaLejana: 0.94,
  opacidadActiva: 1,
  opacidadCercana: 0.9,
  opacidadLejana: 0.78,
  rotacionCercana: 1.2,
  rotacionLejana: 2,
} as const;

const ANGULOS_ORBITA: Record<PosicionOrbital, number> = {
  superior: 180 - ORBITA.separacionAngular * 2,
  superiorIntermedia: 180 - ORBITA.separacionAngular,
  activa: 180,
  intermedia: 180 + ORBITA.separacionAngular,
  inferior: 180 + ORBITA.separacionAngular * 2,
};

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

/** Asistente de alta inicial: empresa, primera sucursal y cierre. solo se pide lo mínimo.*/
@Component({
  selector: 'app-alta-inicial',
  imports: [ReactiveFormsModule, RouterLink, SelectModule],
  templateUrl: './alta-inicial.html',
})
export class AltaInicial {
  private readonly fb = inject(FormBuilder);
  private readonly negocios = inject(NegocioService);
  private readonly sucursales = inject(SucursalService);
  private readonly contexto = inject(ContextoService);
  private readonly router = inject(Router);
  private readonly host = inject<ElementRef<HTMLElement>>(ElementRef);
  private readonly destroyRef = inject(DestroyRef);
  private readonly injector = inject(Injector);

  @ViewChild('vistaPreviaLateral', { read: ElementRef })
  private readonly vistaPreviaLateral?: ElementRef<HTMLElement>;

  private vistaPreviaTimeline?: gsap.core.Timeline;
  private orbitaObserver?: ResizeObserver;
  private mediaVistaCompacta?: MediaQueryList;
  /** Posiciones previas reales: permiten animar sin asumir el ancho del panel. */
  private posicionesCardsAntes = new Map<string, CardAntes>();
  private angulosCardsActuales = new Map<string, number>();
  private cambioVistaPrevia = 0;

  readonly monedas = Object.entries(MONEDAS).map(([valor, etiqueta]) => ({ valor, etiqueta }));
  readonly zonas = Object.entries(ZONAS).map(([valor, etiqueta]) => ({ valor, etiqueta }));
  /** Para invitar primero hay que registrar al empleado: el cierre lleva a ese formulario. */
  readonly puedeRegistrarEquipo = computed(() => this.contexto.puede(PERMISOS.empleadoGestionar));

  // Con la empresa ya registrada se retoma en la sucursal: nunca se duplica.
  readonly paso = signal<Paso>(
    this.contexto.estadoAlta() === 'requiere_sucursal' ? 'sucursal' : 'empresa',
  );
  readonly guardando = signal(false);
  /** Una vez completada el alta, los pasos anteriores se consultan sin editarse. */
  readonly altaCompletada = signal(this.contexto.estadoAlta() === 'listo');
  readonly modoConsulta = computed(() => this.altaCompletada() && this.paso() !== 'listo');
  readonly error = signal<string | null>(null);
  readonly campos = signal<Record<string, string>>({});
  readonly masDatosEmpresa = signal(false);
  readonly masDatosSucursal = signal(false);
  /** Empresas propias archivadas: se ofrece restaurarlas antes de crear otra. */
  readonly archivadas = signal<NegocioResumen[]>([]);
  readonly empresa = signal<{ id: string; nombre: string } | null>(this.empresaDelContexto());
  readonly sucursalCreada = signal<{ nombre: string; codigo: string } | null>(null);
  private readonly ordenPasos: Paso[] = ['empresa', 'sucursal', 'listo'];

  // Mismos límites que `negocio-form` y `sucursal-form`.
  readonly empresaForm = this.fb.nonNullable.group({
    nombre_comercial: [
      '',
      [Validators.required, sinSoloEspacios, Validators.minLength(2), Validators.maxLength(180)],
    ],
    codigo_moneda: ['MXN', [Validators.required, Validators.pattern(/^[A-Z]{3}$/)]],
    zona_horaria: ['America/Mexico_City', Validators.required],
    razon_social: ['', Validators.maxLength(220)],
    rfc: ['', Validators.pattern(/^[A-Za-zÑñ&]{3,4}[0-9]{6}[A-Za-z0-9]{3}$/)],
    telefono: ['', Validators.maxLength(30)],
    correo: ['', [Validators.email, Validators.maxLength(254)]],
  });
  readonly sucursalForm = this.fb.nonNullable.group({
    nombre: [
      '',
      [Validators.required, sinSoloEspacios, Validators.minLength(2), Validators.maxLength(160)],
    ],
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
    () =>
      this.empresa()?.nombre ||
      (this.valoresEmpresa().nombre_comercial ?? '').trim() ||
      'Tu empresa',
  );
  readonly monedaVista = computed(() => this.valoresEmpresa().codigo_moneda ?? 'MXN');
  readonly zonaVista = computed(() =>
    (ZONAS[this.valoresEmpresa().zona_horaria ?? ''] ?? '').replace(/ \(.*\)$/, ''),
  );
  readonly nombreSucursalVista = computed(
    () => (this.valoresSucursal().nombre ?? '').trim() || 'Tu sucursal',
  );
  readonly codigoSucursalVista = computed(() => {
    const control = this.sucursalForm.controls.codigo;
    this.valoresSucursal();
    return control.valid ? control.value.trim().toUpperCase() : 'Código por confirmar';
  });

  constructor() {
    if (this.paso() === 'empresa') this.buscarArchivadas();
    afterNextRender(() => this.iniciarModoResponsivo(), { injector: this.injector });
    this.destroyRef.onDestroy(() => {
      this.vistaPreviaTimeline?.kill();
      this.orbitaObserver?.disconnect();
      this.mediaVistaCompacta?.removeEventListener('change', this.alCambiarModoResponsivo);
    });
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
    if (this.modoConsulta()) return;
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
    this.cambiarPaso('sucursal');
  }

  restaurar(negocio: NegocioResumen): void {
    if (this.guardando()) return;
    this.guardando.set(true);
    this.error.set(null);
    this.negocios.restaurar(negocio.id).subscribe({
      next: (restaurado) => this.empresaLista(restaurado.id, restaurado.nombre_comercial),
      error: (respuesta: HttpErrorResponse) =>
        this.fallo(respuesta, 'No fue posible restaurar la empresa.'),
    });
  }

  crearSucursal(): void {
    if (this.modoConsulta()) return;
    this.sucursalForm.markAllAsTouched();
    this.error.set(null);
    this.campos.set({});
    const empresa = this.empresa();
    if (this.sucursalForm.invalid) {
      const basicos = this.sucursalForm.controls;
      if (!basicos.nombre.invalid && !basicos.codigo.invalid) this.masDatosSucursal.set(true);
      enfocarPrimerInvalido(this.host.nativeElement);
      return;
    }
    if (!empresa && this.empresaForm.invalid) {
      this.empresaForm.markAllAsTouched();
      this.cambiarPaso('empresa');
      enfocarPrimerInvalido(this.host.nativeElement);
      return;
    }
    if (this.guardando()) return;
    this.guardando.set(true);
    if (empresa) {
      this.crearSucursalParaEmpresa(empresa.id);
      return;
    }

    this.crearEmpresaYSucursal();
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
      else this.cambiarPaso('sucursal');
    };
    this.contexto.recargar().subscribe({ next: seguir, error: seguir });
  }

  private crearEmpresaYSucursal(): void {
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
        next: (negocio) => {
          this.contexto.seleccionarNegocio(negocio.id);
          this.empresa.set({ id: negocio.id, nombre: negocio.nombre_comercial });
          this.crearSucursalParaEmpresa(negocio.id);
        },
        error: (respuesta: HttpErrorResponse) =>
          this.fallo(respuesta, 'No fue posible registrar la empresa.'),
      });
  }

  private crearSucursalParaEmpresa(negocioId: string): void {
    const valor = this.sucursalForm.getRawValue();
    const direccion = this.direccion(valor);
    this.sucursales
      .crear(negocioId, {
        codigo: valor.codigo.trim().toUpperCase(),
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
            this.contexto.seleccionarNegocio(negocioId);
            this.contexto.seleccionarSucursal(sucursal.id);
            this.guardando.set(false);
            this.cambiarPaso('listo');
          };
          this.contexto.recargar().subscribe({ next: cerrar, error: cerrar });
        },
        error: (respuesta: HttpErrorResponse) =>
          this.fallo(respuesta, 'No fue posible crear la sucursal.'),
      });
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
      next: ({ items }) =>
        this.archivadas.set(items.filter((item) => item.tipo_miembro === 'propietario')),
      // Sin la lista el asistente sigue sirviendo: solo no ofrece restaurar.
      error: () => this.archivadas.set([]),
    });
  }

  private empresaDelContexto(): { id: string; nombre: string } | null {
    const negocio = this.contexto.negocio();
    return negocio ? { id: negocio.id, nombre: negocio.nombre_comercial } : null;
  }

  private direccion(valor: ReturnType<typeof this.sucursalForm.getRawValue>) {
    const partes = [
      valor.calle,
      valor.numero_exterior,
      valor.colonia,
      valor.ciudad,
      valor.estado,
      valor.codigo_postal,
    ];
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

  irPaso(paso: Paso): void {
    if (this.guardando() || (paso === 'listo' && !this.altaCompletada())) return;
    this.cambiarPaso(paso);
  }

  posicionRelativa(paso: Paso): PosicionRelativa {
    const indice = this.ordenPasos.indexOf(paso);
    const activo = this.ordenPasos.indexOf(this.paso());
    return Math.max(-2, Math.min(2, indice - activo)) as PosicionRelativa;
  }

  private cambiarPaso(paso: Paso): void {
    if (paso === this.paso()) return;
    if (paso === 'listo') this.altaCompletada.set(true);
    if (this.vistaCompacta()) {
      ++this.cambioVistaPrevia;
      this.detenerAnimacionVistaPrevia();
      this.paso.set(paso);
      return;
    }
    this.posicionesCardsAntes = this.medidasCards();
    const cambio = ++this.cambioVistaPrevia;
    this.detenerAnimacionVistaPrevia();
    this.paso.set(paso);
    afterNextRender(
      () => {
        if (cambio === this.cambioVistaPrevia) this.animarVistaPrevia();
      },
      { injector: this.injector },
    );
  }

  private animarVistaPrevia(): void {
    if (this.vistaCompacta()) {
      this.limpiarVistaPrevia();
      return;
    }
    const lateral = this.vistaPreviaLateral?.nativeElement;
    if (!lateral) return;

    const panel = lateral.querySelector<HTMLElement>('.vista-previa');
    if (!panel) return;

    const cards = Array.from(panel.querySelectorAll<HTMLElement>('.vista-previa__card'));

    this.vistaPreviaTimeline?.kill();
    gsap.killTweensOf(cards);
    this.vistaPreviaTimeline = gsap.timeline({ defaults: { ease: 'power2.inOut' } });

    const geometria = this.geometriaOrbital(panel);
    if (!geometria) return;

    const contenedor = panel.querySelector<HTMLElement>('.vista-previa__cards--carrusel');
    if (contenedor) this.aplicarGeometriaCss(contenedor, panel, geometria);

    if (this.movimientoReducido()) {
      gsap.set(cards, { clearProps: 'transform' });
      cards.forEach((card) => {
        const paso = card.dataset['paso'];
        const posicion = this.posicionOrbital(card);
        this.aplicarVisual(card, this.valoresVisuales(posicion));
        if (paso) this.angulosCardsActuales.set(paso, ANGULOS_ORBITA[posicion]);
      });
      return;
    }

    gsap.set(cards, { clearProps: 'transform' });
    cards.forEach((card) => {
      const paso = card.dataset['paso'];
      const antes = paso ? this.posicionesCardsAntes.get(paso) : undefined;
      const despues = card.getBoundingClientRect();
      const posicionDestino = this.posicionOrbital(card);
      if (!antes || antes.posicion === posicionDestino) return;

      const anguloDestino = ANGULOS_ORBITA[posicionDestino];
      const centroDestino = {
        x: despues.left + despues.width / 2,
        y: despues.top + despues.height / 2,
      };
      const angulo = { valor: antes.angulo };
      const visualInicial = this.valoresVisuales(antes.posicion);
      const visualDestino = this.valoresVisuales(posicionDestino);
      if (paso) this.angulosCardsActuales.set(paso, antes.angulo);

      this.aplicarPunto(card, angulo.valor, geometria, centroDestino);
      this.aplicarVisual(card, visualInicial);

      this.vistaPreviaTimeline?.to(
        angulo,
        {
          duration: ORBITA.duracion,
          valor: anguloDestino,
          onUpdate: () => {
            const progreso = gsap.utils.normalize(antes.angulo, anguloDestino, angulo.valor);
            this.aplicarPunto(card, angulo.valor, geometria, centroDestino);
            this.aplicarVisual(card, {
              escala:
                visualInicial.escala + (visualDestino.escala - visualInicial.escala) * progreso,
              opacidad:
                visualInicial.opacidad +
                (visualDestino.opacidad - visualInicial.opacidad) * progreso,
              rotacion:
                visualInicial.rotacion +
                (visualDestino.rotacion - visualInicial.rotacion) * progreso,
            });
            if (paso) this.angulosCardsActuales.set(paso, angulo.valor);
          },
        },
        0,
      );
    });
    this.vistaPreviaTimeline?.eventCallback('onComplete', () => {
      gsap.set(cards, { clearProps: 'transform' });
    });
  }

  private detenerAnimacionVistaPrevia(): void {
    this.vistaPreviaTimeline?.kill();
    const lateral = this.vistaPreviaLateral?.nativeElement;
    if (!lateral) return;
    const cards = Array.from(lateral.querySelectorAll<HTMLElement>('.vista-previa__card'));
    gsap.killTweensOf(cards);
  }

  private medidasCards(): Map<string, CardAntes> {
    const lateral = this.vistaPreviaLateral?.nativeElement;
    if (!lateral) return new Map();
    return new Map(
      Array.from(lateral.querySelectorAll<HTMLElement>('.vista-previa__card'))
        .filter((card): card is HTMLElement & { dataset: DOMStringMap & { paso: string } } =>
          Boolean(card.dataset['paso']),
        )
        .map((card) => [
          card.dataset['paso'],
          (() => {
            const posicion = this.posicionOrbital(card);
            return {
              rect: card.getBoundingClientRect(),
              posicion,
              angulo:
                this.angulosCardsActuales.get(card.dataset['paso']) ?? ANGULOS_ORBITA[posicion],
            };
          })(),
        ]),
    );
  }

  private posicionOrbital(card: HTMLElement): PosicionOrbital {
    if (card.classList.contains('vista-previa__card--activo')) return 'activa';
    const relativo = Number(card.dataset['relativo']);
    if (relativo === -2) return 'superior';
    if (relativo === -1) return 'superiorIntermedia';
    if (relativo === 1) return 'intermedia';
    if (relativo === 2) return 'inferior';
    return 'superior';
  }

  private geometriaOrbital(panel: HTMLElement): GeometriaOrbital | null {
    const halo = panel.querySelector<HTMLElement>('.vista-previa__halo');
    if (!halo) return null;
    const rect = halo.getBoundingClientRect();
    const radioVisual = Math.min(rect.width, rect.height) / 2;
    return {
      x: rect.left + rect.width / 2,
      y: rect.top + rect.height / 2,
      radio: radioVisual * ORBITA.radioHalo,
    };
  }

  private aplicarGeometriaCss(contenedor: HTMLElement, panel: HTMLElement, geometria: GeometriaOrbital): void {
    const rect = panel.getBoundingClientRect();
    contenedor.style.setProperty('--orbita-centro-x', `${geometria.x - rect.left}px`);
    contenedor.style.setProperty('--orbita-centro-y', `${geometria.y - rect.top}px`);
    contenedor.style.setProperty('--orbita-radio', `${geometria.radio}px`);
  }

  private observarGeometriaOrbital(): void {
    if (this.vistaCompacta()) return;
    const lateral = this.vistaPreviaLateral?.nativeElement;
    const panel = lateral?.querySelector<HTMLElement>('.vista-previa');
    const contenedor = panel?.querySelector<HTMLElement>('.vista-previa__cards--carrusel');
    const halo = panel?.querySelector<HTMLElement>('.vista-previa__halo');
    if (!panel || !contenedor || !halo) return;

    const actualizar = () => {
      if (this.vistaCompacta()) return;
      const geometria = this.geometriaOrbital(panel);
      if (geometria) this.aplicarGeometriaCss(contenedor, panel, geometria);
    };
    actualizar();
    this.orbitaObserver?.disconnect();
    if (typeof ResizeObserver !== 'undefined') {
      this.orbitaObserver = new ResizeObserver(actualizar);
      this.orbitaObserver.observe(panel);
      this.orbitaObserver.observe(halo);
    }
  }

  private readonly alCambiarModoResponsivo = (): void => {
    ++this.cambioVistaPrevia;
    this.detenerAnimacionVistaPrevia();
    if (this.vistaCompacta()) {
      this.orbitaObserver?.disconnect();
      this.orbitaObserver = undefined;
      this.limpiarVistaPrevia();
      return;
    }

    this.observarGeometriaOrbital();
    this.restaurarVistaPreviaSinAnimacion();
  };

  private iniciarModoResponsivo(): void {
    if (typeof window === 'undefined') return;
    this.mediaVistaCompacta = window.matchMedia('(max-width: 1199.98px)');
    this.mediaVistaCompacta.addEventListener('change', this.alCambiarModoResponsivo);
    this.alCambiarModoResponsivo();
  }

  private vistaCompacta(): boolean {
    if (typeof window === 'undefined') return true;
    return window.matchMedia('(max-width: 1199.98px)').matches;
  }

  private limpiarVistaPrevia(): void {
    const lateral = this.vistaPreviaLateral?.nativeElement;
    if (!lateral) return;
    const cards = Array.from(lateral.querySelectorAll<HTMLElement>('.vista-previa__card'));
    const carrusel = lateral.querySelector<HTMLElement>('.vista-previa__cards--carrusel');
    gsap.killTweensOf(cards);
    gsap.set(cards, { clearProps: 'transform' });
    cards.forEach((card) => {
      card.style.removeProperty('--card-escala');
      card.style.removeProperty('--card-opacidad');
      card.style.removeProperty('--card-rotacion');
    });
    carrusel?.style.removeProperty('--orbita-centro-x');
    carrusel?.style.removeProperty('--orbita-centro-y');
    carrusel?.style.removeProperty('--orbita-radio');
  }

  private restaurarVistaPreviaSinAnimacion(): void {
    const lateral = this.vistaPreviaLateral?.nativeElement;
    const panel = lateral?.querySelector<HTMLElement>('.vista-previa');
    const carrusel = panel?.querySelector<HTMLElement>('.vista-previa__cards--carrusel');
    if (!panel || !carrusel) return;

    const geometria = this.geometriaOrbital(panel);
    if (geometria) this.aplicarGeometriaCss(carrusel, panel, geometria);
    const cards = Array.from(panel.querySelectorAll<HTMLElement>('.vista-previa__card'));
    gsap.set(cards, { clearProps: 'transform' });
    cards.forEach((card) => {
      card.style.removeProperty('--card-escala');
      card.style.removeProperty('--card-opacidad');
      card.style.removeProperty('--card-rotacion');
      const paso = card.dataset['paso'];
      if (paso) this.angulosCardsActuales.set(paso, ANGULOS_ORBITA[this.posicionOrbital(card)]);
    });
  }

  private aplicarPunto(
    card: HTMLElement,
    angulo: number,
    geometria: GeometriaOrbital,
    centroDestino: { x: number; y: number },
  ): void {
    const punto = this.puntoOrbital(angulo, geometria);
    gsap.set(card, {
      x: punto.x - centroDestino.x,
      y: punto.y - centroDestino.y,
      force3D: true,
    });
  }

  private aplicarVisual(
    card: HTMLElement,
    visual: { escala: number; opacidad: number; rotacion: number },
  ): void {
    card.style.setProperty('--card-escala', String(visual.escala));
    card.style.setProperty('--card-opacidad', String(visual.opacidad));
    card.style.setProperty('--card-rotacion', `${visual.rotacion}deg`);
  }

  private valoresVisuales(posicion: PosicionOrbital): {
    escala: number;
    opacidad: number;
    rotacion: number;
  } {
    if (posicion === 'activa') {
      return { escala: ORBITA.escalaActiva, opacidad: ORBITA.opacidadActiva, rotacion: 0 };
    }
    const lejana = posicion === 'superior' || posicion === 'inferior';
    return {
      escala: lejana ? ORBITA.escalaLejana : ORBITA.escalaCercana,
      opacidad: lejana ? ORBITA.opacidadLejana : ORBITA.opacidadCercana,
      rotacion:
        posicion === 'superior' || posicion === 'superiorIntermedia'
          ? -(lejana ? ORBITA.rotacionLejana : ORBITA.rotacionCercana)
          : lejana
            ? ORBITA.rotacionLejana
            : ORBITA.rotacionCercana,
    };
  }

  private puntoOrbital(angulo: number, geometria: GeometriaOrbital) {
    const radianes = (angulo * Math.PI) / 180;
    return {
      x: geometria.x + geometria.radio * Math.cos(radianes),
      y: geometria.y - geometria.radio * Math.sin(radianes),
    };
  }

  private movimientoReducido(): boolean {
    if (typeof window === 'undefined') return true;
    return window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false;
  }
}

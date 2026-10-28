import { HttpErrorResponse } from '@angular/common/http';
import {
  afterNextRender,
  Component,
  HostListener,
  computed,
  DestroyRef,
  effect,
  ElementRef,
  inject,
  signal,
  untracked,
  viewChild,
} from '@angular/core';
import { RouterLink } from '@angular/router';
import { Subscription, timer } from 'rxjs';
import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { PuedeDirective } from '../../contexto/puede.directive';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import {
  aCentavos,
  aPesos,
  esCodigoEscaneado,
  lineaDesdeProducto,
  margen,
  nivelMargen,
  resumir,
} from './venta-calculos';
import {
  ErrorVenta,
  LineaVenta,
  ProductoVendible,
  ResultadoBusqueda,
  VentaRegistrada,
} from './ventas.models';
import { TicketImpresoComponent } from './ticket-impreso.component';
import { VentasService } from './ventas.service';

const formatoDinero = new Intl.NumberFormat('es-MX', { style: 'currency', currency: 'MXN' });
const formatoPorcentaje = new Intl.NumberFormat('es-MX', {
  style: 'percent',
  maximumFractionDigits: 1,
});

/** Billetes que se ofrecen como recibido rápido; solo los que cubren el total. */
const billetes = [2000, 5000, 10000, 20000, 50000, 100000];
const formatoRecibido = /^\d{1,9}(\.\d{1,2})?$/;

/** Convierte el texto capturado («100», «100.5») a centavos sin pasar por flotantes; null si no es válido. */
export function recibidoEnCentavos(texto: string): number | null {
  const limpio = texto.trim().replace(/^\$/, '').replace(/,/g, '');
  if (!formatoRecibido.test(limpio)) return null;
  const [pesos, fraccion = ''] = limpio.split('.');
  return Number(pesos) * 100 + Number(fraccion.padEnd(2, '0'));
}

interface Aviso {
  tipo: 'error' | 'atencion';
  texto: string;
}

@Component({
  selector: 'app-nueva-venta',
  standalone: true,
  imports: [RouterLink, PuedeDirective, TicketImpresoComponent],
  templateUrl: './nueva-venta.component.html',
  styleUrl: './nueva-venta.component.css',
})
export class NuevaVentaComponent {
  readonly P = PERMISOS;
  private readonly service = inject(VentasService);
  private readonly contexto = inject(ContextoService);
  private readonly feedback = inject(FeedbackService);
  private readonly campo = viewChild<ElementRef<HTMLInputElement>>('campo');
  private readonly dialogoCobro = viewChild<ElementRef<HTMLDialogElement>>('dialogoCobro');
  private readonly campoRecibido = viewChild<ElementRef<HTMLInputElement>>('campoRecibido');
  private readonly botonNuevaVenta = viewChild<ElementRef<HTMLButtonElement>>('botonNuevaVenta');

  readonly negocio = this.contexto.negocio;
  readonly sucursal = this.contexto.sucursal;
  readonly texto = signal('');
  readonly coincidencias = signal<ProductoVendible[]>([]);
  readonly abierto = signal(false);
  readonly resaltado = signal(0);
  readonly buscando = signal(false);
  readonly aviso = signal<Aviso | null>(null);
  readonly lineas = signal<LineaVenta[]>([]);
  /** Última fila agregada o modificada; se resalta en la tabla. */
  readonly ultima = signal<string | null>(null);
  readonly verificarCostos = signal(false);
  readonly puedeVerCostos = computed(() => this.contexto.puede(PERMISOS.ventaVerCostos));
  readonly mostrarCostos = computed(() => this.verificarCostos() && this.puedeVerCostos());
  readonly resumen = computed(() => resumir(this.lineas()));
  readonly bajoCosto = computed(() =>
    this.lineas().some(
      (linea) => nivelMargen(linea.precioCentavos, linea.costoCentavos) === 'negativo',
    ),
  );

  readonly cobroAbierto = signal(false);
  readonly recibidoTexto = signal('');
  readonly cobrando = signal(false);
  readonly errorCobro = signal<string | null>(null);
  /** Última venta registrada; la pantalla de venta terminada llega con su captura de diseño. */
  readonly terminada = signal<VentaRegistrada | null>(null);
  readonly recibido = computed(() => recibidoEnCentavos(this.recibidoTexto()));
  readonly falta = computed(() => Math.max(0, this.resumen().total - (this.recibido() ?? 0)));
  readonly cambio = computed(() => {
    const recibido = this.recibido();
    return recibido === null ? null : Math.max(0, recibido - this.resumen().total);
  });
  readonly puedeConfirmar = computed(
    () =>
      this.recibido() !== null &&
      this.falta() === 0 &&
      !this.cobrando() &&
      this.lineas().length > 0,
  );
  readonly billetesRapidos = computed(() =>
    billetes.filter((billete) => billete >= this.resumen().total).slice(0, 3),
  );

  /** Negocio y sucursal a los que pertenece la lista cargada; se cambian junto con ella. */
  private readonly borrador = signal<{ negocioId: string; sucursalId: string } | null>(null);
  private busquedaDiferida?: Subscription;
  private busquedaEnCurso?: Subscription;

  constructor() {
    effect(() => {
      const negocioId = this.contexto.negocio()?.id;
      const sucursalId = this.contexto.sucursal()?.id;
      if (!negocioId || !sucursalId) return;
      untracked(() => {
        const actual = this.borrador();
        if (actual?.negocioId === negocioId && actual.sucursalId === sucursalId) return;
        this.lineas.set(this.service.leerBorrador(negocioId, sucursalId));
        this.borrador.set({ negocioId, sucursalId });
        this.limpiarCampo();
      });
    });
    effect(() => {
      const borrador = this.borrador();
      const lineas = this.lineas();
      if (borrador) this.service.guardarBorrador(borrador.negocioId, borrador.sucursalId, lineas);
    });
    afterNextRender(() => this.enfocar());
    inject(DestroyRef).onDestroy(() => {
      this.busquedaDiferida?.unsubscribe();
      this.busquedaEnCurso?.unsubscribe();
    });
  }

  dinero(centavos: number): string {
    return formatoDinero.format(aPesos(centavos));
  }

  /** Importes de una venta registrada, que la API devuelve en pesos. */
  dineroPesos(pesos: number): string {
    return formatoDinero.format(pesos);
  }

  centavos(pesos: number): number {
    return aCentavos(pesos);
  }

  porcentaje(valor: number | null): string {
    return valor === null ? '—' : formatoPorcentaje.format(valor);
  }

  margenDe(linea: LineaVenta): number | null {
    return margen(linea.precioCentavos, linea.costoCentavos);
  }

  nivelDe(linea: LineaVenta) {
    return nivelMargen(linea.precioCentavos, linea.costoCentavos);
  }

  utilidadDe(linea: LineaVenta): number | null {
    return linea.costoCentavos === null
      ? null
      : (linea.precioCentavos - linea.costoCentavos) * linea.cantidad;
  }

  alEscribir(valor: string): void {
    this.texto.set(valor);
    this.aviso.set(null);
    this.busquedaDiferida?.unsubscribe();
    const texto = valor.trim();
    // Un código escaneado llega completo con Enter; solo el texto escrito a mano abre coincidencias.
    if (texto.length < 2 || esCodigoEscaneado(texto)) {
      this.abierto.set(false);
      return;
    }
    this.busquedaDiferida = timer(200).subscribe(() => this.buscar(texto, false));
  }

  alPresionarTecla(evento: KeyboardEvent): void {
    switch (evento.key) {
      case 'Enter':
        evento.preventDefault();
        this.confirmarCampo();
        break;
      case 'Escape':
        evento.preventDefault();
        this.limpiarCampo();
        this.aviso.set(null);
        break;
      case 'ArrowDown':
      case 'ArrowUp': {
        const total = this.coincidencias().length;
        if (!this.abierto() || !total) return;
        evento.preventDefault();
        const paso = evento.key === 'ArrowDown' ? 1 : -1;
        this.resaltado.update((actual) => (actual + paso + total) % total);
        break;
      }
    }
  }

  /** Enter agrega la coincidencia resaltada o, si no hay desplegable, busca el texto y agrega el código exacto. */
  private confirmarCampo(): void {
    const texto = this.texto().trim();
    if (!texto) return;
    this.busquedaDiferida?.unsubscribe();
    const resaltada = this.abierto() ? this.coincidencias()[this.resaltado()] : undefined;
    if (resaltada && !esCodigoEscaneado(texto)) {
      this.agregar(resaltada);
      return;
    }
    this.buscar(texto, true);
  }

  private buscar(texto: string, alConfirmar: boolean): void {
    const negocioId = this.contexto.negocio()?.id;
    const sucursalId = this.contexto.sucursal()?.id;
    if (!negocioId || !sucursalId) return;
    // Al escanear se limpia el campo de inmediato para que la siguiente lectura no se mezcle.
    if (alConfirmar) this.limpiarCampo();
    else this.busquedaEnCurso?.unsubscribe();
    this.buscando.set(true);
    const peticion = this.service.buscar(negocioId, sucursalId, texto).subscribe({
      next: (resultado) =>
        alConfirmar
          ? this.resolverConfirmacion(texto, resultado)
          : this.mostrarCoincidencias(texto, resultado),
      error: () =>
        this.aviso.set({
          tipo: 'error',
          texto: 'No fue posible buscar el producto. Intenta de nuevo.',
        }),
      complete: () => this.buscando.set(false),
    });
    if (!alConfirmar) this.busquedaEnCurso = peticion;
  }

  private mostrarCoincidencias(texto: string, resultado: ResultadoBusqueda): void {
    if (this.texto().trim() !== texto) return;
    this.coincidencias.set(resultado.items);
    this.resaltado.set(
      Math.max(
        0,
        resultado.items.findIndex((item) => !item.agotado),
      ),
    );
    this.abierto.set(true);
  }

  private resolverConfirmacion(texto: string, resultado: ResultadoBusqueda): void {
    const disponibles = resultado.items;
    if (disponibles.length === 1 || (resultado.por_codigo && disponibles.length)) {
      this.agregar(disponibles[0]);
      return;
    }
    if (!disponibles.length) {
      this.aviso.set({
        tipo: 'error',
        texto:
          resultado.por_codigo || esCodigoEscaneado(texto)
            ? `No encontramos el código «${texto}» en esta sucursal.`
            : `No encontramos productos con «${texto}».`,
      });
      return;
    }
    // Varias coincidencias por nombre: se muestran para elegir con flechas y Enter.
    this.texto.set(texto);
    this.mostrarCoincidencias(texto, resultado);
  }

  agregar(producto: ProductoVendible): void {
    this.limpiarCampo();
    if (producto.agotado) {
      this.aviso.set({
        tipo: 'atencion',
        texto: `${producto.nombre} no tiene existencias en esta sucursal.`,
      });
      this.enfocar();
      return;
    }
    const existente = this.lineas().find(
      (linea) => linea.productoNegocioId === producto.producto_negocio_id,
    );
    if (existente && existente.cantidad >= producto.existencia) {
      this.aviso.set({
        tipo: 'atencion',
        texto: `Ya agregaste todas las existencias de ${producto.nombre} (${producto.existencia} pz).`,
      });
      this.enfocar();
      return;
    }
    const nueva = lineaDesdeProducto(producto);
    this.lineas.update((lineas) =>
      existente
        ? lineas.map((linea) =>
            linea === existente ? { ...nueva, cantidad: linea.cantidad + 1 } : linea,
          )
        : [...lineas, nueva],
    );
    this.ultima.set(producto.producto_negocio_id);
    this.aviso.set(null);
    this.enfocar();
  }

  cambiarCantidad(linea: LineaVenta, paso: number): void {
    const cantidad = linea.cantidad + paso;
    // Bajar siempre se permite: tras un aviso de existencias la cantidad puede quedar sobre el tope.
    if (cantidad < 1 || (paso > 0 && cantidad > linea.existencia)) return;
    this.lineas.update((lineas) =>
      lineas.map((actual) =>
        actual.productoNegocioId === linea.productoNegocioId ? { ...actual, cantidad } : actual,
      ),
    );
    this.ultima.set(linea.productoNegocioId);
  }

  quitar(linea: LineaVenta): void {
    const posicion = this.lineas().findIndex(
      (actual) => actual.productoNegocioId === linea.productoNegocioId,
    );
    if (posicion < 0) return;
    this.lineas.update((lineas) =>
      lineas.filter((actual) => actual.productoNegocioId !== linea.productoNegocioId),
    );
    this.feedback.conAccion(`Quitaste ${linea.nombre}`, 'Deshacer', () =>
      this.restaurar(linea, posicion),
    );
    this.enfocar();
  }

  private restaurar(linea: LineaVenta, posicion: number): void {
    if (this.lineas().some((actual) => actual.productoNegocioId === linea.productoNegocioId))
      return;
    this.lineas.update((lineas) => [
      ...lineas.slice(0, posicion),
      linea,
      ...lineas.slice(posicion),
    ]);
    this.ultima.set(linea.productoNegocioId);
  }

  async vaciar(): Promise<void> {
    if (!this.lineas().length) return;
    const confirmado = await this.feedback.confirmDanger({
      titulo: 'Cancelar y vaciar la lista',
      descripcion: 'Se quitarán todos los productos de esta venta. No se registra nada.',
      textoConfirmar: 'Vaciar lista',
      textoCancelar: 'Seguir vendiendo',
    });
    if (!confirmado) return;
    this.lineas.set([]);
    this.ultima.set(null);
    this.enfocar();
  }

  abrirCobro(): void {
    if (!this.lineas().length) return;
    this.recibidoTexto.set('');
    this.errorCobro.set(null);
    this.cobroAbierto.set(true);
    const dialogo = this.dialogoCobro()?.nativeElement;
    if (dialogo && !dialogo.open) dialogo.showModal();
    const recibido = this.campoRecibido()?.nativeElement;
    if (recibido) {
      recibido.value = '';
      recibido.focus();
    }
  }

  cerrarCobro(): void {
    if (this.cobrando()) return;
    this.cobroAbierto.set(false);
    const dialogo = this.dialogoCobro()?.nativeElement;
    if (dialogo?.open) dialogo.close();
    this.enfocar();
  }

  usarRecibido(centavos: number): void {
    const texto = (centavos / 100).toFixed(2);
    this.recibidoTexto.set(texto);
    const recibido = this.campoRecibido()?.nativeElement;
    if (recibido) {
      recibido.value = texto;
      recibido.focus();
    }
  }

  confirmarCobro(): void {
    const negocioId = this.contexto.negocio()?.id;
    const sucursalId = this.contexto.sucursal()?.id;
    const recibido = this.recibido();
    if (!negocioId || !sucursalId || recibido === null || !this.puedeConfirmar()) return;
    this.cobrando.set(true);
    this.errorCobro.set(null);
    this.service
      .registrar(negocioId, {
        sucursal_id: sucursalId,
        forma_pago: 'EFECTIVO',
        recibido: (recibido / 100).toFixed(2),
        lineas: this.lineas().map((linea) => ({
          producto_negocio_id: linea.productoNegocioId,
          cantidad: linea.cantidad,
        })),
      })
      .subscribe({
        next: (venta) => {
          this.cobrando.set(false);
          this.terminada.set(venta);
          this.lineas.set([]);
          this.ultima.set(null);
          this.cerrarCobro();
          // Con el foco en «Nueva venta», Enter empieza la siguiente sin tocar el mouse.
          setTimeout(() => this.botonNuevaVenta()?.nativeElement.focus());
        },
        error: (respuesta: HttpErrorResponse) => {
          this.cobrando.set(false);
          this.errorCobro.set(this.mensajeDeError(respuesta.error as ErrorVenta | null));
        },
      });
  }

  imprimir(): void {
    window.print();
  }

  nuevaVenta(): void {
    this.terminada.set(null);
    this.limpiarCampo();
    this.aviso.set(null);
    setTimeout(() => this.enfocar());
  }

  /** Un producto que ya no alcanza ajusta su tope en la lista para corregirlo sin volver a escanear. */
  private mensajeDeError(error: ErrorVenta | null): string {
    if (error?.codigo === 'EXISTENCIA_INSUFICIENTE' && error.producto) {
      const disponible = error.disponible ?? 0;
      this.lineas.update((lineas) =>
        lineas.map((linea) =>
          linea.nombre === error.producto ? { ...linea, existencia: disponible } : linea,
        ),
      );
      return error.mensaje ?? `Ya no hay existencias suficientes de ${error.producto}.`;
    }
    if (error?.codigo === 'PRODUCTO_NO_DISPONIBLE')
      return 'Uno de los productos ya no está disponible para venta en esta sucursal. Quítalo y vuelve a cobrar.';
    return (
      error?.mensaje ?? 'No fue posible registrar la venta. Revisa tu conexión e intenta de nuevo.'
    );
  }

  private limpiarCampo(): void {
    this.texto.set('');
    this.abierto.set(false);
    this.coincidencias.set([]);
    this.resaltado.set(0);
    const campo = this.campo()?.nativeElement;
    if (campo) campo.value = '';
  }

  /** F2 dentro de la captura regresa el foco al campo; tras cobrar empieza la siguiente venta. */
  @HostListener('document:keydown.f2')
  alPresionarF2(): void {
    if (this.cobroAbierto()) return;
    if (this.terminada()) this.nuevaVenta();
    else this.enfocar();
  }

  enfocar(): void {
    this.campo()?.nativeElement.focus();
  }
}

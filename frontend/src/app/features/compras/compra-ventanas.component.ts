import { CommonModule } from '@angular/common';
import { Component, computed, effect, inject, input, output, signal, untracked } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { PuedeDirective } from '../../contexto/puede.directive';
import { TASAS_IEPS, TASAS_IVA } from './impuestos';
import { EditPurchaseRequest, PurchaseRow, PurchaseView, PurchaseViewRow } from './compras.models';
import { ComprasService, guardarArchivo } from './compras.service';
import { VentanaCompraComponent } from './ventana-compra.component';

export type TipoAccion = 'ver' | 'editar' | 'suspender' | 'reactivar';
export interface AccionCompra {
  tipo: TipoAccion;
  compra: PurchaseRow;
}

/** Renglón editable: empaques, costo e impuestos (en %); lo demás se conserva tal cual. */
interface RenglonEditable {
  fila: PurchaseViewRow;
  empaques: number;
  costo: number;
  ieps: number;
  iva: number;
}

const ETIQUETAS_COSTO: Record<string, string> = { PENDIENTE: 'Por validar', AUTORIZADA: 'Validado', CANCELADA: 'Cancelado' };

/** Ventanas Detalle, Editar, Suspender y Reactivar de una compra (tablero 07, formato del tablero 12). */
@Component({
  selector: 'app-compra-ventanas',
  standalone: true,
  imports: [CommonModule, FormsModule, PuedeDirective, VentanaCompraComponent],
  templateUrl: './compra-ventanas.component.html',
})
export class CompraVentanasComponent {
  readonly P = PERMISOS;
  readonly tasasIeps = TASAS_IEPS;
  readonly tasasIva = TASAS_IVA;
  readonly documentTypes = ['TICKET', 'FACTURA'];
  readonly paymentMethods = ['EFECTIVO', 'TRANSFERENCIA', 'TARJETA', 'CREDITO', 'OTRO'];
  private readonly service = inject(ComprasService);
  private readonly context = inject(ContextoService);
  readonly accion = input<AccionCompra | null>(null);
  readonly cerrar = output<void>();
  /** Se emite después de editar, suspender o reactivar para recargar la lista. */
  readonly cambio = output<string>();
  readonly tipo = signal<TipoAccion | null>(null);
  readonly compra = signal<PurchaseView | null>(null);
  readonly cargando = signal(false);
  readonly guardando = signal(false);
  readonly error = signal<string | null>(null);
  readonly motivo = signal('');
  readonly encabezado = signal({ tipo_documento: 'TICKET', folio_documento: '', fecha_documento: '', fecha_recepcion: '', forma_pago: 'EFECTIVO', numero_cuenta: '', observaciones: '' });
  readonly renglones = signal<RenglonEditable[]>([]);

  readonly suspendida = computed(() => this.compra()?.estado === 'SUSPENDIDA');
  readonly pendientes = computed(() => this.compra()?.renglones.filter((r) => r.estado_costo === 'PENDIENTE').length ?? 0);
  readonly validados = computed(() => this.compra()?.renglones.filter((r) => r.estado_costo === 'AUTORIZADA').length ?? 0);
  readonly cancelados = computed(() => this.compra()?.renglones.filter((r) => r.estado_costo === 'CANCELADA').length ?? 0);
  readonly piezas = computed(() => this.compra()?.renglones.reduce((total, r) => total + r.piezas_buenas, 0) ?? 0);
  /** Con algún costo validado los productos quedan bloqueados; solo se corrige el documento. */
  readonly bloqueada = computed(() => this.validados() > 0);
  readonly titulo = computed(() => {
    const compra = this.compra();
    const documento = compra ? `${compra.tipo_documento === 'FACTURA' ? 'Factura' : 'Ticket'} ${compra.folio_documento}` : 'Compra';
    switch (this.tipo()) {
      case 'editar':
        return `Editar ${documento}`;
      case 'suspender':
        return `Suspender ${documento}`;
      case 'reactivar':
        return `Reactivar ${documento}`;
      default:
        return documento;
    }
  });
  readonly subtitulo = computed(() => {
    const compra = this.compra();
    if (!compra) return '';
    return `${compra.proveedor ?? ''} · recibida ${compra.fecha_recepcion.slice(0, 10).split('-').reverse().join('/')}`;
  });

  constructor() {
    effect(() => {
      const accion = this.accion();
      untracked(() => (accion ? this.abrir(accion) : this.tipo.set(null)));
    });
  }

  private abrir(accion: AccionCompra): void {
    const negocio = this.context.negocio()?.id;
    if (!negocio) return;
    this.tipo.set(accion.tipo);
    this.error.set(null);
    this.motivo.set('');
    this.cargando.set(true);
    this.compra.set(null);
    this.service.view(negocio, accion.compra.id).subscribe({
      next: (compra) => {
        this.compra.set(compra);
        this.prepararEdicion(compra);
        this.cargando.set(false);
      },
      error: () => {
        this.error.set('No fue posible cargar la compra.');
        this.cargando.set(false);
      },
    });
  }
  private prepararEdicion(compra: PurchaseView): void {
    this.encabezado.set({
      tipo_documento: compra.tipo_documento,
      folio_documento: compra.folio_documento,
      fecha_documento: compra.fecha_documento.slice(0, 10),
      fecha_recepcion: compra.fecha_recepcion.slice(0, 10),
      forma_pago: compra.forma_pago,
      numero_cuenta: compra.numero_cuenta ?? '',
      observaciones: compra.observaciones ?? '',
    });
    this.renglones.set(
      compra.renglones.map((fila) => ({
        fila,
        empaques: Number(fila.cantidad_empaques),
        costo: Number(fila.costo_empaque),
        // Compra anterior sin tasas por renglón: arranca sin IEPS e IVA 16 %, lo más común.
        ieps: fila.impuestos_por_renglon ? Math.round(Number(fila.tasa_ieps) * 1000) / 10 : 0,
        iva: fila.impuestos_por_renglon ? Math.round(Number(fila.tasa_iva) * 100) : 16,
      })),
    );
  }

  ir(tipo: TipoAccion): void {
    this.error.set(null);
    this.motivo.set('');
    this.tipo.set(tipo);
  }
  cerrarVentana(): void {
    this.tipo.set(null);
    this.cerrar.emit();
  }
  etiquetaCosto(estado: string): string {
    return ETIQUETAS_COSTO[estado] ?? 'Sin costo';
  }
  cambiarEncabezado(campo: string, valor: string): void {
    this.encabezado.update((actual) => ({ ...actual, [campo]: valor }));
  }
  cambiarRenglon(indice: number, cambios: Partial<RenglonEditable>): void {
    this.renglones.update((renglones) => renglones.map((r, i) => (i === indice ? { ...r, ...cambios } : r)));
  }

  guardarEdicion(): void {
    const compra = this.compra();
    const negocio = this.context.negocio()?.id;
    if (!compra || !negocio || this.guardando()) return;
    const encabezado = this.encabezado();
    if (!encabezado.folio_documento.trim()) {
      this.error.set('El folio es obligatorio.');
      return;
    }
    const request: EditPurchaseRequest = {
      ...encabezado,
      numero_cuenta: encabezado.numero_cuenta || null,
      observaciones: encabezado.observaciones || null,
      motivo: this.motivo().trim() || null,
    };
    if (!this.bloqueada()) {
      request.detalles = this.renglones().map(({ fila, empaques, costo, ieps, iva }) => ({
        id: fila.id,
        producto_negocio_id: fila.producto_negocio_id,
        unidad_medida_id: fila.unidad_medida_id,
        cantidad_empaques: Number(empaques),
        costo_empaque: Number(costo),
        descuento: Number(fila.descuento),
        piezas_danadas: fila.piezas_danadas,
        piezas_faltantes: fila.piezas_faltantes,
        numero_lote: fila.numero_lote,
        fecha_caducidad: fila.fecha_caducidad ? fila.fecha_caducidad.slice(0, 10) : null,
        producto_nuevo: false,
        nombre_producto_nuevo: '',
        codigo_barras_nuevo: '',
        tasa_ieps: Number(ieps) / 100,
        tasa_iva: Number(iva) / 100,
      }));
    }
    this.ejecutar(this.service.edit(negocio, compra.id, request), 'Compra actualizada.');
  }
  confirmarSuspension(): void {
    const compra = this.compra();
    const negocio = this.context.negocio()?.id;
    if (!compra || !negocio || !this.motivo().trim()) return;
    this.ejecutar(this.service.suspend(negocio, compra.id, this.motivo().trim()), 'Compra suspendida.');
  }
  confirmarReactivacion(): void {
    const compra = this.compra();
    const negocio = this.context.negocio()?.id;
    if (!compra || !negocio) return;
    this.ejecutar(this.service.reactivate(negocio, compra.id, this.motivo().trim()), 'Compra reactivada.');
  }
  private ejecutar(peticion: ReturnType<ComprasService['edit']>, mensaje: string): void {
    if (this.guardando()) return;
    this.guardando.set(true);
    this.error.set(null);
    peticion.subscribe({
      next: () => {
        this.guardando.set(false);
        this.tipo.set(null);
        this.cambio.emit(mensaje);
        this.cerrar.emit();
      },
      error: (respuesta) => {
        this.guardando.set(false);
        this.error.set(respuesta.error?.mensaje ?? 'No fue posible guardar los cambios.');
      },
    });
  }
  descargarExcel(): void {
    const compra = this.compra();
    const negocio = this.context.negocio()?.id;
    const sucursal = this.context.sucursal()?.id;
    if (!compra || !negocio || !sucursal) return;
    this.service.exportar(negocio, { sucursal_id: sucursal, compra_id: [compra.id], comprobacion: true, incluir_suspendidas: true }).subscribe({
      next: (archivo) => guardarArchivo(archivo, `compra-${compra.folio_documento}.xlsx`),
      error: () => this.error.set('No fue posible generar el Excel.'),
    });
  }
}

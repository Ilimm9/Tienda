import { CommonModule } from '@angular/common';
import { Component, computed, inject, input, output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { PuedeDirective } from '../../contexto/puede.directive';
import { PurchaseRow } from './compras.models';
import { ComprasService, guardarArchivo } from './compras.service';
import { VentanaCompraComponent } from './ventana-compra.component';

/** Columnas de «hoja 1», para que la ventana diga qué trae el archivo. */
export const COLUMNAS_EXCEL: [string, string][] = [
  ['A', 'No. de ticket'], ['B', 'Fecha de recepción'], ['C', 'Código'], ['D', 'Producto'], ['E', 'Proveedor'], ['F', '¿IEPS?'],
  ['G', 'Costo sin impuesto'], ['H', 'Costo con impuesto'], ['I', 'Cantidad comprada'], ['J', 'Unidad'], ['K', 'Cantidad'],
  ['L', 'Total de piezas'], ['M', 'Costo por caja'], ['N', 'Costo por pieza'], ['O', 'Costo en sistema'], ['P', 'Diferencia'],
  ['Q', 'Observaciones'], ['R', 'P.V en sistema'], ['S', 'Nuevo P.V'], ['T', 'Producto nuevo'], ['U', 'Nombre del nuevo'],
  ['V', 'Gramaje'], ['W', 'Precio de venta'], ['X', 'Dañadas'], ['Y', 'Faltantes'], ['Z', 'Buenas'], ['AA', 'Descuento'],
  ['AB', 'IVA'], ['AC', 'Estado de validación'], ['AD', 'Estado de la compra'],
];

export interface OpcionesExcel {
  alcance: 'seleccion' | 'filtro';
  proveedor: string;
  desde: string;
  hasta: string;
  comprobacion: boolean;
  suspendidas: boolean;
}

const ETIQUETAS_PAGO: Record<string, string> = { EFECTIVO: 'Efectivo', TRANSFERENCIA: 'Transferencia', TARJETA: 'Tarjeta', CREDITO: 'Crédito', OTRO: 'Otro' };

/** Lista de compras del tablero 07: filtros, selección, estado de costos y Excel. */
@Component({
  selector: 'app-compras-lista',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, PuedeDirective, VentanaCompraComponent],
  templateUrl: './compras-lista.component.html',
  styleUrl: './compras-lista.component.css',
})
export class ComprasListaComponent {
  readonly P = PERMISOS;
  readonly columnas = COLUMNAS_EXCEL;
  private readonly service = inject(ComprasService);
  private readonly context = inject(ContextoService);
  readonly compras = input.required<PurchaseRow[]>();
  /** Se emite cuando una acción cambia compras y hay que recargar la lista. */
  readonly cambio = output<void>();
  /** Ver, editar, suspender o reactivar: las ventanas las abre el contenedor. */
  readonly accion = output<{ tipo: 'ver' | 'editar' | 'suspender' | 'reactivar'; compra: PurchaseRow }>();
  readonly puedeGestionar = computed(() => this.context.puede(PERMISOS.compraEditar) && this.context.puede(PERMISOS.compraSuspender));
  readonly q = signal('');
  readonly proveedor = signal('');
  readonly desde = signal('');
  readonly hasta = signal('');
  readonly seleccion = signal(new Set<string>());
  readonly ventana = signal<'excel' | null>(null);
  readonly excel = signal<OpcionesExcel>({ alcance: 'filtro', proveedor: '', desde: '', hasta: '', comprobacion: true, suspendidas: false });
  readonly descargando = signal(false);
  readonly error = signal<string | null>(null);

  readonly proveedores = computed(() => {
    const nombres = new Map<string, string>();
    for (const compra of this.compras()) nombres.set(compra.proveedor_id, compra.proveedor || 'Proveedor');
    return [...nombres].map(([id, nombre]) => ({ id, nombre })).sort((a, b) => a.nombre.localeCompare(b.nombre));
  });
  readonly filtradas = computed(() => {
    const texto = this.q().trim().toLowerCase();
    return this.compras().filter((compra) => {
      const fecha = compra.fecha_recepcion.slice(0, 10);
      return (
        (!texto || `${compra.folio_documento} ${compra.proveedor ?? ''}`.toLowerCase().includes(texto)) &&
        (!this.proveedor() || compra.proveedor_id === this.proveedor()) &&
        (!this.desde() || fecha >= this.desde()) &&
        (!this.hasta() || fecha <= this.hasta())
      );
    });
  });
  readonly resumen = computed(() => `${this.filtradas().length} de ${this.compras().length} ${this.compras().length === 1 ? 'compra' : 'compras'}`);
  readonly costosPendientes = computed(() => this.compras().reduce((total, c) => total + (c.costos_pendientes ?? 0), 0));
  readonly comprasPendientes = computed(() => this.compras().filter((c) => (c.costos_pendientes ?? 0) > 0).length);
  readonly todas = computed(() => this.filtradas().length > 0 && this.filtradas().every((c) => this.seleccion().has(c.id)));
  readonly paraExcel = computed(() => {
    const opciones = this.excel();
    if (opciones.alcance === 'seleccion') {
      return this.compras().filter((c) => this.seleccion().has(c.id) && (opciones.suspendidas || !this.suspendida(c)));
    }
    return this.compras().filter((c) => {
      const fecha = c.fecha_recepcion.slice(0, 10);
      return (
        (opciones.suspendidas || !this.suspendida(c)) &&
        (!opciones.proveedor || c.proveedor_id === opciones.proveedor) &&
        (!opciones.desde || fecha >= opciones.desde) &&
        (!opciones.hasta || fecha <= opciones.hasta)
      );
    });
  });
  readonly resumenExcel = computed(() => {
    const compras = this.paraExcel();
    return {
      compras: compras.length,
      renglones: compras.reduce((total, c) => total + (c.productos ?? 0), 0),
      importe: compras.reduce((total, c) => total + Number(c.total), 0),
      pendientes: compras.reduce((total, c) => total + (c.costos_pendientes ?? 0), 0),
    };
  });

  suspendida(compra: PurchaseRow): boolean {
    return compra.estado === 'SUSPENDIDA';
  }
  documento(compra: PurchaseRow): string {
    return `${compra.tipo_documento === 'FACTURA' ? 'Factura' : 'Ticket'} ${compra.folio_documento}`;
  }
  pago(compra: PurchaseRow): string {
    return ETIQUETAS_PAGO[compra.forma_pago] ?? compra.forma_pago;
  }
  tieneIeps(compra: PurchaseRow): boolean {
    return Number(compra.ieps ?? 0) > 0;
  }
  /** «N por validar», «Validados» o «Cancelados». */
  costos(compra: PurchaseRow): { texto: string; tipo: 'pendiente' | 'validado' | 'cancelado' | 'ninguno' } {
    const pendientes = compra.costos_pendientes ?? 0;
    if (pendientes > 0) return { texto: `${pendientes} por validar`, tipo: 'pendiente' };
    if ((compra.costos_validados ?? 0) > 0) return { texto: 'Validados', tipo: 'validado' };
    if ((compra.costos_cancelados ?? 0) > 0) return { texto: 'Cancelados', tipo: 'cancelado' };
    return { texto: 'Sin costos', tipo: 'ninguno' };
  }
  alternar(compra: PurchaseRow, marcar: boolean): void {
    this.seleccion.update((actual) => {
      const siguiente = new Set(actual);
      if (marcar) siguiente.add(compra.id);
      else siguiente.delete(compra.id);
      return siguiente;
    });
  }
  alternarTodas(marcar: boolean): void {
    this.seleccion.set(marcar ? new Set(this.filtradas().map((c) => c.id)) : new Set());
  }
  limpiarSeleccion(): void {
    this.seleccion.set(new Set());
  }

  /** Abre el Excel con las seleccionadas o con los filtros de la tabla. */
  abrirExcel(alcance?: 'seleccion' | 'filtro'): void {
    this.error.set(null);
    this.excel.set({
      alcance: alcance ?? (this.seleccion().size ? 'seleccion' : 'filtro'),
      proveedor: this.proveedor(),
      desde: this.desde(),
      hasta: this.hasta(),
      comprobacion: true,
      suspendidas: false,
    });
    this.ventana.set('excel');
  }
  /** Excel de una sola compra, desde su detalle. */
  descargarCompra(compra: PurchaseRow): void {
    this.descargar({ compra_id: [compra.id], incluir_suspendidas: true });
  }
  cambiarExcel(cambios: Partial<OpcionesExcel>): void {
    this.excel.update((actual) => ({ ...actual, ...cambios }));
  }
  descargarExcel(): void {
    const opciones = this.excel();
    if (opciones.alcance === 'seleccion') {
      this.descargar({ compra_id: [...this.seleccion()], comprobacion: opciones.comprobacion, incluir_suspendidas: opciones.suspendidas });
    } else {
      this.descargar({ proveedor_id: opciones.proveedor || undefined, desde: opciones.desde || undefined, hasta: opciones.hasta || undefined, comprobacion: opciones.comprobacion, incluir_suspendidas: opciones.suspendidas });
    }
  }
  private descargar(filtro: { compra_id?: string[]; proveedor_id?: string; desde?: string; hasta?: string; comprobacion?: boolean; incluir_suspendidas: boolean }): void {
    const negocio = this.context.negocio()?.id;
    const sucursal = this.context.sucursal()?.id;
    if (!negocio || !sucursal || this.descargando()) return;
    this.descargando.set(true);
    this.service.exportar(negocio, { sucursal_id: sucursal, comprobacion: true, ...filtro }).subscribe({
      next: (archivo) => {
        guardarArchivo(archivo, `compras-${new Date().toISOString().slice(0, 10)}.xlsx`);
        this.descargando.set(false);
        this.ventana.set(null);
      },
      error: () => {
        this.descargando.set(false);
        this.error.set('No fue posible generar el Excel.');
      },
    });
  }
}

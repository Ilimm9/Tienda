import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { InputTextModule } from 'primeng/inputtext';
import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { PuedeDirective } from '../../contexto/puede.directive';
import { margenDePrecio, precioSugerido } from './precio-sugerido';
import { ComprasService, guardarArchivo } from '../compras/compras.service';
import { VentanaCompraComponent } from '../compras/ventana-compra.component';
import { Autorizacion, FiltroPendientes, PriceHistoryItem, PriceProposal, PreciosService } from './precios.service';

/** Renglón de Validar costos con los importes ya numéricos y lo que la sucursal edita. */
export interface Renglon {
  id: string;
  compraId: string;
  productoId: string;
  producto: string;
  nota: string;
  nuevo: boolean;
  piezas: number;
  costo: number;
  costoAnterior: number | null;
  precioAnterior: number;
  sugerido: number;
  margen: number;
  precio: number;
}
export interface Ticket {
  compraId: string;
  documento: string;
  proveedor: string;
  impuestos: string;
  fechaRecepcion: string;
  total: number;
  renglones: Renglon[];
}

const numero = (value: unknown): number => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
};
const porcentaje = (tasa: number): string => `${Math.round(tasa * 1000) / 10} %`;

@Component({
  selector: 'app-precios',
  standalone: true,
  imports: [PuedeDirective, CommonModule, FormsModule, InputTextModule, RouterLink, VentanaCompraComponent],
  templateUrl: './precios.component.html',
  styleUrl: './precios.component.css',
})
export class PreciosComponent {
  readonly P = PERMISOS;
  private readonly service = inject(PreciosService);
  private readonly context = inject(ContextoService);
  private readonly compras = inject(ComprasService);
  readonly renglones = signal<Renglon[]>([]);
  readonly tickets = signal<Ticket[]>([]);
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly aviso = signal<string | null>(null);
  readonly saving = signal<string | null>(null);
  readonly seleccion = signal(new Set<string>());
  readonly plegados = signal(new Set<string>());
  readonly margenMasivo = signal(30);
  readonly filtro = signal<FiltroPendientes>({});
  readonly proveedores = signal<{ id: string; nombre: string }[]>([]);
  /** Sin filtros: distingue «Todo validado» de «Ningún ticket coincide». */
  readonly hayFiltro = computed(() => Object.values(this.filtro()).some(Boolean));
  readonly gruposVisibles = computed(() => {
    const porCompra = new Map(this.renglones().map((renglon) => [renglon.id, renglon]));
    return this.tickets()
      .map((ticket) => ({ ...ticket, renglones: ticket.renglones.map((r) => porCompra.get(r.id)).filter((r): r is Renglon => !!r) }))
      .filter((ticket) => ticket.renglones.length > 0);
  });
  readonly resumenFiltro = computed(() => {
    const tickets = this.gruposVisibles().length;
    const productos = this.renglones().length;
    return `${tickets} ${tickets === 1 ? 'ticket' : 'tickets'} · ${productos} ${productos === 1 ? 'producto' : 'productos'}`;
  });
  readonly conCostoMasAlto = computed(() => this.renglones().filter((r) => r.costoAnterior !== null && r.costo > r.costoAnterior).length);
  readonly gananciaTotal = computed(() => this.renglones().reduce((total, r) => total + this.gananciaTotalRenglon(r), 0));
  readonly excelAbierto = signal(false);
  readonly excelAlcance = signal<'filtro' | 'seleccion'>('filtro');
  readonly excelComprobacion = signal(true);
  readonly descargando = signal(false);
  /** Tickets del Excel: los del filtro o los que tienen algún renglón seleccionado. */
  readonly ticketsExcel = computed(() => {
    const tickets = this.gruposVisibles();
    if (this.excelAlcance() === 'filtro') return tickets;
    return tickets.filter((ticket) => ticket.renglones.some((r) => this.seleccion().has(r.id)));
  });
  readonly pendientesExcel = computed(() => this.ticketsExcel().reduce((total, ticket) => total + ticket.renglones.length, 0));
  readonly seleccionados = computed(() => this.renglones().filter((r) => this.seleccion().has(r.id)));
  // Historial oculto temporalmente (2026-10-21); se conserva la lógica.
  readonly history = signal<PriceHistoryItem[]>([]);
  readonly historyProduct = signal<string | null>(null);
  readonly manualPrice = signal<number | null>(null);
  readonly manualReason = signal('');
  private temporizador: ReturnType<typeof setTimeout> | undefined;

  constructor() {
    this.load();
  }

  load(): void {
    const business = this.context.negocio()?.id;
    const branch = this.context.sucursal()?.id;
    if (!business || !branch) {
      this.error.set('Selecciona una sucursal para validar costos.');
      this.loading.set(false);
      return;
    }
    this.service.pending(business, branch, this.filtro()).subscribe({
      next: (response) => {
        this.cargar(response.items);
        this.loading.set(false);
      },
      error: (response) => {
        this.error.set(response.error?.mensaje ?? 'No fue posible cargar las propuestas.');
        this.loading.set(false);
      },
    });
  }

  private cargar(items: PriceProposal[]): void {
    const tickets = new Map<string, Ticket>();
    const renglones: Renglon[] = [];
    const proveedores = new Map(this.proveedores().map((p) => [p.id, p.nombre]));
    for (const item of items) {
      const costo = numero(item.costo_capturado);
      const sugerido = numero(item.precio_sugerido);
      const renglon: Renglon = {
        id: item.id,
        compraId: item.compra_id,
        productoId: item.producto_negocio_id,
        producto: item.producto,
        nota: item.nota,
        nuevo: item.producto_nuevo,
        piezas: numero(item.piezas_buenas),
        costo,
        costoAnterior: item.costo_anterior === null || item.costo_anterior === undefined ? null : numero(item.costo_anterior),
        precioAnterior: numero(item.precio_anterior),
        sugerido,
        margen: Math.round(numero(item.margen_sugerido) * 10000) / 100,
        precio: sugerido,
      };
      renglones.push(renglon);
      proveedores.set(item.proveedor_id, item.proveedor);
      let ticket = tickets.get(item.compra_id);
      if (!ticket) {
        ticket = {
          compraId: item.compra_id,
          documento: `${item.tipo_documento === 'FACTURA' ? 'Factura' : 'Ticket'} ${item.folio_documento}`,
          proveedor: item.proveedor,
          impuestos: '',
          fechaRecepcion: item.fecha_recepcion,
          total: numero(item.total_compra),
          renglones: [],
        };
        tickets.set(item.compra_id, ticket);
      }
      ticket.renglones.push(renglon);
    }
    for (const [compraId, ticket] of tickets) {
      const tasas = items.filter((item) => item.compra_id === compraId && item.impuestos_por_renglon);
      const ieps = [...new Set(tasas.map((item) => numero(item.tasa_ieps)).filter((t) => t > 0))];
      const iva = [...new Set(tasas.map((item) => numero(item.tasa_iva)).filter((t) => t > 0))];
      ticket.impuestos = [...ieps.map((t) => `IEPS ${porcentaje(t)}`), ...iva.map((t) => `IVA ${porcentaje(t)}`)].join(' · ');
    }
    this.renglones.set(renglones);
    this.tickets.set([...tickets.values()]);
    this.proveedores.set([...proveedores].map(([id, nombre]) => ({ id, nombre })).sort((a, b) => a.nombre.localeCompare(b.nombre)));
    this.seleccion.set(new Set());
  }

  filtrar(cambios: Partial<FiltroPendientes>, esperar = false): void {
    this.filtro.update((actual) => ({ ...actual, ...cambios }));
    clearTimeout(this.temporizador);
    if (esperar) this.temporizador = setTimeout(() => this.load(), 300);
    else this.load();
  }
  limpiarFiltros(): void {
    this.filtro.set({});
    this.load();
  }

  private actualizar(id: string, cambios: Partial<Renglon>): void {
    this.renglones.update((renglones) => renglones.map((r) => (r.id === id ? { ...r, ...cambios } : r)));
  }
  /** Cambiar el margen recalcula el precio con la regla del backend: hacia arriba al peso. */
  cambiarMargen(renglon: Renglon, margen: number): void {
    this.actualizar(renglon.id, { margen: numero(margen), precio: precioSugerido(renglon.costo, numero(margen)) });
  }
  /** Un precio escrito a mano se respeta y recalcula el margen. */
  cambiarPrecio(renglon: Renglon, precio: number): void {
    this.actualizar(renglon.id, { precio: numero(precio), margen: margenDePrecio(renglon.costo, numero(precio)) });
  }
  aplicarMargen(): void {
    const ids = this.seleccion();
    const margen = numero(this.margenMasivo());
    this.renglones.update((renglones) => renglones.map((r) => (ids.has(r.id) ? { ...r, margen, precio: precioSugerido(r.costo, margen) } : r)));
  }

  ganancia(r: Renglon): number {
    return Math.round((r.precio - r.costo) * 100) / 100;
  }
  gananciaTotalRenglon(r: Renglon): number {
    return Math.round(this.ganancia(r) * r.piezas * 100) / 100;
  }
  gananciaTicket(ticket: Ticket): number {
    return ticket.renglones.reduce((total, r) => total + this.gananciaTotalRenglon(r), 0);
  }
  /** «▲ 4.2 %», «▼ 1 %», «Sin cambio» o «Primer costo». */
  cambioCosto(r: Renglon): { texto: string; tipo: 'sube' | 'baja' | 'igual' | 'nuevo' } {
    if (r.costoAnterior === null || r.costoAnterior <= 0) return { texto: 'Primer costo', tipo: 'nuevo' };
    const cambio = Math.round(((r.costo - r.costoAnterior) / r.costoAnterior) * 1000) / 10;
    if (Math.abs(r.costo - r.costoAnterior) < 0.005) return { texto: 'Sin cambio', tipo: 'igual' };
    return cambio > 0 ? { texto: `▲ ${cambio} %`, tipo: 'sube' } : { texto: `▼ ${Math.abs(cambio)} %`, tipo: 'baja' };
  }

  estaPlegado(ticket: Ticket): boolean {
    return this.plegados().has(ticket.compraId);
  }
  alternarPlegado(ticket: Ticket): void {
    this.plegados.update((actual) => {
      const siguiente = new Set(actual);
      if (!siguiente.delete(ticket.compraId)) siguiente.add(ticket.compraId);
      return siguiente;
    });
  }
  ticketCompleto(ticket: Ticket): boolean {
    return ticket.renglones.every((r) => this.seleccion().has(r.id));
  }
  alternarTicket(ticket: Ticket, marcar: boolean): void {
    this.seleccion.update((actual) => {
      const siguiente = new Set(actual);
      for (const r of ticket.renglones) if (marcar) siguiente.add(r.id); else siguiente.delete(r.id);
      return siguiente;
    });
  }
  alternarRenglon(r: Renglon, marcar: boolean): void {
    this.seleccion.update((actual) => {
      const siguiente = new Set(actual);
      if (marcar) siguiente.add(r.id); else siguiente.delete(r.id);
      return siguiente;
    });
  }

  private autorizacion(r: Renglon): Autorizacion {
    return { costo: r.costo, margen: r.margen / 100, precio_venta: r.precio };
  }
  private quitar(ids: string[], mensaje: string): void {
    const quitados = new Set(ids);
    this.renglones.update((renglones) => renglones.filter((r) => !quitados.has(r.id)));
    this.seleccion.update((actual) => new Set([...actual].filter((id) => !quitados.has(id))));
    this.aviso.set(mensaje);
  }
  validar(r: Renglon): void {
    const business = this.context.negocio()?.id;
    if (!business || this.saving()) return;
    this.saving.set(r.id);
    this.service.authorize(business, r.id, this.autorizacion(r)).subscribe({
      next: () => {
        this.quitar([r.id], `${r.producto} validado.`);
        this.saving.set(null);
      },
      error: (response) => {
        this.error.set(response.error?.mensaje ?? 'No fue posible validar el renglón.');
        this.saving.set(null);
      },
    });
  }
  /** Valida en una transacción los renglones dados de un mismo ticket. */
  validarTicket(ticket: Ticket, renglones: Renglon[] = ticket.renglones): void {
    const business = this.context.negocio()?.id;
    if (!business || this.saving() || !renglones.length) return;
    this.saving.set(ticket.compraId);
    this.service
      .authorizePurchase(business, ticket.compraId, renglones.map((r) => ({ propuesta_id: r.id, ...this.autorizacion(r) })))
      .subscribe({
        next: () => {
          this.quitar(renglones.map((r) => r.id), `${ticket.documento}: ${renglones.length} ${renglones.length === 1 ? 'producto validado' : 'productos validados'}.`);
          this.saving.set(null);
        },
        error: (response) => {
          this.error.set(response.error?.mensaje ?? 'No fue posible validar el ticket.');
          this.saving.set(null);
        },
      });
  }
  /** Una llamada por ticket con sus renglones seleccionados. */
  validarSeleccionados(): void {
    const elegidos = this.seleccion();
    for (const ticket of this.gruposVisibles()) {
      const renglones = ticket.renglones.filter((r) => elegidos.has(r.id));
      if (!renglones.length) continue;
      const business = this.context.negocio()?.id;
      if (!business) return;
      this.service.authorizePurchase(business, ticket.compraId, renglones.map((r) => ({ propuesta_id: r.id, ...this.autorizacion(r) }))).subscribe({
        next: () => this.quitar(renglones.map((r) => r.id), 'Seleccionados validados.'),
        error: (response) => this.error.set(response.error?.mensaje ?? `No fue posible validar ${ticket.documento}.`),
      });
    }
  }

  abrirExcel(): void {
    this.excelAlcance.set(this.seleccion().size ? 'seleccion' : 'filtro');
    this.excelAbierto.set(true);
  }
  descargarExcel(): void {
    const negocio = this.context.negocio()?.id;
    const sucursal = this.context.sucursal()?.id;
    const tickets = this.ticketsExcel();
    if (!negocio || !sucursal || !tickets.length || this.descargando()) return;
    this.descargando.set(true);
    this.compras
      .exportar(negocio, { sucursal_id: sucursal, compra_id: tickets.map((t) => t.compraId), comprobacion: this.excelComprobacion(), incluir_suspendidas: false })
      .subscribe({
        next: (archivo) => {
          guardarArchivo(archivo, `compras-${new Date().toISOString().slice(0, 10)}.xlsx`);
          this.descargando.set(false);
          this.excelAbierto.set(false);
        },
        error: () => {
          this.descargando.set(false);
          this.error.set('No fue posible generar el Excel.');
        },
      });
  }

  etiquetaOrigen(origen: PriceHistoryItem['origen']): string { return ETIQUETAS_ORIGEN[origen] ?? origen; }
  showHistory(r: Renglon): void { const business = this.context.negocio()?.id; const branch = this.context.sucursal()?.id; if (!business || !branch) return; this.historyProduct.set(r.productoId); this.manualPrice.set(r.precioAnterior); this.manualReason.set(''); this.service.history(business, branch, r.productoId).subscribe({ next: response => this.history.set(response.items), error: response => this.error.set(response.error?.mensaje ?? 'No fue posible cargar el historial.') }); }
  adjustPrice(): void { const business = this.context.negocio()?.id; const branch = this.context.sucursal()?.id; const product = this.historyProduct(); const price = this.manualPrice(); if (!business || !branch || !product || price === null || this.saving()) return; this.saving.set(product); this.service.adjust(business, product, { sucursal_id: branch, precio_venta: price, motivo: this.manualReason() }).subscribe({ next: item => { this.history.update(items => [item, ...items]); this.manualReason.set(''); this.saving.set(null); }, error: response => { this.error.set(response.error?.mensaje ?? 'No fue posible actualizar el precio.'); this.saving.set(null); } }); }
}

const ETIQUETAS_ORIGEN: Record<PriceHistoryItem['origen'], string> = {
  ALTA_INICIAL: 'Alta inicial',
  COMPRA_AUTORIZADA: 'Compra autorizada',
  AJUSTE_PRECIO_MANUAL: 'Ajuste manual',
  CARGA_MASIVA: 'Carga masiva',
};

import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { InputTextModule } from 'primeng/inputtext';
import { ContextoService } from '../../contexto/contexto.service';
import { PriceProposal, PreciosService } from './precios.service';
import { PERMISOS } from '../../contexto/permisos';
import { PuedeDirective } from '../../contexto/puede.directive';

@Component({ selector: 'app-precios', standalone: true, imports: [PuedeDirective, CommonModule, FormsModule, InputTextModule], templateUrl: './precios.component.html', styleUrl: './precios.component.css' })
export class PreciosComponent {
  readonly P = PERMISOS;
  private readonly service = inject(PreciosService); private readonly context = inject(ContextoService);
  readonly items = signal<PriceProposal[]>([]); readonly loading = signal(true); readonly error = signal<string | null>(null); readonly saving = signal<string | null>(null);
  constructor() { this.load(); }
  load(): void { const business = this.context.negocio()?.id; const branch = this.context.sucursal()?.id; if (!business || !branch) { this.error.set('Selecciona una sucursal para validar costos.'); this.loading.set(false); return; } this.service.pending(business, branch).subscribe({ next: r => { this.items.set(r.items.map(item => { const costo = this.number(item.costo_capturado); const margen = this.number(item.margen_sugerido) * 100; return {...item, piezas_buenas: this.number(item.piezas_buenas), costo_capturado: costo, costo_anterior: item.costo_anterior === null ? null : this.number(item.costo_anterior), precio_anterior: this.number(item.precio_anterior), margen_sugerido: margen, precio_sugerido: this.number(item.precio_sugerido), ganancia: Math.round(costo * margen) / 100}; })); this.loading.set(false); }, error: response => { this.error.set(response.error?.mensaje ?? 'No fue posible cargar las propuestas.'); this.loading.set(false); } }); }
  private number(value: unknown): number { const number = Number(value); return Number.isFinite(number) ? number : 0; }
  finalPrice(item: PriceProposal): number { return Number(item.costo_capturado) + Number(item.ganancia ?? 0); }
  totalProfit(item: PriceProposal): number { return Number(item.ganancia ?? 0) * Number(item.piezas_buenas); }
  syncProfitFromPercent(item: PriceProposal): void { item.ganancia = Math.round(Number(item.costo_capturado) * Number(item.margen_sugerido)) / 100; }
  syncPercentFromProfit(item: PriceProposal): void { item.margen_sugerido = Number(item.costo_capturado) ? Math.round(Number(item.ganancia ?? 0) / Number(item.costo_capturado) * 10000) / 100 : 0; }
  authorize(item: PriceProposal): void { const business = this.context.negocio()?.id; if (!business || this.saving()) return; this.saving.set(item.id); this.service.authorize(business, item.id, { costo: Number(item.costo_capturado), margen: Number(item.margen_sugerido) / 100, precio_venta: this.finalPrice(item) }).subscribe({ next: () => { this.items.update(items => items.filter(current => current.id !== item.id)); this.saving.set(null); }, error: response => { this.error.set(response.error?.mensaje ?? 'No fue posible autorizar el renglón.'); this.saving.set(null); } }); }
}

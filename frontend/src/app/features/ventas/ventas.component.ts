import { Component, computed, effect, inject, signal, untracked } from '@angular/core';
import { RouterLink } from '@angular/router';
import { Subscription } from 'rxjs';
import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { PuedeDirective } from '../../contexto/puede.directive';
import { etiquetaPeriodo, fechaHoraCorta, horaVenta } from './venta-fechas';
import {
  FORMAS_PAGO,
  FormaPago,
  ListadoVentas,
  PeriodoVentas,
  VentaRegistrada,
} from './ventas.models';
import { VentasService } from './ventas.service';

const formatoDinero = new Intl.NumberFormat('es-MX', { style: 'currency', currency: 'MXN' });

const normalizar = (texto: string) =>
  texto.normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase().trim();

@Component({
  selector: 'app-ventas',
  standalone: true,
  imports: [RouterLink, PuedeDirective],
  templateUrl: './ventas.component.html',
  styleUrl: './ventas.component.css',
})
export class VentasComponent {
  readonly P = PERMISOS;
  readonly formasPago = FORMAS_PAGO;
  readonly periodos: readonly { valor: PeriodoVentas; etiqueta: string }[] = [
    { valor: 'hoy', etiqueta: 'Hoy' },
    { valor: 'semana', etiqueta: 'Semana' },
    { valor: 'mes', etiqueta: 'Mes' },
  ];
  private readonly service = inject(VentasService);
  private readonly contexto = inject(ContextoService);

  readonly sucursal = this.contexto.sucursal;
  readonly periodo = signal<PeriodoVentas>('hoy');
  readonly busqueda = signal('');
  readonly formaPago = signal<FormaPago | ''>('');
  readonly listado = signal<ListadoVentas | null>(null);
  readonly cargando = signal(true);
  readonly error = signal<string | null>(null);

  readonly etiqueta = computed(() =>
    etiquetaPeriodo(this.periodo(), this.listado()?.desde, this.listado()?.hasta),
  );
  readonly resumen = computed(() => this.listado()?.resumen ?? null);
  /** Folio, cajero y forma de pago acotan la tabla sobre las ventas ya cargadas del periodo. */
  readonly ventas = computed(() => {
    const texto = normalizar(this.busqueda());
    const forma = this.formaPago();
    return (this.listado()?.items ?? []).filter(
      (venta) =>
        (!forma || venta.forma_pago === forma) &&
        (!texto ||
          normalizar(venta.folio).includes(texto) ||
          normalizar(venta.atendio).includes(texto)),
    );
  });
  readonly textoPeriodo = computed(
    () => ({ hoy: 'hoy', semana: 'esta semana', mes: 'este mes' })[this.periodo()],
  );

  private peticion?: Subscription;

  constructor() {
    effect(() => {
      const negocioId = this.contexto.negocio()?.id;
      const sucursalId = this.contexto.sucursal()?.id;
      const periodo = this.periodo();
      if (negocioId && sucursalId) untracked(() => this.cargar(negocioId, sucursalId, periodo));
    });
  }

  private cargar(negocioId: string, sucursalId: string, periodo: PeriodoVentas): void {
    this.peticion?.unsubscribe();
    this.cargando.set(true);
    this.error.set(null);
    this.peticion = this.service.listar(negocioId, sucursalId, periodo).subscribe({
      next: (listado) => {
        this.listado.set(listado);
        this.cargando.set(false);
      },
      error: () => {
        this.error.set('No fue posible cargar las ventas. Intenta de nuevo.');
        this.cargando.set(false);
      },
    });
  }

  reintentar(): void {
    const negocioId = this.contexto.negocio()?.id;
    const sucursalId = this.contexto.sucursal()?.id;
    if (negocioId && sucursalId) this.cargar(negocioId, sucursalId, this.periodo());
  }

  dinero(pesos: number): string {
    return formatoDinero.format(pesos);
  }

  cuando(venta: VentaRegistrada): string {
    return this.periodo() === 'hoy' ? horaVenta(venta.creado_en) : fechaHoraCorta(venta.creado_en);
  }

  claseFormaPago(forma: FormaPago): string {
    return {
      EFECTIVO: 'pastilla--exito',
      TARJETA: 'pastilla--info',
      TRANSFERENCIA: 'pastilla--atencion',
    }[forma];
  }

  alElegirFormaPago(valor: string): void {
    this.formaPago.set(valor as FormaPago | '');
  }
}

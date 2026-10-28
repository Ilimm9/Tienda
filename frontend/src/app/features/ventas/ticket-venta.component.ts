import { HttpErrorResponse } from '@angular/common/http';
import {
  Component,
  computed,
  effect,
  ElementRef,
  inject,
  signal,
  untracked,
  viewChild,
} from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { map } from 'rxjs';
import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { TicketImpresoComponent } from './ticket-impreso.component';
import { fechaHoraLarga, horaVenta } from './venta-fechas';
import {
  ErrorVenta,
  FORMAS_PAGO,
  MOTIVOS_CANCELACION,
  MotivoCancelacion,
  VentaRegistrada,
} from './ventas.models';
import { VentasService } from './ventas.service';

const formatoDinero = new Intl.NumberFormat('es-MX', { style: 'currency', currency: 'MXN' });

@Component({
  selector: 'app-ticket-venta',
  standalone: true,
  imports: [RouterLink, TicketImpresoComponent],
  templateUrl: './ticket-venta.component.html',
  styleUrl: './ticket-venta.component.css',
})
export class TicketVentaComponent {
  readonly motivos = MOTIVOS_CANCELACION;
  readonly formasPago = FORMAS_PAGO;
  private readonly service = inject(VentasService);
  private readonly contexto = inject(ContextoService);
  private readonly feedback = inject(FeedbackService);
  private readonly dialogo = viewChild<ElementRef<HTMLDialogElement>>('dialogoCancelar');
  private readonly campoDetalle = viewChild<ElementRef<HTMLTextAreaElement>>('detalleCancelar');

  readonly ventaId = toSignal(
    inject(ActivatedRoute).paramMap.pipe(map((params) => params.get('ventaId') ?? '')),
    { initialValue: '' },
  );
  readonly venta = signal<VentaRegistrada | null>(null);
  readonly cargando = signal(true);
  readonly noEncontrada = signal(false);
  readonly error = signal<string | null>(null);

  readonly motivo = signal<MotivoCancelacion | null>(null);
  readonly detalle = signal('');
  readonly cancelando = signal(false);
  readonly errorCancelar = signal<string | null>(null);

  readonly puedeCancelar = computed(
    () => this.contexto.puede(PERMISOS.ventaCancelar) && this.venta()?.estado === 'PAGADA',
  );
  readonly motivoCompleto = computed(() => {
    const motivo = this.motivo();
    return !!motivo && (motivo !== 'OTRO' || this.detalle().trim().length > 0);
  });

  constructor() {
    effect(() => {
      const negocioId = this.contexto.negocio()?.id;
      const ventaId = this.ventaId();
      if (negocioId && ventaId) untracked(() => this.cargar(negocioId, ventaId));
    });
  }

  private cargar(negocioId: string, ventaId: string): void {
    this.cargando.set(true);
    this.error.set(null);
    this.noEncontrada.set(false);
    this.service.obtener(negocioId, ventaId).subscribe({
      next: (venta) => {
        this.venta.set(venta);
        this.cargando.set(false);
      },
      error: (respuesta: HttpErrorResponse) => {
        this.cargando.set(false);
        if (respuesta.status === 404) this.noEncontrada.set(true);
        else this.error.set('No fue posible cargar el ticket. Intenta de nuevo.');
      },
    });
  }

  dinero(pesos: number): string {
    return formatoDinero.format(pesos);
  }

  fecha(venta: VentaRegistrada): string {
    return fechaHoraLarga(venta.creado_en);
  }

  hora(iso: string): string {
    return horaVenta(iso);
  }

  reimprimir(): void {
    window.print();
  }

  abrirCancelar(): void {
    if (!this.venta()?.cancelable) return;
    this.motivo.set(null);
    this.detalle.set('');
    this.errorCancelar.set(null);
    const campo = this.campoDetalle()?.nativeElement;
    if (campo) campo.value = '';
    const dialogo = this.dialogo()?.nativeElement;
    if (dialogo && !dialogo.open) dialogo.showModal();
  }

  cerrarCancelar(): void {
    if (this.cancelando()) return;
    this.dialogo()?.nativeElement.close();
  }

  confirmarCancelar(): void {
    const negocioId = this.contexto.negocio()?.id;
    const venta = this.venta();
    const motivo = this.motivo();
    if (!negocioId || !venta || !motivo || !this.motivoCompleto() || this.cancelando()) return;
    this.cancelando.set(true);
    this.errorCancelar.set(null);
    this.service.cancelar(negocioId, venta.id, motivo, this.detalle().trim()).subscribe({
      next: (cancelada) => {
        this.cancelando.set(false);
        this.venta.set(cancelada);
        this.dialogo()?.nativeElement.close();
        this.feedback.success(
          `Venta ${cancelada.folio} cancelada`,
          'Las piezas regresaron al inventario.',
        );
      },
      error: (respuesta: HttpErrorResponse) => {
        this.cancelando.set(false);
        const error = respuesta.error as ErrorVenta | null;
        if (error?.codigo === 'VENTA_FUERA_DE_PLAZO') {
          this.venta.update((actual) => (actual ? { ...actual, cancelable: false } : actual));
          this.errorCancelar.set(
            'Solo se pueden cancelar ventas del mismo día. Esta venta ya no se puede cancelar.',
          );
        } else if (error?.codigo === 'VENTA_YA_CANCELADA') {
          this.errorCancelar.set(
            'Esta venta ya estaba cancelada. Recarga la página para ver el detalle.',
          );
        } else {
          this.errorCancelar.set(
            error?.mensaje ?? 'No fue posible cancelar la venta. Intenta de nuevo.',
          );
        }
      },
    });
  }
}

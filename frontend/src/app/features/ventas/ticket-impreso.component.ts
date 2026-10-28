import { Component, computed, inject, input } from '@angular/core';
import { ContextoService } from '../../contexto/contexto.service';
import { FechaMexicoPipe } from '../../shared/fecha-mexico.pipe';
import { VentaRegistrada } from './ventas.models';

const formatoDinero = new Intl.NumberFormat('es-MX', { style: 'currency', currency: 'MXN' });

/**
 * Ticket de 72 mm que solo aparece al imprimir; una regla global de `styles.css` oculta el resto de la página.
 * Lo usan la venta terminada y la reimpresión desde el detalle.
 */
@Component({
  selector: 'app-ticket-impreso',
  standalone: true,
  imports: [FechaMexicoPipe],
  template: `
    @let v = venta();
    <article class="ticket-imprimible" aria-hidden="true">
      <h2>{{ negocio()?.nombre_comercial }}</h2>
      <p>{{ sucursal() }}</p>
      <p>{{ v.folio }} · {{ v.creado_en | fechaMexico: 'dd/MM/yyyy HH:mm' }}</p>
      @if (v.estado === 'CANCELADA') {
        <p><strong>CANCELADA</strong></p>
      }
      <table>
        <tbody>
          @for (linea of v.lineas ?? []; track linea.id) {
            <tr>
              <td>{{ linea.cantidad }} × {{ linea.nombre }}</td>
              <td>{{ dinero(linea.importe) }}</td>
            </tr>
          }
        </tbody>
      </table>
      <dl>
        <div>
          <dt>Subtotal</dt>
          <dd>{{ dinero(v.subtotal) }}</dd>
        </div>
        <div>
          <dt>IVA (16%)</dt>
          <dd>{{ dinero(v.iva) }}</dd>
        </div>
        <div>
          <dt>Total</dt>
          <dd>{{ dinero(v.total) }}</dd>
        </div>
        <div>
          <dt>{{ v.forma_pago === 'EFECTIVO' ? 'Efectivo' : 'Pagado' }}</dt>
          <dd>{{ dinero(v.recibido) }}</dd>
        </div>
        <div>
          <dt>Cambio</dt>
          <dd>{{ dinero(v.cambio) }}</dd>
        </div>
      </dl>
      <p>Atendió {{ v.atendio }}</p>
      <p>Gracias por su compra</p>
    </article>
  `,
  styleUrl: './ticket-impreso.component.css',
})
export class TicketImpresoComponent {
  private readonly contexto = inject(ContextoService);
  readonly venta = input.required<VentaRegistrada>();
  readonly negocio = this.contexto.negocio;
  /** La sucursal de la venta, que puede no ser la activa al reimprimir. */
  readonly sucursal = computed(
    () =>
      this.negocio()?.sucursales?.find((item) => item.id === this.venta().sucursal_id)?.nombre ??
      this.contexto.sucursal()?.nombre ??
      '',
  );

  dinero(pesos: number): string {
    return formatoDinero.format(pesos);
  }
}

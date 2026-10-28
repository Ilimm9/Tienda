import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import {
  LineaVenta,
  ListadoVentas,
  MotivoCancelacion,
  PeriodoVentas,
  RegistrarVentaRequest,
  ResultadoBusqueda,
  VentaRegistrada,
} from './ventas.models';

@Injectable({ providedIn: 'root' })
export class VentasService {
  private readonly http = inject(HttpClient);
  private readonly storage = typeof localStorage === 'undefined' ? undefined : localStorage;

  private base(id: string): string {
    return `${environment.apiUrl}/negocios/${id}`;
  }

  buscar(negocioId: string, sucursalId: string, q: string): Observable<ResultadoBusqueda> {
    return this.http.get<ResultadoBusqueda>(`${this.base(negocioId)}/ventas/productos`, {
      params: { sucursal_id: sucursalId, q },
    });
  }

  registrar(negocioId: string, request: RegistrarVentaRequest): Observable<VentaRegistrada> {
    return this.http.post<VentaRegistrada>(`${this.base(negocioId)}/ventas`, request);
  }

  listar(negocioId: string, sucursalId: string, periodo: PeriodoVentas): Observable<ListadoVentas> {
    return this.http.get<ListadoVentas>(`${this.base(negocioId)}/ventas`, {
      params: { sucursal_id: sucursalId, periodo },
    });
  }

  obtener(negocioId: string, ventaId: string): Observable<VentaRegistrada> {
    return this.http.get<VentaRegistrada>(`${this.base(negocioId)}/ventas/${ventaId}`);
  }

  cancelar(
    negocioId: string,
    ventaId: string,
    motivo: MotivoCancelacion,
    detalle: string,
  ): Observable<VentaRegistrada> {
    return this.http.post<VentaRegistrada>(`${this.base(negocioId)}/ventas/${ventaId}/cancelar`, {
      motivo,
      detalle,
    });
  }

  /** La venta en captura sobrevive a una recarga accidental; se guarda por negocio y sucursal. */
  leerBorrador(negocioId: string, sucursalId: string): LineaVenta[] {
    try {
      const lineas: unknown = JSON.parse(
        this.storage?.getItem(this.claveBorrador(negocioId, sucursalId)) ?? '[]',
      );
      return Array.isArray(lineas) ? lineas.filter(esLineaValida) : [];
    } catch {
      // Sin almacenamiento disponible o con datos dañados se empieza con la lista vacía.
      return [];
    }
  }

  guardarBorrador(negocioId: string, sucursalId: string, lineas: readonly LineaVenta[]): void {
    try {
      const clave = this.claveBorrador(negocioId, sucursalId);
      if (lineas.length) this.storage?.setItem(clave, JSON.stringify(lineas));
      else this.storage?.removeItem(clave);
    } catch {
      // Sin almacenamiento la lista sigue en memoria mientras la pantalla esté abierta.
    }
  }

  private claveBorrador(negocioId: string, sucursalId: string): string {
    return `tienda.venta.borrador.${negocioId}.${sucursalId}`;
  }
}

function esLineaValida(valor: unknown): valor is LineaVenta {
  const linea = valor as LineaVenta;
  return (
    typeof linea?.productoNegocioId === 'string' &&
    typeof linea.nombre === 'string' &&
    Number.isInteger(linea.precioCentavos) &&
    Number.isInteger(linea.cantidad) &&
    linea.cantidad > 0 &&
    Number.isInteger(linea.existencia)
  );
}

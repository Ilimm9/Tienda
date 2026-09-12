import { HttpClient } from '@angular/common/http';
import { inject, Injectable, signal } from '@angular/core';
import { catchError, map, Observable, of, tap } from 'rxjs';

import { environment } from '../../environments/environment';
import { ContextoNegocio, ContextoOpcionesResponse, ContextoSucursal, EstadoContexto } from './contexto.models';

const negocioKey = 'tienda.contexto.negocio_id';
const sucursalKey = 'tienda.contexto.sucursal_id';

@Injectable({ providedIn: 'root' })
export class ContextoService {
  private readonly http = inject(HttpClient);
  private readonly storage = typeof localStorage === 'undefined' ? undefined : localStorage;
  readonly estado = signal<EstadoContexto>('cargando');
  readonly negocios = signal<ContextoNegocio[]>([]);
  readonly negocio = signal<ContextoNegocio | null>(null);
  readonly sucursal = signal<ContextoSucursal | null>(null);

  inicializar(): Observable<EstadoContexto> {
    this.estado.set('cargando');
    return this.http.get<ContextoOpcionesResponse>(`${environment.apiUrl}/contexto/opciones`).pipe(
      tap(({ items }) => this.resolver(items ?? [])),
      map(() => this.estado()),
      catchError(() => {
        this.limpiarSeleccion();
        this.negocios.set([]);
        this.negocio.set(null);
        this.sucursal.set(null);
        this.estado.set('error');
        return of('error' as EstadoContexto);
      }),
    );
  }

  seleccionarNegocio(id: string): void {
    const negocio = this.negocios().find((item) => item.id === id);
    if (!negocio) return;
    const sucursal = negocio.sucursales.find((item) => item.es_principal) ?? negocio.sucursales[0] ?? null;
    this.aplicar(negocio, sucursal);
  }

  seleccionarSucursal(id: string): void {
    const sucursal = this.negocio()?.sucursales.find((item) => item.id === id) ?? null;
    if (!sucursal || !this.negocio()) return;
    this.aplicar(this.negocio()!, sucursal);
  }

  limpiar(): void {
    this.limpiarSeleccion();
    this.negocios.set([]);
    this.negocio.set(null);
    this.sucursal.set(null);
    this.estado.set('requiere_negocio');
  }

  private resolver(items: ContextoNegocio[]): void {
    this.negocios.set(items);
    if (!items.length) {
      this.limpiarSeleccion();
      this.negocio.set(null);
      this.sucursal.set(null);
      this.estado.set('requiere_negocio');
      return;
    }
    const guardado = this.leer(negocioKey);
    const negocio = items.find((item) => item.id === guardado) ?? (items.length === 1 ? items[0] : null);
    if (!negocio) {
      this.limpiarSeleccion();
      this.negocio.set(null);
      this.sucursal.set(null);
      this.estado.set('requiere_negocio');
      return;
    }
    const sucursalGuardada = this.leer(sucursalKey);
    const sucursal = negocio.sucursales.find((item) => item.id === sucursalGuardada)
      ?? negocio.sucursales.find((item) => item.es_principal)
      ?? negocio.sucursales[0]
      ?? null;
    this.aplicar(negocio, sucursal);
  }

  private aplicar(negocio: ContextoNegocio, sucursal: ContextoSucursal | null): void {
    this.negocio.set(negocio);
    this.sucursal.set(sucursal);
    this.storage?.setItem(negocioKey, negocio.id);
    if (sucursal) this.storage?.setItem(sucursalKey, sucursal.id);
    else this.storage?.removeItem(sucursalKey);
    this.estado.set(sucursal ? 'listo' : 'sin_sucursal');
  }

  private leer(key: string): string | null {
    const value = this.storage?.getItem(key) ?? null;
    return value && /^[0-9a-f]{8}-[0-9a-f-]{27}$/i.test(value) ? value : null;
  }

  private limpiarSeleccion(): void {
    this.storage?.removeItem(negocioKey);
    this.storage?.removeItem(sucursalKey);
  }
}

import { HttpClient } from '@angular/common/http';
import { DestroyRef, inject, Injectable, signal } from '@angular/core';
import { catchError, finalize, map, Observable, of, shareReplay, tap } from 'rxjs';

import { environment } from '../../environments/environment';
import { ContextoNegocio, ContextoOpcionesResponse, ContextoSucursal, EstadoContexto } from './contexto.models';

const negocioKey = 'tienda.contexto.negocio_id';
const sucursalKey = 'tienda.contexto.sucursal_id';
const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

@Injectable({ providedIn: 'root' })
export class ContextoService {
  private readonly http = inject(HttpClient);
  private readonly storage = typeof localStorage === 'undefined' ? undefined : localStorage;
  readonly estado = signal<EstadoContexto>('cargando');
  readonly negocios = signal<ContextoNegocio[]>([]);
  readonly negocio = signal<ContextoNegocio | null>(null);
  readonly sucursal = signal<ContextoSucursal | null>(null);
  readonly inicializado = signal(false);

  // Deduplica inicializaciones concurrentes: varios guards pueden resolverse en la misma navegación.
  private enVuelo: Observable<EstadoContexto> | null = null;

  constructor() {
    if (typeof window === 'undefined') return;
    // Un cambio de contexto en otra pestaña nunca se adopta a ciegas: siempre revalida contra la API.
    const onStorage = (event: StorageEvent) => {
      if (event.key !== negocioKey && event.key !== sucursalKey && event.key !== null) return;
      this.recargar().subscribe();
    };
    window.addEventListener('storage', onStorage);
    inject(DestroyRef).onDestroy(() => window.removeEventListener('storage', onStorage));
  }

  /** Resuelve el contexto una sola vez; reutiliza el resultado en navegaciones posteriores. */
  asegurarInicializado(): Observable<EstadoContexto> {
    if (this.inicializado() && this.estado() !== 'error') return of(this.estado());
    return this.inicializar();
  }

  inicializar(): Observable<EstadoContexto> {
    if (this.enVuelo) return this.enVuelo;
    this.estado.set('cargando');
    this.enVuelo = this.http.get<ContextoOpcionesResponse>(`${environment.apiUrl}/contexto/opciones`).pipe(
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
      tap(() => this.inicializado.set(true)),
      finalize(() => (this.enVuelo = null)),
      shareReplay({ bufferSize: 1, refCount: false }),
    );
    return this.enVuelo;
  }

  /** Fuerza una revalidación contra la API, por ejemplo después de mutar negocios o sucursales. */
  recargar(): Observable<EstadoContexto> {
    this.inicializado.set(false);
    return this.inicializar();
  }

  seleccionarNegocio(id: string): void {
    const negocio = this.negocios().find((item) => item.id === id);
    if (!negocio) return;
    this.aplicar(negocio, this.sucursalPreferida(negocio, null));
  }

  seleccionarSucursal(id: string): void {
    const negocio = this.negocio();
    const sucursal = negocio?.sucursales.find((item) => item.id === id) ?? null;
    if (!negocio || !sucursal) return;
    this.aplicar(negocio, sucursal);
  }

  limpiar(): void {
    this.limpiarSeleccion();
    this.negocios.set([]);
    this.negocio.set(null);
    this.sucursal.set(null);
    this.inicializado.set(false);
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
    this.aplicar(negocio, this.sucursalPreferida(negocio, this.leer(sucursalKey)));
  }

  /** Conserva la sucursal guardada solo si sigue perteneciendo al negocio; si no, cae en la principal. */
  private sucursalPreferida(negocio: ContextoNegocio, guardada: string | null): ContextoSucursal | null {
    return (
      negocio.sucursales.find((item) => item.id === guardada)
      ?? negocio.sucursales.find((item) => item.es_principal)
      ?? negocio.sucursales[0]
      ?? null
    );
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
    let value: string | null = null;
    try {
      value = this.storage?.getItem(key) ?? null;
    } catch {
      return null;
    }
    return value && uuidPattern.test(value) ? value : null;
  }

  private limpiarSeleccion(): void {
    try {
      this.storage?.removeItem(negocioKey);
      this.storage?.removeItem(sucursalKey);
    } catch {
      // localStorage no disponible: el contexto sigue siendo válido solo en memoria.
    }
  }
}

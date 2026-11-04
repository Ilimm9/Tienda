import { HttpClient } from '@angular/common/http';
import { computed, effect, inject, Injectable, signal, untracked } from '@angular/core';
import { catchError, finalize, map, Observable, of, shareReplay, tap } from 'rxjs';

import { environment } from '../../environments/environment';
import { ContextoService } from './contexto.service';

/** Secciones que un negocio puede ocultar en Configuración > Vistas (adecuación 1). */
export type SeccionOcultable = 'categorias' | 'marcas';

export interface VistasNegocio {
  secciones_ocultas: SeccionOcultable[];
  secciones_ocultables: SeccionOcultable[];
}

interface EstadoVistas {
  negocioId: string | null;
  ocultas: ReadonlySet<SeccionOcultable>;
}

/**
 * Preferencia de vistas del negocio activo: qué secciones ve todo el equipo.
 * Es solo visual; ocultar no borra datos ni cambia permisos. Sin respuesta de la API se muestra todo.
 */
@Injectable({ providedIn: 'root' })
export class VistasService {
  private readonly http = inject(HttpClient);
  private readonly contexto = inject(ContextoService);
  private readonly estado = signal<EstadoVistas>({ negocioId: null, ocultas: new Set() });
  private enVuelo: { negocioId: string; peticion: Observable<void> } | null = null;

  /** Secciones ocultas del negocio activo; vacío mientras se cargan. */
  readonly ocultas = computed(() => {
    const estado = this.estado();
    return estado.negocioId === this.contexto.negocio()?.id ? estado.ocultas : new Set<SeccionOcultable>();
  });

  constructor() {
    effect(() => {
      const negocioId = this.contexto.negocio()?.id ?? null;
      untracked(() => this.cargar(negocioId).subscribe());
    });
  }

  visible(seccion: SeccionOcultable | null | undefined): boolean {
    return !seccion || !this.ocultas().has(seccion);
  }

  /** Espera a tener la preferencia del negocio activo, por ejemplo antes de abrir una ruta ocultable. */
  asegurar(): Observable<void> {
    return this.cargar(this.contexto.negocio()?.id ?? null);
  }

  guardar(ocultas: SeccionOcultable[]): Observable<VistasNegocio> {
    const negocioId = this.contexto.negocio()?.id ?? '';
    return this.http
      .put<VistasNegocio>(`${environment.apiUrl}/negocios/${negocioId}/configuracion/vistas`, { secciones_ocultas: ocultas })
      .pipe(tap((vistas) => this.estado.set({ negocioId, ocultas: new Set(vistas.secciones_ocultas ?? []) })));
  }

  private cargar(negocioId: string | null): Observable<void> {
    if (!negocioId) {
      this.estado.set({ negocioId: null, ocultas: new Set() });
      return of(undefined);
    }
    if (this.estado().negocioId === negocioId) return of(undefined);
    if (this.enVuelo?.negocioId === negocioId) return this.enVuelo.peticion;
    const peticion = this.http.get<VistasNegocio>(`${environment.apiUrl}/negocios/${negocioId}/configuracion/vistas`).pipe(
      map((vistas) => new Set<SeccionOcultable>(vistas.secciones_ocultas ?? [])),
      catchError(() => of(new Set<SeccionOcultable>())),
      // Una respuesta tardía de otro negocio no pisa la del negocio activo.
      tap((ocultas) => {
        if (this.contexto.negocio()?.id === negocioId) this.estado.set({ negocioId, ocultas });
      }),
      map(() => undefined),
      finalize(() => {
        if (this.enVuelo?.negocioId === negocioId) this.enVuelo = null;
      }),
      shareReplay({ bufferSize: 1, refCount: false }),
    );
    this.enVuelo = { negocioId, peticion };
    return peticion;
  }
}

import { Component, computed, inject } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { map } from 'rxjs';

import { ContextoService } from './contexto.service';

/** Destino cuando la persona abre una sección para la que su rol no tiene permiso. No muestra datos del negocio. */
@Component({
  selector: 'app-sin-acceso',
  imports: [RouterLink],
  template: `
    <section class="mx-auto flex max-w-[520px] flex-col items-center px-4 py-16 text-center" aria-labelledby="sin-acceso-titulo">
      <span class="grid h-16 w-16 place-items-center rounded-full bg-primary-soft text-primary-hover">
        <i class="pi pi-lock text-[26px]" aria-hidden="true"></i>
      </span>
      <h1 id="sin-acceso-titulo" class="mt-5 mb-0 text-[26px] leading-tight font-bold tracking-[-0.03em] text-ink">
        No tienes acceso a {{ seccion() ? '«' + seccion() + '»' : 'esta sección' }}
      </h1>
      <p class="mt-3 mb-0 text-[14.5px] leading-relaxed text-muted">
        Tu rol en <strong class="font-semibold text-ink">{{ negocio() }}</strong> no incluye este permiso.
        Si lo necesitas, pídelo a quien administra el negocio.
      </p>
      <a class="boton mt-7" routerLink="/inicio"><i class="pi pi-home" aria-hidden="true"></i> Ir a Inicio</a>
    </section>
  `,
})
export class SinAcceso {
  private readonly contexto = inject(ContextoService);
  readonly seccion = toSignal(inject(ActivatedRoute).queryParamMap.pipe(map((params) => params.get('seccion'))), {
    initialValue: null,
  });
  readonly negocio = computed(() => this.contexto.negocio()?.nombre_comercial ?? 'este negocio');
}

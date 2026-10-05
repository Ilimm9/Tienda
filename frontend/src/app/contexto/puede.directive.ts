import { Directive, effect, inject, input, TemplateRef, ViewContainerRef } from '@angular/core';

import { ContextoService } from './contexto.service';
import { PermisoRequerido } from './permisos';

/**
 * Muestra el elemento solo si el negocio activo otorga el permiso: `*appPuede="PERMISOS.rolGestionar"`.
 * Con una lista se exigen todos. Reacciona cuando cambian el negocio o sus permisos.
 */
@Directive({ selector: '[appPuede]' })
export class PuedeDirective {
  private readonly contexto = inject(ContextoService);
  private readonly plantilla = inject(TemplateRef<unknown>);
  private readonly contenedor = inject(ViewContainerRef);

  readonly appPuede = input.required<PermisoRequerido | null | undefined>();

  constructor() {
    effect(() => {
      const permitido = this.contexto.puede(this.appPuede());
      if (permitido && !this.contenedor.length) this.contenedor.createEmbeddedView(this.plantilla);
      else if (!permitido) this.contenedor.clear();
    });
  }
}

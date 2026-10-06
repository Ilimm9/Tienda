import { inject, Injectable } from '@angular/core';
import { ActivatedRouteSnapshot, Router } from '@angular/router';

import { FeedbackService } from '../shared/feedback/feedback.service';
import { ContextoService } from './contexto.service';
import { negocioDeRuta, nombreDeSeccion, permisosDeRuta } from './permiso.guard';

/**
 * Reacciona a un 403 `ACCESO_DENEGADO` de la API: los permisos en memoria quedaron viejos.
 * Recarga el contexto y, si la pantalla actual ya no está permitida, muestra «Sin acceso».
 */
@Injectable({ providedIn: 'root' })
export class AccesoDenegadoService {
  private readonly contexto = inject(ContextoService);
  private readonly router = inject(Router);
  private readonly feedback = inject(FeedbackService);
  private enCurso = false;

  atender(): void {
    // Una pantalla puede lanzar varias peticiones a la vez; basta una recarga.
    if (this.enCurso) return;
    this.enCurso = true;
    this.contexto.recargar().subscribe({
      next: () => {
        this.enCurso = false;
        const ruta = this.rutaActual();
        const negocioId = negocioDeRuta(ruta);
        const permitido = permisosDeRuta(ruta).every((permiso) =>
          negocioId ? this.contexto.puedeEn(negocioId, permiso, true) : this.contexto.puede(permiso),
        );
        if (permitido) return;
        this.feedback.warning('Tu acceso cambió', 'Ya no tienes permiso para esta sección.');
        void this.router.navigate(['/sin-acceso'], {
          queryParams: { seccion: nombreDeSeccion(ruta) },
          skipLocationChange: true,
        });
      },
      error: () => (this.enCurso = false),
    });
  }

  private rutaActual(): ActivatedRouteSnapshot {
    let ruta = this.router.routerState.snapshot.root;
    while (ruta.firstChild) ruta = ruta.firstChild;
    return ruta;
  }
}

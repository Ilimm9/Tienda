import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';

import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { PuedeDirective } from '../../contexto/puede.directive';

@Component({
  selector: 'app-home',
  imports: [RouterLink, PuedeDirective],
  templateUrl: './home.component.html',
  styleUrl: './home.component.css',
})
export class HomeComponent {
  readonly contexto = inject(ContextoService);
  readonly P = PERMISOS;
  readonly alta = this.contexto.estadoAlta;
  /** Fecha de hoy en Ciudad de México, como en el resto de la app. */
  readonly hoy = new Intl.DateTimeFormat('es-MX', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    timeZone: 'America/Mexico_City',
  }).format(new Date());

  irASecciones(evento: Event): void {
    evento.preventDefault();
    document.getElementById('secciones')?.scrollIntoView({ behavior: 'smooth' });
  }
}

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
}

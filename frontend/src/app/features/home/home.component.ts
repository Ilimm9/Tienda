import { Component, inject } from '@angular/core';

import { ContextoService } from '../../contexto/contexto.service';

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrl: './home.component.css',
})
export class HomeComponent {
  readonly contexto = inject(ContextoService);
}

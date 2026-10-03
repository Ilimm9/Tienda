import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';

import { ContextoService } from '../../contexto/contexto.service';

@Component({
  selector: 'app-home',
  imports: [RouterLink],
  templateUrl: './home.component.html',
  styleUrl: './home.component.css',
})
export class HomeComponent {
  readonly contexto = inject(ContextoService);
}

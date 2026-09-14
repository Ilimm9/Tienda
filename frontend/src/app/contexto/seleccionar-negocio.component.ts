import { Component, inject } from '@angular/core';
import { Router } from '@angular/router';

import { ContextoService } from './contexto.service';

@Component({
  selector: 'app-seleccionar-negocio',
  template: `<section class="selection"><h1>Elige un negocio</h1><p>Selecciona contexto para continuar.</p>@for (item of contexto.negocios(); track item.id) {<button type="button" (click)="select(item.id)"><strong>{{ item.nombre_comercial }}</strong><span>{{ item.sucursales.length }} sucursal(es)</span></button>} @empty {<a href="/negocios">Registrar negocio</a>}</section>`,
  styles: `.selection{max-width:640px;margin:32px auto}.selection button{display:flex;width:100%;justify-content:space-between;margin:12px 0;padding:18px;border:1px solid var(--border);border-radius:12px;background:var(--surface-raised);color:var(--text);cursor:pointer}.selection span{color:var(--text-muted)}`,
})
export class SeleccionarNegocioComponent {
  readonly contexto = inject(ContextoService);
  private readonly router = inject(Router);
  constructor() {
    this.contexto.asegurarInicializado().subscribe();
  }
  select(id: string): void {
    this.contexto.seleccionarNegocio(id);
    void this.router.navigate(['/inicio']);
  }
}

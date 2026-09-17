import { Component, inject } from '@angular/core';
import { Router } from '@angular/router';

import { ContextoService } from './contexto.service';

@Component({ selector: 'app-ir-a-sucursales', template: '' })
export class IrASucursalesComponent {
  private readonly contexto = inject(ContextoService);
  private readonly router = inject(Router);
  constructor() {
    const negocioID = this.contexto.negocio()?.id;
    void this.router.navigate(negocioID ? ['/negocios', negocioID, 'sucursales'] : ['/seleccionar-negocio']);
  }
}

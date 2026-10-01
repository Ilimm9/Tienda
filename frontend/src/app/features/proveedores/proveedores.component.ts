import { Component } from '@angular/core';

import { CatalogoComponent } from '../catalogo/catalogo.component';

@Component({
  selector: 'app-proveedores',
  standalone: true,
  imports: [CatalogoComponent],
  template: '<app-catalogo />',
})
export class ProveedoresComponent {}

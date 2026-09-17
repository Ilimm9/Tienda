import { Routes } from '@angular/router';

export const SUCURSALES_ROUTES: Routes = [
  {
    path: '',
    pathMatch: 'full',
    loadComponent: () =>
      import('../../contexto/ir-a-sucursales.component').then(
        (module) => module.IrASucursalesComponent,
      ),
  },
];

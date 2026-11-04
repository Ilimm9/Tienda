import { Routes } from '@angular/router';

import { contextoGuard } from '../../contexto/contexto.guard';
import { soloPropietarioGuard } from '../productos/solo-propietario.guard';

export const CONFIGURACION_ROUTES: Routes = [
  { path: '', redirectTo: 'vistas', pathMatch: 'full' },
  {
    path: 'vistas',
    canActivate: [contextoGuard, soloPropietarioGuard],
    data: { breadcrumb: 'Vistas' },
    loadComponent: () => import('./vistas.component').then((module) => module.VistasComponent),
  },
];

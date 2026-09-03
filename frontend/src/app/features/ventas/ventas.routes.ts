import { Routes } from '@angular/router';

export const VENTAS_ROUTES: Routes = [
  {
    path: '',
    data: {
      breadcrumb: 'Ventas',
      title: 'Ventas',
      description: 'Consulta y administra la actividad comercial de tus sucursales.',
      icon: 'pi pi-shopping-cart',
    },
    loadComponent: () =>
      import('../../shared/ui/section-placeholder/section-placeholder.component').then(
        (module) => module.SectionPlaceholderComponent,
      ),
  },
];

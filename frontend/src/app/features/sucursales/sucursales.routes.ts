import { Routes } from '@angular/router';

export const SUCURSALES_ROUTES: Routes = [
  {
    path: '',
    data: {
      breadcrumb: 'Sucursales',
      title: 'Sucursales',
      description: 'Organiza las ubicaciones y puntos de operación de tus negocios.',
      icon: 'pi pi-map-marker',
    },
    loadComponent: () =>
      import('../../shared/ui/section-placeholder/section-placeholder.component').then(
        (module) => module.SectionPlaceholderComponent,
      ),
  },
];

import { Routes } from '@angular/router';

export const NEGOCIOS_ROUTES: Routes = [
  {
    path: '',
    data: {
      breadcrumb: 'Negocios',
      title: 'Negocios',
      description: 'Administra los negocios asociados a tu cuenta.',
      icon: 'pi pi-briefcase',
    },
    loadComponent: () =>
      import('../../shared/ui/section-placeholder/section-placeholder.component').then(
        (module) => module.SectionPlaceholderComponent,
      ),
  },
];

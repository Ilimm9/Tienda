import { Routes } from '@angular/router';

export const ROLES_PERMISOS_ROUTES: Routes = [
  {
    path: '',
    data: {
      breadcrumb: 'Roles y permisos',
      title: 'Roles y permisos',
      description: 'Define responsabilidades y niveles de acceso dentro de la plataforma.',
      icon: 'pi pi-lock',
    },
    loadComponent: () =>
      import('../../shared/ui/section-placeholder/section-placeholder.component').then(
        (module) => module.SectionPlaceholderComponent,
      ),
  },
];

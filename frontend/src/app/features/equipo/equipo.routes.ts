import { Routes } from '@angular/router';

const placeholder = () =>
  import('../../shared/ui/section-placeholder/section-placeholder.component').then(
    (module) => module.SectionPlaceholderComponent,
  );

export const EQUIPO_ROUTES: Routes = [
  {
    path: '',
    data: { breadcrumb: 'Equipo' },
    children: [
      { path: '', redirectTo: 'empleados', pathMatch: 'full' },
      {
        path: 'empleados',
        data: {
          breadcrumb: 'Empleados',
          title: 'Empleados',
          description: 'Administra las personas que colaboran en tus negocios y sucursales.',
          icon: 'pi pi-users',
        },
        loadComponent: placeholder,
      },
      {
        path: 'invitaciones',
        data: {
          breadcrumb: 'Invitaciones',
          title: 'Invitaciones',
          description: 'Gestiona el acceso y registro de nuevos integrantes del equipo.',
          icon: 'pi pi-send',
        },
        loadComponent: placeholder,
      },
    ],
  },
];

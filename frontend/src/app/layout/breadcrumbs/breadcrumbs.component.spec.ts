import { Component } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter, Router, Routes } from '@angular/router';

import { BreadcrumbsComponent } from './breadcrumbs.component';

@Component({ template: '' })
class EmptyPageComponent {}

const ROUTES: Routes = [
  {
    path: 'equipo',
    data: { breadcrumb: 'Equipo' },
    children: [
      {
        path: 'invitaciones',
        data: { breadcrumb: 'Invitaciones' },
        children: [
          { path: '', component: EmptyPageComponent },
          { path: 'nueva', data: { breadcrumb: 'Nueva invitación' }, component: EmptyPageComponent },
        ],
      },
    ],
  },
  {
    path: 'roles-permisos',
    data: { breadcrumb: 'Roles y permisos' },
    children: [
      { path: '', data: { breadcrumb: 'Roles y permisos' }, component: EmptyPageComponent },
      { path: 'nuevo', data: { breadcrumb: 'Nuevo rol' }, component: EmptyPageComponent },
    ],
  },
  {
    path: 'catalogo',
    children: [
      { path: 'marcas/nuevo', data: { breadcrumb: 'Agregar marca' }, component: EmptyPageComponent },
    ],
  },
  { path: 'proveedores/nuevo', data: { breadcrumb: 'Agregar proveedor' }, component: EmptyPageComponent },
  {
    path: 'negocios',
    data: { breadcrumb: 'Negocios' },
    children: [
      {
        path: ':negocioId/sucursales',
        data: { breadcrumb: 'Sucursales' },
        children: [{ path: 'nueva', data: { breadcrumb: 'Nueva' }, component: EmptyPageComponent }],
      },
    ],
  },
];

async function render(url: string) {
  await TestBed.configureTestingModule({
    imports: [BreadcrumbsComponent],
    providers: [provideRouter(ROUTES)],
  }).compileComponents();
  await TestBed.inject(Router).navigateByUrl(url);
  const fixture = TestBed.createComponent(BreadcrumbsComponent);
  fixture.detectChanges();
  const element: HTMLElement = fixture.nativeElement;
  const crumbs = Array.from(element.querySelectorAll('.crumb')).map((item) =>
    (item.textContent ?? '').replace('/', '').trim(),
  );
  return { element, crumbs, component: fixture.componentInstance };
}

describe('BreadcrumbsComponent', () => {
  it('shows the menu group as text and links the section', async () => {
    const { element, crumbs } = await render('/equipo/invitaciones/nueva');

    expect(crumbs).toEqual(['Equipo', 'Invitaciones', 'Nueva invitación']);
    expect(element.querySelector('.group')?.textContent).toContain('Equipo');
    expect(element.querySelector('a[href="/equipo/invitaciones"]')).not.toBeNull();
    expect(element.querySelector('[aria-current="page"]')?.textContent).toContain('Nueva invitación');
  });

  it('marks the section itself as the current page', async () => {
    const { element, crumbs } = await render('/equipo/invitaciones');

    expect(crumbs).toEqual(['Equipo', 'Invitaciones']);
    expect(element.querySelector('[aria-current="page"]')?.textContent).toContain('Invitaciones');
  });

  it('collapses a parent and its empty child that share the same URL', async () => {
    const { crumbs } = await render('/roles-permisos');

    expect(crumbs).toEqual(['Roles y permisos']);
  });

  it('keeps a single level for the section list and adds child pages after it', async () => {
    const { crumbs } = await render('/roles-permisos/nuevo');

    expect(crumbs).toEqual(['Roles y permisos', 'Nuevo rol']);
  });

  it('adds the menu group and section for flat routes', async () => {
    expect((await render('/catalogo/marcas/nuevo')).crumbs).toEqual(['Catálogo', 'Marcas', 'Agregar marca']);
  });

  it('adds the section for a flat route without a group', async () => {
    expect((await render('/proveedores/nuevo')).crumbs).toEqual(['Proveedores', 'Agregar proveedor']);
  });

  it('keeps nested route levels under their menu section', async () => {
    const { element, crumbs } = await render('/negocios/abc/sucursales/nueva');

    expect(crumbs).toEqual(['Negocios', 'Sucursales', 'Nueva']);
    expect(element.querySelector('a[href="/negocios/abc/sucursales"]')).not.toBeNull();
  });
});

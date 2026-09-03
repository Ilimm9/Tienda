import { Component } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';

import { BreadcrumbsComponent } from './breadcrumbs.component';

@Component({ template: '' })
class EmptyPageComponent {}

describe('BreadcrumbsComponent', () => {
  it('builds linked levels from route metadata', async () => {
    await TestBed.configureTestingModule({
      imports: [BreadcrumbsComponent],
      providers: [
        provideRouter([
          {
            path: 'equipo',
            data: { breadcrumb: 'Equipo' },
            children: [
              {
                path: 'invitaciones',
                data: { breadcrumb: 'Invitaciones' },
                component: EmptyPageComponent,
              },
            ],
          },
        ]),
      ],
    }).compileComponents();
    await TestBed.inject(Router).navigateByUrl('/equipo/invitaciones');
    const fixture = TestBed.createComponent(BreadcrumbsComponent);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Equipo');
    expect(fixture.nativeElement.textContent).toContain('Invitaciones');
    expect(fixture.nativeElement.querySelector('[aria-current="page"]').textContent).toContain(
      'Invitaciones',
    );
  });
});

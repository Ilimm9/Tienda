import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { HomeComponent } from './home.component';

describe('HomeComponent', () => {
  it('renders the dashboard summary inside the administrative shell', async () => {
    await TestBed.configureTestingModule({
      imports: [HomeComponent],
      providers: [provideRouter([])],
    }).compileComponents();
    const fixture = TestBed.createComponent(HomeComponent);
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('h1').textContent).toContain('Bienvenido');
    const tarjetas = fixture.nativeElement.querySelectorAll('a.resumen-tarjeta');
    expect(tarjetas).toHaveLength(3);
    expect(tarjetas[2].getAttribute('href')).toBe('/equipo/empleados');
  });
});

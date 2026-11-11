import { Component, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { ConsejoMascota, MascotaComponent } from './mascota.component';

const CONSEJOS: ConsejoMascota[] = [
  { titulo: 'Uno', texto: 'Primero', accion: 'Ir', ruta: '/productos' },
  { titulo: 'Dos', texto: 'Segundo' },
  { titulo: 'Tres', texto: 'Tercero' },
];

@Component({
  imports: [MascotaComponent],
  template: `<app-mascota [consejos]="consejos" [nombre]="nombre()" [flotante]="false" [(abierto)]="abierto" />`,
})
class AnfitrionComponent {
  consejos = CONSEJOS;
  nombre = signal('Stocki');
  abierto = false;
}

async function crear() {
  await TestBed.configureTestingModule({ imports: [AnfitrionComponent], providers: [provideRouter([])] }).compileComponents();
  const fixture = TestBed.createComponent(AnfitrionComponent);
  fixture.detectChanges();
  const el: HTMLElement = fixture.nativeElement;
  const boton = () => el.querySelector<HTMLButtonElement>('.mascota__boton')!;
  const titulo = () => el.querySelector('.mascota__globo h2')?.textContent?.trim();
  return { fixture, el, boton, titulo };
}

describe('MascotaComponent', () => {
  afterEach(() => TestBed.resetTestingModule());

  it('abre y cierra el globo, y avisa al padre', async () => {
    const { fixture, el, boton, titulo } = await crear();

    expect(el.querySelector('.mascota__globo')).toBeNull();
    expect(el.querySelector('.mascota__invitacion')?.textContent).toContain('¿Te doy un tip?');
    expect(boton().getAttribute('aria-expanded')).toBe('false');

    boton().click();
    fixture.detectChanges();
    expect(titulo()).toBe('Uno');
    expect(el.textContent).toContain('Hola, soy Stocki');
    expect(el.querySelector('.mascota__invitacion')).toBeNull();
    expect(fixture.componentInstance.abierto).toBe(true);

    el.querySelector<HTMLButtonElement>('.mascota__cerrar')!.click();
    fixture.detectChanges();
    expect(el.querySelector('.mascota__globo')).toBeNull();
    expect(fixture.componentInstance.abierto).toBe(false);
  });

  it('recorre los consejos en ciclo y muestra la acción solo si existe', async () => {
    const { fixture, el, boton, titulo } = await crear();
    boton().click();
    fixture.detectChanges();
    const [anterior, siguiente] = el.querySelectorAll<HTMLButtonElement>('.mascota__flecha');

    expect(el.querySelector('.mascota__accion')?.getAttribute('href')).toBe('/productos');
    anterior.click();
    fixture.detectChanges();
    expect(titulo()).toBe('Tres');
    expect(el.textContent).toContain('Tip 3 de 3');
    expect(el.querySelector('.mascota__accion')).toBeNull();
    siguiente.click();
    siguiente.click();
    fixture.detectChanges();
    expect(titulo()).toBe('Dos');
  });

  it('cierra con Escape y recibe otro nombre', async () => {
    const { fixture, el, boton } = await crear();
    fixture.componentInstance.nombre.set('Cajito');
    fixture.detectChanges();
    expect(boton().getAttribute('aria-label')).toBe('Abrir tips de Cajito');

    boton().click();
    fixture.detectChanges();
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));
    fixture.detectChanges();
    expect(el.querySelector('.mascota__globo')).toBeNull();
    expect(document.activeElement).toBe(boton());
  });
});

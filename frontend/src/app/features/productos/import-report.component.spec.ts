import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ImportReportComponent } from './import-report.component';
import { ProductImportIssue } from './product.models';

describe('ImportReportComponent', () => {
  let fixture: ComponentFixture<ImportReportComponent>;
  let page: HTMLElement;

  function errores(stock: number, costo: number): ProductImportIssue[] {
    return [
      ...Array.from({ length: stock }, (_, index) => ({ fila: index + 2, nombre: `Producto ${index + 2}`, campo: 'Stock inicial', motivo: 'negativo' })),
      ...Array.from({ length: costo }, (_, index) => ({ fila: 100 + index, nombre: `Costo ${index}`, campo: 'Costo', motivo: 'inválido' })),
    ];
  }

  function render(inputs: Record<string, unknown>): void {
    fixture = TestBed.createComponent(ImportReportComponent);
    for (const [name, value] of Object.entries(inputs)) fixture.componentRef.setInput(name, value);
    fixture.detectChanges();
    page = fixture.nativeElement;
  }

  const filas = () => page.querySelectorAll('#reporte-panel tbody tr');
  const pestanas = () => [...page.querySelectorAll<HTMLButtonElement>('[role="tab"]')];
  const pestana = (texto: string) => pestanas().find((item) => item.textContent?.includes(texto))!;
  const boton = (texto: string) =>
    [...page.querySelectorAll<HTMLButtonElement>('.reporte-filtro')].find((item) => item.textContent?.trim().startsWith(texto))!;

  it('muestra los contadores como pestañas y abre «Filas con error» primero', () => {
    render({
      errores: errores(1, 0),
      advertencias: [{ fila: 4, campo: 'Código de barras', motivo: 'corto' }],
      correctos: [{ fila: 2, nombre: 'Agua', categoria: 'Bebidas', precio_venta: 12, costo: 8.38, stock_inicial: 4 }],
    });

    expect(pestanas().map((item) => item.textContent?.replace(/\s+/g, ' ').trim())).toEqual(['Listos1', 'Filas con error1', 'Advertencias1']);
    expect(pestana('Filas con error').getAttribute('aria-selected')).toBe('true');
    expect(page.querySelectorAll('[role="tabpanel"]')).toHaveLength(1);
    expect(page.querySelector('#reporte-panel')?.textContent).not.toContain('Agua');

    pestana('Listos').click();
    fixture.detectChanges();

    expect(pestana('Listos').getAttribute('aria-selected')).toBe('true');
    expect(pestana('Filas con error').getAttribute('aria-selected')).toBe('false');
    expect(page.querySelector('#reporte-panel')?.textContent).toContain('Agua');
  });

  it('abre «Listos» cuando no hay errores', () => {
    render({ correctos: [{ fila: 2, nombre: 'Agua', categoria: 'Bebidas', precio_venta: 12, costo: 8.38, stock_inicial: 4 }] });
    expect(pestana('Listos').getAttribute('aria-selected')).toBe('true');
    expect(filas()[0].textContent).toContain('Agua');
  });

  it('cambia de pestaña con las flechas', () => {
    render({ errores: errores(1, 0), catalogos: [{ tipo: 'Marca', nombre: 'Acme', productos: 3 }] });
    pestana('Filas con error').dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight' }));
    fixture.detectChanges();
    expect(pestana('Catálogos nuevos').getAttribute('aria-selected')).toBe('true');
    expect(page.querySelector('#reporte-panel')?.textContent).toContain('Acme');
  });

  it('pagina de 10 en 10 solo cuando hay más de 10 filas', () => {
    render({ errores: errores(10, 0) });
    expect(filas().length).toBe(10);
    expect(page.querySelector('.reporte-paginador')).toBeNull();

    render({ errores: errores(25, 0) });
    expect(page.querySelector('.reporte-paginador')?.textContent).toContain('1–10 de 25');

    page.querySelector<HTMLButtonElement>('[aria-label="Página siguiente"]')!.click();
    fixture.detectChanges();

    expect(page.querySelector('.reporte-paginador')?.textContent).toContain('11–20 de 25');
    expect(filas()[0].textContent).toContain('Producto 12');
  });

  it('filtra por columna y regresa a la primera página', () => {
    render({ errores: errores(25, 2) });
    page.querySelector<HTMLButtonElement>('[aria-label="Página siguiente"]')!.click();
    fixture.detectChanges();

    boton('Costo').click();
    fixture.detectChanges();

    expect(boton('Costo').getAttribute('aria-pressed')).toBe('true');
    expect(filas().length).toBe(2);
    expect(page.querySelector('.reporte-paginador')).toBeNull();

    boton('Todos').click();
    fixture.detectChanges();
    expect(page.querySelector('.reporte-paginador')?.textContent).toContain('1–10 de 27');
  });

  it('muestra los costos faltantes como advertencia y emite la descarga', () => {
    render({ errores: errores(1, 0), costosFaltantes: 3, puedeDescargar: true });
    const descargar = vi.fn();
    fixture.componentInstance.descargarErrores.subscribe(descargar);

    page.querySelector<HTMLButtonElement>('.reporte-descargar')!.click();
    pestana('Advertencias').click();
    fixture.detectChanges();

    expect(descargar).toHaveBeenCalled();
    expect(page.querySelector('#reporte-panel')?.textContent).toContain('3 productos no cuentan con costo');
  });

  it('muestra las pestañas de productos que se agregan o actualizan en la sucursal (28)', () => {
    const fila = { fila: 2, nombre: 'Jabón', categoria: 'Limpieza', precio_venta: 12, costo: 6, stock_inicial: 7 };
    render({ agregarSucursal: [fila], actualizar: [{ ...fila, fila: 3, nombre: 'Cloro' }], sucursal: 'SUR' });

    const tarjetas = [...page.querySelectorAll('.reporte-pestana--sucursal')].map((item) => item.textContent?.replace(/\s+/g, ' ').trim());
    expect(tarjetas).toEqual(['Se agregan a la sucursal1', 'Se actualizan1']);
    expect(page.querySelector('#reporte-panel small')?.textContent).toContain('a SUR');

    pestana('Se actualizan').click();
    fixture.detectChanges();
    expect(page.querySelector('#reporte-panel tbody')?.textContent).toContain('Cloro');
    expect(page.querySelector('#reporte-panel tbody')?.textContent).not.toContain('Jabón');
  });
});

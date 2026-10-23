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

  const filas = () => page.querySelectorAll('#reporte-errores-cuerpo tbody tr');
  const boton = (texto: string) =>
    [...page.querySelectorAll<HTMLButtonElement>('#reporte-errores button')].find((item) => item.textContent?.trim().startsWith(texto))!;

  it('abre errores y compacta las demás secciones', () => {
    render({
      errores: errores(1, 0),
      advertencias: [{ fila: 4, campo: 'Código de barras', motivo: 'corto' }],
      correctos: [{ fila: 2, nombre: 'Agua', categoria: 'Bebidas', precio_venta: 12, costo: 8.38, stock_inicial: 4 }],
    });

    const toggles = page.querySelectorAll<HTMLButtonElement>('.reporte-toggle');
    expect([...toggles].map((item) => item.getAttribute('aria-expanded'))).toEqual(['true', 'false', 'false']);
    expect(page.querySelector('#reporte-correctos-cuerpo')).toBeNull();

    toggles[2].click();
    fixture.detectChanges();

    expect(toggles[2].getAttribute('aria-expanded')).toBe('true');
    expect(page.querySelector('#reporte-correctos-cuerpo')?.textContent).toContain('Agua');
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
    page.querySelectorAll<HTMLButtonElement>('.reporte-toggle')[1].click();
    fixture.detectChanges();

    expect(descargar).toHaveBeenCalled();
    expect(page.querySelector('#reporte-advertencias-cuerpo')?.textContent).toContain('3 productos no cuentan con costo');
  });
  it('muestra las secciones de productos que se agregan o actualizan en la sucursal (28)', () => {
    const fila = { fila: 2, nombre: 'Jabón', categoria: 'Limpieza', precio_venta: 12, costo: 6, stock_inicial: 7 };
    render({ agregarSucursal: [fila], actualizar: [{ ...fila, fila: 3, nombre: 'Cloro' }], sucursal: 'SUR' });

    const tarjetas = [...page.querySelectorAll('.reporte-tarjeta--sucursal')].map((item) => item.textContent?.replace(/\s+/g, ' ').trim());
    expect(tarjetas).toEqual(['Se agregarán a esta sucursal1', 'Se actualizarán1']);
    expect(page.querySelector('#reporte-agregar small')?.textContent).toContain('a SUR');

    page.querySelector<HTMLButtonElement>('#reporte-actualizar .reporte-toggle')!.click();
    fixture.detectChanges();
    expect(page.querySelector('#reporte-actualizar-cuerpo tbody')?.textContent).toContain('Cloro');
    expect(page.querySelector('#reporte-agregar-cuerpo')).toBeNull();
  });
});

import { CurrencyPipe, NgTemplateOutlet } from '@angular/common';
import { Component, computed, input, output, signal } from '@angular/core';

import {
  ProductImportCatalogProposal,
  ProductImportIssue,
  ProductImportValidRow,
} from './product.models';

type Seccion = 'errores' | 'advertencias' | 'catalogos' | 'correctos';

interface Paginado<T> {
  filas: T[];
  total: number;
  pagina: number;
  ultima: number;
  desde: number;
  hasta: number;
  // 0 representa un hueco («…») entre números de página.
  paginas: number[];
}

const TAMANO_PAGINA = 10;

/** Resultado de una importación en secciones compactables y paginadas. */
@Component({
  selector: 'app-import-report',
  imports: [CurrencyPipe, NgTemplateOutlet],
  templateUrl: './import-report.component.html',
  styleUrl: './import-report.component.css',
})
export class ImportReportComponent {
  readonly errores = input<ProductImportIssue[]>([]);
  readonly advertencias = input<ProductImportIssue[]>([]);
  readonly catalogos = input<ProductImportCatalogProposal[]>([]);
  readonly correctos = input<ProductImportValidRow[]>([]);
  readonly costosFaltantes = input(0);
  readonly sucursal = input('');
  readonly puedeDescargar = input(false);
  readonly descargando = input(false);
  readonly descargarErrores = output<void>();

  readonly abiertas = signal<Record<Seccion, boolean>>({
    errores: true,
    advertencias: false,
    catalogos: false,
    correctos: false,
  });
  private readonly paginas = signal<Record<Seccion, number>>({
    errores: 1,
    advertencias: 1,
    catalogos: 1,
    correctos: 1,
  });
  private readonly campoElegido = signal('Todos');

  readonly filasConError = computed(() => new Set(this.errores().map((issue) => issue.fila)).size);
  readonly hayAdvertencias = computed(
    () => this.advertencias().length > 0 || this.costosFaltantes() > 0,
  );
  readonly campos = computed(() => {
    const conteo = new Map<string, number>();
    for (const issue of this.errores()) {
      const campo = issue.campo || 'General';
      conteo.set(campo, (conteo.get(campo) ?? 0) + 1);
    }
    return [
      { campo: 'Todos', total: this.errores().length },
      ...[...conteo].map(([campo, total]) => ({ campo, total })),
    ];
  });
  readonly campo = computed(() =>
    this.campos().some((item) => item.campo === this.campoElegido()) ? this.campoElegido() : 'Todos',
  );
  readonly erroresPagina = computed(() => {
    const campo = this.campo();
    const filas =
      campo === 'Todos'
        ? this.errores()
        : this.errores().filter((issue) => (issue.campo || 'General') === campo);
    return this.paginar('errores', filas);
  });
  readonly advertenciasPagina = computed(() => this.paginar('advertencias', this.advertencias()));
  readonly catalogosPagina = computed(() => this.paginar('catalogos', this.catalogos()));
  readonly correctosPagina = computed(() => this.paginar('correctos', this.correctos()));

  alternar(seccion: Seccion): void {
    this.abiertas.update((actual) => ({ ...actual, [seccion]: !actual[seccion] }));
  }

  abrir(seccion: Seccion): void {
    this.abiertas.update((actual) => ({ ...actual, [seccion]: true }));
    document.getElementById(`reporte-${seccion}`)?.scrollIntoView({ block: 'start' });
  }

  irA(seccion: Seccion, pagina: number): void {
    this.paginas.update((actual) => ({ ...actual, [seccion]: pagina }));
  }

  elegirCampo(campo: string): void {
    this.campoElegido.set(campo);
    this.irA('errores', 1);
  }

  private paginar<T>(seccion: Seccion, filas: T[]): Paginado<T> {
    const total = filas.length;
    const ultima = Math.max(1, Math.ceil(total / TAMANO_PAGINA));
    const pagina = Math.min(this.paginas()[seccion], ultima);
    const desde = (pagina - 1) * TAMANO_PAGINA;
    return {
      filas: filas.slice(desde, desde + TAMANO_PAGINA),
      total,
      pagina,
      ultima,
      desde: total ? desde + 1 : 0,
      hasta: Math.min(desde + TAMANO_PAGINA, total),
      paginas: numerosDePagina(pagina, ultima),
    };
  }
}

function numerosDePagina(pagina: number, ultima: number): number[] {
  if (ultima <= 7) return Array.from({ length: ultima }, (_, index) => index + 1);
  const numeros = [1];
  if (pagina > 3) numeros.push(0);
  for (let numero = Math.max(2, pagina - 1); numero <= Math.min(ultima - 1, pagina + 1); numero++) {
    numeros.push(numero);
  }
  if (pagina < ultima - 2) numeros.push(0);
  numeros.push(ultima);
  return numeros;
}

import { Component, computed, effect, inject, signal, untracked } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import Swal from 'sweetalert2/dist/sweetalert2.esm.js';

import { ContextoService } from '../../contexto/contexto.service';
import { BranchCatalogProduct } from './product.models';
import { ProductosService } from './productos.service';

/**
 * Surtido por sucursal (capability 28): el propietario elige qué productos del
 * catálogo del negocio entran a la sucursal activa. Entran sin precio ni stock.
 */
@Component({
  selector: 'app-agregar-catalogo',
  imports: [FormsModule],
  templateUrl: './agregar-catalogo.component.html',
  styleUrl: './agregar-catalogo.component.css',
})
export class AgregarCatalogoComponent {
  private readonly productosService = inject(ProductosService);
  private readonly contexto = inject(ContextoService);
  private readonly router = inject(Router);

  readonly sucursal = this.contexto.sucursal;
  readonly productos = signal<BranchCatalogProduct[]>([]);
  readonly cargando = signal(true);
  readonly guardando = signal(false);
  readonly error = signal<string | null>(null);
  readonly busqueda = signal('');
  readonly seleccion = signal<ReadonlySet<string>>(new Set());

  readonly visibles = computed(() => {
    const texto = normalizar(this.busqueda());
    if (!texto) return this.productos();
    return this.productos().filter((producto) =>
      [producto.nombre, producto.presentacion, producto.sku, producto.codigo_barras, producto.categoria]
        .some((valor) => normalizar(valor ?? '').includes(texto)),
    );
  });
  readonly todosVisiblesElegidos = computed(() => {
    const visibles = this.visibles();
    const elegidos = this.seleccion();
    return visibles.length > 0 && visibles.every((producto) => elegidos.has(producto.id));
  });

  constructor() {
    effect(() => {
      const negocio = this.contexto.negocio()?.id;
      const sucursal = this.contexto.sucursal()?.id;
      untracked(() => this.cargar(negocio, sucursal));
    });
  }

  private cargar(negocio: string | undefined, sucursal: string | undefined): void {
    this.seleccion.set(new Set());
    this.productos.set([]);
    this.error.set(null);
    if (!negocio || !sucursal) {
      this.cargando.set(false);
      return;
    }
    this.cargando.set(true);
    this.productosService.listNotInBranch(negocio, sucursal).subscribe({
      next: (items) => {
        this.productos.set(items);
        this.cargando.set(false);
      },
      error: (response) => {
        this.error.set(response.error?.mensaje ?? 'No fue posible cargar el catálogo del negocio.');
        this.cargando.set(false);
      },
    });
  }

  alternar(id: string): void {
    this.seleccion.update((actual) => {
      const siguiente = new Set(actual);
      if (siguiente.has(id)) siguiente.delete(id);
      else siguiente.add(id);
      return siguiente;
    });
  }

  alternarVisibles(): void {
    const visibles = this.visibles().map((producto) => producto.id);
    const quitar = this.todosVisiblesElegidos();
    this.seleccion.update((actual) => {
      const siguiente = new Set(actual);
      for (const id of visibles) {
        if (quitar) siguiente.delete(id);
        else siguiente.add(id);
      }
      return siguiente;
    });
  }

  agregarElegidos(): void {
    this.agregar([...this.seleccion()]);
  }

  agregarTodos(): void {
    void Swal.fire({
      icon: 'question',
      title: 'Agregar todo el catálogo',
      text: `Se agregarán ${this.productos().length} productos a ${this.sucursal()?.nombre ?? 'esta sucursal'}, sin precio ni stock.`,
      showCancelButton: true,
      confirmButtonText: 'Agregar todos',
      cancelButtonText: 'Cancelar',
    }).then((respuesta) => {
      if (respuesta.isConfirmed) this.agregar('todos');
    });
  }

  private agregar(ids: string[] | 'todos'): void {
    const negocio = this.contexto.negocio()?.id;
    const sucursal = this.contexto.sucursal()?.id;
    if (!negocio || !sucursal || this.guardando() || (ids !== 'todos' && !ids.length)) return;
    this.guardando.set(true);
    this.error.set(null);
    this.productosService.addToBranch(negocio, sucursal, ids).subscribe({
      next: ({ agregados }) => {
        this.guardando.set(false);
        void Swal.fire({
          icon: 'success',
          title: 'Productos agregados',
          text: `Se agregaron ${agregados} productos sin precio. Complétalos con Carga masiva o en Validar costos y precios.`,
        });
        void this.router.navigate(['/catalogo/productos']);
      },
      error: (response) => {
        this.guardando.set(false);
        this.error.set(response.error?.mensaje ?? 'No fue posible agregar los productos.');
      },
    });
  }

  volver(): void {
    void this.router.navigate(['/catalogo/productos']);
  }
}

function normalizar(valor: string): string {
  return valor.normalize('NFD').replace(/\p{Mn}/gu, '').trim().toLowerCase();
}

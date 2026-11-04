import { Component, computed, inject, OnInit, signal } from '@angular/core';

import { ContextoService } from '../../contexto/contexto.service';
import { SeccionOcultable, VistasService } from '../../contexto/vistas.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';

interface SeccionVista {
  readonly clave: SeccionOcultable | 'productos' | 'unidades';
  readonly nombre: string;
  readonly descripcion: string;
  /** Las secciones fijas las usan compras e inventario: no se pueden ocultar. */
  readonly fija: boolean;
}

const SECCIONES: readonly SeccionVista[] = [
  { clave: 'productos', nombre: 'Productos', descripcion: 'Lo usan compras, precios e inventario.', fija: true },
  { clave: 'marcas', nombre: 'Marcas', descripcion: 'Menú, campo «Marca» del producto y filtro de la lista.', fija: false },
  { clave: 'categorias', nombre: 'Categorías', descripcion: 'Menú, campo «Categoría» del producto, filtro y columna de la lista.', fija: false },
  { clave: 'unidades', nombre: 'Unidades de medida', descripcion: 'Las usan compras (Caja 12, Caja 24) e inventario.', fija: true },
];

/** Configuración > Vistas: el propietario elige qué secciones ve su equipo (adecuación 1). */
@Component({
  selector: 'app-vistas',
  templateUrl: './vistas.component.html',
  styleUrl: './vistas.component.css',
})
export class VistasComponent implements OnInit {
  private readonly vistas = inject(VistasService);
  private readonly feedback = inject(FeedbackService);
  readonly contexto = inject(ContextoService);

  readonly secciones = SECCIONES;
  readonly cargando = signal(true);
  readonly guardando = signal(false);
  readonly error = signal<string | null>(null);
  private readonly guardadas = signal<ReadonlySet<SeccionOcultable>>(new Set());
  readonly borrador = signal<ReadonlySet<SeccionOcultable>>(new Set());

  readonly cambios = computed(() =>
    SECCIONES.filter((seccion) => !seccion.fija && this.oculta(seccion, this.borrador()) !== this.oculta(seccion, this.guardadas())).length,
  );
  readonly ocultasBorrador = computed(() => SECCIONES.filter((seccion) => !seccion.fija && this.oculta(seccion, this.borrador())));
  readonly aviso = computed(() => {
    const nombres = this.ocultasBorrador().map((seccion) => seccion.nombre);
    if (!nombres.length) return null;
    const verbo = nombres.length > 1 ? 'quedarán ocultas' : 'quedará oculta';
    return `${nombres.join(' y ')} ${verbo} para todo el equipo. Los productos conservan sus datos; nadie los pierde.`;
  });
  readonly efectos = computed(() => {
    const efectos: string[] = [];
    if (this.borrador().has('marcas')) efectos.push('Campo y filtro «Marca».');
    if (this.borrador().has('categorias')) efectos.push('Campo, filtro y columna «Categoría».');
    return efectos;
  });

  ngOnInit(): void {
    this.vistas.asegurar().subscribe({
      next: () => {
        const ocultas = new Set(this.vistas.ocultas());
        this.guardadas.set(ocultas);
        this.borrador.set(new Set(ocultas));
        this.cargando.set(false);
      },
      error: () => {
        this.cargando.set(false);
        this.error.set('No fue posible cargar las vistas.');
      },
    });
  }

  visible(seccion: SeccionVista): boolean {
    return seccion.fija || !this.oculta(seccion, this.borrador());
  }

  alternar(seccion: SeccionVista): void {
    if (seccion.fija || this.guardando()) return;
    const clave = seccion.clave as SeccionOcultable;
    const siguiente = new Set(this.borrador());
    if (siguiente.has(clave)) siguiente.delete(clave);
    else siguiente.add(clave);
    this.borrador.set(siguiente);
  }

  descartar(): void {
    this.borrador.set(new Set(this.guardadas()));
  }

  guardar(): void {
    if (!this.cambios() || this.guardando()) return;
    this.guardando.set(true);
    this.error.set(null);
    this.vistas.guardar([...this.borrador()]).subscribe({
      next: (vistas) => {
        const ocultas = new Set(vistas.secciones_ocultas ?? []);
        this.guardadas.set(ocultas);
        this.borrador.set(new Set(ocultas));
        this.guardando.set(false);
        this.feedback.success('Vistas guardadas', 'El menú ya se actualizó para tu equipo.');
      },
      error: (response) => {
        this.guardando.set(false);
        this.error.set(response?.error?.mensaje ?? 'No fue posible guardar las vistas.');
      },
    });
  }

  private oculta(seccion: SeccionVista, ocultas: ReadonlySet<SeccionOcultable>): boolean {
    return !seccion.fija && ocultas.has(seccion.clave as SeccionOcultable);
  }
}

import { Component, computed, effect, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { PuedeDirective } from '../../contexto/puede.directive';
import { ComprasService } from '../compras/compras.service';
import { EmpleadoService } from '../equipo/empleado.service';
import { ProductosService } from '../productos/productos.service';

const ZONA = 'America/Mexico_City';

/** Consejos verificados contra registro-productos.md y modulo-compras.md (capability 32). */
const CONSEJOS = [
  {
    icono: 'pi-file-excel',
    titulo: 'Carga productos sin preparar catálogos',
    texto:
      'No necesitas dar de alta marcas, categorías, unidades ni proveedores antes. Sube tu Excel y los nombres nuevos se crean con una sola confirmación.',
    accion: 'Abrir carga masiva',
    ruta: '/productos/importar',
  },
  {
    icono: 'pi-barcode',
    titulo: 'El código de barras es opcional',
    texto: 'Si un producto no trae código, regístralo igual con su nombre. Puedes agregar el código después, cuando lo tengas a la mano.',
    accion: 'Ver productos',
    ruta: '/productos',
  },
  {
    icono: 'pi-map-marker',
    titulo: 'Una sucursal nueva empieza vacía',
    texto: 'El catálogo es de toda la empresa. Usa «Agregar del catálogo» para surtir otra sucursal sin volver a capturar productos.',
    accion: 'Ver productos',
    ruta: '/productos',
  },
  {
    icono: 'pi-shopping-bag',
    titulo: 'Compra lo que no tienes registrado',
    texto: 'Si al registrar una compra el producto no existe, usa «Crear nuevo…» ahí mismo. Un producto puede tener varios proveedores.',
    accion: 'Ver compras',
    ruta: '/compras',
  },
  {
    icono: 'pi-tag',
    titulo: 'Cada sucursal tiene su precio',
    texto: 'Al editar un producto, el precio se guarda en la sucursal activa y queda registrado en el historial de precios.',
    accion: 'Ver productos',
    ruta: '/productos',
  },
];

interface Avance {
  productos: boolean;
  equipo: boolean;
  compras: boolean;
}

const hayRegistros = (respuesta: { items?: unknown[] | null; total?: number }) =>
  (respuesta.total ?? respuesta.items?.length ?? 0) > 0;

export function saludoPorHora(hora: number): string {
  if (hora >= 5 && hora < 12) return 'Buenos días';
  if (hora >= 12 && hora < 19) return 'Buenas tardes';
  return 'Buenas noches';
}

@Component({
  selector: 'app-home',
  imports: [RouterLink, PuedeDirective],
  templateUrl: './home.component.html',
  styleUrl: './home.component.css',
})
export class HomeComponent {
  readonly contexto = inject(ContextoService);
  private readonly productos = inject(ProductosService);
  private readonly empleados = inject(EmpleadoService);
  private readonly compras = inject(ComprasService);
  readonly P = PERMISOS;
  readonly alta = this.contexto.estadoAlta;
  /** Fecha de hoy en Ciudad de México, como en el resto de la app. */
  readonly hoy = new Intl.DateTimeFormat('es-MX', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    timeZone: ZONA,
  }).format(new Date());
  readonly saludo = saludoPorHora(
    Number(new Intl.DateTimeFormat('es-MX', { hour: 'numeric', hourCycle: 'h23', timeZone: ZONA }).format(new Date())),
  );
  readonly esPropietario = computed(() => this.contexto.negocio()?.tipo_miembro === 'propietario');

  readonly consejos = CONSEJOS;
  readonly indiceConsejo = signal(0);
  readonly consejo = computed(() => CONSEJOS[this.indiceConsejo()]);

  /** null mientras carga o si alguna consulta falla: la sección no se muestra. */
  readonly avance = signal<Avance | null>(null);
  readonly hechos = computed(() => {
    const avance = this.avance();
    return avance ? 2 + Number(avance.productos) + Number(avance.equipo) + Number(avance.compras) : 0;
  });
  readonly mostrarPasos = computed(() => this.avance() !== null && this.hechos() < 5);

  constructor() {
    effect((onCleanup) => {
      this.avance.set(null);
      const negocio = this.contexto.negocio();
      const sucursal = this.contexto.sucursal();
      if (this.alta() !== 'listo' || !negocio || !sucursal || negocio.tipo_miembro !== 'propietario') return;
      // shortcut: las tres listas llegan completas (sin paginar); con catálogos grandes conviene un endpoint de resumen.
      const consulta = forkJoin({
        productos: this.productos.listByBusiness(negocio.id, sucursal.id),
        equipo: this.empleados.listar(negocio.id),
        compras: this.compras.list(negocio.id),
      }).subscribe({
        next: (r) =>
          this.avance.set({ productos: hayRegistros(r.productos), equipo: hayRegistros(r.equipo), compras: hayRegistros(r.compras) }),
        error: () => this.avance.set(null),
      });
      onCleanup(() => consulta.unsubscribe());
    });
  }

  moverConsejo(paso: number): void {
    this.indiceConsejo.update((i) => (i + paso + CONSEJOS.length) % CONSEJOS.length);
  }

  irASecciones(evento: Event): void {
    evento.preventDefault();
    document.getElementById('secciones')?.scrollIntoView({ behavior: 'smooth' });
  }
}

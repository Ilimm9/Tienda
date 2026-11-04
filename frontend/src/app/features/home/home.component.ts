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
    titulo: 'Sube tu Excel tal cual',
    texto: 'Marcas, categorías y proveedores nuevos se crean solos.',
    accion: 'Carga masiva',
    ruta: '/productos/importar',
  },
  {
    icono: 'pi-barcode',
    titulo: '¿Sin código de barras?',
    texto: 'Regístralo igual. El código lo agregas después.',
    accion: 'Ver productos',
    ruta: '/productos',
  },
  {
    icono: 'pi-map-marker',
    titulo: 'Surte otra sucursal',
    texto: 'Toma productos del catálogo, sin volver a capturarlos.',
    accion: 'Ver productos',
    ruta: '/productos',
  },
  {
    icono: 'pi-shopping-bag',
    titulo: '¿El producto no existe?',
    texto: 'Créalo ahí mismo, desde la compra.',
    accion: 'Ver compras',
    ruta: '/compras',
  },
  {
    icono: 'pi-tag',
    titulo: 'Un precio por sucursal',
    texto: 'Cada cambio queda guardado en el historial.',
    accion: 'Ver productos',
    ruta: '/productos',
  },
];

interface Paso {
  titulo: string;
  icono: string;
  hecho: boolean;
  nota: string;
  accion: string;
  ruta: string;
}

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
  readonly pasos = computed<Paso[]>(() => {
    const a = this.avance();
    if (!a) return [];
    return [
      { titulo: 'Empresa', icono: 'pi-building', hecho: true, nota: `${this.contexto.negocio()?.nombre_comercial ?? 'Tu empresa'} ya está registrada.`, accion: 'Ver negocio', ruta: '/negocios' },
      { titulo: 'Sucursal', icono: 'pi-map-marker', hecho: true, nota: `${this.contexto.sucursal()?.nombre ?? 'Tu sucursal'} ya abrió sus puertas.`, accion: 'Ver sucursales', ruta: '/sucursales' },
      a.productos
        ? { titulo: 'Productos', icono: 'pi-box', hecho: true, nota: 'Tu catálogo ya tiene productos.', accion: 'Ver productos', ruta: '/productos' }
        : { titulo: 'Productos', icono: 'pi-box', hecho: false, nota: 'Súbelos uno por uno o todos con tu Excel.', accion: 'Cargar productos', ruta: '/productos/importar' },
      a.equipo
        ? { titulo: 'Equipo', icono: 'pi-users', hecho: true, nota: 'Ya tienes equipo trabajando contigo.', accion: 'Ver equipo', ruta: '/equipo/empleados' }
        : { titulo: 'Equipo', icono: 'pi-users', hecho: false, nota: 'Invita a quien te ayuda en la tienda.', accion: 'Invitar', ruta: '/equipo/empleados/nuevo' },
      a.compras
        ? { titulo: 'Primera compra', icono: 'pi-shopping-bag', hecho: true, nota: 'Ya registraste compras.', accion: 'Ver compras', ruta: '/compras' }
        : { titulo: 'Primera compra', icono: 'pi-shopping-bag', hecho: false, nota: 'Registra lo que compras y tu inventario se arma solo.', accion: 'Registrar compra', ruta: '/compras/nueva' },
    ];
  });
  /** Primer paso pendiente: se resalta y se elige al entrar. */
  readonly pasoSiguiente = computed(() => this.pasos().findIndex((p) => !p.hecho));
  readonly pasoElegido = signal<number | null>(null);
  readonly pasoActivo = computed(() => this.pasos()[this.pasoElegido() ?? this.pasoSiguiente()] ?? null);

  constructor() {
    effect((onCleanup) => {
      this.avance.set(null);
      this.pasoElegido.set(null);
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

  elegirPaso(indice: number): void {
    this.pasoElegido.set(indice);
  }

  moverConsejo(paso: number): void {
    this.indiceConsejo.update((i) => (i + paso + CONSEJOS.length) % CONSEJOS.length);
  }

  irASecciones(evento: Event): void {
    evento.preventDefault();
    document.getElementById('secciones')?.scrollIntoView({ behavior: 'smooth' });
  }
}

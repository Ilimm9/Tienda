import { Component, computed, effect, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { ConsejoMascota, MascotaComponent } from '../../shared/ui/mascota/mascota.component';
import { ComprasService } from '../compras/compras.service';
import { EmpleadoService } from '../equipo/empleado.service';
import { ProductosService } from '../productos/productos.service';

const ZONA = 'America/Mexico_City';

/** Consejos verificados contra registro-productos.md y modulo-compras.md (capability 32). */
const CONSEJOS: ConsejoMascota[] = [
  { titulo: 'Sube tu Excel tal cual', texto: 'Marcas, categorías y proveedores nuevos se crean solos.', accion: 'Ir a carga masiva', ruta: '/productos/importar' },
  { titulo: '¿Sin código de barras?', texto: 'Regístralo igual. El código lo agregas después.', accion: 'Ver productos', ruta: '/productos' },
  { titulo: 'Surte otra sucursal', texto: 'Toma productos del catálogo, sin volver a capturarlos.', accion: 'Ver productos', ruta: '/productos' },
  { titulo: '¿El producto no existe?', texto: 'Créalo ahí mismo, desde la compra. No tienes que salirte.', accion: 'Ver compras', ruta: '/compras' },
  { titulo: 'Un precio por sucursal', texto: 'Cada cambio queda guardado en el historial.', accion: 'Ver productos', ruta: '/productos' },
];

interface PuntoPaso {
  titulo: string;
  texto: string;
}

interface Paso {
  titulo: string;
  /** Título largo para el panel de detalle. */
  encabezado: string;
  icono: string;
  hecho: boolean;
  nota: string;
  accion: string;
  ruta: string;
  puntos: PuntoPaso[];
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
  imports: [RouterLink, MascotaComponent],
  templateUrl: './home.component.html',
  styleUrl: './home.component.css',
})
export class HomeComponent {
  readonly contexto = inject(ContextoService);
  private readonly productos = inject(ProductosService);
  private readonly empleados = inject(EmpleadoService);
  private readonly compras = inject(ComprasService);
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
    const sucursal = this.contexto.sucursal();
    return [
      {
        titulo: 'Empresa', encabezado: 'Tu empresa', icono: 'pi-building', hecho: true,
        nota: `${this.contexto.negocio()?.nombre_comercial ?? 'Tu empresa'} ya está registrada. Puedes cambiar sus datos cuando quieras.`,
        accion: 'Ver negocio', ruta: '/negocios',
        puntos: [
          { titulo: 'Nombre comercial', texto: 'RFC y datos fiscales pueden esperar.' },
          { titulo: 'Catálogo de la empresa', texto: 'Tus productos sirven para todas las sucursales.' },
        ],
      },
      {
        titulo: 'Sucursal', encabezado: 'Tu primera sucursal', icono: 'pi-map-marker', hecho: true,
        nota: `${sucursal?.nombre ?? 'Tu sucursal'} ya abrió sus puertas. Si tienes más tiendas, agrégalas aquí.`,
        accion: 'Ver sucursales', ruta: '/sucursales',
        puntos: [
          { titulo: sucursal?.nombre ?? 'Tu sucursal', texto: sucursal?.codigo ? `Código ${sucursal.codigo}.` : 'Queda como principal.' },
          { titulo: 'Precios por sucursal', texto: 'Cada tienda puede tener los suyos.' },
        ],
      },
      {
        titulo: 'Productos', encabezado: 'Carga tus productos', icono: 'pi-box', hecho: a.productos,
        nota: a.productos
          ? 'Tu catálogo ya tiene productos. Puedes seguir agregando cuando quieras.'
          : 'Súbelos uno por uno o todos de golpe con tu Excel. Con esto ya puedes empezar a vender.',
        accion: a.productos ? 'Ver productos' : 'Cargar productos',
        ruta: a.productos ? '/productos' : '/productos/importar',
        puntos: [
          { titulo: 'Sube tu Excel tal cual', texto: 'Marcas, categorías y proveedores nuevos se crean solos.' },
          { titulo: 'Revisa la vista previa', texto: 'Corriges antes de guardar.' },
          { titulo: 'Sin código de barras, también', texto: 'Lo agregas después.' },
        ],
      },
      {
        titulo: 'Equipo', encabezado: 'Invita a tu equipo', icono: 'pi-users', hecho: a.equipo,
        nota: a.equipo
          ? 'Ya tienes equipo trabajando contigo.'
          : 'Invita a quien te ayuda en la tienda. Cada quien entra con su propia cuenta.',
        accion: a.equipo ? 'Ver equipo' : 'Invitar',
        ruta: a.equipo ? '/equipo/empleados' : '/equipo/empleados/nuevo',
        puntos: [
          { titulo: 'Manda la invitación', texto: 'El enlace vence en 72 horas.' },
          { titulo: 'Asigna un puesto', texto: 'Define qué hace cada quien.' },
          { titulo: 'Ajusta permisos', texto: 'Desde Roles y permisos.' },
        ],
      },
      {
        titulo: 'Primera compra', encabezado: 'Registra tu primera compra', icono: 'pi-shopping-bag', hecho: a.compras,
        nota: a.compras
          ? 'Ya registraste compras.'
          : 'Registra lo que le compras a tus proveedores y tu inventario se arma solo.',
        accion: a.compras ? 'Ver compras' : 'Registrar compra',
        ruta: a.compras ? '/compras' : '/compras/nueva',
        puntos: [
          { titulo: 'Elige al proveedor', texto: 'El que te surte la mercancía.' },
          { titulo: 'Agrega productos y cantidades', texto: 'Si el producto no existe, créalo ahí mismo.' },
          { titulo: 'Valida costos y precios', texto: 'Antes de que lleguen a caja.' },
        ],
      },
    ];
  });
  /** Primer paso pendiente: se resalta y se elige al entrar. */
  readonly pasoSiguiente = computed(() => this.pasos().findIndex((p) => !p.hecho));
  readonly pasoElegido = signal<number | null>(null);
  readonly indiceActivo = computed(() => this.pasoElegido() ?? this.pasoSiguiente());
  readonly pasoActivo = computed(() => this.pasos()[this.indiceActivo()] ?? null);

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
        compras: this.compras.list(negocio.id, sucursal.id),
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

  estadoPaso(indice: number): string {
    const paso = this.pasos()[indice];
    if (!paso) return '';
    return paso.hecho ? 'Listo' : indice === this.pasoSiguiente() ? 'Sigue este' : 'Después';
  }

  irASecciones(evento: Event): void {
    evento.preventDefault();
    document.getElementById('secciones')?.scrollIntoView({ behavior: 'smooth' });
  }
}

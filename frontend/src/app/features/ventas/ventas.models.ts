export type NivelMargen = 'negativo' | 'bajo' | 'ok';

/** Coincidencia de `GET /ventas/productos`; los importes llegan en pesos con dos decimales. */
export interface ProductoVendible {
  producto_negocio_id: string;
  nombre: string;
  codigo_barras: string | null;
  precio: number;
  existencia: number;
  agotado: boolean;
  existencia_baja: boolean;
  /** Solo con `ventas.ver_costos`. */
  costo?: number;
  margen?: number;
  nivel_margen?: NivelMargen;
}

export interface ResultadoBusqueda {
  items: ProductoVendible[];
  por_codigo: boolean;
}

/** Línea de la venta en captura; importes en centavos enteros para no acumular errores de redondeo. */
export interface LineaVenta {
  productoNegocioId: string;
  nombre: string;
  codigoBarras: string | null;
  precioCentavos: number;
  costoCentavos: number | null;
  existencia: number;
  existenciaBaja: boolean;
  cantidad: number;
}

export interface RegistrarVentaRequest {
  sucursal_id: string;
  forma_pago: 'EFECTIVO';
  /** Pesos con dos decimales; la API los convierte a centavos sin redondeos flotantes. */
  recibido: string;
  lineas: { producto_negocio_id: string; cantidad: number }[];
}

export interface VentaLineaRegistrada {
  id: string;
  producto_negocio_id: string;
  codigo_barras?: string;
  nombre: string;
  cantidad: number;
  precio_unitario: number;
  importe: number;
}

export interface VentaRegistrada {
  id: string;
  numero: number;
  folio: string;
  sucursal_id: string;
  usuario_id: string;
  subtotal: number;
  iva: number;
  total: number;
  recibido: number;
  cambio: number;
  piezas: number;
  forma_pago: FormaPago;
  referencia?: string;
  estado: 'PAGADA' | 'CANCELADA';
  creado_en: string;
  /** Nombre del perfil de quien cobró o, sin perfil, su correo. */
  atendio: string;
  /** Pagada y del mismo día en el calendario del negocio. */
  cancelable: boolean;
  motivo_cancelacion?: string;
  motivo_cancelacion_codigo?: MotivoCancelacion;
  cancelada_por?: string;
  cancelada_en?: string;
  /** El listado no trae líneas; el detalle sí. */
  lineas?: VentaLineaRegistrada[];
}

export type FormaPago = 'EFECTIVO' | 'TARJETA' | 'TRANSFERENCIA';
export type PeriodoVentas = 'hoy' | 'semana' | 'mes';
export type MotivoCancelacion =
  'ERROR_CAPTURA' | 'COBRO_DUPLICADO' | 'DEVOLUCION_INMEDIATA' | 'OTRO';

export const MOTIVOS_CANCELACION: readonly { valor: MotivoCancelacion; etiqueta: string }[] = [
  { valor: 'ERROR_CAPTURA', etiqueta: 'Error de captura' },
  { valor: 'COBRO_DUPLICADO', etiqueta: 'Cobro duplicado' },
  { valor: 'DEVOLUCION_INMEDIATA', etiqueta: 'Devolución inmediata del cliente' },
  { valor: 'OTRO', etiqueta: 'Otro' },
];

export const FORMAS_PAGO: Record<FormaPago, string> = {
  EFECTIVO: 'Efectivo',
  TARJETA: 'Tarjeta',
  TRANSFERENCIA: 'Transferencia',
};

export interface ResumenVentas {
  vendido: number;
  tickets: number;
  ticket_promedio: number;
  efectivo: number;
  cancelados: number;
}

export interface ListadoVentas {
  items: VentaRegistrada[];
  total: number;
  resumen: ResumenVentas;
  desde?: string;
  hasta?: string;
}

/** Error de `POST /ventas`; `producto` viene con `EXISTENCIA_INSUFICIENTE`. */
export interface ErrorVenta {
  codigo?:
    | 'EXISTENCIA_INSUFICIENTE'
    | 'PAGO_INSUFICIENTE'
    | 'PRODUCTO_NO_DISPONIBLE'
    | 'VENTA_INVALIDA'
    | 'VENTA_FUERA_DE_PLAZO'
    | 'VENTA_YA_CANCELADA'
    | string;
  mensaje?: string;
  producto?: string;
  disponible?: number;
}

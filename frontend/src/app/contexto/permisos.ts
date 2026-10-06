/**
 * Códigos de permiso, espejo de `CatalogoPermisos` en el backend.
 * La interfaz los usa para ocultar y redirigir; la autorización real ocurre en la API.
 */
export const PERMISOS = {
  negocioVer: 'negocios.ver',
  negocioEditar: 'negocios.editar',
  negocioArchivar: 'negocios.archivar',
  sucursalVer: 'sucursales.ver',
  sucursalCrear: 'sucursales.crear',
  sucursalEditar: 'sucursales.editar',
  sucursalArchivar: 'sucursales.archivar',
  rolVer: 'roles.ver',
  rolGestionar: 'roles.gestionar',
  rolAsignar: 'roles.asignar',
  empleadoVer: 'equipo.empleados.ver',
  empleadoGestionar: 'equipo.empleados.gestionar',
  invitacionVer: 'equipo.invitaciones.ver',
  invitacionEnviar: 'equipo.invitaciones.enviar',
  asignacionVer: 'equipo.asignaciones.ver',
  asignacionEditar: 'equipo.asignaciones.editar',
  catalogoVer: 'catalogo.ver',
  catalogoGestionar: 'catalogo.gestionar',
  compraVer: 'compras.ver',
  compraRegistrar: 'compras.registrar',
  precioVer: 'precios.ver',
  precioAutorizar: 'precios.autorizar',
} as const;

export type CodigoPermiso = (typeof PERMISOS)[keyof typeof PERMISOS];

/** Un permiso o varios; cuando son varios se exigen todos. */
export type PermisoRequerido = CodigoPermiso | readonly CodigoPermiso[];

export function comoLista(requerido: PermisoRequerido | null | undefined): readonly CodigoPermiso[] {
  if (!requerido) return [];
  return typeof requerido === 'string' ? [requerido] : requerido;
}

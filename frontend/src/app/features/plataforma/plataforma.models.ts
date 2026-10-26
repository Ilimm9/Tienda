export interface ActividadDiaria {
  fecha: string;
  usuarios: number;
  solicitudes: number;
}

export interface PlataformaResumen {
  negocios_total: number;
  negocios_activos: number;
  negocios_archivados: number;
  sucursales_activas: number;
  cuentas_registradas: number;
  usuarios_activos_7d: number;
  usuarios_activos_30d: number;
  actividad: ActividadDiaria[];
}

export interface NegocioPlataformaResumen {
  id: string;
  slug: string;
  nombre_comercial: string;
  rfc: string | null;
  estado: string;
  creado_en: string;
  total_sucursales: number;
  total_personal: number;
  total_productos: number;
  cuentas_vinculadas: number;
  ultima_actividad_en: string | null;
}

export interface SucursalPlataformaResumen {
  id: string;
  codigo: string;
  nombre: string;
  es_principal: boolean;
  estado: string;
  personal_asignado: number;
  con_cuenta_activa: number;
}

export interface NegocioPlataformaDetalle extends NegocioPlataformaResumen {
  razon_social: string | null;
  correo: string | null;
  zona_horaria: string;
  total_propietarios: number;
  sucursales: SucursalPlataformaResumen[];
  actividad: ActividadDiaria[];
}

export interface UsuarioPlataformaResumen {
  id: string;
  nombre: string;
  correo: string;
  estado: string;
  tipo_miembro: string;
  roles: string[];
  sucursales: string[];
  ultimo_inicio_sesion_en: string | null;
}

export interface PaginaNegocios {
  items: NegocioPlataformaResumen[];
  total: number;
  page: number;
  page_size: number;
}

export interface UsuariosResponse {
  items: UsuarioPlataformaResumen[];
  total: number;
}

export interface CuentaPlataformaResumen {
  id: string;
  nombre: string;
  correo: string;
  estado: string;
  correo_verificado: boolean;
  creado_en: string;
  ultimo_inicio_sesion_en: string | null;
  total_negocios: number;
  negocios: string[];
}

export interface PaginaCuentas {
  items: CuentaPlataformaResumen[];
  total: number;
  page: number;
  page_size: number;
}

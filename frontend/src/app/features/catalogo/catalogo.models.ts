export interface Marca {
  id: string;
  nombre: string;
  activo: boolean;
}
export interface Categoria {
  id: string;
  nombre: string;
  descripcion?: string | null;
  categoria_padre_id?: string | null;
  activo: boolean;
}
export interface Proveedor {
  id: string;
  nombre: string;
  razon_social?: string | null;
  rfc?: string | null;
  telefono?: string | null;
  email?: string | null;
  direccion?: string | null;
  activo: boolean;
}
export interface UnidadMedida { id: string; nombre: string; simbolo: string; tipo: string; factor_a_base: number; permite_fraccion: boolean; decimales: number; activo: boolean; }
export type CatalogRecord = Marca | Categoria | Proveedor | UnidadMedida;

export interface CatalogImportIssue {
  fila: number;
  campo?: string;
  motivo: string;
}

/** Lo que no se agregó de una lista pegada, con su número de renglón. */
export interface RenglonOmitido {
  numero: number;
  nombre: string;
  motivo: string;
  /** Dato inválido que se puede corregir (no un repetido ni un existente). */
  corregible: boolean;
  linea: string;
}
export interface ReporteLista {
  creadas: number;
  omitidos: RenglonOmitido[];
}
export interface CatalogImportResult {
  procesadas: number;
  creadas: number;
  omitidas: number;
  invalidas: number;
  errores: CatalogImportIssue[];
}

export interface CatalogDeletionCheck {
  puede_eliminarse: boolean;
  vinculaciones: string[];
}

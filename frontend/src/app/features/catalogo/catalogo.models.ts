export interface Marca { id: string; nombre: string; activo: boolean; }
export interface Categoria { id: string; nombre: string; descripcion?: string | null; categoria_padre_id?: string | null; activo: boolean; }
export interface Proveedor { id: string; nombre: string; razon_social?: string | null; rfc?: string | null; telefono?: string | null; email?: string | null; direccion?: string | null; activo: boolean; }
export type CatalogRecord = Marca | Categoria | Proveedor;

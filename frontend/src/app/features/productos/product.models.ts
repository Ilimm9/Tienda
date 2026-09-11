export type ProductStatus = 'En stock' | 'Agotado' | 'Bajo stock';

export interface ProductRow {
  id: string;
  nombre: string;
  imagen_url: string | null;
  sku: string | null;
  precio: number;
  stock: number;
  categoria: string | null;
  categoria_id: string | null;
  marca: string | null;
  marca_id: string | null;
  descripcion: string | null;
  presentacion: string | null;
  contenido: number | null;
  unidad_contenido: string | null;
  unidad_medida: string | null;
  unidad_medida_id: string | null;
  codigo_barras: string | null;
  inventario: ProductBranchStock[];
  estado: ProductStatus;
}

export interface ProductBranchStock {
  sucursal_id: string;
  sucursal: string;
  stock: number;
}

export interface ProductListResponse {
  items: ProductRow[];
  total: number;
}

export interface CatalogOption {
  id: string;
  nombre: string;
}

export interface CreateProductRequest {
  nombre: string;
  sku_interno: string;
  marca_id: string | null;
  categoria_id: string;
  sucursal_id: string;
  descripcion: string | null;
  contenido: number | null;
  unidad_contenido: string | null;
  unidad_medida_id?: string | null;
  presentacion: string | null;
  precio_venta: number;
  stock_inicial: number;
  codigo_barras: string | null;
  imagen_url: string | null;
}

export interface UpdateProductRequest extends Omit<CreateProductRequest, 'sucursal_id' | 'stock_inicial'> {}

export interface ProductLookup {
  codigo_barras: string;
  nombre: string | null;
  descripcion: string | null;
  marca: string | null;
  categoria: string | null;
  precio_sugerido: number | null;
  contenido: number | null;
  unidad_contenido: string | null;
  imagen_url: string | null;
  fuentes: string[];
}

export interface ProductImportIssue {
  fila: number;
  campo?: string;
  motivo: string;
}

export interface ProductImportResult {
  procesadas: number;
  creadas: number;
  omitidas: number;
  invalidas: number;
  errores: ProductImportIssue[];
  advertencias: ProductImportIssue[];
}

export interface ProductImportPreview extends ProductImportResult {
  insertables: number;
}

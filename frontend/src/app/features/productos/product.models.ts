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
  variantes: ProductVariantRow[];
}

export interface ProductVariantAttribute {
  nombre: string;
  valor: string;
}

export interface ProductVariantRow {
  id: string;
  sku: string | null;
  precio: number;
  stock: number;
  codigo_barras: string | null;
  atributos: ProductVariantAttribute[];
}

export interface CreateProductVariantRequest {
  atributos: ProductVariantAttribute[];
  sku_interno: string;
  generar_sku_interno: boolean;
  precio_venta: number;
  stock_inicial: number;
  codigo_barras: string | null;
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
  generar_sku_interno: boolean;
  marca_id: string | null;
  categoria_id: string | null;
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
  variantes?: CreateProductVariantRequest[];
}

export interface UpdateProductRequest extends Omit<CreateProductRequest, 'sucursal_id' | 'stock_inicial' | 'generar_sku_interno' | 'categoria_id'> {
  categoria_id: string;
}

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
  skus_generados: number;
  productos_base_creados: number;
  productos_base_reutilizados: number;
  variantes_creadas: number;
  errores: ProductImportIssue[];
  advertencias: ProductImportIssue[];
}

export interface ProductImportPreview extends ProductImportResult {
  insertables: number;
}

export interface ProductImportJob {
  id: string;
  estado: 'pendiente' | 'procesando' | 'completada' | 'fallida';
  etapa: string;
  porcentaje: number;
  mensaje_error: string | null;
  resultado: ProductImportResult | null;
}

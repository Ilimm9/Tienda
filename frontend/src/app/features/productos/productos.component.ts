import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { FormArray, FormBuilder, FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { ButtonModule } from 'primeng/button';
import { CheckboxModule } from 'primeng/checkbox';
import { InputTextModule } from 'primeng/inputtext';
import { SelectModule } from 'primeng/select';
import { TableModule } from 'primeng/table';
import { TextareaModule } from 'primeng/textarea';

import { ContextoService } from '../../contexto/contexto.service';
import { CatalogOption, ProductImportPreview, ProductImportResult, ProductLookup, ProductRow, ProductVariantRow } from './product.models';
import { ProductosService } from './productos.service';
import Swal from 'sweetalert2/dist/sweetalert2.esm.js';

type VariantFormGroup = FormGroup<{
  atributos: FormControl<string>;
  sku_interno: FormControl<string>;
  generar_sku_interno: FormControl<boolean>;
  precio_venta: FormControl<number>;
  stock_inicial: FormControl<number>;
  codigo_barras: FormControl<string>;
}>;

@Component({
  selector: 'app-productos',
  standalone: true,
  imports: [
    CommonModule,
    ButtonModule,
    CheckboxModule,
    FormsModule,
    ReactiveFormsModule,
    InputTextModule,
    SelectModule,
    TableModule,
    TextareaModule,
  ],
  templateUrl: './productos.component.html',
  styleUrl: './productos.component.css',
})
export class ProductosComponent {
  private readonly productosService = inject(ProductosService);
  private readonly contexto = inject(ContextoService);
  private readonly formBuilder = inject(FormBuilder);
  private readonly route = inject(ActivatedRoute, { optional: true });
  private readonly router = inject(Router, { optional: true });

  readonly pageMode =
    (this.route?.snapshot.data['mode'] as 'list' | 'create' | 'edit' | 'import' | undefined) ?? 'list';

  readonly products = signal<ProductRow[]>([]);
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly categories = signal<CatalogOption[]>([]);
  readonly brands = signal<CatalogOption[]>([]);
  readonly branches = signal<CatalogOption[]>([]);
  readonly units = signal<CatalogOption[]>([]);
  readonly catalogLoadErrors = signal<string[]>([]);
  readonly branchReady = signal(false);
  readonly saving = signal(false);
  readonly imageLookupLoading = signal(false);
  readonly imageLookupError = signal<string | null>(null);
  readonly previewImageURL = signal<string | null>(null);
  readonly lookupSources = signal<string[]>([]);
  readonly catalogLookupWarnings = signal<string[]>([]);
  readonly failedProductImages = signal<ReadonlySet<string>>(new Set());
  readonly importBranches = signal<CatalogOption[]>([]);
  readonly importing = signal(false);
  readonly importError = signal<string | null>(null);
  readonly importResult = signal<ProductImportResult | null>(null);
  readonly importPreview = signal<ProductImportPreview | null>(null);
  readonly editingProduct = signal<ProductRow | null>(null);
  readonly deletingProductId = signal<string | null>(null);
  private lastLookup: ProductLookup | null = null;
  private categoriesLoaded = false;
  private brandsLoaded = false;
  formError: string | null = null;
  importFile: File | null = null;
  importDropActive = false;
  importBranchId = '';

  readonly productForm = this.formBuilder.nonNullable.group({
    nombre: ['', [Validators.required, Validators.maxLength(255)]],
    sku_interno: ['', [Validators.required, Validators.maxLength(120)]],
    generar_sku_interno: [false],
    tiene_variantes: [false],
    variantes: this.formBuilder.array<VariantFormGroup>([]),
    codigo_barras: ['', Validators.pattern(/^\d{8,14}$/)],
    imagen_url: [''],
    marca_id: [''],
    categoria_id: [''],
    sucursal_id: ['', Validators.required],
    descripcion: ['', Validators.maxLength(2000)],
    contenido: this.formBuilder.control<number | null>(null, [Validators.min(0)]),
    unidad_contenido: ['', Validators.maxLength(30)],
    unidad_medida_id: [''],
    presentacion: ['', Validators.maxLength(100)],
    precio_venta: [0, [Validators.required, Validators.min(0)]],
    stock_inicial: [0, [Validators.required, Validators.min(0)]],
  });

  constructor() {
    if (this.pageMode === 'list') {
      this.loadProducts();
    } else if (this.pageMode === 'create') {
      this.openCreateDialog();
    } else if (this.pageMode === 'import') {
      this.openImportDialog();
    } else {
      this.loadProductForEdit();
    }
  }

  private get negocioID(): string {
    return this.contexto.negocio()?.id ?? '';
  }

  loadProducts(): void {
    this.loading.set(true);
    this.error.set(null);
    this.productosService.listByBusiness(this.negocioID).subscribe({
      next: ({ items }) => {
        this.products.set(items ?? []);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('No fue posible cargar los productos.');
        this.loading.set(false);
      },
    });
  }

  isListPage(): boolean {
    return this.pageMode === 'list';
  }

  isFormPage(): boolean {
    return this.pageMode === 'create' || this.pageMode === 'edit';
  }

  isImportPage(): boolean {
    return this.pageMode === 'import';
  }

  private loadProductForEdit(): void {
    const productId = this.route?.snapshot.paramMap.get('productoId');
    if (!productId) {
      this.navigate(['/catalogo/productos']);
      return;
    }
    this.loading.set(true);
    this.productosService.listByBusiness(this.negocioID).subscribe({
      next: ({ items }) => {
        const product = (items ?? []).find((item) => item.id === productId);
        this.loading.set(false);
        if (!product) {
          this.error.set('No fue posible encontrar el producto solicitado.');
          return;
        }
        this.openEditDialog(product);
      },
      error: () => {
        this.loading.set(false);
        this.error.set('No fue posible cargar el producto.');
      },
    });
  }

  openCreateDialog(): void {
    this.loading.set(false);
    const branchControl = this.productForm.controls.sucursal_id;
    this.productForm.reset({
      nombre: '', sku_interno: '', generar_sku_interno: false, tiene_variantes: false, marca_id: '', categoria_id: '', sucursal_id: '',
      codigo_barras: '', imagen_url: '',
      descripcion: '', contenido: null, unidad_contenido: '', presentacion: '',
      unidad_medida_id: '',
      precio_venta: 0, stock_inicial: 0,
    });
    this.productForm.controls.sucursal_id.enable();
    this.productForm.controls.stock_inicial.enable();
    this.productForm.controls.sku_interno.enable();
	this.productForm.controls.precio_venta.enable();
	this.productForm.controls.stock_inicial.enable();
	this.productForm.controls.variantes.clear();
    this.editingProduct.set(null);
    this.categories.set([]);
    this.brands.set([]);
    this.branches.set([]);
    this.branchReady.set(false);
    this.imageLookupLoading.set(false);
    this.imageLookupError.set(null);
    this.previewImageURL.set(null);
    this.lookupSources.set([]);
    this.catalogLookupWarnings.set([]);
    this.catalogLoadErrors.set([]);
    this.lastLookup = null;
    this.categoriesLoaded = false;
    this.brandsLoaded = false;
    this.formError = null;
    this.productosService.listCategories(this.negocioID).subscribe({
      next: (items) => {
        this.categories.set(items);
        this.categoriesLoaded = true;
        this.applyCatalogSuggestions();
      },
      error: () => this.addCatalogLoadError('No fue posible cargar las categorías.'),
    });
    this.productosService.listBrands(this.negocioID).subscribe({
      next: (items) => {
        this.brands.set(items);
        this.brandsLoaded = true;
        this.applyCatalogSuggestions();
      },
      error: () => this.addCatalogLoadError('No fue posible cargar las marcas.'),
    });
    this.productosService.listBranches(this.negocioID).subscribe({
      next: (items) => {
        this.branches.set(items);
        if (items.length === 0) {
          this.formError = 'No se encontró una sucursal activa para registrar el inventario.';
          return;
        }
        const activa = this.contexto.sucursal()?.id;
        branchControl.setValue(items.find((item) => item.id === activa)?.id ?? items[0].id);
        this.branchReady.set(true);
      },
      error: () => {
        this.formError = 'No fue posible cargar las sucursales.';
      },
    });
    this.productosService.listUnits(this.negocioID).subscribe({
      next: (items) => this.units.set(items),
      error: () => this.addCatalogLoadError('No fue posible cargar las unidades de medida.'),
    });
  }

  openEditDialog(product: ProductRow): void {
    this.categories.set([]);
    this.brands.set([]);
    this.units.set([]);
    this.catalogLoadErrors.set([]);
    this.imageLookupLoading.set(false);
    this.imageLookupError.set(null);
    this.previewImageURL.set(product.imagen_url);
    this.lookupSources.set([]);
    this.catalogLookupWarnings.set([]);
    this.lastLookup = null;
    this.formError = null;
    this.editingProduct.set(product);
    this.branchReady.set(true);
    this.productForm.reset({
      nombre: product.nombre,
      sku_interno: product.sku ?? '',
      generar_sku_interno: false,
	  tiene_variantes: false,
      codigo_barras: product.codigo_barras ?? '',
      imagen_url: product.imagen_url ?? '',
      marca_id: product.marca_id ?? '',
      categoria_id: product.categoria_id ?? '',
      sucursal_id: '',
      descripcion: product.descripcion ?? '',
      contenido: product.contenido,
      unidad_contenido: product.unidad_contenido ?? '',
      unidad_medida_id: product.unidad_medida_id ?? '',
      presentacion: product.presentacion ?? '',
      precio_venta: product.precio,
      stock_inicial: 0,
    });
    this.productForm.controls.sucursal_id.disable();
    this.productForm.controls.stock_inicial.disable();
	this.productForm.controls.sku_interno.enable();
	this.productForm.controls.variantes.clear();
    this.productosService.listCategories(this.negocioID).subscribe({
      next: (items) => this.categories.set(items),
      error: () => this.addCatalogLoadError('No fue posible cargar las categorías.'),
    });
    this.productosService.listBrands(this.negocioID).subscribe({
      next: (items) => this.brands.set(items),
      error: () => this.addCatalogLoadError('No fue posible cargar las marcas.'),
    });
    this.productosService.listUnits(this.negocioID).subscribe({
      next: (items) => this.units.set(items),
      error: () => this.addCatalogLoadError('No fue posible cargar las unidades de medida.'),
    });
  }

  confirmDeactivate(product: ProductRow): void {
    if (this.deletingProductId() || !window.confirm(`¿Deseas desactivar "${product.nombre}"? Podrás conservar su historial e inventario.`)) return;
    this.deletingProductId.set(product.id);
    this.productosService.deactivate(this.negocioID, product.id).subscribe({
      next: () => {
        this.deletingProductId.set(null);
        this.loadProducts();
      },
      error: (response) => {
        this.deletingProductId.set(null);
        this.error.set(response.error?.mensaje ?? 'No fue posible desactivar el producto.');
      },
    });
  }

  navigateToCreate(): void {
    this.navigate(['/catalogo/productos/nuevo']);
  }

  navigateToEdit(product: ProductRow): void {
    this.navigate(['/catalogo/productos', product.id, 'editar']);
  }

  navigateToImport(): void {
    this.navigate(['/catalogo/productos/importar']);
  }

  openImportDialog(): void {
    this.loading.set(false);
    this.importFile = null;
    this.importBranchId = '';
    this.importDropActive = false;
    this.importError.set(null);
    this.importResult.set(null);
    this.importPreview.set(null);
    this.importBranches.set([]);
    this.productosService.listBranches(this.negocioID).subscribe({
      next: (items) => {
        this.importBranches.set(items);
        this.importBranchId = items.find((item) => item.id === this.contexto.sucursal()?.id)?.id ?? items[0]?.id ?? '';
      },
      error: () => this.importError.set('No fue posible cargar las sucursales.'),
    });
  }

  onImportFile(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.setImportFile(input.files?.[0] ?? null);
    input.value = '';
  }

  openImportFilePicker(event: MouseEvent, input: HTMLInputElement): void {
    if (event.target !== input && !this.importing()) input.click();
  }

  onImportDragOver(event: DragEvent): void {
    event.preventDefault();
    if (!this.importing()) this.importDropActive = true;
  }

  onImportDragLeave(event: DragEvent): void {
    event.preventDefault();
    if (this.isMovingWithinDropzone(event)) return;
    this.importDropActive = false;
  }

  onImportDrop(event: DragEvent): void {
    event.preventDefault();
    this.importDropActive = false;
    if (!this.importing()) this.setImportFile(event.dataTransfer?.files?.[0] ?? null);
  }

  removeImportFile(event: MouseEvent, input: HTMLInputElement): void {
    event.stopPropagation();
    if (this.importing()) return;
    input.value = '';
    this.setImportFile(null);
  }

  private isMovingWithinDropzone(event: DragEvent): boolean {
    return event.currentTarget instanceof HTMLElement
      && event.relatedTarget instanceof Node
      && event.currentTarget.contains(event.relatedTarget);
  }

  private setImportFile(file: File | null): void {
    this.importFile = file;
    this.importError.set(null);
    this.importResult.set(null);
    this.importPreview.set(null);
  }

  validateProductImport(): void {
    if (!this.importFile || !this.importBranchId || this.importing()) return;
    this.importing.set(true);
    this.importError.set(null);
    this.importResult.set(null);
    this.importPreview.set(null);
    const file = this.importFile;
    const branchId = this.importBranchId;
    this.productosService.previewProductImport(this.negocioID, branchId, file).subscribe({
      next: (preview) => {
        this.importPreview.set(preview);
        this.importing.set(false);
      },
      error: (response) => {
        this.importing.set(false);
        this.importError.set(response.error?.mensaje ?? 'No fue posible validar el archivo.');
      },
    });
  }

  importExistingProducts(): void {
    if (!this.importFile || !this.importBranchId || !this.importPreview()?.insertables || this.importing()) return;
    this.importing.set(true);
    this.importError.set(null);
    this.executeProductImport(this.importFile, this.importBranchId);
  }

  private executeProductImport(file: File, branchId: string): void {
    this.productosService.importProducts(this.negocioID, branchId, file).subscribe({
      next: (job) => {
        Swal.fire({ title: 'Importando archivo', html: '<p id="import-progress-stage">Preparando importación</p><progress id="import-progress" value="0" max="100" style="width:100%"></progress><p id="import-progress-value">0%</p>', allowOutsideClick: false, allowEscapeKey: false, showConfirmButton: false });
        this.pollProductImport(job.id);
      },
      error: (response) => {
        this.importing.set(false);
        this.importError.set(response.error?.mensaje ?? 'No fue posible importar el archivo.');
      },
    });
  }

  private pollProductImport(importId: string): void {
    this.productosService.productImportStatus(this.negocioID, importId).subscribe({
      next: (job) => {
        const stage = document.getElementById('import-progress-stage');
        const progress = document.getElementById('import-progress') as HTMLProgressElement | null;
        const value = document.getElementById('import-progress-value');
        if (stage) stage.textContent = job.etapa;
        if (progress) progress.value = job.porcentaje;
        if (value) value.textContent = `${job.porcentaje}%`;
        if (job.estado === 'pendiente' || job.estado === 'procesando') { window.setTimeout(() => this.pollProductImport(importId), 750); return; }
        this.importing.set(false);
        this.importPreview.set(null);
        if (job.estado === 'completada' && job.resultado) {
          this.importResult.set(job.resultado);
          if (job.resultado.creadas > 0) this.loadProducts();
          const partial = job.resultado.omitidas > 0 || job.resultado.invalidas > 0 || job.resultado.advertencias.length > 0;
          Swal.fire({ icon: partial ? 'warning' : 'success', title: 'Importación terminada', text: `${job.resultado.productos_base_creados} productos base nuevos · ${job.resultado.productos_base_reutilizados} reutilizados · ${job.resultado.variantes_creadas} variantes registradas · ${job.resultado.omitidas} omitidas · ${job.resultado.invalidas} inválidas` });
          return;
        }
        this.importError.set(job.mensaje_error ?? 'No fue posible importar el archivo.');
        Swal.fire({ icon: 'error', title: 'No fue posible importar', text: job.mensaje_error ?? 'Ocurrió un error durante la importación.' });
      },
      error: () => { this.importing.set(false); this.importError.set('No fue posible consultar el progreso de importación.'); Swal.close(); },
    });
  }

  productImportTemplateUrl(): string {
    return this.productosService.productImportTemplateUrl(this.negocioID);
  }

  onBranchSelected(): void {
    this.branchReady.set(!!this.productForm.controls.sucursal_id.value);
  }

  onGenerateSKUChanged(): void {
    const controls = this.productForm.controls;
    if (controls.generar_sku_interno.value) {
      controls.sku_interno.setValue('');
      controls.sku_interno.disable();
      return;
    }
    controls.sku_interno.enable();
  }

  get variantControls(): FormArray<VariantFormGroup> {
    return this.productForm.controls.variantes;
  }

  onVariantsChanged(): void {
    const controls = this.productForm.controls;
    if (controls.tiene_variantes.value) {
      controls.sku_interno.disable();
      controls.precio_venta.disable();
      controls.stock_inicial.disable();
      if (this.variantControls.length === 0) this.addVariant();
      return;
    }
    controls.sku_interno.enable();
    controls.precio_venta.enable();
    controls.stock_inicial.enable();
    this.variantControls.clear();
  }

  addVariant(): void {
    this.variantControls.push(this.createVariantForm());
  }

  private createVariantForm(): VariantFormGroup {
    return this.formBuilder.nonNullable.group({
      atributos: ['', Validators.required],
      sku_interno: ['', Validators.maxLength(120)],
      generar_sku_interno: [true],
      precio_venta: [0, [Validators.required, Validators.min(0)]],
      stock_inicial: [0, [Validators.required, Validators.min(0)]],
      codigo_barras: ['', Validators.pattern(/^\d{8,14}$/)],
    });
  }

  removeVariant(index: number): void {
    this.variantControls.removeAt(index);
    if (this.variantControls.length === 0) {
      this.productForm.controls.tiene_variantes.setValue(false);
      this.onVariantsChanged();
    }
  }

  private addCatalogLoadError(message: string): void {
    this.catalogLoadErrors.update((messages) => messages.includes(message) ? messages : [...messages, message]);
  }

  canSubmitProduct(): boolean {
    return !this.saving() && (this.editingProduct() !== null || this.branchReady());
  }

  canLookupProduct(): boolean {
    return /^\d{8,14}$/.test(this.productForm.controls.codigo_barras.value.trim())
      && !this.imageLookupLoading();
  }

  onBarcodeChanged(): void {
    if (this.lastLookup) {
      const controls = this.productForm.controls;
      if (controls.nombre.pristine) controls.nombre.setValue('');
      if (controls.sku_interno.pristine) controls.sku_interno.setValue('');
      if (controls.descripcion.pristine) controls.descripcion.setValue('');
      if (controls.precio_venta.pristine) controls.precio_venta.setValue(0);
      if (controls.contenido.pristine) controls.contenido.setValue(null);
      if (controls.unidad_contenido.pristine) controls.unidad_contenido.setValue('');
      if (controls.marca_id.pristine) controls.marca_id.setValue('');
      if (controls.categoria_id.pristine) controls.categoria_id.setValue('');
    }
    this.productForm.controls.imagen_url.setValue('');
    this.previewImageURL.set(null);
    this.lookupSources.set([]);
    this.imageLookupError.set(null);
    this.catalogLookupWarnings.set([]);
    this.lastLookup = null;
  }

  lookupProduct(): void {
    if (!this.canLookupProduct()) return;

    const barcode = this.productForm.controls.codigo_barras.value.trim();
    this.imageLookupLoading.set(true);
    this.imageLookupError.set(null);
    this.productosService.lookupProduct(this.negocioID, barcode).subscribe({
      next: (product) => {
        this.imageLookupLoading.set(false);
        if (this.productForm.controls.codigo_barras.value.trim() !== barcode) return;

        this.lastLookup = product;
        this.lookupSources.set(product.fuentes ?? []);
        this.applyProductSuggestions(product);
        this.applyCatalogSuggestions();

        if (!product.imagen_url) {
          this.imageLookupError.set('El producto fue encontrado, pero no tiene una imagen disponible.');
          this.productForm.controls.imagen_url.setValue('');
          this.previewImageURL.set(null);
          return;
        }
        this.productForm.controls.imagen_url.setValue(product.imagen_url);
        this.previewImageURL.set(product.imagen_url);
      },
      error: (response) => {
        this.imageLookupLoading.set(false);
        if (this.productForm.controls.codigo_barras.value.trim() !== barcode) return;
        this.imageLookupError.set(response.error?.mensaje ?? 'No fue posible buscar el producto en PrecioCheck.');
      },
    });
  }

  private applyProductSuggestions(product: ProductLookup): void {
    const controls = this.productForm.controls;

    if (product.nombre && (controls.nombre.pristine || !controls.nombre.value.trim())) {
      controls.nombre.setValue(product.nombre);
    }
    if (product.descripcion && (controls.descripcion.pristine || !controls.descripcion.value.trim())) {
      controls.descripcion.setValue(product.descripcion);
    }
    if (product.precio_sugerido !== null && controls.precio_venta.pristine) {
      controls.precio_venta.setValue(product.precio_sugerido);
    }
    if (product.contenido !== null && controls.contenido.pristine) {
      controls.contenido.setValue(product.contenido);
    }
    if (product.unidad_contenido && (controls.unidad_contenido.pristine || !controls.unidad_contenido.value.trim())) {
      controls.unidad_contenido.setValue(product.unidad_contenido);
    }
    if (!controls.sku_interno.value.trim()) {
      controls.sku_interno.setValue(product.codigo_barras);
    }
  }

  private applyCatalogSuggestions(): void {
    const product = this.lastLookup;
    if (!product) return;

    const warnings: string[] = [];
    const brandControl = this.productForm.controls.marca_id;
    const categoryControl = this.productForm.controls.categoria_id;

    if (product.marca && this.brandsLoaded && (brandControl.pristine || !brandControl.value)) {
      const brand = this.findCatalogMatch(this.brands(), product.marca);
      if (brand) brandControl.setValue(brand.id);
      else warnings.push(`La marca "${product.marca}" no existe en el catálogo local.`);
    }
    if (product.categoria && this.categoriesLoaded && (categoryControl.pristine || !categoryControl.value)) {
      const category = this.findCatalogMatch(this.categories(), product.categoria);
      if (category) categoryControl.setValue(category.id);
      else warnings.push(`La categoría "${product.categoria}" no existe en el catálogo local.`);
    }

    this.catalogLookupWarnings.set(warnings);
  }

  private findCatalogMatch(options: CatalogOption[], suggestion: string): CatalogOption | undefined {
    const normalizedSuggestion = this.normalizeCatalogText(suggestion);
    return options.find((option) => this.normalizeCatalogText(option.nombre) === normalizedSuggestion);
  }

  private normalizeCatalogText(value: string): string {
    return value.trim().toLocaleLowerCase('es-MX').normalize('NFD').replace(/[\u0300-\u036f]/g, '');
  }

  handlePreviewImageError(): void {
    this.productForm.controls.imagen_url.setValue('');
    this.previewImageURL.set(null);
    this.imageLookupError.set('La imagen encontrada no está disponible.');
  }

  hasProductImageFailed(productId: string): boolean {
    return this.failedProductImages().has(productId);
  }

  hasVariants(product: ProductRow): boolean {
    return product.variantes.length > 0;
  }

  variantStockTotal(product: ProductRow): number {
    return product.variantes.reduce((total, variant) => total + variant.stock, 0);
  }

  variantAttributes(variant: ProductVariantRow): string {
    return variant.atributos.map((attribute) => `${attribute.nombre}: ${attribute.valor}`).join(' · ');
  }

  handleProductImageError(productId: string): void {
    this.failedProductImages.update((current) => new Set([...current, productId]));
  }

  closeCreateDialog(): void {
    if (!this.saving()) this.navigate(['/catalogo/productos']);
  }

  private navigate(commands: string[]): void {
    void this.router?.navigate(commands);
  }

  hasError(field: string): boolean {
    const control = this.productForm.get(field);
    return !!control && control.invalid && (control.dirty || control.touched);
  }

  getStatusClass(status: string | null | undefined): string {
    return status?.trim().toLowerCase().replaceAll(' ', '-') ?? 'sin-estado';
  }

  saveProduct(): void {
    if (!this.canSubmitProduct()) {
      return;
    }

    if (this.productForm.invalid) {
      this.productForm.markAllAsTouched();
      return;
    }
    this.saving.set(true);
    this.formError = null;
    const value = this.productForm.getRawValue();
    const payload = {
      ...value,
      marca_id: value.marca_id || null,
      categoria_id: value.categoria_id || null,
      descripcion: value.descripcion || null,
      contenido: value.contenido,
      unidad_contenido: value.unidad_contenido || null,
      unidad_medida_id: value.unidad_medida_id || null,
      presentacion: value.presentacion || null,
      codigo_barras: value.codigo_barras.trim() || null,
      imagen_url: value.imagen_url || null,
	  variantes: value.tiene_variantes ? value.variantes.map((variant) => ({
		atributos: this.parseVariantAttributes(variant.atributos),
		sku_interno: variant.sku_interno.trim(),
		generar_sku_interno: variant.generar_sku_interno,
		precio_venta: variant.precio_venta,
		stock_inicial: variant.stock_inicial,
		codigo_barras: variant.codigo_barras.trim() || null,
	  })) : undefined,
    };
    const editing = this.editingProduct();
    const request = editing
      ? this.productosService.update(this.negocioID, editing.id, {
          nombre: payload.nombre,
          sku_interno: payload.sku_interno,
          marca_id: payload.marca_id,
          categoria_id: payload.categoria_id as string,
          descripcion: payload.descripcion,
          contenido: payload.contenido,
          unidad_contenido: payload.unidad_contenido,
          unidad_medida_id: payload.unidad_medida_id,
          presentacion: payload.presentacion,
          precio_venta: payload.precio_venta,
          codigo_barras: payload.codigo_barras,
          imagen_url: payload.imagen_url,
        })
      : this.productosService.create(this.negocioID, payload);
    request.subscribe({
      next: () => {
        this.saving.set(false);
        this.navigate(['/catalogo/productos']);
      },
      error: (response) => {
        this.saving.set(false);
        this.formError = response.error?.mensaje ?? 'No fue posible guardar el producto.';
      },
    });
  }

  private parseVariantAttributes(value: string): { nombre: string; valor: string }[] {
	return value.split(';').map((item) => {
		const [nombre, ...valueParts] = item.split('=');
		return { nombre: nombre.trim(), valor: valueParts.join('=').trim() };
	}).filter((attribute) => attribute.nombre && attribute.valor);
	}
}

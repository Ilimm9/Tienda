import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import {
  FormArray,
  FormBuilder,
  FormControl,
  FormGroup,
  FormsModule,
  ReactiveFormsModule,
  Validators,
} from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { ButtonModule } from 'primeng/button';
import { CheckboxModule } from 'primeng/checkbox';
import { InputTextModule } from 'primeng/inputtext';
import { SelectModule } from 'primeng/select';
import { TableModule } from 'primeng/table';
import { TextareaModule } from 'primeng/textarea';

import { ContextoService } from '../../contexto/contexto.service';
import {
  CatalogOption,
  CreateUnitRequest,
  ProductImportPreview,
  ProductImportResult,
  ProductLookup,
  ProductRow,
  ProductVariantRow,
  UnitOption,
} from './product.models';
import { ProductosService } from './productos.service';
import Swal from 'sweetalert2/dist/sweetalert2.esm.js';
import { forkJoin, of, switchMap } from 'rxjs';
import { PERMISOS } from '../../contexto/permisos';
import { PuedeDirective } from '../../contexto/puede.directive';

type VariantFormGroup = FormGroup<{
  id: FormControl<string>;
  atributos: FormControl<string>;
  sku_interno: FormControl<string>;
  generar_sku_interno: FormControl<boolean>;
  precio_venta: FormControl<number>;
  stock_inicial: FormControl<number>;
  codigo_barras: FormControl<string>;
}>;

interface VariantDimension {
  nombre: string;
  valores: string[];
  nuevoValor: string;
}

@Component({
  selector: 'app-productos',
  standalone: true,
  imports: [
    PuedeDirective,
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
  readonly P = PERMISOS;
  private readonly productosService = inject(ProductosService);
  private readonly contexto = inject(ContextoService);
  private readonly formBuilder = inject(FormBuilder);
  private readonly route = inject(ActivatedRoute, { optional: true });
  private readonly router = inject(Router, { optional: true });

  readonly pageMode =
    (this.route?.snapshot.data['mode'] as 'list' | 'create' | 'edit' | 'import' | undefined) ??
    'list';

  readonly products = signal<ProductRow[]>([]);
  readonly productSearch = signal('');
  readonly productPageFirst = signal(0);
  readonly productCategoryFilter = signal<string | null>(null);
  readonly productBrandFilter = signal<string | null>(null);
  readonly productProviderFilter = signal<string | null>(null);
  readonly productStatusFilter = signal<string | null>(null);
  readonly productTypeFilter = signal<'all' | 'simple' | 'variants'>('all');
  readonly productCategoryOptions = computed(() =>
    this.filterOptions(this.products(), 'categoria'),
  );
  readonly productBrandOptions = computed(() => this.filterOptions(this.products(), 'marca'));
  readonly productProviderOptions = computed(() => [
    { id: '__sin_proveedor__', nombre: 'Sin proveedor' },
    ...this.providers(),
  ]);
  readonly productStatusOptions = [
    { id: 'En stock', nombre: 'En stock' },
    { id: 'Bajo stock', nombre: 'Bajo stock' },
    { id: 'Agotado', nombre: 'Agotado' },
  ];
  readonly productTypeOptions = [
    { id: 'all', nombre: 'Todos' },
    { id: 'simple', nombre: 'Productos simples' },
    { id: 'variants', nombre: 'Con variantes' },
  ];
  readonly filteredProducts = computed(() =>
    this.products().filter((product) => this.matchesProductFilters(product)),
  );
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly categories = signal<CatalogOption[]>([]);
  readonly brands = signal<CatalogOption[]>([]);
  readonly categoryOptions = computed(() => [...this.categories(), { id: '__new_category__', nombre: 'Crear nueva…' }]);
  readonly brandOptions = computed(() => [
    ...this.brands(),
    { id: '__new_brand__', nombre: 'Crear nueva…' },
  ]);
  readonly providers = signal<CatalogOption[]>([]);
  readonly branches = signal<CatalogOption[]>([]);
  readonly units = signal<UnitOption[]>([]);
  readonly unitOptions = computed(() => [...this.units(), { id: '__new_unit__', nombre: 'Crear nueva…', simbolo: '' }]);
  readonly catalogLoadErrors = signal<string[]>([]);
  readonly branchReady = signal(false);
  readonly saving = signal(false);
  readonly brandCreating = signal(false);
  readonly brandCreationVisible = signal(false);
  readonly brandCreationError = signal<string | null>(null);
  readonly categoryCreating = signal(false);
  readonly categoryCreationVisible = signal(false);
  readonly categoryCreationError = signal<string | null>(null);
  readonly unitCreating = signal(false);
  readonly unitCreationVisible = signal(false);
  readonly unitCreationError = signal<string | null>(null);
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
  readonly variantDimensions = signal<VariantDimension[]>([]);
  readonly newDimensionName = signal('');
  readonly deletingProductId = signal<string | null>(null);
  readonly stockEditing = signal(false);
  private lastLookup: ProductLookup | null = null;
  private categoriesLoaded = false;
  private brandsLoaded = false;
  formError: string | null = null;
  importFile: File | null = null;
  importDropActive = false;
  importBranchId = '';

  readonly productForm = this.formBuilder.nonNullable.group({
    nombre: ['', [Validators.required, Validators.maxLength(255)]],
    sku_interno: ['', Validators.maxLength(120)],
    generar_sku_interno: [false],
    tiene_variantes: [false],
    variantes: this.formBuilder.array<VariantFormGroup>([]),
    codigo_barras: ['', Validators.pattern(/^\d{8,14}$/)],
    imagen_url: [''],
    marca_id: [''],
    proveedor_id: [''],
    categoria_id: [''],
    sucursal_id: ['', Validators.required],
    descripcion: ['', Validators.maxLength(2000)],
    contenido: this.formBuilder.control<number | null>(null, [Validators.min(0)]),
    unidad_contenido: ['', Validators.maxLength(30)],
    unidad_medida_id: [''],
    presentacion: ['', Validators.maxLength(100)],
    precio_venta: [0, [Validators.required, Validators.min(0)]],
    costo_proveedor: [0, [Validators.required, Validators.min(0)]],
    stock_inicial: [0, [Validators.required, Validators.min(0)]],
  });
  readonly brandForm = this.formBuilder.nonNullable.group({
    nombre: ['', [Validators.required, Validators.maxLength(180)]],
  });
  readonly categoryForm = this.formBuilder.nonNullable.group({ nombre: ['', [Validators.required, Validators.maxLength(180)]] });
  readonly unitForm = this.formBuilder.nonNullable.group({
    nombre: ['', [Validators.required, Validators.maxLength(100)]],
    simbolo: ['', [Validators.required, Validators.maxLength(20)]],
    tipo: ['VOLUMEN', Validators.required],
  });

  constructor() {
    this.productForm.controls.precio_venta.valueChanges.subscribe(() => {
      if (this.productForm.controls.tiene_variantes.value) this.applyBasePrice();
    });
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
  private get sucursalID(): string { return this.contexto.sucursal()?.id ?? ''; }

  loadProducts(): void {
    this.loading.set(true);
    this.error.set(null);
    this.productosService.listByBusiness(this.negocioID, this.sucursalID).subscribe({
      next: ({ items }) => {
        this.products.set(items ?? []);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('No fue posible cargar los productos.');
        this.loading.set(false);
      },
    });
    this.productosService.listProviders(this.negocioID).subscribe({
      next: (items) => this.providers.set(items),
      error: () => this.providers.set([]),
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
    this.productosService.listByBusiness(this.negocioID, this.sucursalID).subscribe({
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
      nombre: '',
      sku_interno: '',
      generar_sku_interno: true,
      tiene_variantes: false,
      marca_id: '',
      proveedor_id: '',
      categoria_id: '',
      sucursal_id: '',
      codigo_barras: '',
      imagen_url: '',
      descripcion: '',
      contenido: null,
      unidad_contenido: '',
      presentacion: '',
      unidad_medida_id: '',
      precio_venta: 0,
      costo_proveedor: 0,
      stock_inicial: 0,
    });
    this.productForm.controls.sucursal_id.enable();
    this.productForm.controls.stock_inicial.enable();
    this.productForm.controls.sku_interno.disable();
    this.onGenerateSKUChanged();
    this.productForm.controls.precio_venta.enable();
    this.productForm.controls.stock_inicial.enable();
    this.productForm.controls.variantes.clear();
    this.variantDimensions.set([]);
    this.editingProduct.set(null);
    this.stockEditing.set(false);
    this.categories.set([]);
    this.brands.set([]);
    this.providers.set([]);
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
    this.resetBrandCreation();
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
    this.productosService.listProviders(this.negocioID).subscribe({
      next: (items) => this.providers.set(items),
      error: () => this.addCatalogLoadError('No fue posible cargar los proveedores.'),
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
    this.resetBrandCreation();
    this.editingProduct.set(product);
    this.stockEditing.set(false);
    this.branchReady.set(true);
    const branchID = this.contexto.sucursal()?.id;
    const currentStock = product.inventario?.find((item) => item.sucursal_id === branchID)?.stock ?? product.stock;
    this.productForm.reset({
      nombre: product.nombre,
      sku_interno: product.sku ?? '',
      generar_sku_interno: false,
      tiene_variantes: product.variantes.length > 0,
      codigo_barras: product.codigo_barras ?? '',
      imagen_url: product.imagen_url ?? '',
      marca_id: product.marca_id ?? '',
      proveedor_id: product.proveedor_id ?? '',
      categoria_id: product.categoria_id ?? '',
      sucursal_id: '',
      descripcion: product.descripcion ?? '',
      contenido: product.contenido,
      unidad_contenido: product.unidad_contenido ?? '',
      unidad_medida_id: product.unidad_medida_id ?? '',
      presentacion: product.presentacion ?? '',
      precio_venta: product.precio,
      stock_inicial: currentStock,
    });
    this.productForm.controls.sucursal_id.disable();
    this.productForm.controls.stock_inicial.disable();
    this.productForm.controls.sku_interno.disable();
    this.productForm.controls.variantes.clear();
    this.variantDimensions.set([]);
    if (product.variantes.length > 0) {
      const dimensions = new Map<string, string[]>();
      product.variantes.forEach((variant) => variant.atributos.forEach((attribute) => {
        const values = dimensions.get(attribute.nombre) ?? [];
        if (!values.includes(attribute.valor)) values.push(attribute.valor);
        dimensions.set(attribute.nombre, values);
      }));
      this.variantDimensions.set([...dimensions.entries()].map(([nombre, valores]) => ({ nombre, valores, nuevoValor: '' })));
      product.variantes.forEach((variant) => {
        const control = this.createVariantForm();
        const variantStock = variant.inventario?.find((item) => item.sucursal_id === branchID)?.stock ?? variant.stock;
        control.patchValue({ id: variant.id, atributos: this.serializeVariantAttributes(variant.atributos), sku_interno: variant.sku ?? '', precio_venta: variant.precio, stock_inicial: variantStock, codigo_barras: variant.codigo_barras ?? '', generar_sku_interno: !variant.sku });
        control.controls.stock_inicial.disable();
        this.variantControls.push(control);
      });
      this.onVariantsChanged();
    }
    this.productForm.markAsPristine();
    this.productForm.markAsUntouched();
    this.productosService.listCategories(this.negocioID).subscribe({
      next: (items) => this.categories.set(items),
      error: () => this.addCatalogLoadError('No fue posible cargar las categorías.'),
    });
    this.productosService.listBrands(this.negocioID).subscribe({
      next: (items) => this.brands.set(items),
      error: () => this.addCatalogLoadError('No fue posible cargar las marcas.'),
    });
    this.productosService.listProviders(this.negocioID).subscribe({
      next: (items) => this.providers.set(items),
      error: () => this.addCatalogLoadError('No fue posible cargar los proveedores.'),
    });
    this.productosService.listUnits(this.negocioID).subscribe({
      next: (items) => this.units.set(items),
      error: () => this.addCatalogLoadError('No fue posible cargar las unidades de medida.'),
    });
  }

  enableStockEditing(): void {
    if (!this.editingProduct()) return;
    if (!this.contexto.puede(PERMISOS.inventarioAjustar)) {
      void Swal.fire({
        icon: 'error',
        title: 'Sin permiso para ajustar inventario',
        text: 'Necesitas el permiso inventario.ajustar para modificar el stock.',
      });
      return;
    }
    Swal.fire({
      icon: 'warning',
      title: 'Ajustar existencias',
      text: 'Este cambio modificará el inventario de la sucursal activa y quedará registrado. ¿Deseas continuar?',
      showCancelButton: true,
      confirmButtonText: 'Sí, editar stock',
      cancelButtonText: 'Cancelar',
      reverseButtons: true,
    }).then((result) => {
      if (!result.isConfirmed) return;
      this.stockEditing.set(true);
      this.productForm.controls.stock_inicial.enable();
      this.syncVariantStockEditing();
    });
  }

  confirmDeactivate(product: ProductRow): void {
    if (
      this.deletingProductId() ||
      !window.confirm(
        `¿Deseas desactivar "${product.nombre}"? Podrás conservar su historial e inventario.`,
      )
    )
      return;
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
        this.importBranchId =
          items.find((item) => item.id === this.contexto.sucursal()?.id)?.id ?? items[0]?.id ?? '';
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
    return (
      event.currentTarget instanceof HTMLElement &&
      event.relatedTarget instanceof Node &&
      event.currentTarget.contains(event.relatedTarget)
    );
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
        this.importPreview.set(this.normalizeImportPreview(preview));
        this.importing.set(false);
      },
      error: (response) => {
        this.importing.set(false);
        this.importError.set(response.error?.mensaje ?? 'No fue posible validar el archivo.');
      },
    });
  }

  importExistingProducts(): void {
    if (
      !this.importFile ||
      !this.importBranchId ||
      !this.importPreview()?.insertables ||
      this.importing()
    )
      return;
    this.importing.set(true);
    this.importError.set(null);
    this.executeProductImport(
      this.importFile,
      this.importBranchId,
      (this.importPreview()!.catalogos_nuevos ?? []).length > 0,
    );
  }

  private executeProductImport(file: File, branchId: string, authorizeNewCatalogs: boolean): void {
    this.productosService.importProducts(this.negocioID, branchId, file, authorizeNewCatalogs).subscribe({
      next: (job) => {
        Swal.fire({
          title: 'Importando archivo',
          html: '<p id="import-progress-stage">Preparando importación</p><progress id="import-progress" value="0" max="100" style="width:100%"></progress><p id="import-progress-value">0%</p>',
          allowOutsideClick: false,
          allowEscapeKey: false,
          showConfirmButton: false,
        });
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
        if (job.estado === 'pendiente' || job.estado === 'procesando') {
          window.setTimeout(() => this.pollProductImport(importId), 750);
          return;
        }
        this.importing.set(false);
        this.importPreview.set(null);
        if (job.estado === 'completada' && job.resultado) {
          const result = this.normalizeImportResult(job.resultado);
          this.importResult.set(result);
          if (result.creadas > 0) this.loadProducts();
          const partial =
            result.omitidas > 0 || result.invalidas > 0 || result.advertencias.length > 0;
          Swal.fire({
            icon: partial ? 'warning' : 'success',
            title: 'Importación terminada',
            text: `${result.creadas} filas importadas correctamente · ${result.productos_simples_creados ?? 0} productos simples · ${result.productos_base_creados} familias nuevas · ${result.productos_base_reutilizados} familias reutilizadas · ${result.variantes_creadas} variantes · ${result.omitidas} omitidas · ${result.invalidas} inválidas`,
            confirmButtonText: 'Ver registros',
          }).then(() => {
            this.navigate(['/catalogo/productos']);
          });
          return;
        }
        this.importError.set(job.mensaje_error ?? 'No fue posible importar el archivo.');
        Swal.fire({
          icon: 'error',
          title: 'No fue posible importar',
          text: job.mensaje_error ?? 'Ocurrió un error durante la importación.',
        });
      },
      error: () => {
        this.importing.set(false);
        this.importError.set('No fue posible consultar el progreso de importación.');
        Swal.close();
      },
    });
  }

  private normalizeImportResult(result: ProductImportResult): ProductImportResult {
    return {
      ...result,
      errores: result.errores ?? [],
      advertencias: result.advertencias ?? [],
    };
  }

  private normalizeImportPreview(preview: ProductImportPreview): ProductImportPreview {
    return {
      ...preview,
      errores: preview.errores ?? [],
      advertencias: preview.advertencias ?? [],
      catalogos_nuevos: preview.catalogos_nuevos ?? [],
    };
  }

  productImportTemplateUrl(): string {
    return this.productosService.productImportTemplateUrl(this.negocioID);
  }

  onBranchSelected(): void {
    this.branchReady.set(!!this.productForm.controls.sucursal_id.value);
  }

  onUnitContentSelected(unitID: string): void {
    if (unitID === '__new_unit__') {
      this.productForm.controls.unidad_medida_id.setValue('');
      this.unitCreationVisible.set(true);
      this.unitCreationError.set(null);
      return;
    }
    const unit = this.units().find((item) => item.id === unitID);
    this.productForm.controls.unidad_contenido.setValue(unit?.simbolo ?? '');
  }

  onCategorySelected(categoryID: string): void {
    if (categoryID === '__new_category__') {
      this.productForm.controls.categoria_id.setValue('');
      this.categoryCreationVisible.set(true);
      this.categoryCreationError.set(null);
    }
  }

  createCategory(): void {
    this.categoryForm.markAllAsTouched();
    if (this.categoryForm.invalid || this.categoryCreating()) return;
    const nombre = this.categoryForm.controls.nombre.value.trim();
    this.categoryCreating.set(true);
    this.productosService.createCategory(this.negocioID, nombre).subscribe({
      next: () => this.productosService.listCategories(this.negocioID).subscribe({
        next: (items) => { this.categories.set(items); this.categoryCreating.set(false); const created = items.find((item) => this.normalizeCatalogText(item.nombre) === this.normalizeCatalogText(nombre)); if (!created) { this.categoryCreationError.set('La categoría fue creada, pero no fue posible seleccionarla.'); return; } this.productForm.controls.categoria_id.setValue(created.id); this.categoryCreationVisible.set(false); this.categoryForm.reset(); },
        error: () => { this.categoryCreating.set(false); this.categoryCreationError.set('La categoría fue creada, pero no fue posible recargarla.'); },
      }),
      error: (response) => { this.categoryCreating.set(false); this.categoryCreationError.set(response.error?.mensaje ?? 'No fue posible crear la categoría.'); },
    });
  }

  createUnit(): void {
    this.unitForm.markAllAsTouched();
    if (this.unitForm.invalid || this.unitCreating()) return;
    const value = this.unitForm.getRawValue();
    const simbolo = value.simbolo.trim();
    const payload: CreateUnitRequest = { ...value, nombre: value.nombre.trim(), simbolo, factor_a_base: 1, permite_fraccion: true, decimales: 2 };
    this.unitCreating.set(true);
    this.productosService.createUnit(this.negocioID, payload).subscribe({
      next: () => this.productosService.listUnits(this.negocioID).subscribe({
        next: (items) => { this.units.set(items); this.unitCreating.set(false); const created = items.find((item) => this.normalizeCatalogText(item.simbolo) === this.normalizeCatalogText(payload.simbolo)); if (!created) { this.unitCreationError.set('La unidad fue creada, pero no fue posible seleccionarla.'); return; } this.productForm.controls.unidad_medida_id.setValue(created.id); this.productForm.controls.unidad_contenido.setValue(created.simbolo); this.unitCreationVisible.set(false); this.unitForm.reset({ nombre: '', simbolo: '', tipo: 'VOLUMEN' }); },
        error: () => { this.unitCreating.set(false); this.unitCreationError.set('La unidad fue creada, pero no fue posible recargarla.'); },
      }),
      error: (response) => { this.unitCreating.set(false); this.unitCreationError.set(response.error?.mensaje ?? 'No fue posible crear la unidad.'); },
    });
  }

  onGenerateSKUChanged(): void {
    const controls = this.productForm.controls;
    if (controls.generar_sku_interno.value) {
      controls.sku_interno.setValue('');
      controls.sku_interno.disable();
      return;
    }
    if (this.editingProduct()) controls.sku_interno.disable();
    else controls.sku_interno.enable();
  }

  get variantControls(): FormArray<VariantFormGroup> {
    return this.productForm.controls.variantes;
  }

  onVariantsChanged(): void {
    const controls = this.productForm.controls;
    if (controls.tiene_variantes.value) {
      controls.sku_interno.disable();
      controls.stock_inicial.disable();
      controls.codigo_barras.setValue('');
      const editing = this.editingProduct();
      if (editing && editing.variantes.length === 0 && this.variantControls.length === 0) {
        this.variantDimensions.set([]);
      }
      this.syncVariantStockEditing();
      return;
    }
    if (this.editingProduct()) controls.sku_interno.disable();
    else controls.sku_interno.enable();
    controls.stock_inicial.enable();
    this.variantControls.clear();
    this.variantDimensions.set([]);
  }

  addVariant(): void {
    this.variantControls.push(this.createVariantForm());
  }

  addDimension(): void {
    const nombre = this.newDimensionName().trim();
    if (!nombre || this.variantDimensions().some((dimension) => dimension.nombre.toLocaleLowerCase() === nombre.toLocaleLowerCase())) return;
    this.variantDimensions.update((items) => [...items, { nombre, valores: [], nuevoValor: '' }]);
    this.newDimensionName.set('');
  }

  removeDimension(index: number): void {
    this.variantDimensions.update((items) => items.filter((_, itemIndex) => itemIndex !== index));
  }

  setDimensionValue(index: number, value: string): void {
    this.variantDimensions.update((items) => items.map((item, itemIndex) => itemIndex === index ? { ...item, nuevoValor: value } : item));
  }

  addDimensionValue(index: number): void {
    this.variantDimensions.update((items) => items.map((item, itemIndex) => {
      if (itemIndex !== index) return item;
      const value = item.nuevoValor.trim();
      return !value || item.valores.some((current) => current.toLocaleLowerCase() === value.toLocaleLowerCase())
        ? item
        : { ...item, valores: [...item.valores, value], nuevoValor: '' };
    }));
  }

  removeDimensionValue(dimensionIndex: number, valueIndex: number): void {
    this.variantDimensions.update((items) => items.map((item, itemIndex) => itemIndex === dimensionIndex ? { ...item, valores: item.valores.filter((_, index) => index !== valueIndex) } : item));
  }

  canGenerateCombinations(): boolean {
    return this.variantDimensions().length > 0 && this.variantDimensions().every((dimension) => dimension.valores.length > 0);
  }

  generateCombinations(): void {
    if (!this.canGenerateCombinations()) return;
    const dimensions = this.variantDimensions();
    const combinations = dimensions.reduce<{ nombre: string; valor: string }[][]>((rows, dimension) => rows.flatMap((row) => dimension.valores.map((valor) => [...row, { nombre: dimension.nombre, valor }])), [[]]);
    const existing = new Set(this.variantControls.controls.map((control) => control.controls.atributos.value));
    for (const attributes of combinations) {
      const serialized = this.serializeVariantAttributes(attributes);
      if (!existing.has(serialized)) {
        const control = this.createVariantForm();
        control.controls.atributos.setValue(serialized);
        control.controls.precio_venta.setValue(this.productForm.controls.precio_venta.value);
        if (this.editingProduct()) this.syncVariantControlStockEditing(control);
        this.variantControls.push(control);
      }
    }
  }

  applyBasePrice(): void {
    const price = this.productForm.controls.precio_venta.value;
    this.variantControls.controls.forEach((control) => control.controls.precio_venta.setValue(price));
  }

  variantAttributeValues(index: number): { nombre: string; valor: string }[] {
    return this.parseVariantAttributes(this.variantControls.at(index).controls.atributos.value);
  }

  private serializeVariantAttributes(attributes: { nombre: string; valor: string }[]): string {
    return attributes.map((attribute) => `${attribute.nombre}=${attribute.valor}`).join(';');
  }

  private createVariantForm(): VariantFormGroup {
    return this.formBuilder.nonNullable.group({
      id: [''],
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
    this.catalogLoadErrors.update((messages) =>
      messages.includes(message) ? messages : [...messages, message],
    );
  }

  canSubmitProduct(): boolean {
    if (this.saving()) return false;
    if (this.productForm.controls.tiene_variantes.value && !this.hasValidVariants()) return false;
    if (this.editingProduct()) return this.productForm.dirty;
    return this.branchReady();
  }

  canLookupProduct(): boolean {
    return (
      /^\d{8,14}$/.test(this.productForm.controls.codigo_barras.value.trim()) &&
      !this.imageLookupLoading()
    );
  }

  onBarcodeEnter(event: Event): void {
    event.preventDefault();
    if (this.canLookupProduct()) this.lookupProduct();
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
          this.imageLookupError.set(
            'El producto fue encontrado, pero no tiene una imagen disponible.',
          );
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
        this.imageLookupError.set(
          response.error?.mensaje ?? 'No fue posible buscar el producto en PrecioCheck.',
        );
      },
    });
  }

  private applyProductSuggestions(product: ProductLookup): void {
    const controls = this.productForm.controls;

    if (product.nombre && (controls.nombre.pristine || !controls.nombre.value.trim())) {
      controls.nombre.setValue(product.nombre);
    }
    if (
      product.descripcion &&
      (controls.descripcion.pristine || !controls.descripcion.value.trim())
    ) {
      controls.descripcion.setValue(product.descripcion);
    }
    if (product.precio_sugerido !== null && controls.precio_venta.pristine) {
      controls.precio_venta.setValue(product.precio_sugerido);
    }
    if (product.contenido !== null && controls.contenido.pristine) {
      controls.contenido.setValue(product.contenido);
    }
    if (
      product.unidad_contenido &&
      (controls.unidad_contenido.pristine || !controls.unidad_contenido.value.trim())
    ) {
      controls.unidad_contenido.setValue(product.unidad_contenido);
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
    if (
      product.categoria &&
      this.categoriesLoaded &&
      (categoryControl.pristine || !categoryControl.value)
    ) {
      const category = this.findCatalogMatch(this.categories(), product.categoria);
      if (category) categoryControl.setValue(category.id);
      else warnings.push(`La categoría "${product.categoria}" no existe en el catálogo local.`);
    }

    this.catalogLookupWarnings.set(warnings);
  }

  private findCatalogMatch(
    options: CatalogOption[],
    suggestion: string,
  ): CatalogOption | undefined {
    const normalizedSuggestion = this.normalizeCatalogText(suggestion);
    return options.find(
      (option) => this.normalizeCatalogText(option.nombre) === normalizedSuggestion,
    );
  }

  private normalizeCatalogText(value: string): string {
    return value
      .trim()
      .toLocaleLowerCase('es-MX')
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '');
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
    return variant.atributos
      .map((attribute) => `${attribute.nombre}: ${attribute.valor}`)
      .join(' · ');
  }

  setProductSearch(value: string): void {
    this.productSearch.set(value);
    this.resetProductPagination();
  }

  setProductCategoryFilter(value: string | null): void {
    this.productCategoryFilter.set(value);
    this.resetProductPagination();
  }

  setProductBrandFilter(value: string | null): void {
    this.productBrandFilter.set(value);
    this.resetProductPagination();
  }
  setProductProviderFilter(value: string | null): void {
    this.productProviderFilter.set(value);
    this.resetProductPagination();
  }

  setProductStatusFilter(value: string | null): void {
    this.productStatusFilter.set(value);
    this.resetProductPagination();
  }

  setProductTypeFilter(value: 'all' | 'simple' | 'variants' | null): void {
    this.productTypeFilter.set(value ?? 'all');
    this.resetProductPagination();
  }

  clearProductFilters(): void {
    this.productSearch.set('');
    this.productCategoryFilter.set(null);
    this.productBrandFilter.set(null);
    this.productProviderFilter.set(null);
    this.productStatusFilter.set(null);
    this.productTypeFilter.set('all');
    this.resetProductPagination();
  }

  onProductTablePage(first: number): void {
    this.productPageFirst.set(first);
  }

  private resetProductPagination(): void {
    this.productPageFirst.set(0);
  }

  private filterOptions(products: ProductRow[], field: 'categoria' | 'marca'): CatalogOption[] {
    const values = new Map<string, CatalogOption>();
    for (const product of products) {
      const value = product[field];
      const id = field === 'categoria' ? product.categoria_id : product.marca_id;
      if (value && id) values.set(id, { id, nombre: value });
    }
    return [...values.values()].sort((first, second) =>
      first.nombre.localeCompare(second.nombre, 'es'),
    );
  }

  private matchesProductFilters(product: ProductRow): boolean {
    const query = this.normalizeProductFilter(this.productSearch());
    const category = this.productCategoryFilter();
    const brand = this.productBrandFilter();
    const provider = this.productProviderFilter();
    const status = this.productStatusFilter();
    const type = this.productTypeFilter();
    if (category && product.categoria_id !== category) return false;
    if (brand && product.marca_id !== brand) return false;
    if (provider === '__sin_proveedor__' && product.proveedor_id) return false;
    if (provider && provider !== '__sin_proveedor__' && product.proveedor_id !== provider)
      return false;
    if (status && product.estado !== status) return false;
    if (type === 'simple' && this.hasVariants(product)) return false;
    if (type === 'variants' && !this.hasVariants(product)) return false;
    if (!query) return true;
    const searchable = [
      product.nombre,
      product.presentacion,
      product.descripcion,
      product.marca,
      product.categoria,
      product.sku,
      product.codigo_barras,
      ...product.variantes.flatMap((variant) => [
        variant.sku,
        variant.codigo_barras,
        ...variant.atributos.flatMap((attribute) => [
          attribute.nombre,
          attribute.valor,
          `${attribute.nombre}=${attribute.valor}`,
        ]),
      ]),
    ];
    return searchable.some((value) => this.normalizeProductFilter(value).includes(query));
  }

  private normalizeProductFilter(value: string | null | undefined): string {
    return (value ?? '')
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .trim()
      .toLocaleLowerCase('es');
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

  onBrandSelected(value: string): void {
    if (value === '__new_brand__') {
      this.productForm.controls.marca_id.setValue('');
      this.brandCreationVisible.set(true);
      this.brandForm.reset();
      this.brandCreationError.set(null);
      return;
    }
    this.brandCreationVisible.set(false);
  }

  cancelBrandCreation(): void {
    if (this.brandCreating()) return;
    this.brandCreationVisible.set(false);
    this.brandForm.reset();
    this.brandCreationError.set(null);
  }

  createBrand(): void {
    this.brandForm.markAllAsTouched();
    if (this.brandForm.invalid || this.brandCreating()) return;

    const nombre = this.brandForm.controls.nombre.value.trim();
    this.brandCreating.set(true);
    this.brandCreationError.set(null);
    this.productosService.createBrand(this.negocioID, nombre).subscribe({
      next: () =>
        this.productosService.listBrands(this.negocioID).subscribe({
          next: (items) => {
            this.brands.set(items);
            const created = items.find(
              (brand) => brand.nombre.trim().toLocaleLowerCase() === nombre.toLocaleLowerCase(),
            );
            this.brandCreating.set(false);
            if (!created) {
              this.brandCreationError.set(
                'La marca fue creada, pero no fue posible encontrarla al recargar el catálogo.',
              );
              return;
            }
            this.productForm.controls.marca_id.setValue(created.id);
            this.brandCreationVisible.set(false);
            this.brandForm.reset();
          },
          error: () => {
            this.brandCreating.set(false);
            this.brandCreationError.set(
              'La marca fue creada, pero no fue posible recargar el catálogo.',
            );
          },
        }),
      error: (response) => {
        this.brandCreating.set(false);
        this.brandCreationError.set(response.error?.mensaje ?? 'No fue posible crear la marca.');
      },
    });
  }

  private resetBrandCreation(): void {
    this.brandCreating.set(false);
    this.brandCreationVisible.set(false);
    this.brandCreationError.set(null);
    this.brandForm.reset();
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
    const editing = this.editingProduct();
    if (editing && editing.variantes.length === 0 && this.productForm.controls.tiene_variantes.value) {
      void Swal.fire({
        icon: 'warning',
        title: 'Convertir a producto con variantes',
        text: 'El stock, SKU y código de barras actuales no se copiarán. Deberás asignarlos manualmente a las nuevas variantes; el historial anterior se conservará.',
        showCancelButton: true,
        confirmButtonText: 'Convertir y guardar',
        cancelButtonText: 'Cancelar',
        reverseButtons: true,
      }).then((result) => {
        if (result.isConfirmed) this.persistProduct();
      });
      return;
    }
    this.persistProduct();
  }

  private persistProduct(): void {
    this.saving.set(true);
    this.formError = null;
    const value = this.productForm.getRawValue();
    const payload = {
      ...value,
      marca_id: value.marca_id || null,
      proveedor_id: value.proveedor_id || null,
      categoria_id: value.categoria_id || null,
      descripcion: value.descripcion || null,
      contenido: value.contenido,
      unidad_contenido: value.unidad_contenido || null,
      unidad_medida_id: value.unidad_medida_id || null,
      presentacion: value.presentacion || null,
      codigo_barras: value.codigo_barras.trim() || null,
      imagen_url: value.imagen_url || null,
      variantes: value.tiene_variantes
        ? value.variantes.map((variant) => ({
            id: variant.id || null,
            atributos: this.parseVariantAttributes(variant.atributos),
            sku_interno: variant.sku_interno.trim(),
            generar_sku_interno: variant.generar_sku_interno,
            precio_venta: variant.precio_venta,
            stock_inicial: variant.stock_inicial,
            codigo_barras: variant.codigo_barras.trim() || null,
          }))
        : undefined,
    };
    const editing = this.editingProduct();
    const request = editing
      ? this.productosService.update(this.negocioID, editing.id, {
          nombre: payload.nombre,
          marca_id: payload.marca_id,
          proveedor_id: payload.proveedor_id,
          categoria_id: payload.categoria_id,
          descripcion: payload.descripcion,
          contenido: payload.contenido,
          unidad_contenido: payload.unidad_contenido,
          unidad_medida_id: payload.unidad_medida_id,
          presentacion: payload.presentacion,
          precio_venta: payload.precio_venta,
          codigo_barras: payload.codigo_barras,
          imagen_url: payload.imagen_url,
          variantes: payload.tiene_variantes
            ? payload.variantes!.map((variant) => ({
                id: variant.id,
                atributos: variant.atributos,
                precio_venta: variant.precio_venta,
                codigo_barras: variant.codigo_barras,
              }))
            : [],
        })
      : this.productosService.create(this.negocioID, payload);
    const directStockRequests = this.stockEditing() && editing && !payload.tiene_variantes
      ? this.inventoryAdjustments(editing, payload)
      : [];
    const stockRequests$ = this.stockEditing() && editing
      ? payload.tiene_variantes
        ? this.inventoryAdjustmentsAfterSave(editing, payload)
        : (directStockRequests.length ? forkJoin(directStockRequests) : of([]))
      : of([]);
    request.pipe(
      switchMap(() => stockRequests$),
    ).subscribe({
      next: () => {
        this.saving.set(false);
        this.navigate(['/catalogo/productos']);
      },
      error: (response) => {
        this.saving.set(false);
        this.formError = this.productSaveError(response);
      },
    });
  }

  private inventoryAdjustments(editing: ProductRow | null, payload: { stock_inicial: number; tiene_variantes: boolean; variantes?: Array<{ id: string | null; stock_inicial: number }> }) {
    if (!editing) return [];
    const branchID = this.contexto.sucursal()?.id;
    if (!branchID) return [];
    const requests = [] as ReturnType<ProductosService['adjustInventory']>[];
    if (!payload.tiene_variantes) {
      const current = editing.inventario?.find((item) => item.sucursal_id === branchID)?.stock ?? editing.stock;
      if (payload.stock_inicial !== current) requests.push(this.productosService.adjustInventory(this.negocioID, editing.id, { sucursal_id: branchID, stock: payload.stock_inicial }));
    } else {
      payload.variantes?.forEach((variant, index) => {
        const current = editing.variantes[index]?.inventario?.find((item) => item.sucursal_id === branchID)?.stock ?? editing.variantes[index]?.stock ?? 0;
        if (variant.id && variant.stock_inicial !== current) requests.push(this.productosService.adjustInventory(this.negocioID, editing.id, { sucursal_id: branchID, variante_id: variant.id, stock: variant.stock_inicial }));
      });
    }
    return requests;
  }

  private inventoryAdjustmentsAfterSave(
    editing: ProductRow,
    payload: { tiene_variantes: boolean; variantes?: Array<{ id: string | null; atributos: { nombre: string; valor: string }[]; stock_inicial: number }> },
  ) {
    const branchID = this.contexto.sucursal()?.id;
    if (!branchID || !payload.tiene_variantes) return of([]);
    return this.productosService.listByBusiness(this.negocioID, branchID).pipe(
      switchMap(({ items }) => {
        const saved = (items ?? []).find((product) => product.id === editing.id);
        if (!saved) throw new Error('No fue posible recargar las variantes del producto.');
        const requests = (payload.variantes ?? []).flatMap((variant) => {
          const savedVariant = variant.id
            ? saved.variantes.find((item) => item.id === variant.id)
            : saved.variantes.find((item) => this.variantKey(item.atributos) === this.variantKey(variant.atributos));
          if (!savedVariant) return [];
          const current = savedVariant.inventario?.find((item) => item.sucursal_id === branchID)?.stock ?? (variant.id ? savedVariant.stock : 0);
          if (variant.stock_inicial === current) return [];
          return [this.productosService.adjustInventory(this.negocioID, saved.id, {
            sucursal_id: branchID,
            variante_id: savedVariant.id,
            stock: variant.stock_inicial,
          })];
        });
        return requests.length ? forkJoin(requests) : of([]);
      }),
    );
  }

  private variantKey(attributes: { nombre: string; valor: string }[]): string {
    return attributes
      .map((attribute) => `${this.normalizeCatalogText(attribute.nombre)}=${this.normalizeCatalogText(attribute.valor)}`)
      .sort()
      .join(';');
  }

  private hasValidVariants(): boolean {
    if (this.variantControls.length === 0) return false;
    const keys = new Set<string>();
    return this.variantControls.controls.every((control) => {
      const attributes = this.parseVariantAttributes(control.controls.atributos.value);
      if (attributes.length === 0) return false;
      const names = new Set(attributes.map((attribute) => this.normalizeCatalogText(attribute.nombre)));
      if (names.size !== attributes.length) return false;
      const key = this.variantKey(attributes);
      if (keys.has(key)) return false;
      keys.add(key);
      return true;
    });
  }

  private syncVariantStockEditing(): void {
    this.variantControls.controls.forEach((control) => this.syncVariantControlStockEditing(control));
  }

  private syncVariantControlStockEditing(control: VariantFormGroup): void {
    if (!this.editingProduct()) return;
    if (this.stockEditing()) control.controls.stock_inicial.enable();
    else control.controls.stock_inicial.disable();
  }

  private productSaveError(response: { error?: unknown; message?: string }): string {
    const error = response.error;
    if (error && typeof error === 'object' && 'mensaje' in error) {
      const message = (error as { mensaje?: unknown }).mensaje;
      if (typeof message === 'string' && message.trim()) return message;
    }
    if (typeof error === 'string' && error.trim()) return error;
    return response.message?.trim() || 'No fue posible guardar el producto.';
  }

  private parseVariantAttributes(value: string): { nombre: string; valor: string }[] {
    return value
      .split(';')
      .map((item) => {
        const [nombre, ...valueParts] = item.split('=');
        return { nombre: nombre.trim(), valor: valueParts.join('=').trim() };
      })
      .filter((attribute) => attribute.nombre && attribute.valor);
  }
}

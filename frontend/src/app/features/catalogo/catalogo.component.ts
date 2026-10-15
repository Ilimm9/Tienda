import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { FormBuilder, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { InputTextModule } from 'primeng/inputtext';
import { TableModule } from 'primeng/table';
import { TextareaModule } from 'primeng/textarea';
import { SelectModule } from 'primeng/select';
import { ContextoService } from '../../contexto/contexto.service';
import { CatalogoService } from './catalogo.service';
import { CatalogImportResult, CatalogRecord, Categoria, Proveedor, UnidadMedida } from './catalogo.models';
import { PERMISOS } from '../../contexto/permisos';
import { PuedeDirective } from '../../contexto/puede.directive';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import Swal from 'sweetalert2/dist/sweetalert2.esm.js';

interface SelectOption {
  label: string;
  value: string;
}

type CatalogStatusFilter = 'all' | 'active' | 'inactive';
const NO_PARENT_CATEGORY = '__sin_categoria_padre__';

@Component({
  selector: 'app-catalogo',
  standalone: true,
  imports: [PuedeDirective,
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    InputTextModule,
    TableModule,
    TextareaModule,
    SelectModule,
  ],
  templateUrl: './catalogo.component.html',
  styleUrl: './catalogo.component.css',
})
export class CatalogoComponent {
  readonly P = PERMISOS;
  private readonly service = inject(CatalogoService);
  private readonly fb = inject(FormBuilder);
  private readonly route = inject(ActivatedRoute);
  private readonly contexto = inject(ContextoService);
  private readonly router = inject(Router, { optional: true });
  private readonly feedback = inject(FeedbackService);
  readonly section = this.route.snapshot.data['section'] as 'marcas' | 'categorias' | 'proveedores' | 'unidades';
  readonly pageMode = (this.route.snapshot.data['mode'] as 'list' | 'create' | 'edit' | 'import' | undefined) ?? 'list';
  readonly items = signal<CatalogRecord[]>([]);
  readonly parents = signal<Categoria[]>([]);
  readonly catalogSearch = signal('');
  readonly catalogStatusFilter = signal<CatalogStatusFilter>('all');
  readonly categoryParentFilter = signal<string | null>(null);
  readonly unitTypeFilter = signal<string | null>(null);
  readonly catalogPageFirst = signal(0);
  readonly catalogStatusOptions: SelectOption[] = [
    { label: 'Todos', value: 'all' },
    { label: 'Activos', value: 'active' },
    { label: 'Inactivos', value: 'inactive' },
  ];
  readonly unitTypeOptions: SelectOption[] = [
    { label: 'Peso', value: 'PESO' },
    { label: 'Volumen', value: 'VOLUMEN' },
    { label: 'Longitud', value: 'LONGITUD' },
    { label: 'Área', value: 'AREA' },
    { label: 'Cantidad', value: 'CANTIDAD' },
    { label: 'Empaque', value: 'EMPAQUE' },
  ];
  readonly categoryParentFilterOptions = computed<SelectOption[]>(() => [
    { label: 'Sin categoría padre', value: NO_PARENT_CATEGORY },
    ...this.items()
      .filter((item): item is Categoria => 'categoria_padre_id' in item)
      .map((category) => ({ label: category.nombre, value: category.id })),
  ]);
  readonly filteredItems = computed(() =>
    this.items().filter((item) => this.matchesCatalogFilters(item)),
  );
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  editingId: string | null = null;
  readonly saving = signal(false);
  readonly deletingId = signal<string | null>(null);
  readonly formError = signal<string | null>(null);
  importFile: File | null = null;
  importDropActive = false;
  readonly importing = signal(false);
  readonly importError = signal<string | null>(null);
  readonly importResult = signal<CatalogImportResult | null>(null);
  readonly form = this.fb.nonNullable.group({
    nombre: ['', [Validators.required, Validators.maxLength(180)]],
    descripcion: [''],
    categoria_padre_id: [''],
    razon_social: [''],
    rfc: [''],
    telefono: [''],
    email: ['', Validators.email],
    direccion: [''],
    simbolo: [''], tipo: ['PESO'], factor_a_base: [1, [Validators.min(0.000001)]], decimales: [2, [Validators.min(0), Validators.max(6)]],
  });
  constructor() {
    if (this.pageMode === 'list') this.load();
    else if (this.pageMode === 'create') {
      this.prepareCreate();
      if (this.section === 'categorias') this.loadParentCategories();
    }
    else if (this.pageMode === 'import') this.prepareImport();
    else this.loadForEdit();
  }
  get title(): string {
    return this.section === 'marcas'
      ? 'Marcas'
      : this.section === 'categorias'
        ? 'Categorías'
        : this.section === 'proveedores' ? 'Proveedores' : 'Unidades de medida';
  }
  get singularTitle(): string {
    return this.section === 'marcas'
      ? 'Marca'
      : this.section === 'categorias'
        ? 'Categoría'
        : this.section === 'proveedores'
          ? 'Proveedor'
          : 'Unidad de medida';
  }
  get description(): string {
    return this.section === 'marcas'
      ? 'Administra las marcas de tus productos.'
      : this.section === 'categorias'
        ? 'Organiza los productos por categorías.'
        : this.section === 'proveedores' ? 'Administra tus proveedores de inventario.' : 'Administra las unidades normalizadas del catálogo.';
  }
  get parentCategoryOptions(): SelectOption[] {
    return this.parents()
      .filter((category) => category.id !== this.editingId)
      .map((category) => ({ label: category.nombre, value: category.id }));
  }
  load(): void {
    this.loading.set(true);
    const negocioId = this.contexto.negocio()?.id ?? '';
    const request =
      this.section === 'marcas'
        ? this.service.marcas(negocioId)
        : this.section === 'categorias'
          ? this.service.categorias(negocioId)
          : this.section === 'proveedores' ? this.service.proveedores(negocioId) : this.service.unidades(negocioId);
    request.subscribe({
      next: (v) => {
        this.items.set(v);
        this.loading.set(false);
        if (this.section === 'categorias') this.parents.set(v as Categoria[]);
      },
      error: () => {
        this.error.set(`No fue posible cargar ${this.title.toLowerCase()}.`);
        this.loading.set(false);
      },
    });
  }

  isListPage(): boolean {
    return this.pageMode === 'list';
  }

  setCatalogSearch(value: string): void {
    this.catalogSearch.set(value);
    this.resetCatalogPagination();
  }

  setCatalogStatusFilter(value: CatalogStatusFilter | null): void {
    this.catalogStatusFilter.set(value ?? 'all');
    this.resetCatalogPagination();
  }

  setCategoryParentFilter(value: string | null): void {
    this.categoryParentFilter.set(value);
    this.resetCatalogPagination();
  }

  setUnitTypeFilter(value: string | null): void {
    this.unitTypeFilter.set(value);
    this.resetCatalogPagination();
  }

  clearCatalogFilters(): void {
    this.catalogSearch.set('');
    this.catalogStatusFilter.set('all');
    this.categoryParentFilter.set(null);
    this.unitTypeFilter.set(null);
    this.resetCatalogPagination();
  }

  onCatalogTablePage(first: number): void {
    this.catalogPageFirst.set(first);
  }

  isFormPage(): boolean {
    return this.pageMode === 'create' || this.pageMode === 'edit';
  }

  isImportPage(): boolean {
    return this.pageMode === 'import';
  }

  private loadForEdit(): void {
    const itemId = this.route.snapshot.paramMap.get('id');
    if (!itemId) {
      this.goToList();
      return;
    }
    const negocioId = this.contexto.negocio()?.id ?? '';
    const request = this.section === 'marcas'
      ? this.service.marcas(negocioId)
      : this.section === 'categorias'
        ? this.service.categorias(negocioId)
        : this.section === 'proveedores'
          ? this.service.proveedores(negocioId)
          : this.service.unidades(negocioId);
    this.loading.set(true);
    request.subscribe({
      next: (items) => {
        this.items.set(items);
        if (this.section === 'categorias') this.parents.set(items as Categoria[]);
        this.loading.set(false);
        const item = items.find((candidate) => candidate.id === itemId);
        if (!item) {
          this.error.set(`No fue posible encontrar ${this.title.toLowerCase().slice(0, -1)}.`);
          return;
        }
        this.prepareEdit(item);
      },
      error: () => {
        this.error.set(`No fue posible cargar ${this.title.toLowerCase()}.`);
        this.loading.set(false);
      },
    });
  }

  openCreate(): void {
    this.prepareCreate();
    this.navigateToCreate();
  }

  private prepareCreate(): void {
    this.loading.set(false);
    this.editingId = null;
    this.form.reset();
    this.formError.set(null);
  }

  private loadParentCategories(): void {
    this.loading.set(true);
    this.service.categorias(this.contexto.negocio()?.id ?? '').subscribe({
      next: (categories) => {
        this.parents.set(categories);
        this.loading.set(false);
      },
      error: () => {
        this.parents.set([]);
        this.loading.set(false);
        this.formError.set('No fue posible cargar las categorías padre. Puedes crear la categoría sin asignarle una.');
      },
    });
  }

  openImport(): void {
    this.prepareImport();
    this.navigateToImport();
  }

  private prepareImport(): void {
    this.loading.set(false);
    this.importFile = null;
    this.importDropActive = false;
    this.importError.set(null);
    this.importResult.set(null);
  }
  onImportFile(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.setImportFile(input.files?.[0] ?? null);
    input.value = '';
  }
  openFilePicker(event: MouseEvent, input: HTMLInputElement): void {
    if (event.target === input || this.importing()) return;
    input.click();
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
  }
  templateUrl(): string {
    return this.service.plantillaUrl(this.contexto.negocio()?.id ?? '', this.section as 'marcas' | 'categorias' | 'proveedores' | 'unidades');
  }
  importCatalog(): void {
    if (!this.importFile || !this.canImport()) return;
    this.importing.set(true);
    this.importError.set(null);
    this.importResult.set(null);
    this.service.importar(this.contexto.negocio()?.id ?? '', this.section as 'marcas' | 'categorias' | 'proveedores' | 'unidades', this.importFile).subscribe({
      next: (result) => {
        this.importing.set(false);
        this.importResult.set(result);
        const partial = result.omitidas > 0 || result.invalidas > 0;
        void Swal.fire({
          icon: partial ? 'warning' : 'success',
          title: 'Importación terminada',
          text: `${result.creadas} registros insertados · ${result.omitidas} omitidos · ${result.invalidas} inválidos`,
          confirmButtonText: 'Ver registros',
        }).then(() => this.goToList());
      },
      error: (error) => {
        this.importing.set(false);
        this.importError.set(error.error?.mensaje ?? 'No fue posible importar el archivo.');
      },
    });
  }
  openEdit(item: CatalogRecord): void {
    this.prepareEdit(item);
    this.navigateToEdit(item);
  }

  private prepareEdit(item: CatalogRecord): void {
    this.editingId = item.id;
    this.form.reset({
      nombre: item.nombre,
      descripcion: 'descripcion' in item ? (item.descripcion ?? '') : '',
      categoria_padre_id: 'categoria_padre_id' in item ? (item.categoria_padre_id ?? '') : '',
      razon_social: 'razon_social' in item ? (item.razon_social ?? '') : '',
      rfc: 'rfc' in item ? (item.rfc ?? '') : '',
      telefono: 'telefono' in item ? (item.telefono ?? '') : '',
      email: 'email' in item ? (item.email ?? '') : '',
      direccion: 'direccion' in item ? (item.direccion ?? '') : '',
      simbolo: 'simbolo' in item ? item.simbolo : '',
      tipo: 'tipo' in item ? item.tipo : 'PESO',
      factor_a_base: 'factor_a_base' in item ? item.factor_a_base : 1,
      decimales: 'decimales' in item ? item.decimales : 2,
    });
    this.formError.set(null);
  }
  save(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.saving.set(true);
    this.formError.set(null);
    const v = this.form.getRawValue();
    if (this.section === 'unidades' && (!v.simbolo.trim() || !v.tipo.trim())) {
      this.formError.set('Símbolo y tipo son obligatorios para la unidad.');
      this.saving.set(false);
      return;
    }
    const payload: any = { nombre: v.nombre };
    if (this.section === 'categorias')
      Object.assign(payload, {
        descripcion: v.descripcion || null,
        categoria_padre_id: v.categoria_padre_id || null,
      });
    if (this.section === 'proveedores')
      Object.assign(payload, {
        razon_social: v.razon_social || null,
        rfc: v.rfc || null,
        telefono: v.telefono || null,
        email: v.email || null,
        direccion: v.direccion || null,
      });
    if (this.section === 'unidades') Object.assign(payload, { simbolo: v.simbolo, tipo: v.tipo, factor_a_base: v.factor_a_base, decimales: v.decimales });
    const negocioId = this.contexto.negocio()?.id ?? '';
    const requestPath = this.section === 'unidades' ? 'unidades-medida' : this.section;
    const request = this.editingId
      ? this.service.actualizar(negocioId, requestPath, this.editingId, payload)
      : this.service.crear(negocioId, requestPath, payload);
    request.subscribe({
      next: () => {
        this.saving.set(false);
        this.goToList();
      },
      error: (e) => {
        this.saving.set(false);
        this.formError.set(
          e.error?.mensaje ?? 'No fue posible comunicarse con el servidor. Intenta de nuevo.',
        );
      },
    });
  }
  hasError(name: string): boolean {
    const c = this.form.get(name);
    return !!c && c.invalid && (c.dirty || c.touched);
  }
  isCategory(): boolean {
    return this.section === 'categorias';
  }
  isProvider(): boolean {
    return this.section === 'proveedores';
  }
  isUnit(): boolean {
    return this.section === 'unidades';
  }

  private resetCatalogPagination(): void {
    this.catalogPageFirst.set(0);
  }

  private matchesCatalogFilters(item: CatalogRecord): boolean {
    const status = this.catalogStatusFilter();
    if (status === 'active' && !item.activo) return false;
    if (status === 'inactive' && item.activo) return false;

    if (this.section === 'categorias') {
      const parent = this.categoryParentFilter();
      const category = item as Categoria;
      if (parent === NO_PARENT_CATEGORY && category.categoria_padre_id) return false;
      if (parent && parent !== NO_PARENT_CATEGORY && category.categoria_padre_id !== parent) return false;
    }
    if (this.section === 'unidades' && this.unitTypeFilter() && (item as UnidadMedida).tipo !== this.unitTypeFilter()) return false;

    const query = this.normalizeCatalogFilter(this.catalogSearch());
    if (!query) return true;
    return this.catalogSearchableValues(item).some((value) =>
      this.normalizeCatalogFilter(value).includes(query),
    );
  }

  private catalogSearchableValues(item: CatalogRecord): Array<string | null | undefined> {
    if (this.section === 'categorias') {
      const category = item as Categoria;
      return [category.nombre, category.descripcion];
    }
    if (this.section === 'unidades') {
      const unit = item as UnidadMedida;
      return [unit.nombre, unit.simbolo];
    }
    if (this.section === 'proveedores') {
      const provider = item as Proveedor;
      return [provider.nombre, provider.razon_social, provider.rfc, provider.email, provider.telefono];
    }
    return [item.nombre];
  }

  private normalizeCatalogFilter(value: string | null | undefined): string {
    return (value ?? '')
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .trim()
      .toLocaleLowerCase('es');
  }

  canImport(): boolean {
    return this.section === 'marcas' || this.section === 'categorias' || this.section === 'proveedores' || this.section === 'unidades';
  }

  navigateToCreate(): void {
    this.navigate([this.catalogBasePath(), 'nuevo']);
  }

  navigateToEdit(item: CatalogRecord): void {
    this.navigate([this.catalogBasePath(), item.id, 'editar']);
  }

  async deleteCatalogItem(item: CatalogRecord): Promise<void> {
    if (this.deletingId()) return;
    this.deletingId.set(item.id);
    const negocioId = this.contexto.negocio()?.id ?? '';
    const requestPath = this.section === 'unidades' ? 'unidades-medida' : this.section;
    this.service.comprobarEliminacion(negocioId, requestPath, item.id).subscribe({
      next: async (check) => {
        this.deletingId.set(null);
        if (!check.puede_eliminarse) {
          this.feedback.warning(
            'No se puede eliminar',
            `Tiene vinculaciones con: ${check.vinculaciones.join(', ')}.`,
          );
          return;
        }
        const confirmed = await this.feedback.confirmDanger({
          titulo: `Eliminar ${this.singularTitle.toLowerCase()}`,
          descripcion: `Se eliminará “${item.nombre}” porque no tiene vinculaciones.`,
          textoConfirmar: 'Sí, eliminar',
        });
        if (!confirmed) return;
        this.deletingId.set(item.id);
        this.service.eliminar(negocioId, requestPath, item.id).subscribe({
          next: () => {
            this.deletingId.set(null);
            const feminine = this.section === 'marcas' || this.section === 'categorias';
            this.feedback.success(`${this.singularTitle} ${feminine ? 'eliminada' : 'eliminado'}`);
            this.load();
          },
          error: (response) => {
            this.deletingId.set(null);
            if (response.status === 409) {
              this.feedback.warning('No se puede eliminar', response.error?.mensaje ?? 'El registro tiene vinculaciones.');
              return;
            }
            this.feedback.error('No fue posible eliminar el registro', response.error?.mensaje ?? 'Intenta nuevamente.');
          },
        });
      },
      error: (response) => {
        this.deletingId.set(null);
        this.feedback.error('No fue posible comprobar las vinculaciones', response.error?.mensaje ?? 'Intenta nuevamente.');
      },
    });
  }

  navigateToImport(): void {
    this.navigate([this.catalogBasePath(), 'importar']);
  }

  goToList(): void {
    this.navigate([this.catalogBasePath()]);
  }

  private catalogBasePath(): string {
    if (this.section === 'proveedores') return '/proveedores';
    return this.section === 'unidades' ? '/catalogo/unidades-medida' : `/catalogo/${this.section}`;
  }

  private navigate(commands: string[]): void {
    void this.router?.navigate(commands);
  }
}

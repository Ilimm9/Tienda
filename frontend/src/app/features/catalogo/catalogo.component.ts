import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { DialogModule } from 'primeng/dialog';
import { InputTextModule } from 'primeng/inputtext';
import { TableModule } from 'primeng/table';
import { TextareaModule } from 'primeng/textarea';
import { SelectModule } from 'primeng/select';
import { ContextoService } from '../../contexto/contexto.service';
import { CatalogoService } from './catalogo.service';
import { CatalogImportResult, CatalogRecord, Categoria } from './catalogo.models';

interface SelectOption {
  label: string;
  value: string;
}

@Component({
  selector: 'app-catalogo',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    DialogModule,
    InputTextModule,
    TableModule,
    TextareaModule,
    SelectModule,
  ],
  templateUrl: './catalogo.component.html',
  styleUrl: './catalogo.component.css',
})
export class CatalogoComponent {
  private readonly service = inject(CatalogoService);
  private readonly fb = inject(FormBuilder);
  private readonly route = inject(ActivatedRoute);
  private readonly contexto = inject(ContextoService);
  private readonly router = inject(Router, { optional: true });
  readonly section = this.route.snapshot.data['section'] as 'marcas' | 'categorias' | 'proveedores' | 'unidades';
  readonly pageMode = (this.route.snapshot.data['mode'] as 'list' | 'create' | 'edit' | 'import' | undefined) ?? 'list';
  readonly items = signal<CatalogRecord[]>([]);
  readonly parents = signal<Categoria[]>([]);
  readonly unitTypeOptions: SelectOption[] = [
    { label: 'Peso', value: 'PESO' },
    { label: 'Volumen', value: 'VOLUMEN' },
    { label: 'Longitud', value: 'LONGITUD' },
    { label: 'Área', value: 'AREA' },
    { label: 'Cantidad', value: 'CANTIDAD' },
    { label: 'Empaque', value: 'EMPAQUE' },
  ];
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly dialogVisible = signal(false);
  editingId: string | null = null;
  readonly saving = signal(false);
  readonly formError = signal<string | null>(null);
  importDialogVisible = false;
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
    codigo: [''], simbolo: [''], tipo: ['PESO'], factor_a_base: [1, [Validators.min(0.000001)]], decimales: [2, [Validators.min(0), Validators.max(6)]],
  });
  constructor() {
    if (this.pageMode === 'list') this.load();
    else if (this.pageMode === 'create') this.prepareCreate();
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
    const request =
      this.section === 'marcas'
        ? this.service.marcas()
        : this.section === 'categorias'
          ? this.service.categorias()
          : this.section === 'proveedores' ? this.service.proveedores(this.contexto.negocio()?.id ?? '') : this.service.unidades();
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
    const request = this.section === 'marcas'
      ? this.service.marcas()
      : this.section === 'categorias'
        ? this.service.categorias()
        : this.service.unidades();
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
    if (!this.isProvider()) this.navigateToCreate();
    else this.dialogVisible.set(true);
  }

  private prepareCreate(): void {
    this.loading.set(false);
    this.editingId = null;
    this.form.reset();
    this.formError.set(null);
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
    return this.service.plantillaUrl(this.section as 'marcas' | 'categorias' | 'unidades');
  }
  importCatalog(): void {
    if (!this.importFile || !this.canImport()) return;
    this.importing.set(true);
    this.importError.set(null);
    this.importResult.set(null);
    this.service.importar(this.section as 'marcas' | 'categorias' | 'unidades', this.importFile).subscribe({
      next: (result) => {
        this.importing.set(false);
        this.importResult.set(result);
        if (result.creadas > 0) this.load();
      },
      error: (error) => {
        this.importing.set(false);
        this.importError.set(error.error?.mensaje ?? 'No fue posible importar el archivo.');
      },
    });
  }
  openEdit(item: CatalogRecord): void {
    this.prepareEdit(item);
    if (!this.isProvider()) this.navigateToEdit(item);
    else this.dialogVisible.set(true);
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
      codigo: 'codigo' in item ? item.codigo : '',
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
    if (this.section === 'unidades' && (!v.codigo.trim() || !v.simbolo.trim() || !v.tipo.trim())) {
      this.formError.set('Código, símbolo y tipo son obligatorios para la unidad.');
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
    if (this.section === 'unidades') Object.assign(payload, { codigo: v.codigo, simbolo: v.simbolo, tipo: v.tipo, factor_a_base: v.factor_a_base, decimales: v.decimales });
    const path =
      this.section === 'proveedores'
        ? `negocios/${this.contexto.negocio()?.id ?? ''}/catalogo/proveedores`
        : `catalogo/${this.section}`;
    const requestPath = this.section === 'unidades' ? 'catalogo/unidades-medida' : path;
    const request = this.editingId
      ? this.service.actualizar(requestPath, this.editingId, payload)
      : this.service.crear(requestPath, payload);
    request.subscribe({
      next: () => {
        this.saving.set(false);
        if (this.isFormPage()) this.goToList();
        else {
          this.dialogVisible.set(false);
          this.load();
        }
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
  canImport(): boolean {
    return this.section === 'marcas' || this.section === 'categorias' || this.section === 'unidades';
  }

  navigateToCreate(): void {
    this.navigate([this.catalogBasePath(), 'nuevo']);
  }

  navigateToEdit(item: CatalogRecord): void {
    this.navigate([this.catalogBasePath(), item.id, 'editar']);
  }

  navigateToImport(): void {
    this.navigate([this.catalogBasePath(), 'importar']);
  }

  goToList(): void {
    this.navigate([this.catalogBasePath()]);
  }

  private catalogBasePath(): string {
    return this.section === 'unidades' ? '/catalogo/unidades-medida' : `/catalogo/${this.section}`;
  }

  private navigate(commands: string[]): void {
    void this.router?.navigate(commands);
  }
}

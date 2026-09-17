import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
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
  readonly section = this.route.snapshot.data['section'] as 'marcas' | 'categorias' | 'proveedores' | 'unidades';
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
    this.load();
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
  openCreate(): void {
    this.editingId = null;
    this.form.reset();
    this.formError.set(null);
    this.dialogVisible.set(true);
  }
  openImport(): void {
    this.importFile = null;
    this.importDropActive = false;
    this.importError.set(null);
    this.importResult.set(null);
    this.importDialogVisible = true;
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
    this.importDropActive = false;
  }
  onImportDrop(event: DragEvent): void {
    event.preventDefault();
    this.importDropActive = false;
    if (!this.importing()) this.setImportFile(event.dataTransfer?.files?.[0] ?? null);
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
    this.dialogVisible.set(true);
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
        this.dialogVisible.set(false);
        this.load();
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
}

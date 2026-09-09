import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { DialogModule } from 'primeng/dialog';
import { InputTextModule } from 'primeng/inputtext';
import { TableModule } from 'primeng/table';
import { TextareaModule } from 'primeng/textarea';
import { SelectModule } from 'primeng/select';
import { environment } from '../../../environments/environment';
import { CatalogoService } from './catalogo.service';
import { CatalogImportResult, CatalogRecord, Categoria } from './catalogo.models';

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
  readonly section = this.route.snapshot.data['section'] as 'marcas' | 'categorias' | 'proveedores';
  readonly items = signal<CatalogRecord[]>([]);
  readonly parents = signal<Categoria[]>([]);
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  dialogVisible = false;
  editingId: string | null = null;
  saving = false;
  formError: string | null = null;
  importDialogVisible = false;
  importFile: File | null = null;
  importDropActive = false;
  importing = false;
  importError: string | null = null;
  importResult: CatalogImportResult | null = null;
  readonly form = this.fb.nonNullable.group({
    nombre: ['', [Validators.required, Validators.maxLength(180)]],
    descripcion: [''],
    categoria_padre_id: [''],
    razon_social: [''],
    rfc: [''],
    telefono: [''],
    email: ['', Validators.email],
    direccion: [''],
  });
  constructor() {
    this.load();
  }
  get title(): string {
    return this.section === 'marcas'
      ? 'Marcas'
      : this.section === 'categorias'
        ? 'Categorías'
        : 'Proveedores';
  }
  get description(): string {
    return this.section === 'marcas'
      ? 'Administra las marcas de tus productos.'
      : this.section === 'categorias'
        ? 'Organiza los productos por categorías.'
        : 'Administra tus proveedores de inventario.';
  }
  load(): void {
    this.loading.set(true);
    const request =
      this.section === 'marcas'
        ? this.service.marcas()
        : this.section === 'categorias'
          ? this.service.categorias()
          : this.service.proveedores(environment.defaultBusinessId);
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
    this.formError = null;
    this.dialogVisible = true;
  }
  openImport(): void {
    this.importFile = null;
    this.importDropActive = false;
    this.importError = null;
    this.importResult = null;
    this.importDialogVisible = true;
  }
  onImportFile(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.setImportFile(input.files?.[0] ?? null);
    input.value = '';
  }
  openFilePicker(event: MouseEvent, input: HTMLInputElement): void {
    if (event.target === input || this.importing) return;
    input.click();
  }
  onImportDragOver(event: DragEvent): void {
    event.preventDefault();
    if (!this.importing) this.importDropActive = true;
  }
  onImportDragLeave(event: DragEvent): void {
    event.preventDefault();
    this.importDropActive = false;
  }
  onImportDrop(event: DragEvent): void {
    event.preventDefault();
    this.importDropActive = false;
    if (!this.importing) this.setImportFile(event.dataTransfer?.files?.[0] ?? null);
  }
  private setImportFile(file: File | null): void {
    this.importFile = file;
    this.importError = null;
    this.importResult = null;
  }
  templateUrl(): string {
    return this.service.plantillaUrl(this.section as 'marcas' | 'categorias');
  }
  importCatalog(): void {
    if (!this.importFile || !this.canImport()) return;
    this.importing = true;
    this.importError = null;
    this.importResult = null;
    this.service.importar(this.section as 'marcas' | 'categorias', this.importFile).subscribe({
      next: (result) => {
        this.importing = false;
        this.importResult = result;
        if (result.creadas > 0) this.load();
      },
      error: (error) => {
        this.importing = false;
        this.importError = error.error?.mensaje ?? 'No fue posible importar el archivo.';
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
    });
    this.formError = null;
    this.dialogVisible = true;
  }
  save(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.saving = true;
    this.formError = null;
    const v = this.form.getRawValue();
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
    const path =
      this.section === 'proveedores'
        ? `negocios/${environment.defaultBusinessId}/catalogo/proveedores`
        : `catalogo/${this.section}`;
    const request = this.editingId
      ? this.service.actualizar(path, this.editingId, payload)
      : this.service.crear(path, payload);
    request.subscribe({
      next: () => {
        this.saving = false;
        this.dialogVisible = false;
        this.load();
      },
      error: (e) => {
        this.saving = false;
        this.formError = e.error?.mensaje ?? 'No fue posible guardar los cambios.';
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
  canImport(): boolean {
    return this.section === 'marcas' || this.section === 'categorias';
  }
}

import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { DialogModule } from 'primeng/dialog';
import { InputTextModule } from 'primeng/inputtext';
import { SelectModule } from 'primeng/select';
import { TableModule } from 'primeng/table';
import { TextareaModule } from 'primeng/textarea';

import { environment } from '../../../environments/environment';
import { CatalogOption, ProductRow } from './product.models';
import { ProductosService } from './productos.service';

@Component({
  selector: 'app-productos',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    DialogModule,
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
  private readonly formBuilder = inject(FormBuilder);

  readonly products = signal<ProductRow[]>([]);
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly categories = signal<CatalogOption[]>([]);
  readonly brands = signal<CatalogOption[]>([]);
  readonly branches = signal<CatalogOption[]>([]);
  readonly saving = signal(false);
  dialogVisible = false;
  formError: string | null = null;

  readonly productForm = this.formBuilder.nonNullable.group({
    nombre: ['', [Validators.required, Validators.maxLength(255)]],
    sku_interno: ['', [Validators.required, Validators.maxLength(120)]],
    marca_id: [''],
    categoria_id: ['', Validators.required],
    sucursal_id: ['', Validators.required],
    descripcion: ['', Validators.maxLength(2000)],
    contenido: this.formBuilder.control<number | null>(null, [Validators.min(0)]),
    unidad_contenido: ['', Validators.maxLength(30)],
    presentacion: ['', Validators.maxLength(100)],
    precio_venta: [0, [Validators.required, Validators.min(0)]],
    stock_inicial: [0, [Validators.required, Validators.min(0)]],
  });

  constructor() {
    this.loadProducts();
  }

  loadProducts(): void {
    this.loading.set(true);
    this.error.set(null);
    this.productosService.listByBusiness(environment.defaultBusinessId).subscribe({
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

  openCreateDialog(): void {
    this.productForm.reset({
      nombre: '', sku_interno: '', marca_id: '', categoria_id: '', sucursal_id: '',
      descripcion: '', contenido: null, unidad_contenido: '', presentacion: '',
      precio_venta: 0, stock_inicial: 0,
    });
    this.formError = null;
    this.dialogVisible = true;
    this.productosService.listCategories(environment.defaultBusinessId).subscribe({ next: (items) => this.categories.set(items) });
    this.productosService.listBrands(environment.defaultBusinessId).subscribe({ next: (items) => this.brands.set(items) });
    this.productosService.listBranches(environment.defaultBusinessId).subscribe({ next: (items) => this.branches.set(items) });
  }

  closeCreateDialog(): void {
    if (!this.saving()) this.dialogVisible = false;
  }

  hasError(field: string): boolean {
    const control = this.productForm.get(field);
    return !!control && control.invalid && (control.dirty || control.touched);
  }

  saveProduct(): void {
    if (this.productForm.invalid) {
      this.productForm.markAllAsTouched();
      return;
    }
    this.saving.set(true);
    this.formError = null;
    const value = this.productForm.getRawValue();
    this.productosService.create(environment.defaultBusinessId, {
      ...value,
      marca_id: value.marca_id || null,
      descripcion: value.descripcion || null,
      contenido: value.contenido,
      unidad_contenido: value.unidad_contenido || null,
      presentacion: value.presentacion || null,
    }).subscribe({
      next: () => {
        this.saving.set(false);
        this.dialogVisible = false;
        this.loadProducts();
      },
      error: (response) => {
        this.saving.set(false);
        this.formError = response.error?.mensaje ?? 'No fue posible guardar el producto.';
      },
    });
  }
}

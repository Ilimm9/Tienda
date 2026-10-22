import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import {
  AbstractControl,
  FormArray,
  FormBuilder,
  FormsModule,
  ReactiveFormsModule,
  Validators,
} from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { InputTextModule } from 'primeng/inputtext';
import { SelectModule } from 'primeng/select';
import { TableModule } from 'primeng/table';
import { TextareaModule } from 'primeng/textarea';
import { forkJoin } from 'rxjs';
import { ContextoService } from '../../contexto/contexto.service';
import { CatalogOption, PackagingUnit, PurchaseProduct, PurchaseRow } from './compras.models';
import { ComprasService } from './compras.service';
import { PERMISOS } from '../../contexto/permisos';
import { PuedeDirective } from '../../contexto/puede.directive';

@Component({
  selector: 'app-compras',
  standalone: true,
  imports: [PuedeDirective, 
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    RouterLink,
    InputTextModule,
    SelectModule,
    TableModule,
    TextareaModule,
  ],
  templateUrl: './compras.component.html',
  styleUrl: './compras.component.css',
})
export class ComprasComponent {
  readonly P = PERMISOS;
  private readonly service = inject(ComprasService);
  private readonly context = inject(ContextoService);
  private readonly fb = inject(FormBuilder);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  readonly createMode = this.route.snapshot.data['mode'] === 'create';
  readonly purchases = signal<PurchaseRow[]>([]);
  readonly providers = signal<CatalogOption[]>([]);
  readonly branches = signal<CatalogOption[]>([]);
  readonly units = signal<PackagingUnit[]>([]);
  readonly products = signal<PurchaseProduct[]>([]);
  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly creatingProvider = signal(false);
  readonly providerSaving = signal(false);
  readonly providerError = signal<string | null>(null);
  readonly unitDetailIndex = signal<number | null>(null);
  // Detalles compactados; se guarda el control para que quitar uno no altere a los demás.
  readonly contraidos = signal(new Set<AbstractControl>());
  readonly unitSaving = signal(false);
  readonly documentTypes = ['TICKET', 'FACTURA'];
  readonly paymentMethods = ['EFECTIVO', 'TRANSFERENCIA', 'TARJETA', 'CREDITO', 'OTRO'];
  private readonly calculationVersion = signal(0);
  readonly form = this.fb.nonNullable.group({
    proveedor_id: ['', Validators.required],
    sucursal_id: ['', Validators.required],
    tipo_documento: ['TICKET' as 'TICKET' | 'FACTURA', Validators.required],
    folio_documento: ['', Validators.required],
    numero_cuenta: [''],
    fecha_documento: [this.today(), Validators.required],
    fecha_recepcion: [this.today(), Validators.required],
    forma_pago: ['EFECTIVO', Validators.required],
    incluye_ieps: [false],
    ieps: [0, Validators.min(0)],
    incluye_iva: [false],
    observaciones: [''],
    detalles: this.fb.array([this.newDetail()]),
  });
  readonly unitForm = this.fb.nonNullable.group({
    nombre: ['', Validators.required],
    simbolo: ['', Validators.required],
    tipo: ['EMPAQUE', Validators.required],
    factor_a_base: [1, [Validators.required, Validators.min(1), Validators.pattern('^[0-9]+$')]],
    decimales: [0, [Validators.required, Validators.min(0), Validators.max(6)]],
  });
  readonly providerForm = this.fb.nonNullable.group({ nombre: ['', Validators.required] });
  readonly packagingUnits = computed(() =>
    this.units().filter(
      (unit) =>
        unit.tipo === 'EMPAQUE' &&
        Number.isInteger(Number(unit.factor_a_base)) &&
        unit.factor_a_base >= 1,
    ),
  );
  readonly productOptions = computed(() => [
    ...this.products(),
    { id: '__new_product__', nombre: 'Crear nuevo…', sku: null, codigo_barras: null },
  ]);
  readonly providerOptions = computed(() => [
    ...this.providers(),
    ...(this.context.puede(this.P.catalogoGestionar)
      ? [{ id: '__new_provider__', nombre: 'Crear nuevo…' }]
      : []),
  ]);
  readonly unitOptions = computed(() => [
    ...this.packagingUnits(),
    { id: '__new_unit__', nombre: 'Crear nuevo…', tipo: 'EMPAQUE', factor_a_base: 0 },
  ]);
  readonly subtotal = computed(() => {
    this.calculationVersion();
    return this.details.controls.reduce(
      (total, control) =>
        total +
        Math.max(
          0,
          Number(control.get('cantidad_empaques')?.value ?? 0) *
            Number(control.get('costo_empaque')?.value ?? 0) -
            Number(control.get('descuento')?.value ?? 0),
        ),
      0,
    );
  });
  readonly ieps = computed(() => {
    this.calculationVersion();
    return this.form.controls.incluye_ieps.value ? this.money(this.form.controls.ieps.value) : 0;
  });
  readonly iva = computed(() => {
    this.calculationVersion();
    return this.form.controls.incluye_iva.value
      ? this.money((this.subtotal() + this.ieps()) * 0.16)
      : 0;
  });
  readonly taxes = computed(() => this.money(this.ieps() + this.iva()));
  readonly total = computed(() => this.money(this.subtotal() + this.taxes()));
  get details(): FormArray {
    return this.form.controls.detalles;
  }
  constructor() {
    this.form.valueChanges.subscribe(() => this.calculationVersion.update((value) => value + 1));
    this.form.controls.incluye_ieps.valueChanges.subscribe((included) => {
      this.form.controls.ieps.setValue(included ? this.money(this.subtotal() * 0.08) : 0, {
        emitEvent: false,
      });
      this.form.controls.ieps.markAsPristine();
    });
    this.createMode ? this.loadForm() : this.loadList();
  }
  private get businessId(): string {
    return this.context.negocio()?.id ?? '';
  }
  private today(): string {
    return new Date().toISOString().slice(0, 10);
  }
  private money(value: number): number {
    return Math.round(value * 100) / 100;
  }
  private newDetail() {
    return this.fb.nonNullable.group({
      producto_negocio_id: ['', Validators.required],
      unidad_medida_id: ['', Validators.required],
      cantidad_empaques: [1, [Validators.required, Validators.min(1)]],
      costo_empaque: [0, [Validators.required, Validators.min(0.01)]],
      descuento: [0, Validators.min(0)],
      tiene_anomalias: [false],
      piezas_danadas: [0, Validators.min(0)],
      piezas_faltantes: [0, Validators.min(0)],
      numero_lote: [''],
      fecha_caducidad: [''],
      producto_nuevo: [false],
      nombre_producto_nuevo: [''],
      codigo_barras_nuevo: [''],
    });
  }
  private loadList(): void {
    this.service.list(this.businessId).subscribe({
      next: (r) => {
        this.purchases.set(r.items);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('No fue posible cargar las compras.');
        this.loading.set(false);
      },
    });
  }
  private loadForm(): void {
    forkJoin({
      providers: this.service.providers(this.businessId),
      branches: this.service.branches(this.businessId),
      units: this.service.units(this.businessId),
    }).subscribe({
      next: (r) => {
        this.providers.set(r.providers);
        this.branches.set(r.branches);
        this.units.set(r.units);
        const activeBranch = r.branches.find((branch) => branch.id === this.context.sucursal()?.id);
        if (!activeBranch) {
          this.error.set('No se encontró la sucursal seleccionada en el contexto actual.');
          this.loading.set(false);
          return;
        }
        this.form.controls.sucursal_id.setValue(activeBranch.id);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('No fue posible cargar los catálogos.');
        this.loading.set(false);
      },
    });
  }
  loadProducts(): void {
    const providerId = this.form.controls.proveedor_id.value;
    this.products.set([]);
    if (!providerId) return;
    this.service
      .products(this.businessId, providerId)
      .subscribe({
        next: (items) => this.products.set(items),
        error: () => this.error.set('No fue posible cargar los productos del proveedor.'),
      });
  }
  selectProvider(value: string): void {
    if (value !== '__new_provider__') {
      this.creatingProvider.set(false);
      this.loadProducts();
      return;
    }
    this.form.controls.proveedor_id.setValue('');
    this.providerForm.reset({ nombre: '' });
    this.providerError.set(null);
    this.creatingProvider.set(true);
  }
  createProvider(): void {
    this.providerForm.markAllAsTouched();
    if (this.providerForm.invalid || this.providerSaving()) return;
    const nombre = this.providerForm.controls.nombre.value.trim();
    const previousIDs = new Set(this.providers().map((provider) => provider.id));
    this.providerSaving.set(true);
    this.providerError.set(null);
    this.service.createProvider(this.businessId, { nombre }).subscribe({
      next: () =>
        this.service.providers(this.businessId).subscribe({
          next: (providers) => {
            this.providers.set(providers);
            const created = providers.find((provider) => !previousIDs.has(provider.id))
              ?? providers.find((provider) => provider.nombre.trim().toLowerCase() === nombre.toLowerCase());
            if (!created) {
              this.providerSaving.set(false);
              this.providerError.set('El proveedor fue creado, pero no fue posible seleccionarlo.');
              return;
            }
            this.form.controls.proveedor_id.setValue(created.id);
            this.creatingProvider.set(false);
            this.providerSaving.set(false);
            this.loadProducts();
          },
          error: () => {
            this.providerSaving.set(false);
            this.providerError.set('El proveedor fue creado, pero no fue posible recargarlo.');
          },
        }),
      error: (response) => {
        this.providerSaving.set(false);
        this.providerError.set(response.error?.mensaje ?? 'No fue posible crear el proveedor.');
      },
    });
  }
  addDetail(): void {
    this.contraidos.set(new Set(this.details.controls));
    this.details.push(this.newDetail());
  }
  removeDetail(index: number): void {
    if (this.details.length === 1) return;
    const control = this.details.at(index);
    this.contraidos.update((actuales) => new Set([...actuales].filter((item) => item !== control)));
    this.details.removeAt(index);
  }
  estaContraido(index: number): boolean {
    return this.contraidos().has(this.details.at(index));
  }
  alternarDetalle(index: number): void {
    const control = this.details.at(index);
    this.contraidos.update((actuales) => {
      const siguientes = new Set(actuales);
      if (!siguientes.delete(control)) siguientes.add(control);
      return siguientes;
    });
  }
  private abrirDetalles(controles: AbstractControl[]): void {
    this.contraidos.update((actuales) => new Set([...actuales].filter((item) => !controles.includes(item))));
  }
  selectProduct(index: number, value: string): void {
    const row = this.details.at(index);
    const isNew = value === '__new_product__';
    const product = row.get('producto_negocio_id');
    const newProductName = row.get('nombre_producto_nuevo');
    row.get('producto_nuevo')?.setValue(isNew);
    if (isNew) {
      product?.setValue('');
      product?.clearValidators();
      newProductName?.setValidators(Validators.required);
    } else {
      product?.setValidators(Validators.required);
      newProductName?.clearValidators();
    }
    product?.updateValueAndValidity();
    newProductName?.updateValueAndValidity();
  }
  selectUnit(index: number, value: string): void {
    if (value === '__new_unit__') {
      this.details.at(index).get('unidad_medida_id')?.setValue('');
      this.startUnit(index);
    } else if (this.unitDetailIndex() === index) this.unitDetailIndex.set(null);
  }
  startUnit(index: number): void {
    this.unitDetailIndex.set(index);
    this.unitForm.reset({
      nombre: '',
      simbolo: '',
      tipo: 'EMPAQUE',
      factor_a_base: 1,
      decimales: 0,
    });
  }
  createUnit(): void {
    const index = this.unitDetailIndex();
    this.unitForm.markAllAsTouched();
    if (index === null || this.unitForm.invalid || this.unitSaving()) return;
    const value = this.unitForm.getRawValue();
    this.unitSaving.set(true);
    this.service.createUnit(this.businessId, value).subscribe({
      next: () =>
        this.service.units(this.businessId).subscribe({
          next: (items) => {
            this.units.set(items);
            const added = items.find(
              (item) => item.nombre.toLowerCase() === value.nombre.trim().toLowerCase(),
            );
            if (added) this.details.at(index).get('unidad_medida_id')?.setValue(added.id);
            this.unitSaving.set(false);
            this.unitDetailIndex.set(null);
          },
          error: () => {
            this.unitSaving.set(false);
            this.error.set('La unidad fue creada, pero no fue posible recargarla.');
          },
        }),
      error: (r) => {
        this.unitSaving.set(false);
        this.error.set(r.error?.mensaje ?? 'No fue posible crear la unidad.');
      },
    });
  }
  isNewProduct(index: number): boolean {
    return Boolean(this.details.at(index).get('producto_nuevo')?.value);
  }
  hasAnomalies(index: number): boolean {
    return Boolean(this.details.at(index).get('tiene_anomalias')?.value);
  }
  selectedUnit(index: number): PackagingUnit | undefined {
    return this.units().find(
      (unit) => unit.id === this.details.at(index).get('unidad_medida_id')?.value,
    );
  }
  unitFactor(index: number): number {
    return Number(this.selectedUnit(index)?.factor_a_base ?? 0);
  }
  pieces(index: number): number {
    const row = this.details.at(index).value as {
      cantidad_empaques: number;
      piezas_danadas: number;
      piezas_faltantes: number;
    };
    return Math.max(
      0,
      (row.cantidad_empaques || 0) * this.unitFactor(index) -
        (row.piezas_danadas || 0) -
        (row.piezas_faltantes || 0),
    );
  }
  detailSubtotal(index: number): number {
    const row = this.details.at(index).value as {
      cantidad_empaques: number;
      costo_empaque: number;
      descuento: number;
    };
    return Math.max(
      0,
      (row.cantidad_empaques || 0) * (row.costo_empaque || 0) - (row.descuento || 0),
    );
  }
  productSummaryName(index: number): string {
    const row = this.details.at(index);
    if (this.isNewProduct(index))
      return row.get('nombre_producto_nuevo')?.value || 'Producto nuevo';
    return (
      this.products().find((product) => product.id === row.get('producto_negocio_id')?.value)
        ?.nombre || 'Sin producto'
    );
  }
  submit(): void {
    this.form.markAllAsTouched();
    if (this.form.invalid) this.abrirDetalles(this.details.controls.filter((detail) => detail.invalid));
    if (this.form.invalid || this.saving()) return;
    for (let i = 0; i < this.details.length; i++) {
      const v = this.details.at(i).value as {
        piezas_danadas: number;
        piezas_faltantes: number;
        cantidad_empaques: number;
      };
      if (
        (v.piezas_danadas || 0) + (v.piezas_faltantes || 0) >
        (v.cantidad_empaques || 0) * this.unitFactor(i)
      ) {
        this.abrirDetalles([this.details.at(i)]);
        this.error.set(`El detalle ${i + 1} tiene incidencias mayores al total.`);
        return;
      }
    }
    const value = this.form.getRawValue();
    this.saving.set(true);
    this.error.set(null);
    this.service
      .create(this.businessId, {
        ...value,
        numero_cuenta: value.numero_cuenta || null,
        observaciones: value.observaciones || null,
        detalles: value.detalles.map(({ tiene_anomalias, ...d }) => ({
          ...d,
          numero_lote: d.numero_lote || null,
          fecha_caducidad: d.fecha_caducidad || null,
        })),
      })
      .subscribe({
        next: () => void this.router.navigate(['/compras']),
        error: (r) => {
          this.saving.set(false);
          this.error.set(r.error?.mensaje ?? 'No fue posible confirmar la compra.');
        },
      });
  }
}

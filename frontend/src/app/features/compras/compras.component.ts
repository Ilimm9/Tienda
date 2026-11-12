import { CommonModule } from '@angular/common';
import { Component, computed, effect, inject, signal, untracked, viewChild } from '@angular/core';
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
import { TextareaModule } from 'primeng/textarea';
import { forkJoin } from 'rxjs';
import { ContextoService } from '../../contexto/contexto.service';
import { CatalogOption, PackagingUnit, PurchaseProduct, PurchaseProductGroup, PurchaseRow } from './compras.models';
import { ComprasService } from './compras.service';
import { PERMISOS } from '../../contexto/permisos';
import { PuedeDirective } from '../../contexto/puede.directive';
import { VentanaCompraComponent } from './ventana-compra.component';
import { ComprasListaComponent } from './compras-lista.component';
import { CompraVentanasComponent } from './compra-ventanas.component';
import { calcularImpuestos, desglosar, money, TASAS_IEPS, TASAS_IVA } from './impuestos';

@Component({
  selector: 'app-compras',
  standalone: true,
  imports: [PuedeDirective, VentanaCompraComponent, ComprasListaComponent, CompraVentanasComponent,
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    RouterLink,
    InputTextModule,
    SelectModule,
    TextareaModule,
  ],
  templateUrl: './compras.component.html',
  styleUrl: './compras.component.css',
})
export class ComprasComponent {
  readonly P = PERMISOS;
  readonly Number = Number;
  private readonly service = inject(ComprasService);
  private readonly context = inject(ContextoService);
  private readonly fb = inject(FormBuilder);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  readonly createMode = this.route.snapshot.data['mode'] === 'create';
  readonly purchases = signal<PurchaseRow[]>([]);
  readonly lista = viewChild(ComprasListaComponent);
  /** Compra abierta en Ver, Editar, Suspender o Reactivar. */
  readonly accion = signal<{ tipo: 'ver' | 'editar' | 'suspender' | 'reactivar'; compra: PurchaseRow } | null>(null);
  /** Mensaje tras editar, suspender o reactivar. */
  readonly aviso = signal<string | null>(null);
  readonly providers = signal<CatalogOption[]>([]);
  readonly branches = signal<CatalogOption[]>([]);
  readonly units = signal<PackagingUnit[]>([]);
  readonly products = signal<PurchaseProduct[]>([]);
  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  /** Ventana abierta para crear proveedor o capturar producto nuevo; la de unidad usa `unitDetailIndex`. */
  readonly ventana = signal<'proveedor' | 'producto' | null>(null);
  readonly productoIndex = signal<number | null>(null);
  readonly unitError = signal<string | null>(null);
  readonly providerSaving = signal(false);
  readonly providerError = signal<string | null>(null);
  readonly unitDetailIndex = signal<number | null>(null);
  // Detalles compactados; se guarda el control para que quitar uno no altere a los demás.
  readonly contraidos = signal(new Set<AbstractControl>());
  readonly unitSaving = signal(false);
  readonly documentTypes = ['TICKET', 'FACTURA'];
  readonly tasasIeps = TASAS_IEPS;
  readonly tasasIva = TASAS_IVA;
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
    ajustar_ieps: [false],
    ieps_ajustado: [0, Validators.min(0)],
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
  readonly productForm = this.fb.nonNullable.group({ nombre: ['', Validators.required], codigo_barras: [''] });
  /** Atajos de la ventana de unidad: solo prellenan nombre, símbolo y piezas. */
  readonly unitPresets = [
    { nombre: 'Pieza', simbolo: 'pz', factor_a_base: 1 },
    { nombre: 'Caja 6', simbolo: 'cj6', factor_a_base: 6 },
    { nombre: 'Caja 12', simbolo: 'cj12', factor_a_base: 12 },
    { nombre: 'Caja 24', simbolo: 'cj24', factor_a_base: 24 },
  ];
  /** El número de cuenta solo aplica a pagos que salen de una cuenta. */
  readonly pideCuenta = computed(() => {
    this.calculationVersion();
    return ['TRANSFERENCIA', 'TARJETA'].includes(this.form.controls.forma_pago.value);
  });
  readonly packagingUnits = computed(() =>
    this.units().filter(
      (unit) =>
        unit.tipo === 'EMPAQUE' &&
        Number.isInteger(Number(unit.factor_a_base)) &&
        unit.factor_a_base >= 1,
    ),
  );
  // Todos los productos del negocio: primero los que surte el proveedor elegido.
  readonly productOptions = computed<PurchaseProductGroup[]>(() => {
    const products = this.products();
    const groups: PurchaseProductGroup[] = [
      { label: 'De este proveedor', items: products.filter((product) => product.del_proveedor) },
      { label: 'Otros productos', items: products.filter((product) => !product.del_proveedor) },
    ];
    return groups.filter((group) => group.items.length > 0);
  });
  readonly unitOptions = computed(() =>
    this.packagingUnits().map((unit) => {
      const piezas = Number(unit.factor_a_base);
      return { ...unit, etiqueta: `${unit.nombre} (${piezas} ${piezas === 1 ? 'pieza' : 'piezas'})` };
    }),
  );
  /** Unidad por omisión de un renglón nuevo: «Pieza», o la primera de una pieza. */
  readonly piezaId = computed(() => {
    const unidades = this.packagingUnits().filter((unit) => Number(unit.factor_a_base) === 1);
    return (unidades.find((unit) => unit.nombre.trim().toLowerCase() === 'pieza') ?? unidades[0])?.id ?? '';
  });
  readonly sucursalNombre = computed(() => this.context.sucursal()?.nombre ?? 'la sucursal');
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
  /** IEPS e IVA por producto, con la misma regla que el backend. */
  readonly filasImpuestos = computed(() => {
    this.calculationVersion();
    const ajuste = this.form.controls.ajustar_ieps.value ? Number(this.form.controls.ieps_ajustado.value) : null;
    return calcularImpuestos(
      this.details.controls.map((control, index) => ({
        subtotal: this.detailSubtotal(index),
        piezas: this.pieces(index),
        tasaIeps: control.get('aplica_ieps')?.value ? Number(control.get('tasa_ieps')?.value ?? 0) : 0,
        tasaIva: Number(control.get('tasa_iva')?.value ?? 0),
      })),
      ajuste,
    );
  });
  /** IEPS calculado con las tasas, sin el ajuste del ticket. */
  readonly iepsCalculado = computed(() => {
    this.calculationVersion();
    return money(
      this.details.controls.reduce(
        (total, control, index) =>
          total + (control.get('aplica_ieps')?.value ? money(this.detailSubtotal(index) * Number(control.get('tasa_ieps')?.value ?? 0) / 100) : 0),
        0,
      ),
    );
  });
  readonly ieps = computed(() => money(this.filasImpuestos().reduce((total, fila) => total + fila.ieps, 0)));
  readonly iva = computed(() => money(this.filasImpuestos().reduce((total, fila) => total + fila.iva, 0)));
  readonly desgloseIeps = computed(() => desglosar(this.filasImpuestos(), 'ieps'));
  readonly desgloseIva = computed(() => desglosar(this.filasImpuestos(), 'iva'));
  readonly diferenciaIeps = computed(() => money(this.ieps() - this.iepsCalculado()));
  readonly taxes = computed(() => this.money(this.ieps() + this.iva()));
  readonly total = computed(() => this.money(this.subtotal() + this.taxes()));
  get details(): FormArray {
    return this.form.controls.detalles;
  }
  constructor() {
    this.form.valueChanges.subscribe(() => this.calculationVersion.update((value) => value + 1));
    this.form.controls.forma_pago.valueChanges.subscribe((forma) => {
      if (!['TRANSFERENCIA', 'TARJETA'].includes(forma)) this.form.controls.numero_cuenta.setValue('');
    });
    if (this.createMode) {
      this.loadForm();
    } else {
      // Solo un cambio real de negocio o sucursal recarga; la revalidación del contexto
      // reemplaza los objetos con los mismos IDs.
      const contextoLista = computed(() => `${this.businessId}:${this.context.sucursal()?.id ?? ''}`);
      effect(() => {
        contextoLista();
        untracked(() => this.loadList());
      });
    }
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
      aplica_ieps: [false],
      tasa_ieps: [8],
      tasa_iva: [16],
    });
  }
  abrirAccion(accion: { tipo: 'ver' | 'editar' | 'suspender' | 'reactivar'; compra: PurchaseRow }): void {
    this.accion.set(accion);
  }
  recargar(mensaje?: string): void {
    this.aviso.set(mensaje ?? null);
    this.loadList();
  }
  /** «Ver compra» desde Validar costos llega con `?compra=<id>`. */
  private abrirDesdeEnlace(compras: PurchaseRow[]): void {
    const id = this.route.snapshot.queryParamMap?.get('compra');
    if (!id || this.accion()) return;
    const compra = compras.find((c) => c.id === id);
    if (compra) this.accion.set({ tipo: 'ver', compra });
  }
  private loadList(): void {
    const branchId = this.context.sucursal()?.id;
    if (!branchId) {
      this.purchases.set([]);
      this.loading.set(false);
      return;
    }
    this.loading.set(true);
    this.error.set(null);
    this.service.list(this.businessId, branchId).subscribe({
      next: (r) => {
        this.purchases.set(r.items);
        this.loading.set(false);
        this.abrirDesdeEnlace(r.items);
        if (!r.items.length) this.cargarListo();
      },
      error: () => {
        this.error.set('No fue posible cargar las compras.');
        this.loading.set(false);
      },
    });
  }
  /** Sin compras: muestra lo que el negocio ya tiene listo (lo sembrado o lo que haya creado). */
  private cargarListo(): void {
    forkJoin({ units: this.service.units(this.businessId), providers: this.service.providers(this.businessId) }).subscribe({
      next: (r) => {
        this.units.set(r.units);
        this.providers.set(r.providers);
      },
      error: () => undefined,
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
        this.details.controls.forEach((detail) => this.preseleccionarPieza(detail));
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
        error: () => this.error.set('No fue posible cargar los productos.'),
      });
  }
  abrirProveedor(): void {
    this.providerForm.reset({ nombre: '' });
    this.providerError.set(null);
    this.ventana.set('proveedor');
  }
  cerrarVentana(): void {
    this.ventana.set(null);
    this.productoIndex.set(null);
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
            this.ventana.set(null);
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
    const detail = this.newDetail();
    this.preseleccionarPieza(detail);
    this.details.push(detail);
  }
  private preseleccionarPieza(detail: AbstractControl): void {
    const unidad = detail.get('unidad_medida_id');
    if (unidad && !unidad.value && this.piezaId()) unidad.setValue(this.piezaId());
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
      this.llenarTasas(index, this.products().find((item) => item.id === value));
    }
    product?.updateValueAndValidity();
    newProductName?.updateValueAndValidity();
  }
  /** Llena IEPS e IVA con lo guardado en el producto; sin dato: sin IEPS e IVA 16 %. */
  private llenarTasas(index: number, product: PurchaseProduct | undefined): void {
    const ieps = Number(product?.tasa_ieps ?? 0) * 100;
    const iva = product?.tasa_iva === null || product?.tasa_iva === undefined ? 16 : Number(product.tasa_iva) * 100;
    this.details.at(index).patchValue({
      aplica_ieps: ieps > 0,
      tasa_ieps: ieps > 0 ? Math.round(ieps * 10) / 10 : 8,
      tasa_iva: Math.round(iva),
    });
  }
  /** Abre o cierra el importe de IEPS del ticket; al abrir arranca con lo calculado. */
  alternarAjusteIeps(): void {
    const abrir = !this.form.controls.ajustar_ieps.value;
    this.form.patchValue({ ajustar_ieps: abrir, ieps_ajustado: this.iepsCalculado() });
  }
  costoPieza(index: number): number {
    return this.filasImpuestos()[index]?.costoPieza ?? 0;
  }
  iepsRenglon(index: number): number {
    return this.filasImpuestos()[index]?.ieps ?? 0;
  }
  ivaRenglon(index: number): number {
    return this.filasImpuestos()[index]?.iva ?? 0;
  }
  /** Captura el producto nuevo del renglón; se crea al guardar la compra, como antes. */
  abrirProducto(index: number): void {
    const row = this.details.at(index);
    this.productForm.reset({
      nombre: this.isNewProduct(index) ? row.get('nombre_producto_nuevo')?.value ?? '' : '',
      codigo_barras: this.isNewProduct(index) ? row.get('codigo_barras_nuevo')?.value ?? '' : '',
    });
    this.productoIndex.set(index);
    this.ventana.set('producto');
  }
  usarProducto(): void {
    const index = this.productoIndex();
    this.productForm.markAllAsTouched();
    if (index === null || this.productForm.invalid) return;
    const row = this.details.at(index);
    const value = this.productForm.getRawValue();
    row.get('nombre_producto_nuevo')?.setValue(value.nombre.trim());
    row.get('codigo_barras_nuevo')?.setValue(value.codigo_barras.trim());
    this.selectProduct(index, '__new_product__');
    this.cerrarVentana();
  }
  quitarProductoNuevo(index: number): void {
    const row = this.details.at(index);
    row.get('nombre_producto_nuevo')?.setValue('');
    row.get('codigo_barras_nuevo')?.setValue('');
    this.selectProduct(index, '');
  }
  usarPreset(preset: { nombre: string; simbolo: string; factor_a_base: number }): void {
    this.unitForm.patchValue(preset);
  }
  startUnit(index: number): void {
    this.unitError.set(null);
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
            this.unitError.set('La unidad fue creada, pero no fue posible recargarla.');
          },
        }),
      error: (r) => {
        this.unitSaving.set(false);
        this.unitError.set(r.error?.mensaje ?? 'No fue posible crear la unidad.');
      },
    });
  }
  // El producto elegido aún no está ligado al proveedor; se liga al guardar la compra.
  seLigaraAlProveedor(index: number): boolean {
    const id = this.details.at(index).get('producto_negocio_id')?.value;
    return !this.isNewProduct(index) && this.products().some((product) => product.id === id && !product.del_proveedor);
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
    const { ajustar_ieps, ieps_ajustado, ...rest } = value;
    this.saving.set(true);
    this.error.set(null);
    this.service
      .create(this.businessId, {
        ...rest,
        // Las casillas globales quedan en el modelo por compatibilidad; los impuestos van por renglón.
        incluye_ieps: false,
        ieps: 0,
        incluye_iva: false,
        ieps_ajustado: ajustar_ieps ? money(Number(ieps_ajustado)) : null,
        numero_cuenta: value.numero_cuenta || null,
        observaciones: value.observaciones || null,
        detalles: value.detalles.map(({ tiene_anomalias, aplica_ieps, tasa_ieps, tasa_iva, ...d }) => ({
          ...d,
          tasa_ieps: aplica_ieps ? Number(tasa_ieps) / 100 : 0,
          tasa_iva: Number(tasa_iva) / 100,
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

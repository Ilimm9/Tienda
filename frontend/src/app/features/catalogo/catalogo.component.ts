import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal, viewChild } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute, Router } from '@angular/router';
import { FormBuilder, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { SelectModule } from 'primeng/select';
import { firstValueFrom, startWith } from 'rxjs';
import { ContextoService } from '../../contexto/contexto.service';
import { CatalogoService } from './catalogo.service';
import { CatalogImportResult, CatalogRecord, Categoria, Proveedor, UnidadMedida } from './catalogo.models';
import { PERMISOS } from '../../contexto/permisos';
import { PuedeDirective } from '../../contexto/puede.directive';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { ModalComponent } from '../../shared/ui/modal/modal.component';
import {
  EstadoRenglon,
  normalizar,
  RenglonLista,
  revisarLista,
  SeccionCatalogo,
} from './lista-pegada';

interface SelectOption {
  label: string;
  value: string;
}

type CatalogStatusFilter = 'all' | 'active' | 'inactive';
type Dialogo = 'formulario' | 'masiva' | null;
type OrigenMasiva = 'pegar' | 'archivo';
const NO_PARENT_CATEGORY = '__sin_categoria_padre__';

const ETIQUETAS_ESTADO: Record<EstadoRenglon, string> = {
  nuevo: 'Nuevo',
  existe: 'Ya existe',
  repetido: 'Repetido',
  error: 'Revisar',
};

@Component({
  selector: 'app-catalogo',
  standalone: true,
  imports: [PuedeDirective, CommonModule, FormsModule, ReactiveFormsModule, SelectModule, ModalComponent],
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
  private readonly modal = viewChild(ModalComponent);
  readonly section = this.route.snapshot.data['section'] as SeccionCatalogo;
  /** Las rutas …/nuevo, …/importar y …/:id/editar abren la ventana encima de la lista. */
  readonly pageMode = (this.route.snapshot.data['mode'] as 'list' | 'create' | 'edit' | 'import' | undefined) ?? 'list';
  readonly items = signal<CatalogRecord[]>([]);
  readonly parents = signal<Categoria[]>([]);
  readonly catalogSearch = signal('');
  readonly catalogStatusFilter = signal<CatalogStatusFilter>('all');
  readonly categoryParentFilter = signal<string | null>(null);
  readonly unitTypeFilter = signal<string | null>(null);
  readonly catalogPageFirst = signal(0);
  readonly catalogStatusOptions: Array<{ label: string; value: CatalogStatusFilter }> = [
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
  readonly catalogRowsOptions = [10, 25, 50];
  readonly catalogRows = signal(10);
  readonly paginatedItems = computed(() =>
    this.filteredItems().slice(this.catalogPageFirst(), this.catalogPageFirst() + this.catalogRows()),
  );
  readonly catalogPage = computed(() => Math.floor(this.catalogPageFirst() / this.catalogRows()));
  readonly catalogTotalPages = computed(() =>
    Math.max(1, Math.ceil(this.filteredItems().length / this.catalogRows())),
  );
  readonly catalogRange = computed(() => {
    const total = this.filteredItems().length;
    if (!total) return 'Mostrando 0 registros';
    return `Mostrando ${this.catalogPageFirst() + 1} a ${Math.min(total, this.catalogPageFirst() + this.catalogRows())} de ${total} registros`;
  });
  /** Cuántos registros caen en cada pestaña de estado. */
  readonly catalogStatusCounts = computed<Record<CatalogStatusFilter, number>>(() => {
    const activos = this.items().filter((item) => item.activo).length;
    return { all: this.items().length, active: activos, inactive: this.items().length - activos };
  });
  private readonly categoryNames = computed(
    () => new Map(this.items().map((item) => [item.id, item as Categoria])),
  );
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  editingId: string | null = null;
  readonly saving = signal(false);
  readonly deletingId = signal<string | null>(null);
  readonly formError = signal<string | null>(null);

  readonly dialogo = signal<Dialogo>(null);
  readonly seguirAgregando = signal(false);
  readonly origenMasiva = signal<OrigenMasiva>('pegar');
  readonly textoLista = signal('');

  importFile: File | null = null;
  importDropActive = false;
  readonly importing = signal(false);
  readonly importError = signal<string | null>(null);
  readonly importResult = signal<CatalogImportResult | null>(null);
  readonly pegadoFallas = signal<Array<{ nombre: string; motivo: string }>>([]);

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
  private readonly valores = toSignal(this.form.valueChanges.pipe(startWith(this.form.getRawValue())), {
    initialValue: this.form.getRawValue(),
  });

  /** Aviso inmediato de nombre repetido; el servidor lo vuelve a comprobar. */
  readonly nombreDuplicado = computed<string | null>(() => {
    const valores = this.valores();
    const nombre = normalizar(valores.nombre);
    if (!nombre) return null;
    const padre = valores.categoria_padre_id || null;
    const igual = this.items().find(
      (item) =>
        item.id !== this.editingId &&
        normalizar(item.nombre) === nombre &&
        (this.section !== 'categorias' || ((item as Categoria).categoria_padre_id ?? null) === padre),
    );
    return igual?.nombre ?? null;
  });

  readonly rutaCategoria = computed(() => {
    const valores = this.valores();
    const nombre = valores.nombre?.trim() || 'Nueva categoría';
    const padre = this.categoryNames().get(valores.categoria_padre_id ?? '') as Categoria | undefined;
    return padre ? [...this.categoryPath(padre), nombre].join(' › ') : `${nombre} (principal)`;
  });

  readonly simboloBase = computed(() => {
    const tipo = this.valores().tipo;
    const base = this.items().find(
      (item) => (item as UnidadMedida).tipo === tipo && Number((item as UnidadMedida).factor_a_base) === 1,
    ) as UnidadMedida | undefined;
    return base?.simbolo ?? 'de la unidad base';
  });

  readonly ejemploCaja = computed(() => {
    const valores = this.valores();
    const decimales = Math.min(6, Math.max(0, Number(valores.decimales) || 0));
    return `${(1.25).toFixed(decimales)} ${valores.simbolo?.trim() || 'unidad'}`;
  });

  readonly renglones = computed<RenglonLista[]>(() =>
    revisarLista(this.section, this.textoLista(), this.items(), (categoria) => this.categoryPath(categoria)),
  );
  readonly resumenLista = computed(() => {
    const conteo: Record<EstadoRenglon, number> = { nuevo: 0, existe: 0, repetido: 0, error: 0 };
    for (const renglon of this.renglones()) conteo[renglon.estado]++;
    return conteo;
  });
  readonly chipsResumen = computed(() =>
    (['nuevo', 'existe', 'repetido', 'error'] as EstadoRenglon[])
      .filter((estado) => this.resumenLista()[estado] > 0)
      .map((estado) => ({ estado, texto: `${this.resumenLista()[estado]} ${this.textoResumen(estado, this.resumenLista()[estado])}` })),
  );
  readonly nuevosLista = computed(() => this.renglones().filter((renglon) => renglon.estado === 'nuevo'));

  readonly dialogoSucio = computed(() => {
    if (this.dialogo() === 'masiva') return Boolean(this.textoLista().trim()) || Boolean(this.importFile);
    return this.dialogo() === 'formulario' && this.form.dirty;
  });

  constructor() {
    this.load();
    if (this.pageMode === 'create') this.openCreate();
    else if (this.pageMode === 'import') this.openImport();
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
  /** «marca», «categoría», «proveedor» o «unidad de medida», para frases. */
  get singularLower(): string {
    return this.singularTitle.toLowerCase();
  }
  get esFemenino(): boolean {
    return this.section !== 'proveedores';
  }
  get description(): string {
    return this.section === 'marcas'
      ? 'Administra las marcas de tus productos.'
      : this.section === 'categorias'
        ? 'Organiza los productos por categorías.'
        : this.section === 'proveedores' ? 'Administra tus proveedores de inventario.' : 'Administra las unidades normalizadas del catálogo.';
  }
  get tituloFormulario(): string {
    if (this.editingId) return `Editar ${this.singularLower}`;
    return `${this.esFemenino ? 'Nueva' : 'Nuevo'} ${this.singularLower}`;
  }
  get subtituloFormulario(): string {
    return this.section === 'marcas'
      ? 'Aparecerá al registrar productos.'
      : this.section === 'categorias'
        ? 'Agrupa tus productos para encontrarlos y reportarlos.'
        : this.section === 'proveedores'
          ? 'Solo el nombre es obligatorio; lo demás puede esperar.'
          : 'Cómo se mide lo que compras y vendes.';
  }
  get formatoLista(): { ayuda: string; ejemplo: string } {
    if (this.section === 'categorias') return { ayuda: 'Una por renglón. Usa > para indicar dónde va.', ejemplo: 'Bebidas > Aguas' };
    if (this.section === 'unidades') return { ayuda: 'Nombre, símbolo, tipo, equivalencia y decimales (opcional).', ejemplo: 'Caja, cja, cantidad, 12' };
    if (this.section === 'proveedores') return { ayuda: 'Nombre, RFC, teléfono, correo. Solo el nombre es obligatorio.', ejemplo: 'Lala, , 81 2233 4455, ventas@lala.mx' };
    return { ayuda: 'Una marca por renglón.', ejemplo: 'Peñafiel' };
  }
  get columnasPlantilla(): string {
    if (this.section === 'categorias') return 'Nombre, Categoría padre, Descripción';
    if (this.section === 'unidades') return 'Nombre, Símbolo, Tipo, Factor a base, Decimales';
    if (this.section === 'proveedores') return 'Nombre, Razón social, RFC, Teléfono, Correo, Dirección';
    return 'Nombre';
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
        if (this.pageMode === 'edit' && !this.dialogo() && !this.editingId) this.abrirEdicionDeRuta(v);
      },
      error: () => {
        this.error.set(`No fue posible cargar ${this.title.toLowerCase()}.`);
        this.loading.set(false);
      },
    });
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

  setCatalogRows(rows: number | string): void {
    this.catalogRows.set(Number(rows) || 10);
    this.resetCatalogPagination();
  }

  goToCatalogPage(page: number): void {
    const last = this.catalogTotalPages() - 1;
    this.onCatalogTablePage(Math.min(Math.max(0, page), last) * this.catalogRows());
  }

  initials(item: CatalogRecord): string {
    return item.nombre
      .trim()
      .split(/\s+/)
      .map((word) => word.charAt(0))
      .join('')
      .slice(0, 2)
      .toUpperCase();
  }

  unitTypeLabel(tipo: string): string {
    return this.unitTypeOptions.find((option) => option.value === tipo)?.label ?? tipo;
  }

  /** «1 kg = 1000 g» usando la unidad base del mismo tipo, o «Unidad base» si factor es 1. */
  unitEquivalence(unit: UnidadMedida): string {
    if (Number(unit.factor_a_base) === 1) return 'Unidad base';
    const base = this.items().find(
      (item) => (item as UnidadMedida).tipo === unit.tipo && Number((item as UnidadMedida).factor_a_base) === 1,
    ) as UnidadMedida | undefined;
    return `1 ${unit.simbolo} = ${unit.factor_a_base} ${base?.simbolo ?? 'de la base'}`;
  }

  /** Ruta «Padre › Hija» para mostrar dónde vive cada categoría. */
  categoryPath(category: Categoria): string[] {
    const path = [category.nombre];
    let parentId = category.categoria_padre_id;
    // Tope de niveles para no ciclar si los datos traen una referencia circular.
    while (parentId && path.length < 6) {
      const parent = this.categoryNames().get(parentId);
      if (!parent) break;
      path.unshift(parent.nombre);
      parentId = parent.categoria_padre_id;
    }
    return path;
  }

  /* ---------- Ventana individual ---------- */

  openCreate(): void {
    this.editingId = null;
    this.form.reset();
    this.formError.set(null);
    this.dialogo.set('formulario');
  }

  openEdit(item: CatalogRecord): void {
    this.prepareEdit(item);
    this.dialogo.set('formulario');
  }

  private abrirEdicionDeRuta(items: CatalogRecord[]): void {
    const itemId = this.route.snapshot.paramMap?.get('id');
    const item = items.find((candidate) => candidate.id === itemId);
    if (!item) {
      this.feedback.error(`No encontramos ${this.esFemenino ? 'esa' : 'ese'} ${this.singularLower}`, 'Puede que se haya eliminado.');
      this.goToList();
      return;
    }
    this.openEdit(item);
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
    const duplicado = this.nombreDuplicado();
    if (duplicado) {
      this.formError.set(`Ya existe «${duplicado}». Usa ${this.esFemenino ? 'la' : 'el'} que ya tienes o cambia el nombre.`);
      return;
    }
    const v = this.form.getRawValue();
    if (this.section === 'unidades' && (!v.simbolo.trim() || !v.tipo.trim())) {
      this.formError.set('Símbolo y tipo son obligatorios para la unidad.');
      return;
    }
    this.saving.set(true);
    this.formError.set(null);
    const payload = this.payloadFormulario();
    const negocioId = this.contexto.negocio()?.id ?? '';
    const editando = Boolean(this.editingId);
    const request = this.editingId
      ? this.service.actualizar(negocioId, this.requestPath(), this.editingId, payload)
      : this.service.crear(negocioId, this.requestPath(), payload);
    request.subscribe({
      next: () => {
        this.saving.set(false);
        const nombre = v.nombre.trim();
        this.feedback.success(
          editando ? 'Cambios guardados' : `${this.singularTitle} ${this.esFemenino ? 'agregada' : 'agregado'}`,
          `«${nombre}» ya está en ${this.title.toLowerCase()}.`,
        );
        this.load();
        if (!editando && this.seguirAgregando()) {
          this.form.reset({ tipo: v.tipo, factor_a_base: 1, decimales: v.decimales, categoria_padre_id: v.categoria_padre_id });
          this.modal()?.enfocarInicio();
          return;
        }
        this.cerrarDialogo();
      },
      error: (e) => {
        this.saving.set(false);
        this.formError.set(
          e.error?.mensaje ?? 'No fue posible comunicarse con el servidor. Intenta de nuevo.',
        );
      },
    });
  }

  private payloadFormulario(): Record<string, unknown> {
    const v = this.form.getRawValue();
    const payload: Record<string, unknown> = { nombre: v.nombre.trim() };
    if (this.section === 'categorias')
      Object.assign(payload, {
        descripcion: v.descripcion || null,
        categoria_padre_id: v.categoria_padre_id || null,
      });
    if (this.section === 'proveedores')
      Object.assign(payload, {
        razon_social: v.razon_social || null,
        rfc: v.rfc ? v.rfc.toUpperCase() : null,
        telefono: v.telefono || null,
        email: v.email || null,
        direccion: v.direccion || null,
      });
    if (this.section === 'unidades')
      Object.assign(payload, { simbolo: v.simbolo.trim(), tipo: v.tipo, factor_a_base: v.factor_a_base, decimales: v.decimales });
    return payload;
  }

  hasError(name: string): boolean {
    const c = this.form.get(name);
    return !!c && c.invalid && (c.dirty || c.touched);
  }

  /* ---------- Ventana masiva ---------- */

  openImport(): void {
    this.prepareImport();
    this.textoLista.set('');
    this.pegadoFallas.set([]);
    this.origenMasiva.set('pegar');
    this.dialogo.set('masiva');
  }

  private prepareImport(): void {
    this.importFile = null;
    this.importDropActive = false;
    this.importError.set(null);
    this.importResult.set(null);
  }

  etiquetaEstado(estado: EstadoRenglon): string {
    return ETIQUETAS_ESTADO[estado];
  }

  textoResumen(estado: EstadoRenglon, cantidad: number): string {
    if (estado === 'nuevo') return cantidad === 1 ? 'nuevo' : 'nuevos';
    if (estado === 'existe') return cantidad === 1 ? 'ya existe' : 'ya existen';
    if (estado === 'repetido') return cantidad === 1 ? 'repetido' : 'repetidos';
    return 'por revisar';
  }

  get textoBotonMasiva(): string {
    if (this.origenMasiva() === 'archivo') return this.importing() ? 'Importando…' : 'Importar archivo';
    if (this.importing()) return 'Agregando…';
    const n = this.nuevosLista().length;
    return n ? `Agregar ${n} ${n === 1 ? this.singularLower : this.title.toLowerCase()}` : 'Nada que agregar';
  }

  get notaMasiva(): string {
    if (this.origenMasiva() === 'archivo') return 'Revisamos duplicados al importar.';
    const omitidos = this.renglones().length - this.nuevosLista().length;
    if (!this.renglones().length) return '';
    return omitidos ? `${omitidos} ${omitidos === 1 ? 'renglón se omite' : 'renglones se omiten'}` : 'Todo listo para agregar';
  }

  bloquearMasiva(): boolean {
    if (this.importing()) return true;
    return this.origenMasiva() === 'archivo' ? !this.importFile : this.nuevosLista().length === 0;
  }

  ejecutarMasiva(): void {
    if (this.origenMasiva() === 'archivo') this.importCatalog();
    else void this.agregarLista();
  }

  /** Crea en orden los renglones nuevos; en categorías resuelve o crea primero los padres. */
  async agregarLista(): Promise<void> {
    const nuevos = this.nuevosLista();
    if (!nuevos.length || this.importing()) return;
    const omitidos = this.renglones().length - nuevos.length;
    this.importing.set(true);
    this.pegadoFallas.set([]);
    const negocioId = this.contexto.negocio()?.id ?? '';
    const fallas: Array<{ nombre: string; motivo: string }> = [];
    let creados = 0;
    let categorias = this.section === 'categorias' ? [...(this.items() as Categoria[])] : [];

    for (const renglon of nuevos) {
      try {
        const payload: Record<string, unknown> = { ...renglon.payload };
        if (this.section === 'categorias' && renglon.padres?.length) {
          let padreId: string | null = null;
          for (const nombrePadre of renglon.padres) {
            let padre = categorias.find(
              (c) => normalizar(c.nombre) === normalizar(nombrePadre) && (c.categoria_padre_id ?? null) === padreId,
            );
            if (!padre) {
              await firstValueFrom(this.service.crear(negocioId, 'categorias', { nombre: nombrePadre, categoria_padre_id: padreId }));
              creados++;
              categorias = await firstValueFrom(this.service.categorias(negocioId));
              padre = categorias.find(
                (c) => normalizar(c.nombre) === normalizar(nombrePadre) && (c.categoria_padre_id ?? null) === padreId,
              );
            }
            padreId = padre?.id ?? null;
          }
          payload['categoria_padre_id'] = padreId;
        }
        await firstValueFrom(this.service.crear(negocioId, this.requestPath(), payload));
        creados++;
        if (this.section === 'categorias') categorias = await firstValueFrom(this.service.categorias(negocioId));
      } catch (e: unknown) {
        const mensaje = (e as { error?: { mensaje?: string } })?.error?.mensaje;
        fallas.push({ nombre: renglon.nombre, motivo: mensaje ?? 'No se pudo guardar.' });
      }
    }

    this.importing.set(false);
    this.load();
    const plural = this.title.toLowerCase();
    if (fallas.length) {
      this.pegadoFallas.set(fallas);
      this.textoLista.set(fallas.map((f) => nuevos.find((r) => r.nombre === f.nombre)?.linea ?? f.nombre).join('\n'));
      this.feedback.warning(
        `Se agregaron ${creados} de ${creados + fallas.length}`,
        'Dejamos en la lista los renglones que no se pudieron guardar.',
      );
      return;
    }
    this.textoLista.set('');
    this.feedback.success(
      `${creados} ${creados === 1 ? this.singularLower : plural} ${creados === 1 ? (this.esFemenino ? 'agregada' : 'agregado') : this.esFemenino ? 'agregadas' : 'agregados'}`,
      omitidos ? 'Los repetidos y los que ya existían se omitieron.' : undefined,
    );
    this.cerrarDialogo();
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
    return this.service.plantillaUrl(this.contexto.negocio()?.id ?? '', this.section);
  }
  importCatalog(): void {
    if (!this.importFile || !this.canImport()) return;
    this.importing.set(true);
    this.importError.set(null);
    this.importResult.set(null);
    this.service.importar(this.contexto.negocio()?.id ?? '', this.section, this.importFile).subscribe({
      next: (result) => {
        this.importing.set(false);
        this.load();
        const resumen = `${result.creadas} agregados · ${result.omitidas} omitidos · ${result.invalidas} con errores`;
        if (result.invalidas > 0 || result.errores.length) {
          // Se queda abierta para que se vean los renglones con error.
          this.importResult.set(result);
          this.importFile = null;
          this.feedback.warning('Importación con observaciones', resumen);
          return;
        }
        if (result.omitidas > 0) this.feedback.warning('Importación terminada', resumen);
        else this.feedback.success('Importación terminada', resumen);
        this.cerrarDialogo();
      },
      error: (error) => {
        this.importing.set(false);
        this.importError.set(error.error?.mensaje ?? 'No fue posible importar el archivo.');
      },
    });
  }

  /* ---------- Cierre y navegación ---------- */

  /** Cancelar pasa por la ventana para confirmar si hay datos sin guardar. */
  modalCerrar(): void {
    const modal = this.modal();
    if (modal) void modal.solicitarCierre();
    else this.cerrarDialogo();
  }


  cerrarDialogo(): void {
    this.dialogo.set(null);
    this.editingId = null;
    this.seguirAgregando.set(false);
    // Si se entró por …/nuevo, …/importar o …/editar, la URL vuelve a la lista.
    if (this.pageMode !== 'list') this.goToList();
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

  private requestPath(): string {
    return this.section === 'unidades' ? 'unidades-medida' : this.section;
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

    const query = normalizar(this.catalogSearch());
    if (!query) return true;
    return this.catalogSearchableValues(item).some((value) => normalizar(value).includes(query));
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

  canImport(): boolean {
    return this.section === 'marcas' || this.section === 'categorias' || this.section === 'proveedores' || this.section === 'unidades';
  }

  /** Abre la ventana para agregar sin salir de la lista. */
  navigateToCreate(): void {
    this.openCreate();
  }

  navigateToEdit(item: CatalogRecord): void {
    this.openEdit(item);
  }

  navigateToImport(): void {
    this.openImport();
  }

  async deleteCatalogItem(item: CatalogRecord): Promise<void> {
    if (this.deletingId()) return;
    this.deletingId.set(item.id);
    const negocioId = this.contexto.negocio()?.id ?? '';
    const requestPath = this.requestPath();
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
          titulo: `Eliminar ${this.singularLower}`,
          descripcion: `Se eliminará “${item.nombre}” porque no tiene vinculaciones.`,
          textoConfirmar: 'Sí, eliminar',
        });
        if (!confirmed) return;
        this.deletingId.set(item.id);
        this.service.eliminar(negocioId, requestPath, item.id).subscribe({
          next: () => {
            this.deletingId.set(null);
            this.feedback.success(`${this.singularTitle} ${this.esFemenino ? 'eliminada' : 'eliminado'}`);
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

  goToList(): void {
    void this.router?.navigate([this.catalogBasePath()], { replaceUrl: true });
  }

  private catalogBasePath(): string {
    if (this.section === 'proveedores') return '/proveedores';
    return this.section === 'unidades' ? '/catalogo/unidades-medida' : `/catalogo/${this.section}`;
  }
}

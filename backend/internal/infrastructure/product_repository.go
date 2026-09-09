package infrastructure

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"tienda/backend/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var productBarcodePattern = regexp.MustCompile(`^\d{8,14}$`)

type ProductRepository struct {
	db *gorm.DB
}

func (r *ProductRepository) ListCategories() ([]domain.CatalogOption, error) {
	var options []domain.CatalogOption
	err := r.db.Table("categorias").Select("id, nombre").Where("activo = TRUE").Order("nombre ASC").Scan(&options).Error
	return options, err
}

func (r *ProductRepository) ListBrands() ([]domain.CatalogOption, error) {
	var options []domain.CatalogOption
	err := r.db.Table("marcas").Select("id, nombre").Where("activo = TRUE").Order("nombre ASC").Scan(&options).Error
	return options, err
}

func (r *ProductRepository) ListUnits() ([]domain.UnidadMedida, error) {
	items := make([]domain.UnidadMedida, 0)
	err := r.db.Where("activo = TRUE").Order("tipo ASC, nombre ASC").Find(&items).Error
	return items, err
}

func (r *ProductRepository) CreateUnit(i domain.CreateUnidadMedidaInput) error {
	i.Codigo = strings.TrimSpace(strings.ToLower(i.Codigo))
	i.Nombre = strings.TrimSpace(i.Nombre)
	i.Simbolo = strings.TrimSpace(i.Simbolo)
	i.Tipo = strings.TrimSpace(strings.ToUpper(i.Tipo))
	if i.Codigo == "" || i.Nombre == "" || i.Simbolo == "" || i.Tipo == "" || i.FactorABase <= 0 || i.Decimales < 0 || i.Decimales > 6 {
		return errors.New("los datos de la unidad no son válidos")
	}
	fraccion := true
	if i.PermiteFraccion != nil {
		fraccion = *i.PermiteFraccion
	}
	return r.db.Create(&domain.UnidadMedida{Codigo: i.Codigo, Nombre: i.Nombre, Simbolo: i.Simbolo, Tipo: i.Tipo, UnidadBaseID: i.UnidadBaseID, FactorABase: i.FactorABase, PermiteFraccion: fraccion, Decimales: i.Decimales, Activo: true}).Error
}

func (r *ProductRepository) UpdateUnit(id uuid.UUID, i domain.UpdateUnidadMedidaInput) error {
	values := map[string]interface{}{}
	if i.Codigo != nil {
		values["codigo"] = strings.TrimSpace(strings.ToLower(*i.Codigo))
	}
	if i.Nombre != nil {
		values["nombre"] = strings.TrimSpace(*i.Nombre)
	}
	if i.Simbolo != nil {
		values["simbolo"] = strings.TrimSpace(*i.Simbolo)
	}
	if i.Tipo != nil {
		values["tipo"] = strings.TrimSpace(strings.ToUpper(*i.Tipo))
	}
	if i.UnidadBaseID != nil {
		values["unidad_base_id"] = i.UnidadBaseID
	}
	if i.FactorABase != nil {
		values["factor_a_base"] = *i.FactorABase
	}
	if i.PermiteFraccion != nil {
		values["permite_fraccion"] = *i.PermiteFraccion
	}
	if i.Decimales != nil {
		values["decimales"] = *i.Decimales
	}
	if i.Activo != nil {
		values["activo"] = *i.Activo
	}
	if len(values) == 0 {
		return nil
	}
	result := r.db.Model(&domain.UnidadMedida{}).Where("id = ?", id).Updates(values)
	if result.RowsAffected == 0 && result.Error == nil {
		return errors.New("unidad no encontrada")
	}
	return result.Error
}

func (r *ProductRepository) ImportUnits(rows []domain.CatalogImportUnitRow) (domain.CatalogImportResult, error) {
	result := domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0)}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		current := make([]domain.UnidadMedida, 0)
		if err := tx.Find(&current).Error; err != nil {
			return err
		}
		existing := make(map[string]struct{}, len(current)*3)
		for _, unit := range current {
			existing[catalogImportKey(unit.Codigo)] = struct{}{}
			existing[catalogImportKey(unit.Nombre)] = struct{}{}
			existing[catalogImportKey(unit.Simbolo)] = struct{}{}
		}
		seen := make(map[string]struct{}, len(rows))
		created := make([]domain.UnidadMedida, 0, len(rows))
		for _, row := range rows {
			key := catalogImportKey(row.Codigo)
			if _, duplicate := seen[key]; duplicate {
				result.Omitidas++
				result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Motivo: "La unidad está repetida en el archivo"})
				continue
			}
			seen[key] = struct{}{}
			if _, exists := existing[key]; exists || containsUnitKey(existing, row.Nombre, row.Simbolo) {
				result.Omitidas++
				result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Motivo: "El código, nombre o símbolo de la unidad ya existe"})
				continue
			}
			created = append(created, domain.UnidadMedida{Codigo: strings.ToLower(strings.TrimSpace(row.Codigo)), Nombre: strings.TrimSpace(row.Nombre), Simbolo: strings.TrimSpace(row.Simbolo), Tipo: strings.ToUpper(strings.TrimSpace(row.Tipo)), FactorABase: row.FactorABase, Decimales: row.Decimales, PermiteFraccion: true, Activo: true})
			existing[key] = struct{}{}
			existing[catalogImportKey(row.Nombre)] = struct{}{}
			existing[catalogImportKey(row.Simbolo)] = struct{}{}
		}
		if len(created) > 0 {
			if err := tx.Create(&created).Error; err != nil {
				return err
			}
			result.Creadas = len(created)
		}
		return nil
	})
	return result, err
}

func containsUnitKey(existing map[string]struct{}, values ...string) bool {
	for _, value := range values {
		if _, ok := existing[catalogImportKey(value)]; ok {
			return true
		}
	}
	return false
}

func (r *ProductRepository) ListBrandsAdmin() ([]domain.Marca, error) {
	v := make([]domain.Marca, 0)
	e := r.db.Order("nombre ASC").Find(&v).Error
	return v, e
}
func (r *ProductRepository) CreateBrand(i domain.CreateMarcaInput) error {
	i.Nombre = strings.TrimSpace(i.Nombre)
	if i.Nombre == "" {
		return errors.New("el nombre de la marca es obligatorio")
	}
	return r.db.Create(&domain.Marca{Nombre: i.Nombre, Activo: true}).Error
}
func (r *ProductRepository) UpdateBrand(id uuid.UUID, i domain.UpdateMarcaInput) error {
	values := map[string]interface{}{}
	if i.Nombre != nil {
		*i.Nombre = strings.TrimSpace(*i.Nombre)
		values["nombre"] = *i.Nombre
	}
	if i.Activo != nil {
		values["activo"] = *i.Activo
	}
	if len(values) == 0 {
		return nil
	}
	result := r.db.Model(&domain.Marca{}).Where("id = ?", id).Updates(values)
	if result.RowsAffected == 0 && result.Error == nil {
		return errors.New("marca no encontrada")
	}
	return result.Error
}

func (r *ProductRepository) ImportBrands(rows []domain.CatalogImportBrandRow) (domain.CatalogImportResult, error) {
	result := domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0)}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current []domain.Marca
		if err := tx.Find(&current).Error; err != nil {
			return err
		}
		existing := make(map[string]struct{}, len(current))
		for _, brand := range current {
			existing[catalogImportKey(brand.Nombre)] = struct{}{}
		}
		seen := make(map[string]struct{}, len(rows))
		brands := make([]domain.Marca, 0, len(rows))
		for _, row := range rows {
			key := catalogImportKey(row.Nombre)
			if _, ok := seen[key]; ok {
				result.Omitidas++
				result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Motivo: "La marca está repetida en el archivo"})
				continue
			}
			seen[key] = struct{}{}
			if _, ok := existing[key]; ok {
				result.Omitidas++
				result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Motivo: "La marca ya existe"})
				continue
			}
			brands = append(brands, domain.Marca{Nombre: strings.TrimSpace(row.Nombre), Activo: true})
		}
		if len(brands) == 0 {
			return nil
		}
		if err := tx.Create(&brands).Error; err != nil {
			return err
		}
		result.Creadas = len(brands)
		return nil
	})
	return result, err
}

func (r *ProductRepository) ListCategoriesAdmin() ([]domain.Categoria, error) {
	v := make([]domain.Categoria, 0)
	e := r.db.Order("nombre ASC").Find(&v).Error
	return v, e
}
func (r *ProductRepository) CreateCategory(i domain.CreateCategoriaInput) error {
	i.Nombre = strings.TrimSpace(i.Nombre)
	if i.Nombre == "" {
		return errors.New("el nombre de la categoría es obligatorio")
	}
	return r.db.Create(&domain.Categoria{Nombre: i.Nombre, CategoriaPadreID: i.CategoriaPadreID, Descripcion: i.Descripcion, Activo: true}).Error
}
func (r *ProductRepository) UpdateCategory(id uuid.UUID, i domain.UpdateCategoriaInput) error {
	values := map[string]interface{}{}
	if i.Nombre != nil {
		*i.Nombre = strings.TrimSpace(*i.Nombre)
		values["nombre"] = *i.Nombre
	}
	if i.CategoriaPadreID != nil {
		values["categoria_padre_id"] = i.CategoriaPadreID
	}
	if i.Descripcion != nil {
		values["descripcion"] = i.Descripcion
	}
	if i.Activo != nil {
		values["activo"] = *i.Activo
	}
	if len(values) == 0 {
		return nil
	}
	result := r.db.Model(&domain.Categoria{}).Where("id = ?", id).Updates(values)
	if result.RowsAffected == 0 && result.Error == nil {
		return errors.New("categoría no encontrada")
	}
	return result.Error
}

func (r *ProductRepository) ImportCategories(rows []domain.CatalogImportCategoryRow) (domain.CatalogImportResult, error) {
	result := domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0)}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current []domain.Categoria
		if err := tx.Find(&current).Error; err != nil {
			return err
		}
		existing := make(map[string]uuid.UUID, len(current))
		for _, category := range current {
			existing[catalogImportKey(category.Nombre)] = category.ID
		}

		candidates := make(map[string]domain.CatalogImportCategoryRow, len(rows))
		for _, row := range rows {
			key := catalogImportKey(row.Nombre)
			if _, duplicate := candidates[key]; duplicate {
				result.Omitidas++
				result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Motivo: "La categoría está repetida en el archivo"})
				continue
			}
			if _, exists := existing[key]; exists {
				result.Omitidas++
				result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Motivo: "La categoría ya existe"})
				continue
			}
			candidates[key] = row
		}

		states := make(map[string]int, len(candidates))
		ordered := make([]string, 0, len(candidates))
		var visit func(string) bool
		visit = func(key string) bool {
			switch states[key] {
			case 2:
				return true
			case 3:
				return false
			case 1:
				return false
			}
			states[key] = 1
			row := candidates[key]
			parent := catalogImportKey(row.CategoriaPadre)
			if parent != "" {
				if _, exists := existing[parent]; !exists {
					if _, included := candidates[parent]; !included || !visit(parent) {
						states[key] = 3
						return false
					}
				}
			}
			states[key] = 2
			ordered = append(ordered, key)
			return true
		}
		for key := range candidates {
			visit(key)
		}

		for key, row := range candidates {
			if states[key] != 2 {
				result.Invalidas++
				result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Motivo: "La categoría padre no existe o forma una jerarquía circular"})
			}
		}

		created := make(map[string]uuid.UUID, len(ordered))
		for _, key := range ordered {
			row := candidates[key]
			parentKey := catalogImportKey(row.CategoriaPadre)
			var parentID *uuid.UUID
			if id, exists := existing[parentKey]; parentKey != "" && exists {
				parentID = &id
			} else if id, exists := created[parentKey]; parentKey != "" && exists {
				parentID = &id
			}
			category := domain.Categoria{ID: uuid.New(), Nombre: strings.TrimSpace(row.Nombre), CategoriaPadreID: parentID, Activo: true}
			if description := strings.TrimSpace(row.Descripcion); description != "" {
				category.Descripcion = &description
			}
			if err := tx.Create(&category).Error; err != nil {
				return err
			}
			created[key] = category.ID
			result.Creadas++
		}
		return nil
	})
	return result, err
}

func catalogImportKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func (r *ProductRepository) ListProviders(businessID uuid.UUID) ([]domain.Proveedor, error) {
	v := make([]domain.Proveedor, 0)
	e := r.db.Where("negocio_id = ?", businessID).Order("nombre ASC").Find(&v).Error
	return v, e
}
func (r *ProductRepository) CreateProvider(id uuid.UUID, i domain.CreateProveedorInput) error {
	i.Nombre = strings.TrimSpace(i.Nombre)
	if i.Nombre == "" {
		return errors.New("el nombre del proveedor es obligatorio")
	}
	return r.db.Create(&domain.Proveedor{NegocioID: id, Nombre: i.Nombre, RazonSocial: i.RazonSocial, RFC: i.RFC, Telefono: i.Telefono, Email: i.Email, Direccion: i.Direccion, Activo: true}).Error
}
func (r *ProductRepository) UpdateProvider(businessID, id uuid.UUID, i domain.UpdateProveedorInput) error {
	values := map[string]interface{}{}
	if i.Nombre != nil {
		*i.Nombre = strings.TrimSpace(*i.Nombre)
		values["nombre"] = *i.Nombre
	}
	if i.RazonSocial != nil {
		values["razon_social"] = i.RazonSocial
	}
	if i.RFC != nil {
		values["rfc"] = i.RFC
	}
	if i.Telefono != nil {
		values["telefono"] = i.Telefono
	}
	if i.Email != nil {
		values["email"] = i.Email
	}
	if i.Direccion != nil {
		values["direccion"] = i.Direccion
	}
	if i.Activo != nil {
		values["activo"] = *i.Activo
	}
	if len(values) == 0 {
		return nil
	}
	result := r.db.Model(&domain.Proveedor{}).Where("id = ? AND negocio_id = ?", id, businessID).Updates(values)
	if result.RowsAffected == 0 && result.Error == nil {
		return errors.New("proveedor no encontrado")
	}
	return result.Error
}

func (r *ProductRepository) ListBranches(businessID uuid.UUID) ([]domain.CatalogOption, error) {
	var options []domain.CatalogOption
	err := r.db.Table("sucursales").Select("id, nombre").Where("negocio_id = ? AND activo = TRUE", businessID).Order("nombre ASC").Scan(&options).Error
	return options, err
}

func (r *ProductRepository) Create(businessID uuid.UUID, input domain.CreateProductInput) error {
	input.Nombre = strings.TrimSpace(input.Nombre)
	input.SKUInterno = strings.TrimSpace(input.SKUInterno)
	if input.Nombre == "" || input.SKUInterno == "" || input.PrecioVenta < 0 || input.StockInicial < 0 {
		return errors.New("los datos del producto no son válidos")
	}
	barcode := optionalString(input.CodigoBarras)
	imageURL := optionalString(input.ImagenURL)
	if barcode != "" && !productBarcodePattern.MatchString(barcode) {
		return errors.New("el código de barras debe contener entre 8 y 14 dígitos")
	}
	if imageURL != "" {
		parsed, err := url.Parse(imageURL)
		if barcode == "" || err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return errors.New("la URL de imagen no es válida")
		}
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		var branch domain.Sucursal
		if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", input.SucursalID, businessID).First(&branch).Error; err != nil {
			return errors.New("la sucursal no pertenece al negocio o está inactiva")
		}
		var category domain.Categoria
		if err := tx.Where("id = ? AND activo = TRUE", input.CategoriaID).First(&category).Error; err != nil {
			return errors.New("la categoría no existe o está inactiva")
		}
		if input.UnidadMedidaID != nil {
			var unit domain.UnidadMedida
			if err := tx.Where("id = ? AND activo = TRUE", input.UnidadMedidaID).First(&unit).Error; err != nil {
				return errors.New("la unidad de medida no existe o está inactiva")
			}
		}
		var duplicate int64
		if err := tx.Model(&domain.ProductoNegocio{}).Where("negocio_id = ? AND sku_interno = ?", businessID, input.SKUInterno).Count(&duplicate).Error; err != nil {
			return err
		}
		if duplicate > 0 {
			return errors.New("el SKU ya existe en este negocio")
		}
		if barcode != "" {
			if err := tx.Model(&domain.ProductoCodigo{}).Where("codigo = ?", barcode).Count(&duplicate).Error; err != nil {
				return err
			}
			if duplicate > 0 {
				return errors.New("el código de barras ya está asignado a otro producto")
			}
		}

		product := domain.Producto{
			Nombre: input.Nombre, Descripcion: input.Descripcion, MarcaID: input.MarcaID, UnidadMedidaID: input.UnidadMedidaID,
			Contenido: input.Contenido, UnidadContenido: input.UnidadContenido,
			Presentacion: input.Presentacion, Activo: true,
		}
		if err := tx.Create(&product).Error; err != nil {
			return err
		}
		if barcode != "" {
			if err := tx.Create(&domain.ProductoCodigo{ProductoID: product.ID, Tipo: "GTIN", Codigo: barcode, EsPrincipal: true}).Error; err != nil {
				return err
			}
		}
		if imageURL != "" {
			if err := tx.Create(&domain.ProductoImagen{ProductoID: product.ID, URL: imageURL, EsPrincipal: true, Orden: 0}).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&domain.ProductoCategoria{ProductoID: product.ID, CategoriaID: input.CategoriaID, EsPrincipal: true}).Error; err != nil {
			return err
		}
		commercial := domain.ProductoNegocio{NegocioID: businessID, ProductoID: product.ID, SKUInterno: &input.SKUInterno, PrecioVenta: input.PrecioVenta, PrecioIncluyeImpuestos: true, Activo: true}
		if err := tx.Create(&commercial).Error; err != nil {
			return err
		}
		inventory := domain.InventarioSucursal{SucursalID: input.SucursalID, ProductoNegocioID: commercial.ID, StockActual: input.StockInicial, StockMinimo: 0}
		if err := tx.Create(&inventory).Error; err != nil {
			return err
		}
		if input.StockInicial > 0 {
			movement := domain.MovimientoInventario{SucursalID: input.SucursalID, ProductoNegocioID: commercial.ID, Tipo: "AJUSTE_ENTRADA", Cantidad: input.StockInicial, StockAnterior: 0, StockNuevo: input.StockInicial}
			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) ListByBusiness(businessID uuid.UUID) ([]domain.ProductRow, error) {
	products := make([]domain.ProductRow, 0)
	query := r.db.Table("productos AS p").
		Select(`
			p.id,
			p.nombre,
			(SELECT pi.url FROM producto_imagenes pi WHERE pi.producto_id = p.id ORDER BY pi.es_principal DESC, pi.orden ASC LIMIT 1) AS imagen_url,
			pn.sku_interno AS sku,
			pn.precio_venta AS precio,
			COALESCE((SELECT SUM(isu.stock_actual) FROM inventario_sucursal isu WHERE isu.producto_negocio_id = pn.id), 0) AS stock,
			(SELECT c.nombre FROM producto_categorias pc JOIN categorias c ON c.id = pc.categoria_id WHERE pc.producto_id = p.id ORDER BY pc.es_principal DESC, c.nombre ASC LIMIT 1) AS categoria,
			CASE
				WHEN COALESCE((SELECT SUM(isu.stock_actual) FROM inventario_sucursal isu WHERE isu.producto_negocio_id = pn.id), 0) <= 0 THEN 'Agotado'
				WHEN COALESCE((SELECT SUM(isu.stock_actual) FROM inventario_sucursal isu WHERE isu.producto_negocio_id = pn.id), 0) <= COALESCE((SELECT SUM(isu.stock_minimo) FROM inventario_sucursal isu WHERE isu.producto_negocio_id = pn.id), 0) THEN 'Bajo stock'
				ELSE 'En stock'
			END AS estado`).
		Joins("JOIN producto_negocio pn ON pn.producto_id = p.id AND pn.negocio_id = ? AND pn.activo = TRUE", businessID).
		Where("p.activo = TRUE").
		Order("p.nombre ASC")

	if err := query.Scan(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

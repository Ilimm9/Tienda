package infrastructure

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"tienda/backend/internal/domain"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"
)

var productBarcodePattern = regexp.MustCompile(`^\d{8,14}$`)

type ProductRepository struct {
	db *gorm.DB
}

func (r *ProductRepository) ListCategories(businessID uuid.UUID) ([]domain.CatalogOption, error) {
	var options []domain.CatalogOption
	err := r.db.Table("categorias").Select("id, nombre").Where("negocio_id = ? AND activo = TRUE", businessID).Order("nombre ASC").Scan(&options).Error
	return options, err
}

func (r *ProductRepository) ListBrands(businessID uuid.UUID) ([]domain.CatalogOption, error) {
	var options []domain.CatalogOption
	err := r.db.Table("marcas").Select("id, nombre").Where("negocio_id = ? AND activo = TRUE", businessID).Order("nombre ASC").Scan(&options).Error
	return options, err
}

func (r *ProductRepository) ListUnits(businessID uuid.UUID) ([]domain.UnidadMedida, error) {
	items := make([]domain.UnidadMedida, 0)
	err := r.db.Where("negocio_id = ? AND activo = TRUE", businessID).Order("tipo ASC, nombre ASC").Find(&items).Error
	return items, err
}

func (r *ProductRepository) CreateUnit(businessID uuid.UUID, i domain.CreateUnidadMedidaInput) error {
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
	if i.UnidadBaseID != nil {
		var total int64
		if err := r.db.Model(&domain.UnidadMedida{}).Where("id = ? AND negocio_id = ?", *i.UnidadBaseID, businessID).Count(&total).Error; err != nil {
			return err
		}
		if total == 0 {
			return errors.New("la unidad base no pertenece al negocio")
		}
	}
	return r.db.Create(&domain.UnidadMedida{NegocioID: businessID, Codigo: i.Codigo, Nombre: i.Nombre, Simbolo: i.Simbolo, Tipo: i.Tipo, UnidadBaseID: i.UnidadBaseID, FactorABase: i.FactorABase, PermiteFraccion: fraccion, Decimales: i.Decimales, Activo: true}).Error
}

func (r *ProductRepository) UpdateUnit(businessID, id uuid.UUID, i domain.UpdateUnidadMedidaInput) error {
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
		var total int64
		if err := r.db.Model(&domain.UnidadMedida{}).Where("id = ? AND negocio_id = ?", *i.UnidadBaseID, businessID).Count(&total).Error; err != nil {
			return err
		}
		if total == 0 {
			return errors.New("la unidad base no pertenece al negocio")
		}
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
	result := r.db.Model(&domain.UnidadMedida{}).Where("id = ? AND negocio_id = ?", id, businessID).Updates(values)
	if result.RowsAffected == 0 && result.Error == nil {
		return errors.New("unidad no encontrada")
	}
	return result.Error
}

func (r *ProductRepository) ImportUnits(businessID uuid.UUID, rows []domain.CatalogImportUnitRow) (domain.CatalogImportResult, error) {
	result := domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0)}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		current := make([]domain.UnidadMedida, 0)
		if err := tx.Where("negocio_id = ?", businessID).Find(&current).Error; err != nil {
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
			created = append(created, domain.UnidadMedida{NegocioID: businessID, Codigo: strings.ToLower(strings.TrimSpace(row.Codigo)), Nombre: strings.TrimSpace(row.Nombre), Simbolo: strings.TrimSpace(row.Simbolo), Tipo: strings.ToUpper(strings.TrimSpace(row.Tipo)), FactorABase: row.FactorABase, Decimales: row.Decimales, PermiteFraccion: true, Activo: true})
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

func (r *ProductRepository) ListBrandsAdmin(businessID uuid.UUID) ([]domain.Marca, error) {
	v := make([]domain.Marca, 0)
	e := r.db.Where("negocio_id = ?", businessID).Order("nombre ASC").Find(&v).Error
	return v, e
}
func (r *ProductRepository) CreateBrand(businessID uuid.UUID, i domain.CreateMarcaInput) error {
	i.Nombre = strings.TrimSpace(i.Nombre)
	if i.Nombre == "" {
		return errors.New("el nombre de la marca es obligatorio")
	}
	return r.db.Create(&domain.Marca{NegocioID: businessID, Nombre: i.Nombre, Activo: true}).Error
}
func (r *ProductRepository) UpdateBrand(businessID, id uuid.UUID, i domain.UpdateMarcaInput) error {
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
	result := r.db.Model(&domain.Marca{}).Where("id = ? AND negocio_id = ?", id, businessID).Updates(values)
	if result.RowsAffected == 0 && result.Error == nil {
		return errors.New("marca no encontrada")
	}
	return result.Error
}

func (r *ProductRepository) ImportBrands(businessID uuid.UUID, rows []domain.CatalogImportBrandRow) (domain.CatalogImportResult, error) {
	result := domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0)}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current []domain.Marca
		if err := tx.Where("negocio_id = ?", businessID).Find(&current).Error; err != nil {
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
			brands = append(brands, domain.Marca{NegocioID: businessID, Nombre: strings.TrimSpace(row.Nombre), Activo: true})
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

func (r *ProductRepository) ListCategoriesAdmin(businessID uuid.UUID) ([]domain.Categoria, error) {
	v := make([]domain.Categoria, 0)
	e := r.db.Where("negocio_id = ?", businessID).Order("nombre ASC").Find(&v).Error
	return v, e
}
func (r *ProductRepository) CreateCategory(businessID uuid.UUID, i domain.CreateCategoriaInput) error {
	i.Nombre = strings.TrimSpace(i.Nombre)
	if i.Nombre == "" {
		return errors.New("el nombre de la categoría es obligatorio")
	}
	if err := validarCategoriaPadre(r.db, businessID, uuid.Nil, i.CategoriaPadreID); err != nil {
		return err
	}
	return r.db.Create(&domain.Categoria{NegocioID: businessID, Nombre: i.Nombre, CategoriaPadreID: i.CategoriaPadreID, Descripcion: i.Descripcion, Activo: true}).Error
}
func (r *ProductRepository) UpdateCategory(businessID, id uuid.UUID, i domain.UpdateCategoriaInput) error {
	if err := validarCategoriaPadre(r.db, businessID, id, i.CategoriaPadreID); err != nil {
		return err
	}
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
	result := r.db.Model(&domain.Categoria{}).Where("id = ? AND negocio_id = ?", id, businessID).Updates(values)
	if result.RowsAffected == 0 && result.Error == nil {
		return errors.New("categoría no encontrada")
	}
	return result.Error
}

func (r *ProductRepository) ImportCategories(businessID uuid.UUID, rows []domain.CatalogImportCategoryRow) (domain.CatalogImportResult, error) {
	result := domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0)}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current []domain.Categoria
		if err := tx.Where("negocio_id = ?", businessID).Find(&current).Error; err != nil {
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
			category := domain.Categoria{ID: uuid.New(), NegocioID: businessID, Nombre: strings.TrimSpace(row.Nombre), CategoriaPadreID: parentID, Activo: true}
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
	return strings.ToLower(strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, norm.NFD.String(strings.TrimSpace(value))))
}

func validarCategoriaPadre(db *gorm.DB, businessID, categoryID uuid.UUID, parentID *uuid.UUID) error {
	if parentID == nil {
		return nil
	}
	if categoryID != uuid.Nil && *parentID == categoryID {
		return errors.New("una categoría no puede ser su propia categoría padre")
	}
	var total int64
	if err := db.Model(&domain.Categoria{}).
		Where("id = ? AND negocio_id = ?", *parentID, businessID).
		Count(&total).Error; err != nil {
		return err
	}
	if total == 0 {
		return errors.New("la categoría padre no pertenece al negocio")
	}
	return nil
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
		if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", input.CategoriaID, businessID).First(&category).Error; err != nil {
			return errors.New("la categoría no existe o está inactiva")
		}
		if input.MarcaID != nil {
			var brand domain.Marca
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", *input.MarcaID, businessID).First(&brand).Error; err != nil {
				return errors.New("la marca no existe o está inactiva")
			}
		}
		if input.UnidadMedidaID != nil {
			var unit domain.UnidadMedida
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", *input.UnidadMedidaID, businessID).First(&unit).Error; err != nil {
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
			if err := tx.Model(&domain.ProductoCodigo{}).Where("negocio_id = ? AND codigo = ?", businessID, barcode).Count(&duplicate).Error; err != nil {
				return err
			}
			if duplicate > 0 {
				return errors.New("el código de barras ya está asignado a otro producto")
			}
		}

		product := domain.Producto{
			NegocioID: businessID, Nombre: input.Nombre, Descripcion: input.Descripcion, MarcaID: input.MarcaID, UnidadMedidaID: input.UnidadMedidaID,
			Contenido: input.Contenido, UnidadContenido: input.UnidadContenido,
			Presentacion: input.Presentacion, Activo: true,
		}
		if err := tx.Create(&product).Error; err != nil {
			return err
		}
		if barcode != "" {
			if err := tx.Create(&domain.ProductoCodigo{NegocioID: businessID, ProductoID: product.ID, Tipo: "GTIN", Codigo: barcode, EsPrincipal: true}).Error; err != nil {
				return err
			}
		}
		if imageURL != "" {
			if err := tx.Create(&domain.ProductoImagen{ProductoID: product.ID, URL: imageURL, EsPrincipal: true, Orden: 0}).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&domain.ProductoCategoria{NegocioID: businessID, ProductoID: product.ID, CategoriaID: input.CategoriaID, EsPrincipal: true}).Error; err != nil {
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

func (r *ProductRepository) Update(businessID, productID uuid.UUID, input domain.UpdateProductInput) error {
	input.Nombre = strings.TrimSpace(input.Nombre)
	input.SKUInterno = strings.TrimSpace(input.SKUInterno)
	if input.Nombre == "" || input.SKUInterno == "" || input.PrecioVenta < 0 {
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
		var commercial domain.ProductoNegocio
		if err := tx.Where("negocio_id = ? AND producto_id = ? AND activo = TRUE", businessID, productID).First(&commercial).Error; err != nil {
			return errors.New("producto no encontrado")
		}
		var category domain.Categoria
		if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", input.CategoriaID, businessID).First(&category).Error; err != nil {
			return errors.New("la categoría no existe o está inactiva")
		}
		if input.MarcaID != nil {
			var brand domain.Marca
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", *input.MarcaID, businessID).First(&brand).Error; err != nil {
				return errors.New("la marca no existe o está inactiva")
			}
		}
		if input.UnidadMedidaID != nil {
			var unit domain.UnidadMedida
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", *input.UnidadMedidaID, businessID).First(&unit).Error; err != nil {
				return errors.New("la unidad de medida no existe o está inactiva")
			}
		}
		var duplicates int64
		if err := tx.Model(&domain.ProductoNegocio{}).Where("negocio_id = ? AND sku_interno = ? AND producto_id <> ?", businessID, input.SKUInterno, productID).Count(&duplicates).Error; err != nil {
			return err
		}
		if duplicates > 0 {
			return errors.New("el SKU ya existe en este negocio")
		}
		if barcode != "" {
			if err := tx.Model(&domain.ProductoCodigo{}).Where("negocio_id = ? AND codigo = ? AND producto_id <> ?", businessID, barcode, productID).Count(&duplicates).Error; err != nil {
				return err
			}
			if duplicates > 0 {
				return errors.New("el código de barras ya está asignado a otro producto")
			}
		}

		productValues := map[string]interface{}{
			"nombre": input.Nombre, "descripcion": input.Descripcion, "marca_id": input.MarcaID,
			"unidad_medida_id": input.UnidadMedidaID, "contenido": input.Contenido,
			"unidad_contenido": input.UnidadContenido, "presentacion": input.Presentacion,
		}
		if err := tx.Model(&domain.Producto{}).Where("id = ? AND negocio_id = ?", productID, businessID).Updates(productValues).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.ProductoNegocio{}).Where("id = ?", commercial.ID).Updates(map[string]interface{}{"sku_interno": input.SKUInterno, "precio_venta": input.PrecioVenta}).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.ProductoCategoria{}).Where("negocio_id = ? AND producto_id = ? AND es_principal = TRUE", businessID, productID).Updates(map[string]interface{}{"categoria_id": input.CategoriaID}).Error; err != nil {
			return err
		}

		if err := tx.Where("negocio_id = ? AND producto_id = ? AND es_principal = TRUE", businessID, productID).Delete(&domain.ProductoCodigo{}).Error; err != nil {
			return err
		}
		if barcode != "" {
			if err := tx.Create(&domain.ProductoCodigo{NegocioID: businessID, ProductoID: productID, Tipo: "GTIN", Codigo: barcode, EsPrincipal: true}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("producto_id = ? AND es_principal = TRUE", productID).Delete(&domain.ProductoImagen{}).Error; err != nil {
			return err
		}
		if imageURL != "" {
			if err := tx.Create(&domain.ProductoImagen{ProductoID: productID, URL: imageURL, EsPrincipal: true, Orden: 0}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ProductRepository) Deactivate(businessID, productID uuid.UUID) error {
	result := r.db.Model(&domain.ProductoNegocio{}).
		Where("negocio_id = ? AND producto_id = ? AND activo = TRUE", businessID, productID).
		Update("activo", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("producto no encontrado")
	}
	return nil
}

func (r *ProductRepository) ValidateProductImport(businessID, branchID uuid.UUID, rows []domain.ProductImportRow) ([]domain.ValidatedProductImportRow, domain.CatalogImportResult, error) {
	result := domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0), Advertencias: make([]domain.CatalogImportIssue, 0)}
	var branch domain.Sucursal
	if err := r.db.Where("id = ? AND negocio_id = ? AND activo = TRUE", branchID, businessID).First(&branch).Error; err != nil {
		return nil, result, errors.New("la sucursal no pertenece al negocio o está inactiva")
	}
	var categories []domain.Categoria
	var brands []domain.Marca
	var units []domain.UnidadMedida
	if err := r.db.Where("negocio_id = ? AND activo = TRUE", businessID).Find(&categories).Error; err != nil {
		return nil, result, err
	}
	if err := r.db.Where("negocio_id = ? AND activo = TRUE", businessID).Find(&brands).Error; err != nil {
		return nil, result, err
	}
	if err := r.db.Where("negocio_id = ? AND activo = TRUE", businessID).Find(&units).Error; err != nil {
		return nil, result, err
	}
	categoryIDs := make(map[string]uuid.UUID, len(categories))
	brandIDs := make(map[string]uuid.UUID, len(brands))
	unitIDs := make(map[string]uuid.UUID, len(units))
	for _, item := range categories {
		categoryIDs[catalogImportKey(item.Nombre)] = item.ID
	}
	for _, item := range brands {
		brandIDs[catalogImportKey(item.Nombre)] = item.ID
	}
	for _, item := range units {
		unitIDs[catalogImportKey(item.Nombre)] = item.ID
	}
	var existingSKUs, existingCodes []string
	if err := r.db.Model(&domain.ProductoNegocio{}).Where("negocio_id = ? AND sku_interno IS NOT NULL", businessID).Pluck("sku_interno", &existingSKUs).Error; err != nil {
		return nil, result, err
	}
	if err := r.db.Model(&domain.ProductoCodigo{}).Where("negocio_id = ?", businessID).Pluck("codigo", &existingCodes).Error; err != nil {
		return nil, result, err
	}
	type productIdentity struct {
		Nombre       string
		Presentacion string
	}
	var existingProducts []productIdentity
	if err := r.db.Table("productos AS p").
		Select("p.nombre, COALESCE(p.presentacion, '') AS presentacion").
		Joins("JOIN producto_negocio pn ON pn.producto_id = p.id AND pn.negocio_id = ?", businessID).
		Scan(&existingProducts).Error; err != nil {
		return nil, result, err
	}
	seenSKUs, seenCodes := make(map[string]struct{}, len(existingSKUs)), make(map[string]struct{}, len(existingCodes))
	seenNamePresentations := make(map[string]struct{}, len(existingProducts))
	for _, sku := range existingSKUs {
		seenSKUs[catalogImportKey(sku)] = struct{}{}
	}
	for _, code := range existingCodes {
		seenCodes[code] = struct{}{}
	}
	for _, product := range existingProducts {
		seenNamePresentations[productImportNamePresentationKey(product.Nombre, product.Presentacion)] = struct{}{}
	}
	prepared := make([]domain.ValidatedProductImportRow, 0, len(rows))

	for _, row := range rows {
		hasError := false
		hasDuplicate := false
		addError := func(field, reason string) {
			hasError = true
			result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Campo: field, Motivo: reason})
		}
		skuKey := catalogImportKey(row.SKUInterno)
		namePresentationKey := productImportNamePresentationKey(row.Nombre, row.Presentacion)
		if _, exists := seenNamePresentations[namePresentationKey]; exists {
			hasDuplicate = true
			addError("Nombre y presentación", "ya existe en el negocio o está repetido en el archivo")
		}
		if row.SKUInterno != "" {
			if _, exists := seenSKUs[skuKey]; exists {
				hasDuplicate = true
				addError("SKU interno", "ya existe en este negocio o está repetido en el archivo")
			}
		}
		if row.CodigoBarras != "" {
			if _, exists := seenCodes[row.CodigoBarras]; exists {
				hasDuplicate = true
				addError("Código de barras", "ya está asignado o está repetido en el archivo")
			}
		}
		var categoryID *uuid.UUID
		if row.Categoria != "" {
			id, ok := categoryIDs[catalogImportKey(row.Categoria)]
			if !ok {
				addError("Categoría", "no existe o está inactiva")
			} else {
				categoryID = &id
			}
		}
		var brandID *uuid.UUID
		if row.Marca != "" {
			id, ok := brandIDs[catalogImportKey(row.Marca)]
			if !ok {
				addError("Marca", "no existe o está inactiva")
			} else {
				brandID = &id
			}
		}
		var unitID *uuid.UUID
		if row.UnidadMedida != "" {
			id, ok := unitIDs[catalogImportKey(row.UnidadMedida)]
			if !ok {
				addError("Unidad de medida", "no existe o está inactiva")
			} else {
				unitID = &id
			}
		}
		if hasError {
			if hasDuplicate {
				result.Omitidas++
			} else {
				result.Invalidas++
			}
			continue
		}
		input := domain.CreateImportedProductInput{Nombre: row.Nombre, SKUInterno: importOptionalString(row.SKUInterno), CategoriaID: categoryID, MarcaID: brandID, SucursalID: branchID, Contenido: row.Contenido, PrecioVenta: row.PrecioVenta, StockInicial: row.StockInicial}
		input.Descripcion = importOptionalString(row.Descripcion)
		input.Presentacion = importOptionalString(row.Presentacion)
		input.UnidadContenido = importOptionalString(row.UnidadContenido)
		input.UnidadMedidaID = unitID
		input.CodigoBarras = importOptionalString(row.CodigoBarras)
		if row.SKUInterno != "" {
			seenSKUs[skuKey] = struct{}{}
		}
		if row.CodigoBarras != "" {
			seenCodes[row.CodigoBarras] = struct{}{}
		}
		seenNamePresentations[namePresentationKey] = struct{}{}
		prepared = append(prepared, domain.ValidatedProductImportRow{Fila: row.Fila, Input: input})
	}
	return prepared, result, nil
}

func (r *ProductRepository) CreateImportedProducts(businessID uuid.UUID, rows []domain.ValidatedProductImportRow) (domain.CatalogImportResult, error) {
	result := domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0), Advertencias: make([]domain.CatalogImportIssue, 0)}
	for _, row := range rows {
		if err := r.createImportedProduct(businessID, row.Input); err != nil {
			if strings.Contains(err.Error(), "SKU ya existe") || strings.Contains(err.Error(), "código de barras ya está asignado") || strings.Contains(err.Error(), "nombre y presentación ya existen") {
				result.Omitidas++
			} else {
				result.Invalidas++
			}
			result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Campo: importErrorField(err.Error()), Motivo: err.Error()})
			continue
		}
		result.Creadas++
	}
	return result, nil
}

func importErrorField(message string) string {
	switch {
	case strings.Contains(message, "SKU"):
		return "SKU interno"
	case strings.Contains(message, "nombre y presentación"):
		return "Nombre y presentación"
	case strings.Contains(message, "código de barras"):
		return "Código de barras"
	case strings.Contains(message, "categoría"):
		return "Categoría"
	case strings.Contains(message, "unidad de medida"):
		return "Unidad de medida"
	case strings.Contains(message, "sucursal"):
		return "Sucursal"
	default:
		return "Producto"
	}
}

func productImportNamePresentationKey(name, presentation string) string {
	return catalogImportKey(name) + "\x00" + catalogImportKey(presentation)
}

func (r *ProductRepository) createImportedProduct(businessID uuid.UUID, input domain.CreateImportedProductInput) error {
	input.Nombre = strings.TrimSpace(input.Nombre)
	if input.Nombre == "" || input.PrecioVenta < 0 || input.StockInicial < 0 {
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
		if input.CategoriaID != nil {
			var category domain.Categoria
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", *input.CategoriaID, businessID).First(&category).Error; err != nil {
				return errors.New("la categoría no existe o está inactiva")
			}
		}
		if input.MarcaID != nil {
			var brand domain.Marca
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", *input.MarcaID, businessID).First(&brand).Error; err != nil {
				return errors.New("la marca no existe o está inactiva")
			}
		}
		if input.UnidadMedidaID != nil {
			var unit domain.UnidadMedida
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", *input.UnidadMedidaID, businessID).First(&unit).Error; err != nil {
				return errors.New("la unidad de medida no existe o está inactiva")
			}
		}
		type productIdentity struct {
			Nombre       string
			Presentacion string
		}
		var existingProducts []productIdentity
		if err := tx.Table("productos AS p").
			Select("p.nombre, COALESCE(p.presentacion, '') AS presentacion").
			Joins("JOIN producto_negocio pn ON pn.producto_id = p.id AND pn.negocio_id = ?", businessID).
			Scan(&existingProducts).Error; err != nil {
			return err
		}
		for _, product := range existingProducts {
			if productImportNamePresentationKey(product.Nombre, product.Presentacion) == productImportNamePresentationKey(input.Nombre, optionalString(input.Presentacion)) {
				return errors.New("el nombre y presentación ya existen en este negocio")
			}
		}
		var duplicate int64
		if input.SKUInterno != nil {
			if err := tx.Model(&domain.ProductoNegocio{}).Where("negocio_id = ? AND sku_interno = ?", businessID, *input.SKUInterno).Count(&duplicate).Error; err != nil {
				return err
			}
			if duplicate > 0 {
				return errors.New("el SKU ya existe en este negocio")
			}
		}
		if barcode != "" {
			if err := tx.Model(&domain.ProductoCodigo{}).Where("negocio_id = ? AND codigo = ?", businessID, barcode).Count(&duplicate).Error; err != nil {
				return err
			}
			if duplicate > 0 {
				return errors.New("el código de barras ya está asignado a otro producto")
			}
		}
		product := domain.Producto{NegocioID: businessID, Nombre: input.Nombre, Descripcion: input.Descripcion, MarcaID: input.MarcaID, UnidadMedidaID: input.UnidadMedidaID, Contenido: input.Contenido, UnidadContenido: input.UnidadContenido, Presentacion: input.Presentacion, Activo: true}
		if err := tx.Create(&product).Error; err != nil {
			return err
		}
		if barcode != "" {
			if err := tx.Create(&domain.ProductoCodigo{NegocioID: businessID, ProductoID: product.ID, Tipo: "GTIN", Codigo: barcode, EsPrincipal: true}).Error; err != nil {
				return err
			}
		}
		if imageURL != "" {
			if err := tx.Create(&domain.ProductoImagen{ProductoID: product.ID, URL: imageURL, EsPrincipal: true, Orden: 0}).Error; err != nil {
				return err
			}
		}
		if input.CategoriaID != nil {
			if err := tx.Create(&domain.ProductoCategoria{NegocioID: businessID, ProductoID: product.ID, CategoriaID: *input.CategoriaID, EsPrincipal: true}).Error; err != nil {
				return err
			}
		}
		commercial := domain.ProductoNegocio{NegocioID: businessID, ProductoID: product.ID, SKUInterno: input.SKUInterno, PrecioVenta: input.PrecioVenta, PrecioIncluyeImpuestos: true, Activo: true}
		if err := tx.Create(&commercial).Error; err != nil {
			return err
		}
		inventory := domain.InventarioSucursal{SucursalID: input.SucursalID, ProductoNegocioID: commercial.ID, StockActual: input.StockInicial, StockMinimo: 0}
		if err := tx.Create(&inventory).Error; err != nil {
			return err
		}
		if input.StockInicial > 0 {
			if err := tx.Create(&domain.MovimientoInventario{SucursalID: input.SucursalID, ProductoNegocioID: commercial.ID, Tipo: "AJUSTE_ENTRADA", Cantidad: input.StockInicial, StockAnterior: 0, StockNuevo: input.StockInicial}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func importOptionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
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
			(SELECT pc.categoria_id FROM producto_categorias pc WHERE pc.producto_id = p.id ORDER BY pc.es_principal DESC LIMIT 1) AS categoria_id,
			m.nombre AS marca,
			p.marca_id,
			p.descripcion,
			p.presentacion,
			p.contenido,
			p.unidad_contenido,
			u.nombre AS unidad_medida,
			p.unidad_medida_id,
			(SELECT pcodigo.codigo FROM producto_codigos pcodigo WHERE pcodigo.producto_id = p.id ORDER BY pcodigo.es_principal DESC LIMIT 1) AS codigo_barras,
			CASE
				WHEN COALESCE((SELECT SUM(isu.stock_actual) FROM inventario_sucursal isu WHERE isu.producto_negocio_id = pn.id), 0) <= 0 THEN 'Agotado'
				WHEN COALESCE((SELECT SUM(isu.stock_actual) FROM inventario_sucursal isu WHERE isu.producto_negocio_id = pn.id), 0) <= COALESCE((SELECT SUM(isu.stock_minimo) FROM inventario_sucursal isu WHERE isu.producto_negocio_id = pn.id), 0) THEN 'Bajo stock'
				ELSE 'En stock'
			END AS estado`).
		Joins("JOIN producto_negocio pn ON pn.producto_id = p.id AND pn.negocio_id = ? AND pn.activo = TRUE", businessID).
		Joins("LEFT JOIN marcas m ON m.id = p.marca_id").
		Joins("LEFT JOIN unidades_medida u ON u.id = p.unidad_medida_id").
		Where("p.activo = TRUE").
		Order("p.nombre ASC")

	if err := query.Scan(&products).Error; err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return products, nil
	}
	stockByProduct := make(map[string][]domain.ProductBranchStock, len(products))
	var branches []struct {
		ProductoID uuid.UUID `gorm:"column:producto_id"`
		SucursalID uuid.UUID `gorm:"column:sucursal_id"`
		Sucursal   string    `gorm:"column:sucursal"`
		Stock      float64   `gorm:"column:stock"`
	}
	if err := r.db.Table("inventario_sucursal AS i").
		Select("pn.producto_id, s.id AS sucursal_id, s.nombre AS sucursal, i.stock_actual AS stock").
		Joins("JOIN producto_negocio pn ON pn.id = i.producto_negocio_id AND pn.negocio_id = ? AND pn.activo = TRUE", businessID).
		Joins("JOIN sucursales s ON s.id = i.sucursal_id").
		Order("s.nombre ASC").Scan(&branches).Error; err != nil {
		return nil, err
	}
	for _, branch := range branches {
		key := branch.ProductoID.String()
		stockByProduct[key] = append(stockByProduct[key], domain.ProductBranchStock{SucursalID: branch.SucursalID.String(), Sucursal: branch.Sucursal, Stock: branch.Stock})
	}
	for index := range products {
		products[index].Inventario = stockByProduct[products[index].ID.String()]
		if products[index].Inventario == nil {
			products[index].Inventario = make([]domain.ProductBranchStock, 0)
		}
	}
	return products, nil
}

package infrastructure

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"tienda/backend/internal/domain"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var productBarcodePattern = regexp.MustCompile(`^\d{8,14}$`)
var importedProductBarcodePattern = regexp.MustCompile(`^\d{1,14}$`)

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
	i.Nombre = strings.TrimSpace(i.Nombre)
	i.Simbolo = strings.TrimSpace(i.Simbolo)
	i.Tipo = strings.TrimSpace(strings.ToUpper(i.Tipo))
	if i.Nombre == "" || i.Simbolo == "" || i.Tipo == "" || i.FactorABase <= 0 || i.Decimales < 0 || i.Decimales > 6 {
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
	return r.db.Create(&domain.UnidadMedida{NegocioID: businessID, Nombre: i.Nombre, Simbolo: i.Simbolo, Tipo: i.Tipo, UnidadBaseID: i.UnidadBaseID, FactorABase: i.FactorABase, PermiteFraccion: fraccion, Decimales: i.Decimales, Activo: true}).Error
}

func (r *ProductRepository) UpdateUnit(businessID, id uuid.UUID, i domain.UpdateUnidadMedidaInput) error {
	values := map[string]interface{}{}
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

func (r *ProductRepository) DeleteUnit(businessID, id uuid.UUID) error {
	return r.deleteCatalogItem(&domain.UnidadMedida{}, businessID, id, func(tx *gorm.DB) error {
		if err := r.ensureCatalogReference(tx, `SELECT count(*) FROM productos WHERE negocio_id = ? AND unidad_medida_id = ?`, businessID, id, "productos"); err != nil {
			return err
		}
		if err := r.ensureCatalogReference(tx, `SELECT count(*) FROM compras_proveedor_detalles d JOIN compras_proveedor c ON c.id = d.compra_id WHERE c.negocio_id = ? AND d.unidad_medida_id = ?`, businessID, id, "compras"); err != nil {
			return err
		}
		return r.ensureCatalogReference(tx, `SELECT count(*) FROM unidades_medida WHERE negocio_id = ? AND unidad_base_id = ?`, businessID, id, "unidades derivadas")
	})
}

func (r *ProductRepository) CheckUnitDeletion(businessID, id uuid.UUID) (domain.CatalogDeletionCheck, error) {
	return r.catalogDeletionCheck(&domain.UnidadMedida{}, businessID, id, []catalogReference{
		{query: `SELECT count(*) FROM productos WHERE negocio_id = ? AND unidad_medida_id = ?`, label: "productos"},
		{query: `SELECT count(*) FROM compras_proveedor_detalles d JOIN compras_proveedor c ON c.id = d.compra_id WHERE c.negocio_id = ? AND d.unidad_medida_id = ?`, label: "compras"},
		{query: `SELECT count(*) FROM unidades_medida WHERE negocio_id = ? AND unidad_base_id = ?`, label: "unidades derivadas"},
	})
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
			existing[catalogImportKey(unit.Nombre)] = struct{}{}
			existing[catalogImportKey(unit.Simbolo)] = struct{}{}
		}
		seen := make(map[string]struct{}, len(rows))
		created := make([]domain.UnidadMedida, 0, len(rows))
		for _, row := range rows {
			key := catalogImportKey(row.Simbolo)
			if _, duplicate := seen[key]; duplicate {
				result.Omitidas++
				result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Motivo: "La unidad está repetida en el archivo"})
				continue
			}
			seen[key] = struct{}{}
			if _, exists := existing[key]; exists || containsUnitKey(existing, row.Nombre, row.Simbolo) {
				result.Omitidas++
				result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Motivo: "El nombre o símbolo de la unidad ya existe"})
				continue
			}
			created = append(created, domain.UnidadMedida{NegocioID: businessID, Nombre: strings.TrimSpace(row.Nombre), Simbolo: strings.TrimSpace(row.Simbolo), Tipo: strings.ToUpper(strings.TrimSpace(row.Tipo)), FactorABase: row.FactorABase, Decimales: row.Decimales, PermiteFraccion: true, Activo: true})
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

func (r *ProductRepository) DeleteBrand(businessID, id uuid.UUID) error {
	return r.deleteCatalogItem(&domain.Marca{}, businessID, id, func(tx *gorm.DB) error {
		return r.ensureCatalogReference(tx, `SELECT count(*) FROM productos WHERE negocio_id = ? AND marca_id = ?`, businessID, id, "productos")
	})
}

func (r *ProductRepository) CheckBrandDeletion(businessID, id uuid.UUID) (domain.CatalogDeletionCheck, error) {
	return r.catalogDeletionCheck(&domain.Marca{}, businessID, id, []catalogReference{
		{query: `SELECT count(*) FROM productos WHERE negocio_id = ? AND marca_id = ?`, label: "productos"},
	})
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

func (r *ProductRepository) DeleteCategory(businessID, id uuid.UUID) error {
	return r.deleteCatalogItem(&domain.Categoria{}, businessID, id, func(tx *gorm.DB) error {
		if err := r.ensureCatalogReference(tx, `SELECT count(*) FROM producto_categorias WHERE negocio_id = ? AND categoria_id = ?`, businessID, id, "productos"); err != nil {
			return err
		}
		return r.ensureCatalogReference(tx, `SELECT count(*) FROM categorias WHERE negocio_id = ? AND categoria_padre_id = ?`, businessID, id, "categorías hijas")
	})
}

func (r *ProductRepository) CheckCategoryDeletion(businessID, id uuid.UUID) (domain.CatalogDeletionCheck, error) {
	return r.catalogDeletionCheck(&domain.Categoria{}, businessID, id, []catalogReference{
		{query: `SELECT count(*) FROM producto_categorias WHERE negocio_id = ? AND categoria_id = ?`, label: "productos"},
		{query: `SELECT count(*) FROM categorias WHERE negocio_id = ? AND categoria_padre_id = ?`, label: "categorías hijas"},
	})
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
func (r *ProductRepository) ImportProviders(businessID uuid.UUID, rows []domain.CatalogImportProviderRow) (domain.CatalogImportResult, error) {
	result := domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0)}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current []domain.Proveedor
		if err := tx.Where("negocio_id = ?", businessID).Find(&current).Error; err != nil {
			return err
		}
		existing := make(map[string]struct{}, len(current))
		for _, provider := range current {
			existing[catalogImportKey(provider.Nombre)] = struct{}{}
		}
		seen := make(map[string]struct{}, len(rows))
		providers := make([]domain.Proveedor, 0, len(rows))
		for _, row := range rows {
			key := catalogImportKey(row.Nombre)
			if _, duplicate := seen[key]; duplicate {
				result.Omitidas++
				result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Motivo: "El proveedor está repetido en el archivo"})
				continue
			}
			seen[key] = struct{}{}
			if _, exists := existing[key]; exists {
				result.Omitidas++
				result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Motivo: "El proveedor ya existe"})
				continue
			}
			provider := domain.Proveedor{NegocioID: businessID, Nombre: strings.TrimSpace(row.Nombre), Activo: true}
			if value := strings.TrimSpace(row.RazonSocial); value != "" {
				provider.RazonSocial = &value
			}
			if value := strings.TrimSpace(row.RFC); value != "" {
				provider.RFC = &value
			}
			if value := strings.TrimSpace(row.Telefono); value != "" {
				provider.Telefono = &value
			}
			if value := strings.TrimSpace(row.Email); value != "" {
				provider.Email = &value
			}
			if value := strings.TrimSpace(row.Direccion); value != "" {
				provider.Direccion = &value
			}
			providers = append(providers, provider)
		}
		if len(providers) == 0 {
			return nil
		}
		if err := tx.Create(&providers).Error; err != nil {
			return err
		}
		result.Creadas = len(providers)
		return nil
	})
	return result, err
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

func (r *ProductRepository) DeleteProvider(businessID, id uuid.UUID) error {
	return r.deleteCatalogItem(&domain.Proveedor{}, businessID, id, func(tx *gorm.DB) error {
		if err := r.ensureCatalogReference(tx, `SELECT count(*) FROM producto_proveedor pp JOIN producto_negocio pn ON pn.id = pp.producto_negocio_id WHERE pn.negocio_id = ? AND pp.proveedor_id = ?`, businessID, id, "productos"); err != nil {
			return err
		}
		if err := r.ensureCatalogReference(tx, `SELECT count(*) FROM compras_proveedor WHERE negocio_id = ? AND proveedor_id = ?`, businessID, id, "compras"); err != nil {
			return err
		}
		return r.ensureCatalogReference(tx, `SELECT count(*) FROM lotes l JOIN sucursales s ON s.id = l.sucursal_id WHERE s.negocio_id = ? AND l.proveedor_id = ?`, businessID, id, "lotes de inventario")
	})
}

func (r *ProductRepository) CheckProviderDeletion(businessID, id uuid.UUID) (domain.CatalogDeletionCheck, error) {
	return r.catalogDeletionCheck(&domain.Proveedor{}, businessID, id, []catalogReference{
		{query: `SELECT count(*) FROM producto_proveedor pp JOIN producto_negocio pn ON pn.id = pp.producto_negocio_id WHERE pn.negocio_id = ? AND pp.proveedor_id = ?`, label: "productos"},
		{query: `SELECT count(*) FROM compras_proveedor WHERE negocio_id = ? AND proveedor_id = ?`, label: "compras"},
		{query: `SELECT count(*) FROM lotes l JOIN sucursales s ON s.id = l.sucursal_id WHERE s.negocio_id = ? AND l.proveedor_id = ?`, label: "lotes de inventario"},
	})
}

func (r *ProductRepository) deleteCatalogItem(model interface{}, businessID, id uuid.UUID, check func(*gorm.DB) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var exists int64
		if err := tx.Model(model).Where("id = ? AND negocio_id = ?", id, businessID).Count(&exists).Error; err != nil {
			return err
		}
		if exists == 0 {
			return domain.ErrCatalogItemNotFound
		}
		if err := check(tx); err != nil {
			return err
		}
		result := tx.Where("id = ? AND negocio_id = ?", id, businessID).Delete(model)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrCatalogItemNotFound
		}
		return nil
	})
}

func (r *ProductRepository) ensureCatalogReference(tx *gorm.DB, query string, businessID, id uuid.UUID, relation string) error {
	var count int64
	if err := tx.Raw(query, businessID, id).Scan(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: el registro está vinculado a %s", domain.ErrCatalogItemInUse, relation)
	}
	return nil
}

type catalogReference struct {
	query string
	label string
}

func (r *ProductRepository) catalogDeletionCheck(model interface{}, businessID, id uuid.UUID, references []catalogReference) (domain.CatalogDeletionCheck, error) {
	var exists int64
	if err := r.db.Model(model).Where("id = ? AND negocio_id = ?", id, businessID).Count(&exists).Error; err != nil {
		return domain.CatalogDeletionCheck{}, err
	}
	if exists == 0 {
		return domain.CatalogDeletionCheck{}, domain.ErrCatalogItemNotFound
	}
	result := domain.CatalogDeletionCheck{Vinculaciones: make([]string, 0)}
	for _, reference := range references {
		var count int64
		if err := r.db.Raw(reference.query, businessID, id).Scan(&count).Error; err != nil {
			return domain.CatalogDeletionCheck{}, err
		}
		if count > 0 {
			result.Vinculaciones = append(result.Vinculaciones, reference.label)
		}
	}
	result.PuedeEliminarse = len(result.Vinculaciones) == 0
	return result, nil
}

func (r *ProductRepository) ListBranches(businessID uuid.UUID) ([]domain.CatalogOption, error) {
	var options []domain.CatalogOption
	err := r.db.Table("sucursales").Select("id, nombre").Where("negocio_id = ? AND activo = TRUE", businessID).Order("nombre ASC").Scan(&options).Error
	return options, err
}

func (r *ProductRepository) Create(businessID uuid.UUID, input domain.CreateProductInput) error {
	if len(input.Variantes) > 0 {
		return r.createProductFamily(businessID, input)
	}
	input.Nombre = strings.TrimSpace(input.Nombre)
	input.SKUInterno = strings.TrimSpace(input.SKUInterno)
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
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", input.CategoriaID, businessID).First(&category).Error; err != nil {
				return errors.New("la categoría no existe o está inactiva")
			}
		}
		if input.MarcaID != nil {
			var brand domain.Marca
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", *input.MarcaID, businessID).First(&brand).Error; err != nil {
				return errors.New("la marca no existe o está inactiva")
			}
		}
		if err := validateProductProvider(tx, businessID, input.ProveedorID); err != nil {
			return err
		}
		if input.UnidadMedidaID != nil {
			var unit domain.UnidadMedida
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", *input.UnidadMedidaID, businessID).First(&unit).Error; err != nil {
				return errors.New("la unidad de medida no existe o está inactiva")
			}
		}
		if input.GenerarSKUInterno || input.SKUInterno == "" {
			generatedSKU, err := reserveGeneratedProductSKU(tx, businessID)
			if err != nil {
				return err
			}
			input.SKUInterno = generatedSKU
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
			if err := tx.Create(&domain.ProductoCodigo{NegocioID: businessID, ProductoID: &product.ID, Tipo: "GTIN", Codigo: barcode, EsPrincipal: true}).Error; err != nil {
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
		commercial := domain.ProductoNegocio{NegocioID: businessID, ProductoID: &product.ID, SKUInterno: &input.SKUInterno, PrecioVenta: input.PrecioVenta, PrecioIncluyeImpuestos: true, Activo: true}
		if err := tx.Create(&commercial).Error; err != nil {
			return err
		}
		if err := replaceProductProvider(tx, commercial.ID, input.ProveedorID); err != nil {
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

func (r *ProductRepository) createProductFamily(businessID uuid.UUID, input domain.CreateProductInput) error {
	input.Nombre = strings.TrimSpace(input.Nombre)
	if input.Nombre == "" || input.SucursalID == uuid.Nil || len(input.Variantes) == 0 {
		return errors.New("los datos del producto con variantes no son válidos")
	}
	if input.UnidadMedidaID != nil {
		var unit domain.UnidadMedida
		if err := r.db.Where("id = ? AND negocio_id = ? AND activo = TRUE", input.UnidadMedidaID, businessID).First(&unit).Error; err != nil {
			return errors.New("la unidad de medida no existe o está inactiva")
		}
	}
	if imageURL := optionalString(input.ImagenURL); imageURL != "" {
		parsed, err := url.Parse(imageURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
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
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", input.CategoriaID, businessID).First(&category).Error; err != nil {
				return errors.New("la categoría no existe o está inactiva")
			}
		}
		if input.MarcaID != nil {
			var brand domain.Marca
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", *input.MarcaID, businessID).First(&brand).Error; err != nil {
				return errors.New("la marca no existe o está inactiva")
			}
		}
		if err := validateProductProvider(tx, businessID, input.ProveedorID); err != nil {
			return err
		}
		family, product, _, err := r.resolveVariantBase(tx, businessID, variantBaseInput{
			Nombre: input.Nombre, Descripcion: input.Descripcion, MarcaID: input.MarcaID, CategoriaID: input.CategoriaID,
			UnidadMedidaID: input.UnidadMedidaID, Contenido: input.Contenido, UnidadContenido: input.UnidadContenido,
			Presentacion: input.Presentacion, ImagenURL: input.ImagenURL,
		})
		if err != nil {
			return err
		}
		seenVariants := make(map[string]struct{}, len(input.Variantes))
		for _, variant := range input.Variantes {
			variantKey, err := productVariantKey(variant.Atributos)
			if err != nil {
				return err
			}
			if _, exists := seenVariants[variantKey]; exists {
				return errors.New("no puede repetir la misma combinación de variantes")
			}
			seenVariants[variantKey] = struct{}{}
			if err := r.createFamilyVariant(tx, businessID, input.SucursalID, product, family, variant, variantKey, input.ProveedorID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ProductRepository) createFamilyVariant(tx *gorm.DB, businessID, branchID uuid.UUID, product domain.Producto, family domain.FamiliaProducto, variant domain.CreateProductVariantInput, variantKey string, providerID *uuid.UUID) error {
	sku := strings.TrimSpace(variant.SKUInterno)
	if !variant.GenerarSKUInterno && sku == "" {
		return errors.New("cada variante requiere un SKU o generación automática")
	}
	if variant.PrecioVenta < 0 || variant.StockInicial < 0 {
		return errors.New("el precio y stock de la variante deben ser válidos")
	}
	if variant.GenerarSKUInterno {
		generatedSKU, err := reserveGeneratedProductSKU(tx, businessID)
		if err != nil {
			return err
		}
		sku = generatedSKU
	}
	var duplicate int64
	if err := tx.Model(&domain.ProductoNegocio{}).Where("negocio_id = ? AND sku_interno = ?", businessID, sku).Count(&duplicate).Error; err != nil {
		return err
	}
	if duplicate > 0 {
		return errors.New("el SKU ya existe en este negocio")
	}
	barcode := optionalString(variant.CodigoBarras)
	if barcode != "" {
		if !productBarcodePattern.MatchString(barcode) {
			return errors.New("el código de barras debe contener entre 8 y 14 dígitos")
		}
		if err := tx.Model(&domain.ProductoCodigo{}).Where("negocio_id = ? AND codigo = ?", businessID, barcode).Count(&duplicate).Error; err != nil {
			return err
		}
		if duplicate > 0 {
			return errors.New("el código de barras ya está asignado a otro producto")
		}
	}
	productVariant := domain.ProductoVariante{ProductoID: product.ID, FamiliaProductoID: family.ID, Clave: variantKey, Activo: true}
	if err := tx.Create(&productVariant).Error; err != nil {
		return err
	}
	for _, attribute := range variant.Atributos {
		if err := tx.Create(&domain.ProductoVarianteAtributo{ProductoVarianteID: productVariant.ID, Nombre: strings.TrimSpace(attribute.Nombre), Valor: strings.TrimSpace(attribute.Valor)}).Error; err != nil {
			return err
		}
	}
	if barcode != "" {
		if err := tx.Create(&domain.ProductoCodigo{NegocioID: businessID, ProductoVarianteID: &productVariant.ID, Tipo: "GTIN", Codigo: barcode, EsPrincipal: true}).Error; err != nil {
			return err
		}
	}
	commercial := domain.ProductoNegocio{NegocioID: businessID, ProductoVarianteID: &productVariant.ID, SKUInterno: &sku, PrecioVenta: variant.PrecioVenta, PrecioIncluyeImpuestos: true, Activo: true}
	if err := tx.Create(&commercial).Error; err != nil {
		return err
	}
	if err := replaceProductProvider(tx, commercial.ID, providerID); err != nil {
		return err
	}
	inventory := domain.InventarioSucursal{SucursalID: branchID, ProductoNegocioID: commercial.ID, StockActual: variant.StockInicial, StockMinimo: 0}
	if err := tx.Create(&inventory).Error; err != nil {
		return err
	}
	if variant.StockInicial > 0 {
		return tx.Create(&domain.MovimientoInventario{SucursalID: branchID, ProductoNegocioID: commercial.ID, Tipo: "AJUSTE_ENTRADA", Cantidad: variant.StockInicial, StockAnterior: 0, StockNuevo: variant.StockInicial}).Error
	}
	return nil
}

type variantBaseInput struct {
	Nombre          string
	Descripcion     *string
	MarcaID         *uuid.UUID
	CategoriaID     *uuid.UUID
	UnidadMedidaID  *uuid.UUID
	Contenido       *float64
	UnidadContenido *string
	Presentacion    *string
	ImagenURL       *string
}

// resolveVariantBase is the sole owner of a variant family base product.  It
// reuses the parent on subsequent manual entries and spreadsheet rows.
func (r *ProductRepository) resolveVariantBase(tx *gorm.DB, businessID uuid.UUID, input variantBaseInput) (domain.FamiliaProducto, domain.Producto, bool, error) {
	familyKey := productBaseKey(input.Nombre, input.Presentacion, input.MarcaID, input.CategoriaID)
	var family domain.FamiliaProducto
	err := tx.Where("negocio_id = ? AND clave = ?", businessID, familyKey).First(&family).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		code, codeErr := reserveGeneratedFamilyCode(tx, businessID, input.Nombre)
		if codeErr != nil {
			return family, domain.Producto{}, false, codeErr
		}
		family = domain.FamiliaProducto{NegocioID: businessID, Nombre: input.Nombre, Codigo: code, Clave: familyKey, Activo: true}
		if err := tx.Create(&family).Error; err != nil {
			return family, domain.Producto{}, false, err
		}
	} else if err != nil {
		return family, domain.Producto{}, false, err
	}

	var product domain.Producto
	err = tx.Where("negocio_id = ? AND familia_producto_id = ?", businessID, family.ID).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		product = domain.Producto{NegocioID: businessID, Nombre: input.Nombre, Descripcion: input.Descripcion, MarcaID: input.MarcaID, UnidadMedidaID: input.UnidadMedidaID, FamiliaProductoID: &family.ID, Contenido: input.Contenido, UnidadContenido: input.UnidadContenido, Presentacion: input.Presentacion, Activo: true}
		if err := tx.Create(&product).Error; err != nil {
			return family, domain.Producto{}, false, err
		}
		if input.CategoriaID != nil {
			if err := tx.Create(&domain.ProductoCategoria{NegocioID: businessID, ProductoID: product.ID, CategoriaID: *input.CategoriaID, EsPrincipal: true}).Error; err != nil {
				return family, domain.Producto{}, false, err
			}
		}
		if imageURL := optionalString(input.ImagenURL); imageURL != "" {
			if err := tx.Create(&domain.ProductoImagen{ProductoID: product.ID, URL: imageURL, EsPrincipal: true, Orden: 0}).Error; err != nil {
				return family, domain.Producto{}, false, err
			}
		}
		return family, product, true, nil
	}
	if err != nil {
		return family, domain.Producto{}, false, err
	}
	if !sameVariantBaseData(tx, product, input) {
		return family, domain.Producto{}, false, errors.New("los datos compartidos no coinciden con el producto base existente")
	}
	return family, product, false, nil
}

func sameVariantBaseData(tx *gorm.DB, product domain.Producto, input variantBaseInput) bool {
	if !sameOptionalString(product.Descripcion, input.Descripcion) || !sameOptionalUUID(product.UnidadMedidaID, input.UnidadMedidaID) ||
		!sameOptionalFloat(product.Contenido, input.Contenido) || !sameOptionalString(product.UnidadContenido, input.UnidadContenido) {
		return false
	}
	var imageURL string
	if err := tx.Table("producto_imagenes").Select("url").Where("producto_id = ? AND es_principal = TRUE", product.ID).Limit(1).Scan(&imageURL).Error; err != nil {
		return false
	}
	return catalogImportKey(imageURL) == catalogImportKey(optionalString(input.ImagenURL))
}

func sameVariantBaseInput(first, second variantBaseInput) bool {
	return sameOptionalString(first.Descripcion, second.Descripcion) &&
		sameOptionalUUID(first.UnidadMedidaID, second.UnidadMedidaID) &&
		sameOptionalFloat(first.Contenido, second.Contenido) &&
		sameOptionalString(first.UnidadContenido, second.UnidadContenido) &&
		sameOptionalString(first.ImagenURL, second.ImagenURL)
}

func sameOptionalString(first, second *string) bool {
	return catalogImportKey(optionalString(first)) == catalogImportKey(optionalString(second))
}

func sameOptionalUUID(first, second *uuid.UUID) bool {
	if first == nil || second == nil {
		return first == nil && second == nil
	}
	return *first == *second
}

func sameOptionalFloat(first, second *float64) bool {
	if first == nil || second == nil {
		return first == nil && second == nil
	}
	return *first == *second
}

func productFamilyKey(name string, brandID *uuid.UUID, presentation *string, categoryID *uuid.UUID) string {
	brand := ""
	if brandID != nil {
		brand = brandID.String()
	}
	category := ""
	if categoryID != nil {
		category = categoryID.String()
	}
	canonical := strings.Join([]string{catalogImportKey(name), catalogImportKey(optionalString(presentation)), brand, category}, "|")
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

func productBaseKey(name string, presentation *string, brandID *uuid.UUID, categoryID *uuid.UUID) string {
	return productFamilyKey(name, brandID, presentation, categoryID)
}

func reserveGeneratedFamilyCode(tx *gorm.DB, businessID uuid.UUID, name string) (string, error) {
	var sequence struct {
		Numero int64 `gorm:"column:numero"`
	}
	if err := tx.Raw(`INSERT INTO familia_producto_consecutivos (id, negocio_id, siguiente_numero) VALUES (?, ?, 2) ON CONFLICT (negocio_id) DO UPDATE SET siguiente_numero = familia_producto_consecutivos.siguiente_numero + 1 RETURNING siguiente_numero - 1 AS numero`, uuid.New(), businessID).Scan(&sequence).Error; err != nil {
		return "", err
	}
	slug := strings.ToUpper(catalogImportKey(name))
	slug = strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, slug)
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "PRODUCTO"
	}
	if len(slug) > 80 {
		slug = slug[:80]
	}
	return fmt.Sprintf("FAM-%s-%06d", slug, sequence.Numero), nil
}

func productVariantKey(attributes []domain.ProductVariantAttributeInput) (string, error) {
	if len(attributes) == 0 {
		return "", errors.New("cada variante requiere al menos un atributo")
	}
	items := make([]string, 0, len(attributes))
	seen := make(map[string]struct{}, len(attributes))
	for _, attribute := range attributes {
		name, value := catalogImportKey(attribute.Nombre), catalogImportKey(attribute.Valor)
		if name == "" || value == "" {
			return "", errors.New("los atributos de variante requieren nombre y valor")
		}
		if _, exists := seen[name]; exists {
			return "", errors.New("no puede repetir el mismo atributo en una variante")
		}
		seen[name] = struct{}{}
		items = append(items, name+"="+value)
	}
	sort.Strings(items)
	return strings.Join(items, ";"), nil
}

func reserveGeneratedProductSKU(tx *gorm.DB, businessID uuid.UUID) (string, error) {
	for {
		now := time.Now().UTC()
		var sequence struct {
			Numero int64 `gorm:"column:numero"`
		}
		if err := tx.Raw(`
			INSERT INTO producto_sku_consecutivos (id, negocio_id, siguiente_numero, creado_en, actualizado_en)
			VALUES (?, ?, 2, ?, ?)
			ON CONFLICT (negocio_id) DO UPDATE
			SET siguiente_numero = producto_sku_consecutivos.siguiente_numero + 1,
				actualizado_en = ?
			RETURNING siguiente_numero - 1 AS numero
		`, uuid.New(), businessID, now, now, now).Scan(&sequence).Error; err != nil {
			return "", err
		}
		sku := fmt.Sprintf("PROD-%06d", sequence.Numero)
		var duplicate int64
		if err := tx.Model(&domain.ProductoNegocio{}).Where("negocio_id = ? AND sku_interno = ?", businessID, sku).Count(&duplicate).Error; err != nil {
			return "", err
		}
		if duplicate == 0 {
			return sku, nil
		}
	}
}

func (r *ProductRepository) Update(businessID, productID uuid.UUID, input domain.UpdateProductInput) error {
	input.Nombre = strings.TrimSpace(input.Nombre)
	if input.Nombre == "" || input.PrecioVenta < 0 {
		return errors.New("los datos del producto no son válidos")
	}
	barcode := optionalString(input.CodigoBarras)
	imageURL := optionalString(input.ImagenURL)
	if barcode != "" && !productBarcodePattern.MatchString(barcode) {
		return errors.New("el código de barras debe contener entre 8 y 14 dígitos")
	}
	if imageURL != "" {
		parsed, err := url.Parse(imageURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return errors.New("la URL de imagen no es válida")
		}
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		var product domain.Producto
		if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", productID, businessID).First(&product).Error; err != nil {
			return errors.New("producto no encontrado")
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
		if err := validateProductProvider(tx, businessID, input.ProveedorID); err != nil {
			return err
		}
		if input.UnidadMedidaID != nil {
			var unit domain.UnidadMedida
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", *input.UnidadMedidaID, businessID).First(&unit).Error; err != nil {
				return errors.New("la unidad de medida no existe o está inactiva")
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
		if err := replaceProductCategory(tx, businessID, productID, input.CategoriaID); err != nil {
			return err
		}
		if err := replaceProductImage(tx, productID, imageURL); err != nil {
			return err
		}

		if product.FamiliaProductoID == nil {
			var commercial domain.ProductoNegocio
			if err := tx.Where("negocio_id = ? AND producto_id = ? AND activo = TRUE", businessID, productID).First(&commercial).Error; err != nil {
				return errors.New("producto no encontrado")
			}
			if len(input.Variantes) > 0 {
				return r.convertSimpleProductToFamily(tx, businessID, product, commercial, input)
			}
			if err := tx.Model(&domain.ProductoNegocio{}).Where("id = ?", commercial.ID).Update("precio_venta", input.PrecioVenta).Error; err != nil {
				return err
			}
			if err := replaceProductProvider(tx, commercial.ID, input.ProveedorID); err != nil {
				return err
			}
			return replaceSimpleProductBarcode(tx, businessID, productID, barcode)
		}

		if barcode != "" {
			return errors.New("el código de barras de una familia se captura en cada variante")
		}
		return r.updateProductFamily(tx, businessID, product, input)
	})
}

func (r *ProductRepository) convertSimpleProductToFamily(tx *gorm.DB, businessID uuid.UUID, product domain.Producto, commercial domain.ProductoNegocio, input domain.UpdateProductInput) error {
	if len(input.Variantes) == 0 {
		return errors.New("se requiere al menos una variante")
	}
	seenKeys := make(map[string]struct{}, len(input.Variantes))
	for _, variant := range input.Variantes {
		if variant.ID != nil || variant.PrecioVenta < 0 {
			return errors.New("los datos de cada variante deben ser válidos")
		}
		key, err := productVariantKey(variant.Atributos)
		if err != nil {
			return err
		}
		if _, duplicate := seenKeys[key]; duplicate {
			return errors.New("no puede repetir la misma combinación de variantes")
		}
		seenKeys[key] = struct{}{}
	}
	familyKey := productBaseKey(input.Nombre, input.Presentacion, input.MarcaID, input.CategoriaID)
	var duplicate int64
	if err := tx.Model(&domain.FamiliaProducto{}).Where("negocio_id = ? AND clave = ?", businessID, familyKey).Count(&duplicate).Error; err != nil {
		return err
	}
	if duplicate > 0 {
		return errors.New("ya existe una familia con el mismo nombre, marca, categoría y presentación")
	}
	code, err := reserveGeneratedFamilyCode(tx, businessID, input.Nombre)
	if err != nil {
		return err
	}
	family := domain.FamiliaProducto{NegocioID: businessID, Nombre: input.Nombre, Codigo: code, Clave: familyKey, Activo: true}
	if err := tx.Create(&family).Error; err != nil {
		return err
	}
	// The former simple commercial record remains as historical data.  It is not a
	// sellable variant: stock, SKU and barcode must be deliberately assigned to a
	// new variant by the user.
	if err := tx.Where("negocio_id = ? AND producto_id = ?", businessID, product.ID).Delete(&domain.ProductoCodigo{}).Error; err != nil {
		return err
	}
	if err := tx.Model(&domain.ProductoNegocio{}).Where("id = ?", commercial.ID).Update("activo", false).Error; err != nil {
		return err
	}
	if err := tx.Model(&domain.Producto{}).Where("id = ?", product.ID).Update("familia_producto_id", family.ID).Error; err != nil {
		return err
	}
	for _, variant := range input.Variantes {
		key, err := productVariantKey(variant.Atributos)
		if err != nil {
			return err
		}
		if err := r.createFamilyVariantOnUpdate(tx, businessID, product, family, variant, key, input.ProveedorID); err != nil {
			return err
		}
	}
	return nil
}

func (r *ProductRepository) AdjustInventory(businessID, productID, userID uuid.UUID, input domain.AdjustInventoryInput) error {
	if input.Stock < 0 {
		return errors.New("el stock no puede ser negativo")
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var product domain.Producto
		if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", productID, businessID).First(&product).Error; err != nil {
			return errors.New("producto no encontrado")
		}
		query := tx.Table("producto_negocio AS pn").Where("pn.negocio_id = ? AND pn.activo = TRUE", businessID)
		if input.VarianteID != nil {
			query = query.Where("pn.producto_variante_id = ? AND pn.producto_variante_id IN (SELECT id FROM producto_variantes WHERE producto_id = ? AND activo = TRUE)", *input.VarianteID, productID)
		} else {
			query = query.Where("pn.producto_id = ?", productID)
		}
		var commercial domain.ProductoNegocio
		if err := query.First(&commercial).Error; err != nil {
			return errors.New("la variante o presentación del producto no existe")
		}
		var inventory domain.InventarioSucursal
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sucursal_id = ? AND producto_negocio_id = ?", input.SucursalID, commercial.ID).First(&inventory).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			inventory = domain.InventarioSucursal{SucursalID: input.SucursalID, ProductoNegocioID: commercial.ID, StockActual: 0}
		} else if err != nil {
			return err
		}
		previous := inventory.StockActual
		if previous == input.Stock {
			return nil
		}
		if inventory.ID == uuid.Nil {
			inventory.StockActual = input.Stock
			if err := tx.Create(&inventory).Error; err != nil {
				return err
			}
		} else if err := tx.Model(&inventory).Update("stock_actual", input.Stock).Error; err != nil {
			return err
		}
		delta := input.Stock - previous
		return tx.Create(&domain.MovimientoInventario{SucursalID: input.SucursalID, ProductoNegocioID: commercial.ID, Tipo: "AJUSTE_MANUAL", Cantidad: delta, StockAnterior: previous, StockNuevo: input.Stock, UsuarioID: &userID}).Error
	})
}

func replaceProductCategory(tx *gorm.DB, businessID, productID uuid.UUID, categoryID *uuid.UUID) error {
	if err := tx.Where("negocio_id = ? AND producto_id = ? AND es_principal = TRUE", businessID, productID).Delete(&domain.ProductoCategoria{}).Error; err != nil {
		return err
	}
	if categoryID == nil {
		return nil
	}
	return tx.Create(&domain.ProductoCategoria{NegocioID: businessID, ProductoID: productID, CategoriaID: *categoryID, EsPrincipal: true}).Error
}

func replaceProductImage(tx *gorm.DB, productID uuid.UUID, imageURL string) error {
	if err := tx.Where("producto_id = ? AND es_principal = TRUE", productID).Delete(&domain.ProductoImagen{}).Error; err != nil {
		return err
	}
	if imageURL == "" {
		return nil
	}
	return tx.Create(&domain.ProductoImagen{ProductoID: productID, URL: imageURL, EsPrincipal: true, Orden: 0}).Error
}

func replaceSimpleProductBarcode(tx *gorm.DB, businessID, productID uuid.UUID, barcode string) error {
	if barcode != "" {
		if err := ensureBarcodeAvailable(tx, businessID, barcode, &productID, nil); err != nil {
			return err
		}
	}
	if err := tx.Where("negocio_id = ? AND producto_id = ? AND es_principal = TRUE", businessID, productID).Delete(&domain.ProductoCodigo{}).Error; err != nil {
		return err
	}
	if barcode == "" {
		return nil
	}
	return tx.Create(&domain.ProductoCodigo{NegocioID: businessID, ProductoID: &productID, Tipo: "GTIN", Codigo: barcode, EsPrincipal: true}).Error
}

func ensureBarcodeAvailable(tx *gorm.DB, businessID uuid.UUID, barcode string, productID, variantID *uuid.UUID) error {
	var codes []domain.ProductoCodigo
	if err := tx.Where("negocio_id = ? AND codigo = ?", businessID, barcode).Find(&codes).Error; err != nil {
		return err
	}
	for _, code := range codes {
		if productID != nil && code.ProductoID != nil && *productID == *code.ProductoID {
			continue
		}
		if variantID != nil && code.ProductoVarianteID != nil && *variantID == *code.ProductoVarianteID {
			continue
		}
		return errors.New("el código de barras ya está asignado a otro producto")
	}
	return nil
}

func (r *ProductRepository) updateProductFamily(tx *gorm.DB, businessID uuid.UUID, product domain.Producto, input domain.UpdateProductInput) error {
	if len(input.Variantes) == 0 {
		return errors.New("una familia debe conservar al menos una variante")
	}

	var family domain.FamiliaProducto
	if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", *product.FamiliaProductoID, businessID).First(&family).Error; err != nil {
		return errors.New("la familia del producto no existe")
	}
	newFamilyKey := productBaseKey(input.Nombre, input.Presentacion, input.MarcaID, input.CategoriaID)
	if newFamilyKey != family.Clave {
		var duplicates int64
		if err := tx.Model(&domain.FamiliaProducto{}).Where("negocio_id = ? AND clave = ? AND id <> ?", businessID, newFamilyKey, family.ID).Count(&duplicates).Error; err != nil {
			return err
		}
		if duplicates > 0 {
			return errors.New("ya existe una familia con el mismo nombre, marca, categoría y presentación")
		}
	}
	if err := tx.Model(&domain.FamiliaProducto{}).Where("id = ?", family.ID).Updates(map[string]interface{}{"nombre": input.Nombre, "clave": newFamilyKey}).Error; err != nil {
		return err
	}

	type existingVariant struct {
		ID           uuid.UUID `gorm:"column:id"`
		CommercialID uuid.UUID `gorm:"column:commercial_id"`
	}
	var existingRows []existingVariant
	if err := tx.Table("producto_variantes AS pv").
		Select("pv.id, pn.id AS commercial_id").
		Joins("JOIN producto_negocio pn ON pn.producto_variante_id = pv.id AND pn.negocio_id = ? AND pn.activo = TRUE", businessID).
		Where("pv.producto_id = ? AND pv.activo = TRUE", product.ID).
		Scan(&existingRows).Error; err != nil {
		return err
	}
	existing := make(map[uuid.UUID]existingVariant, len(existingRows))
	for _, row := range existingRows {
		existing[row.ID] = row
	}
	seenIDs := make(map[uuid.UUID]struct{}, len(input.Variantes))
	seenKeys := make(map[string]struct{}, len(input.Variantes))

	for _, variant := range input.Variantes {
		if variant.PrecioVenta < 0 {
			return errors.New("el precio de cada variante debe ser válido")
		}
		key, err := productVariantKey(variant.Atributos)
		if err != nil {
			return err
		}
		if _, duplicate := seenKeys[key]; duplicate {
			return errors.New("no puede repetir la misma combinación de variantes")
		}
		seenKeys[key] = struct{}{}

		if variant.ID == nil {
			if err := r.createFamilyVariantOnUpdate(tx, businessID, product, family, variant, key, input.ProveedorID); err != nil {
				return err
			}
			continue
		}
		current, exists := existing[*variant.ID]
		if !exists {
			return errors.New("una de las variantes no pertenece a este producto")
		}
		if _, duplicate := seenIDs[*variant.ID]; duplicate {
			return errors.New("una variante fue enviada más de una vez")
		}
		seenIDs[*variant.ID] = struct{}{}
		if err := updateExistingFamilyVariant(tx, businessID, current, variant, key, input.ProveedorID); err != nil {
			return err
		}
	}

	for id, current := range existing {
		if _, kept := seenIDs[id]; kept {
			continue
		}
		if err := tx.Model(&domain.ProductoVariante{}).Where("id = ?", id).Update("activo", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.ProductoNegocio{}).Where("id = ?", current.CommercialID).Update("activo", false).Error; err != nil {
			return err
		}
	}
	return nil
}

func updateExistingFamilyVariant(tx *gorm.DB, businessID uuid.UUID, current struct {
	ID           uuid.UUID `gorm:"column:id"`
	CommercialID uuid.UUID `gorm:"column:commercial_id"`
}, input domain.UpdateProductVariantInput, key string, providerID *uuid.UUID) error {
	barcode := optionalString(input.CodigoBarras)
	if barcode != "" {
		if !productBarcodePattern.MatchString(barcode) {
			return errors.New("el código de barras debe contener entre 8 y 14 dígitos")
		}
		if err := ensureBarcodeAvailable(tx, businessID, barcode, nil, &current.ID); err != nil {
			return err
		}
	}
	if err := tx.Model(&domain.ProductoVariante{}).Where("id = ?", current.ID).Update("clave", key).Error; err != nil {
		return err
	}
	if err := tx.Where("producto_variante_id = ?", current.ID).Delete(&domain.ProductoVarianteAtributo{}).Error; err != nil {
		return err
	}
	for _, attribute := range input.Atributos {
		if err := tx.Create(&domain.ProductoVarianteAtributo{ProductoVarianteID: current.ID, Nombre: strings.TrimSpace(attribute.Nombre), Valor: strings.TrimSpace(attribute.Valor)}).Error; err != nil {
			return err
		}
	}
	if err := tx.Model(&domain.ProductoNegocio{}).Where("id = ?", current.CommercialID).Update("precio_venta", input.PrecioVenta).Error; err != nil {
		return err
	}
	if err := replaceProductProvider(tx, current.CommercialID, providerID); err != nil {
		return err
	}
	if err := tx.Where("negocio_id = ? AND producto_variante_id = ? AND es_principal = TRUE", businessID, current.ID).Delete(&domain.ProductoCodigo{}).Error; err != nil {
		return err
	}
	if barcode == "" {
		return nil
	}
	return tx.Create(&domain.ProductoCodigo{NegocioID: businessID, ProductoVarianteID: &current.ID, Tipo: "GTIN", Codigo: barcode, EsPrincipal: true}).Error
}

func (r *ProductRepository) createFamilyVariantOnUpdate(tx *gorm.DB, businessID uuid.UUID, product domain.Producto, family domain.FamiliaProducto, input domain.UpdateProductVariantInput, key string, providerID *uuid.UUID) error {
	if input.PrecioVenta < 0 {
		return errors.New("el precio de cada variante debe ser válido")
	}
	barcode := optionalString(input.CodigoBarras)
	if barcode != "" && !productBarcodePattern.MatchString(barcode) {
		return errors.New("el código de barras debe contener entre 8 y 14 dígitos")
	}
	if barcode != "" {
		if err := ensureBarcodeAvailable(tx, businessID, barcode, nil, nil); err != nil {
			return err
		}
	}
	sku, err := reserveGeneratedProductSKU(tx, businessID)
	if err != nil {
		return err
	}
	variant := domain.ProductoVariante{ProductoID: product.ID, FamiliaProductoID: family.ID, Clave: key, Activo: true}
	if err := tx.Create(&variant).Error; err != nil {
		return err
	}
	for _, attribute := range input.Atributos {
		if err := tx.Create(&domain.ProductoVarianteAtributo{ProductoVarianteID: variant.ID, Nombre: strings.TrimSpace(attribute.Nombre), Valor: strings.TrimSpace(attribute.Valor)}).Error; err != nil {
			return err
		}
	}
	if barcode != "" {
		if err := tx.Create(&domain.ProductoCodigo{NegocioID: businessID, ProductoVarianteID: &variant.ID, Tipo: "GTIN", Codigo: barcode, EsPrincipal: true}).Error; err != nil {
			return err
		}
	}
	commercial := domain.ProductoNegocio{NegocioID: businessID, ProductoVarianteID: &variant.ID, SKUInterno: &sku, PrecioVenta: input.PrecioVenta, PrecioIncluyeImpuestos: true, Activo: true}
	if err := tx.Create(&commercial).Error; err != nil {
		return err
	}
	return replaceProductProvider(tx, commercial.ID, providerID)
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

func (r *ProductRepository) ValidateProductImport(businessID, branchID uuid.UUID, rows []domain.ProductImportRow) ([]domain.ValidatedProductImportRow, domain.CatalogImportResult, []domain.ProductImportCatalogProposal, error) {
	result := domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0), Advertencias: make([]domain.CatalogImportIssue, 0)}
	var branch domain.Sucursal
	if err := r.db.Where("id = ? AND negocio_id = ? AND activo = TRUE", branchID, businessID).First(&branch).Error; err != nil {
		return nil, result, nil, errors.New("la sucursal no pertenece al negocio o está inactiva")
	}
	var categories []domain.Categoria
	var brands []domain.Marca
	var providers []domain.Proveedor
	var units []domain.UnidadMedida
	if err := r.db.Where("negocio_id = ? AND activo = TRUE", businessID).Find(&categories).Error; err != nil {
		return nil, result, nil, err
	}
	if err := r.db.Where("negocio_id = ? AND activo = TRUE", businessID).Find(&brands).Error; err != nil {
		return nil, result, nil, err
	}
	if err := r.db.Where("negocio_id = ? AND activo = TRUE", businessID).Find(&providers).Error; err != nil {
		return nil, result, nil, err
	}
	if err := r.db.Where("negocio_id = ? AND activo = TRUE", businessID).Find(&units).Error; err != nil {
		return nil, result, nil, err
	}
	categoryIDs := make(map[string]uuid.UUID, len(categories))
	brandIDs := make(map[string]uuid.UUID, len(brands))
	providerIDs := make(map[string]uuid.UUID, len(providers))
	unitIDs := make(map[string]uuid.UUID, len(units))
	for _, item := range categories {
		categoryIDs[catalogImportKey(item.Nombre)] = item.ID
	}
	for _, item := range brands {
		brandIDs[catalogImportKey(item.Nombre)] = item.ID
	}
	for _, item := range providers {
		providerIDs[catalogImportKey(item.Nombre)] = item.ID
	}
	for _, item := range units {
		unitIDs[catalogImportKey(item.Nombre)] = item.ID
	}
	var existingCodes []string
	if err := r.db.Model(&domain.ProductoCodigo{}).Where("negocio_id = ?", businessID).Pluck("codigo", &existingCodes).Error; err != nil {
		return nil, result, nil, err
	}
	type productIdentity struct {
		Nombre            string
		Presentacion      string
		FamiliaProductoID *uuid.UUID
	}
	var existingProducts []productIdentity
	if err := r.db.Table("productos AS p").
		Select("p.nombre, COALESCE(p.presentacion, '') AS presentacion, p.familia_producto_id").
		Joins("JOIN producto_negocio pn ON pn.producto_id = p.id AND pn.negocio_id = ?", businessID).
		Scan(&existingProducts).Error; err != nil {
		return nil, result, nil, err
	}
	existingCodeSet := make(map[string]struct{}, len(existingCodes))
	seenFileCodes := make(map[string]struct{})
	seenNamePresentations := make(map[string]struct{}, len(existingProducts))
	seenVariantIdentities := make(map[string]struct{})
	for _, code := range existingCodes {
		existingCodeSet[code] = struct{}{}
	}
	for _, product := range existingProducts {
		if product.FamiliaProductoID == nil {
			seenNamePresentations[productImportNamePresentationKey(product.Nombre, product.Presentacion)] = struct{}{}
		}
	}
	var existingVariants []struct {
		Nombre       string     `gorm:"column:nombre"`
		Presentacion string     `gorm:"column:presentacion"`
		MarcaID      *uuid.UUID `gorm:"column:marca_id"`
		CategoriaID  *uuid.UUID `gorm:"column:categoria_id"`
		Clave        string     `gorm:"column:clave"`
	}
	if err := r.db.Table("producto_variantes AS pv").
		Select("p.nombre, COALESCE(p.presentacion, '') AS presentacion, p.marca_id, pc.categoria_id, pv.clave").
		Joins("JOIN productos p ON p.id = pv.producto_id").
		Joins("LEFT JOIN producto_categorias pc ON pc.producto_id = p.id AND pc.es_principal = TRUE").
		Joins("JOIN producto_negocio pn ON pn.producto_variante_id = pv.id AND pn.negocio_id = ?", businessID).
		Where("pv.activo = TRUE").Scan(&existingVariants).Error; err != nil {
		return nil, result, nil, err
	}
	for _, variant := range existingVariants {
		seenVariantIdentities[productBaseKey(variant.Nombre, importOptionalString(variant.Presentacion), variant.MarcaID, variant.CategoriaID)+"\x00"+variant.Clave] = struct{}{}
	}
	seenVariantBases := make(map[string]variantBaseInput)
	seenVariantProviders := make(map[string]*uuid.UUID)
	prepared := make([]domain.ValidatedProductImportRow, 0, len(rows))
	proposals := make(map[string]domain.ProductImportCatalogProposal)

	for _, row := range rows {
		hasError := false
		hasDuplicate := false
		pendingCatalogs := make([]domain.ProductImportCatalogProposal, 0, 4)
		addError := func(field, reason string) {
			hasError = true
			result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Campo: field, Motivo: reason})
		}
		namePresentationKey := productImportNamePresentationKey(row.Nombre, row.Presentacion)
		// A row with variant attributes is identified by its attribute combination.
		// Empty barcode values must not turn different flavours/colours into
		// duplicate simple products.
		if len(row.Variantes) == 0 {
			if _, exists := seenNamePresentations[namePresentationKey]; exists {
				hasDuplicate = true
				field, reason := simpleProductDuplicateIssue(row.CodigoBarras, existingCodeSet, seenFileCodes)
				addError(field, reason)
			}
		}
		if row.CodigoBarras != "" {
			if _, exists := existingCodeSet[row.CodigoBarras]; exists {
				hasDuplicate = true
				addError("Código de barras", "ya está registrado en el negocio")
			} else if _, exists := seenFileCodes[row.CodigoBarras]; exists {
				hasDuplicate = true
				addError("Código de barras", "está repetido en el archivo")
			}
		}
		var categoryID *uuid.UUID
		if row.Categoria != "" {
			id, ok := categoryIDs[catalogImportKey(row.Categoria)]
			if !ok {
				pendingCatalogs = append(pendingCatalogs, domain.ProductImportCatalogProposal{Tipo: "Categoría", Nombre: strings.TrimSpace(row.Categoria)})
			} else {
				categoryID = &id
			}
		}
		var brandID *uuid.UUID
		if row.Marca != "" {
			id, ok := brandIDs[catalogImportKey(row.Marca)]
			if !ok {
				pendingCatalogs = append(pendingCatalogs, domain.ProductImportCatalogProposal{Tipo: "Marca", Nombre: strings.TrimSpace(row.Marca)})
			} else {
				brandID = &id
			}
		}
		var providerID *uuid.UUID
		if row.Proveedor != "" {
			id, ok := providerIDs[catalogImportKey(row.Proveedor)]
			if !ok {
				pendingCatalogs = append(pendingCatalogs, domain.ProductImportCatalogProposal{Tipo: "Proveedor", Nombre: strings.TrimSpace(row.Proveedor)})
			} else {
				providerID = &id
			}
		}
		var unitID *uuid.UUID
		if row.UnidadMedida != "" {
			id, ok := unitIDs[catalogImportKey(row.UnidadMedida)]
			if !ok {
				pendingCatalogs = append(pendingCatalogs, domain.ProductImportCatalogProposal{Tipo: "Unidad de medida", Nombre: strings.TrimSpace(row.UnidadMedida)})
			} else {
				unitID = &id
			}
		}
		if len(row.Variantes) > 0 {
			variantKey, variantErr := productVariantKey(row.Variantes)
			if variantErr != nil {
				addError("Variantes", variantErr.Error())
			} else {
				base := variantBaseInput{Nombre: row.Nombre, Descripcion: importOptionalString(row.Descripcion), MarcaID: brandID, CategoriaID: categoryID, UnidadMedidaID: unitID, Contenido: row.Contenido, UnidadContenido: importOptionalString(row.UnidadContenido), Presentacion: importOptionalString(row.Presentacion), ImagenURL: importOptionalString(row.ImagenURL)}
				baseKey := productBaseKey(base.Nombre, base.Presentacion, base.MarcaID, base.CategoriaID)
				if previous, exists := seenVariantBases[baseKey]; exists && !sameVariantBaseInput(previous, base) {
					addError("Producto base", "los datos compartidos no coinciden con otra fila del archivo")
				} else {
					seenVariantBases[baseKey] = base
				}
				if previous, exists := seenVariantProviders[baseKey]; exists && !sameOptionalUUID(previous, providerID) {
					addError("Proveedor", "debe ser el mismo para todas las variantes del producto")
				} else {
					seenVariantProviders[baseKey] = providerID
				}
				variantIdentity := baseKey + "\x00" + variantKey
				if _, exists := seenVariantIdentities[variantIdentity]; exists {
					hasDuplicate = true
					addError("Variantes", "la combinación ya existe o está repetida en el archivo")
				}
				seenVariantIdentities[variantIdentity] = struct{}{}
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
		for _, proposal := range pendingCatalogs {
			key := proposal.Tipo + "\x00" + catalogImportKey(proposal.Nombre)
			current := proposals[key]
			current.Tipo = proposal.Tipo
			current.Nombre = proposal.Nombre
			current.Productos++
			proposals[key] = current
		}
		input := domain.CreateImportedProductInput{Nombre: row.Nombre, CategoriaID: categoryID, MarcaID: brandID, ProveedorID: providerID, SucursalID: branchID, Contenido: row.Contenido, PrecioVenta: row.PrecioVenta, StockInicial: row.StockInicial}
		input.Descripcion = importOptionalString(row.Descripcion)
		input.Presentacion = importOptionalString(row.Presentacion)
		input.UnidadContenido = importOptionalString(row.UnidadContenido)
		input.UnidadMedidaID = unitID
		input.CodigoBarras = importOptionalString(row.CodigoBarras)
		input.ImagenURL = importOptionalString(row.ImagenURL)
		input.Variantes = row.Variantes
		if row.CodigoBarras != "" {
			seenFileCodes[row.CodigoBarras] = struct{}{}
		}
		if len(row.Variantes) == 0 {
			seenNamePresentations[namePresentationKey] = struct{}{}
		}
		prepared = append(prepared, domain.ValidatedProductImportRow{Fila: row.Fila, Input: input})
	}
	items := make([]domain.ProductImportCatalogProposal, 0, len(proposals))
	for _, proposal := range proposals {
		items = append(items, proposal)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Tipo == items[j].Tipo {
			return catalogImportKey(items[i].Nombre) < catalogImportKey(items[j].Nombre)
		}
		return items[i].Tipo < items[j].Tipo
	})
	return prepared, result, items, nil
}

func simpleProductDuplicateIssue(barcode string, existingCodes, fileCodes map[string]struct{}) (string, string) {
	if barcode != "" {
		_, existsInBusiness := existingCodes[barcode]
		_, existsInFile := fileCodes[barcode]
		if !existsInBusiness && !existsInFile {
			return "Variantes", "el producto tiene otro código de barras; indica la variante con el formato Atributo: valor"
		}
	}
	return "Producto", "ya existe en el negocio o está repetido en el archivo"
}

func (r *ProductRepository) CreateImportedCatalogs(businessID uuid.UUID, proposals []domain.ProductImportCatalogProposal) error {
	if len(proposals) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var categories []domain.Categoria
		var brands []domain.Marca
		var providers []domain.Proveedor
		var units []domain.UnidadMedida
		if err := tx.Where("negocio_id = ? AND activo = TRUE", businessID).Find(&categories).Error; err != nil {
			return err
		}
		if err := tx.Where("negocio_id = ? AND activo = TRUE", businessID).Find(&brands).Error; err != nil {
			return err
		}
		if err := tx.Where("negocio_id = ? AND activo = TRUE", businessID).Find(&providers).Error; err != nil {
			return err
		}
		if err := tx.Where("negocio_id = ? AND activo = TRUE", businessID).Find(&units).Error; err != nil {
			return err
		}
		categoryKeys, brandKeys, providerKeys, unitKeys := map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}
		for _, item := range categories {
			categoryKeys[catalogImportKey(item.Nombre)] = struct{}{}
		}
		for _, item := range brands {
			brandKeys[catalogImportKey(item.Nombre)] = struct{}{}
		}
		for _, item := range providers {
			providerKeys[catalogImportKey(item.Nombre)] = struct{}{}
		}
		for _, item := range units {
			unitKeys[catalogImportKey(item.Nombre)] = struct{}{}
			unitKeys[catalogImportKey(item.Simbolo)] = struct{}{}
		}
		for _, proposal := range proposals {
			name, key := strings.TrimSpace(proposal.Nombre), catalogImportKey(proposal.Nombre)
			if name == "" {
				continue
			}
			switch proposal.Tipo {
			case "Categoría":
				if _, exists := categoryKeys[key]; !exists {
					if err := tx.Create(&domain.Categoria{NegocioID: businessID, Nombre: name, Activo: true}).Error; err != nil {
						return err
					}
					categoryKeys[key] = struct{}{}
				}
			case "Marca":
				if _, exists := brandKeys[key]; !exists {
					if err := tx.Create(&domain.Marca{NegocioID: businessID, Nombre: name, Activo: true}).Error; err != nil {
						return err
					}
					brandKeys[key] = struct{}{}
				}
			case "Proveedor":
				if _, exists := providerKeys[key]; !exists {
					if err := tx.Create(&domain.Proveedor{NegocioID: businessID, Nombre: name, Activo: true}).Error; err != nil {
						return err
					}
					providerKeys[key] = struct{}{}
				}
			case "Unidad de medida":
				if _, exists := unitKeys[key]; !exists {
					symbol := importedUnitSymbol(name, unitKeys)
					if err := tx.Create(&domain.UnidadMedida{NegocioID: businessID, Nombre: name, Simbolo: symbol, Tipo: "CANTIDAD", FactorABase: 1, PermiteFraccion: true, Decimales: 2, Activo: true}).Error; err != nil {
						return err
					}
					unitKeys[key] = struct{}{}
					unitKeys[catalogImportKey(symbol)] = struct{}{}
				}
			default:
				return fmt.Errorf("tipo de catálogo de importación no reconocido: %s", proposal.Tipo)
			}
		}
		return nil
	})
}

func importedUnitSymbol(name string, used map[string]struct{}) string {
	sum := sha256.Sum256([]byte(catalogImportKey(name)))
	hexValue := strings.ToUpper(hex.EncodeToString(sum[:]))
	for suffix := 0; ; suffix++ {
		symbol := "IMP-" + hexValue[:12]
		if suffix > 0 {
			symbol = fmt.Sprintf("IMP-%s%d", hexValue[:10], suffix)
		}
		if _, exists := used[catalogImportKey(symbol)]; !exists {
			return symbol
		}
	}
}

func (r *ProductRepository) CreateImportedProducts(businessID uuid.UUID, rows []domain.ValidatedProductImportRow) (domain.CatalogImportResult, error) {
	result := domain.CatalogImportResult{Errores: make([]domain.CatalogImportIssue, 0), Advertencias: make([]domain.CatalogImportIssue, 0)}
	accountedVariantBases := make(map[string]struct{})
	for _, row := range rows {
		baseCreated := false
		var err error
		if len(row.Input.Variantes) > 0 {
			baseCreated, err = r.createImportedVariant(businessID, row.Input)
		} else {
			err = r.createImportedProduct(businessID, row.Input)
		}
		if err != nil {
			if strings.Contains(err.Error(), "producto ya existe") || strings.Contains(err.Error(), "código de barras ya está asignado") {
				result.Omitidas++
			} else {
				result.Invalidas++
			}
			result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Campo: importErrorField(err.Error()), Motivo: err.Error()})
			continue
		}
		result.Creadas++
		if len(row.Input.Variantes) > 0 {
			result.VariantesCreadas++
			baseKey := productBaseKey(row.Input.Nombre, row.Input.Presentacion, row.Input.MarcaID, row.Input.CategoriaID)
			if _, counted := accountedVariantBases[baseKey]; !counted {
				if baseCreated {
					result.ProductosBaseCreados++
				} else {
					result.ProductosBaseReutilizados++
				}
				accountedVariantBases[baseKey] = struct{}{}
			}
		} else {
			result.ProductosSimplesCreados++
		}
		result.SKUsGenerados++
	}
	return result, nil
}

func importErrorField(message string) string {
	switch {
	case strings.Contains(message, "producto ya existe"):
		return "Producto"
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
	if len(input.Variantes) > 0 {
		_, err := r.createImportedVariant(businessID, input)
		return err
	}
	input.Nombre = strings.TrimSpace(input.Nombre)
	if input.Nombre == "" || input.PrecioVenta < 0 || input.StockInicial < 0 {
		return errors.New("los datos del producto no son válidos")
	}
	barcode := optionalString(input.CodigoBarras)
	imageURL := optionalString(input.ImagenURL)
	if barcode != "" && !importedProductBarcodePattern.MatchString(barcode) {
		return errors.New("el código de barras debe contener entre 1 y 14 dígitos")
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
		if err := validateProductProvider(tx, businessID, input.ProveedorID); err != nil {
			return err
		}
		if input.UnidadMedidaID != nil {
			var unit domain.UnidadMedida
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", *input.UnidadMedidaID, businessID).First(&unit).Error; err != nil {
				return errors.New("la unidad de medida no existe o está inactiva")
			}
		}
		generatedSKU, err := reserveGeneratedProductSKU(tx, businessID)
		if err != nil {
			return err
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
				return errors.New("el producto ya existe en este negocio")
			}
		}
		var duplicate int64
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
			if err := tx.Create(&domain.ProductoCodigo{NegocioID: businessID, ProductoID: &product.ID, Tipo: "GTIN", Codigo: barcode, EsPrincipal: true}).Error; err != nil {
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
		commercial := domain.ProductoNegocio{NegocioID: businessID, ProductoID: &product.ID, SKUInterno: &generatedSKU, PrecioVenta: input.PrecioVenta, PrecioIncluyeImpuestos: true, Activo: true}
		if err := tx.Create(&commercial).Error; err != nil {
			return err
		}
		if err := replaceProductProvider(tx, commercial.ID, input.ProveedorID); err != nil {
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

func (r *ProductRepository) createImportedVariant(businessID uuid.UUID, input domain.CreateImportedProductInput) (bool, error) {
	input.Nombre = strings.TrimSpace(input.Nombre)
	if input.Nombre == "" || input.PrecioVenta < 0 || input.StockInicial < 0 {
		return false, errors.New("los datos de la variante no son válidos")
	}
	variantKey, err := productVariantKey(input.Variantes)
	if err != nil {
		return false, err
	}
	if imageURL := optionalString(input.ImagenURL); imageURL != "" {
		parsed, parseErr := url.Parse(imageURL)
		if parseErr != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return false, errors.New("la URL de imagen no es válida")
		}
	}
	baseCreated := false
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var branch domain.Sucursal
		if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", input.SucursalID, businessID).First(&branch).Error; err != nil {
			return errors.New("la sucursal no pertenece al negocio o está inactiva")
		}
		if input.CategoriaID != nil {
			var category domain.Categoria
			if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", input.CategoriaID, businessID).First(&category).Error; err != nil {
				return errors.New("la categoría no existe o está inactiva")
			}
		}
		if err := validateProductProvider(tx, businessID, input.ProveedorID); err != nil {
			return err
		}
		family, product, created, err := r.resolveVariantBase(tx, businessID, variantBaseInput{
			Nombre: input.Nombre, Descripcion: input.Descripcion, MarcaID: input.MarcaID, CategoriaID: input.CategoriaID,
			UnidadMedidaID: input.UnidadMedidaID, Contenido: input.Contenido, UnidadContenido: input.UnidadContenido,
			Presentacion: input.Presentacion, ImagenURL: input.ImagenURL,
		})
		if err != nil {
			return err
		}
		baseCreated = created
		var duplicate int64
		if err := tx.Model(&domain.ProductoVariante{}).Where("familia_producto_id = ? AND clave = ?", family.ID, variantKey).Count(&duplicate).Error; err != nil {
			return err
		}
		if duplicate > 0 {
			return errors.New("la combinación de variantes ya existe en este producto")
		}
		sku, err := reserveGeneratedProductSKU(tx, businessID)
		if err != nil {
			return err
		}
		if err := tx.Model(&domain.ProductoNegocio{}).Where("negocio_id = ? AND sku_interno = ?", businessID, sku).Count(&duplicate).Error; err != nil {
			return err
		}
		if duplicate > 0 {
			return errors.New("el SKU ya existe en este negocio")
		}
		barcode := optionalString(input.CodigoBarras)
		if barcode != "" {
			if !importedProductBarcodePattern.MatchString(barcode) {
				return errors.New("el código de barras debe contener entre 1 y 14 dígitos")
			}
			if err := tx.Model(&domain.ProductoCodigo{}).Where("negocio_id = ? AND codigo = ?", businessID, barcode).Count(&duplicate).Error; err != nil {
				return err
			}
			if duplicate > 0 {
				return errors.New("el código de barras ya está asignado a otro producto")
			}
		}
		productVariant := domain.ProductoVariante{ProductoID: product.ID, FamiliaProductoID: family.ID, Clave: variantKey, Activo: true}
		if err := tx.Create(&productVariant).Error; err != nil {
			return err
		}
		for _, attribute := range input.Variantes {
			if err := tx.Create(&domain.ProductoVarianteAtributo{ProductoVarianteID: productVariant.ID, Nombre: strings.TrimSpace(attribute.Nombre), Valor: strings.TrimSpace(attribute.Valor)}).Error; err != nil {
				return err
			}
		}
		if barcode != "" {
			if err := tx.Create(&domain.ProductoCodigo{NegocioID: businessID, ProductoVarianteID: &productVariant.ID, Tipo: "GTIN", Codigo: barcode, EsPrincipal: true}).Error; err != nil {
				return err
			}
		}
		commercial := domain.ProductoNegocio{NegocioID: businessID, ProductoVarianteID: &productVariant.ID, SKUInterno: &sku, PrecioVenta: input.PrecioVenta, PrecioIncluyeImpuestos: true, Activo: true}
		if err := tx.Create(&commercial).Error; err != nil {
			return err
		}
		if err := replaceProductProvider(tx, commercial.ID, input.ProveedorID); err != nil {
			return err
		}
		inventory := domain.InventarioSucursal{SucursalID: input.SucursalID, ProductoNegocioID: commercial.ID, StockActual: input.StockInicial, StockMinimo: 0}
		if err := tx.Create(&inventory).Error; err != nil {
			return err
		}
		if input.StockInicial > 0 {
			return tx.Create(&domain.MovimientoInventario{SucursalID: input.SucursalID, ProductoNegocioID: commercial.ID, Tipo: "AJUSTE_ENTRADA", Cantidad: input.StockInicial, StockAnterior: 0, StockNuevo: input.StockInicial}).Error
		}
		return nil
	})
	return baseCreated, err
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

func validateProductProvider(tx *gorm.DB, businessID uuid.UUID, providerID *uuid.UUID) error {
	if providerID == nil {
		return nil
	}
	var provider domain.Proveedor
	if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", *providerID, businessID).First(&provider).Error; err != nil {
		return errors.New("el proveedor no existe o está inactivo")
	}
	return nil
}

func replaceProductProvider(tx *gorm.DB, commercialID uuid.UUID, providerID *uuid.UUID) error {
	if err := tx.Where("producto_negocio_id = ?", commercialID).Delete(&domain.ProductoProveedor{}).Error; err != nil {
		return err
	}
	if providerID == nil {
		return nil
	}
	return tx.Create(&domain.ProductoProveedor{ProductoNegocioID: commercialID, ProveedorID: *providerID, EsPrincipal: true}).Error
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) CreateProductImportJob(businessID, branchID uuid.UUID) (domain.ProductImportJob, error) {
	job := domain.ImportacionProducto{ID: uuid.New(), NegocioID: businessID, SucursalID: branchID, Estado: "pendiente", Etapa: "Preparando importación"}
	if err := r.db.Create(&job).Error; err != nil {
		return domain.ProductImportJob{}, err
	}
	return productImportJobView(job), nil
}

func (r *ProductRepository) GetProductImportJob(businessID, jobID uuid.UUID) (domain.ProductImportJob, error) {
	var job domain.ImportacionProducto
	if err := r.db.Where("id = ? AND negocio_id = ?", jobID, businessID).First(&job).Error; err != nil {
		return domain.ProductImportJob{}, err
	}
	return productImportJobView(job), nil
}

func (r *ProductRepository) UpdateProductImportJob(jobID uuid.UUID, state, stage string, progress int, result *domain.CatalogImportResult, message *string) error {
	updates := map[string]interface{}{"estado": state, "etapa": stage, "porcentaje": progress, "mensaje_error": message}
	if result != nil {
		encoded, err := json.Marshal(result)
		if err != nil {
			return err
		}
		updates["resultado_json"] = encoded
	}
	return r.db.Model(&domain.ImportacionProducto{}).Where("id = ?", jobID).Updates(updates).Error
}

func productImportJobView(job domain.ImportacionProducto) domain.ProductImportJob {
	view := domain.ProductImportJob{ID: job.ID, Estado: job.Estado, Etapa: job.Etapa, Porcentaje: job.Porcentaje, MensajeError: job.MensajeError}
	if len(job.ResultadoJSON) > 0 {
		var result domain.CatalogImportResult
		if json.Unmarshal(job.ResultadoJSON, &result) == nil {
			view.Resultado = &result
		}
	}
	return view
}

func (r *ProductRepository) ListByBusiness(businessID, branchID uuid.UUID) ([]domain.ProductRow, error) {
	products := make([]domain.ProductRow, 0)
	query := r.db.Table("productos AS p").
		Select(`
			p.id,
			p.nombre,
			(SELECT pi.url FROM producto_imagenes pi WHERE pi.producto_id = p.id ORDER BY pi.es_principal DESC, pi.orden ASC LIMIT 1) AS imagen_url,
			pn.sku_interno AS sku,
			pn.precio_venta AS precio,
			COALESCE(existencias.stock, 0) AS stock,
			(SELECT c.nombre FROM producto_categorias pc JOIN categorias c ON c.id = pc.categoria_id WHERE pc.producto_id = p.id ORDER BY pc.es_principal DESC, c.nombre ASC LIMIT 1) AS categoria,
			(SELECT pc.categoria_id FROM producto_categorias pc WHERE pc.producto_id = p.id ORDER BY pc.es_principal DESC LIMIT 1) AS categoria_id,
			m.nombre AS marca,
			p.marca_id,
			(SELECT pr.nombre FROM producto_proveedor pp JOIN proveedores pr ON pr.id = pp.proveedor_id WHERE pp.producto_negocio_id = pn.id ORDER BY pp.es_principal DESC, pp.creado_en ASC LIMIT 1) AS proveedor,
			(SELECT pp.proveedor_id FROM producto_proveedor pp WHERE pp.producto_negocio_id = pn.id ORDER BY pp.es_principal DESC, pp.creado_en ASC LIMIT 1) AS proveedor_id,
			p.descripcion,
			p.presentacion,
			p.contenido,
			p.unidad_contenido,
			u.nombre AS unidad_medida,
			p.unidad_medida_id,
			(SELECT pcodigo.codigo FROM producto_codigos pcodigo WHERE pcodigo.producto_id = p.id ORDER BY pcodigo.es_principal DESC LIMIT 1) AS codigo_barras,
			CASE
				WHEN COALESCE(existencias.stock, 0) <= 0 THEN 'Agotado'
				WHEN COALESCE(existencias.stock, 0) <= COALESCE(existencias.minimo, 0) THEN 'Bajo stock'
				ELSE 'En stock'
			END AS estado`).
		Joins(`JOIN LATERAL (
			SELECT pn.*
			FROM producto_negocio pn
			WHERE pn.negocio_id = ? AND pn.activo = TRUE
				AND (pn.producto_id = p.id OR pn.producto_variante_id IN (SELECT pv.id FROM producto_variantes pv WHERE pv.producto_id = p.id AND pv.activo = TRUE))
			ORDER BY CASE WHEN pn.producto_id IS NOT NULL THEN 0 ELSE 1 END, pn.creado_en ASC
			LIMIT 1
		 ) pn ON TRUE`, businessID).
		Joins(`LEFT JOIN LATERAL (
			SELECT SUM(i.stock_actual) AS stock, SUM(i.stock_minimo) AS minimo
			FROM inventario_sucursal i
			JOIN producto_negocio oferta ON oferta.id = i.producto_negocio_id AND oferta.negocio_id = ? AND oferta.activo = TRUE
			LEFT JOIN producto_variantes variante ON variante.id = oferta.producto_variante_id AND variante.activo = TRUE
			WHERE i.sucursal_id = ? AND (oferta.producto_id = p.id OR variante.producto_id = p.id)
		) existencias ON TRUE`, businessID, branchID).
		Joins("LEFT JOIN marcas m ON m.id = p.marca_id").
		Joins("LEFT JOIN unidades_medida u ON u.id = p.unidad_medida_id").
		Where("p.negocio_id = ? AND p.activo = TRUE", businessID).
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
		Select("COALESCE(pn.producto_id, pv.producto_id) AS producto_id, s.id AS sucursal_id, s.nombre AS sucursal, SUM(i.stock_actual) AS stock").
		Joins("JOIN producto_negocio pn ON pn.id = i.producto_negocio_id AND pn.negocio_id = ? AND pn.activo = TRUE", businessID).
		Joins("LEFT JOIN producto_variantes pv ON pv.id = pn.producto_variante_id").
		Joins("JOIN sucursales s ON s.id = i.sucursal_id").Where("i.sucursal_id = ? AND (pn.producto_variante_id IS NULL OR pv.activo = TRUE)", branchID).
		Group("COALESCE(pn.producto_id, pv.producto_id), s.id, s.nombre").
		Order("s.nombre ASC").Scan(&branches).Error; err != nil {
		return nil, err
	}
	for _, branch := range branches {
		key := branch.ProductoID.String()
		stockByProduct[key] = append(stockByProduct[key], domain.ProductBranchStock{SucursalID: branch.SucursalID.String(), Sucursal: branch.Sucursal, Stock: branch.Stock})
	}
	variantsByProduct := make(map[string][]domain.ProductVariantRow)
	var variants []struct {
		ProductoID   uuid.UUID `gorm:"column:producto_id"`
		ID           uuid.UUID `gorm:"column:id"`
		SKU          *string   `gorm:"column:sku"`
		Precio       float64   `gorm:"column:precio"`
		Stock        float64   `gorm:"column:stock"`
		CodigoBarras *string   `gorm:"column:codigo_barras"`
	}
	if err := r.db.Table("producto_variantes AS pv").
		Select(`pv.producto_id, pv.id, pn.sku_interno AS sku, pn.precio_venta AS precio,
			COALESCE((SELECT SUM(i.stock_actual) FROM inventario_sucursal i WHERE i.producto_negocio_id = pn.id AND i.sucursal_id = ?), 0) AS stock,
			(SELECT pc.codigo FROM producto_codigos pc WHERE pc.producto_variante_id = pv.id ORDER BY pc.es_principal DESC LIMIT 1) AS codigo_barras`, branchID).
		Joins("JOIN producto_negocio pn ON pn.producto_variante_id = pv.id AND pn.negocio_id = ? AND pn.activo = TRUE", businessID).
		Where("pv.activo = TRUE").Scan(&variants).Error; err != nil {
		return nil, err
	}
	attributesByVariant := make(map[string][]domain.ProductVariantAttributeInput, len(variants))
	var attributes []struct {
		ProductoVarianteID uuid.UUID `gorm:"column:producto_variante_id"`
		Nombre             string    `gorm:"column:nombre"`
		Valor              string    `gorm:"column:valor"`
	}
	if len(variants) > 0 {
		variantIDs := make([]uuid.UUID, 0, len(variants))
		for _, variant := range variants {
			variantIDs = append(variantIDs, variant.ID)
		}
		if err := r.db.Table("producto_variante_atributos").Where("producto_variante_id IN ?", variantIDs).Order("nombre ASC").Scan(&attributes).Error; err != nil {
			return nil, err
		}
	}
	for _, attribute := range attributes {
		key := attribute.ProductoVarianteID.String()
		attributesByVariant[key] = append(attributesByVariant[key], domain.ProductVariantAttributeInput{Nombre: attribute.Nombre, Valor: attribute.Valor})
	}
	variantStockByID := make(map[string][]domain.ProductBranchStock)
	var variantBranches []struct {
		VarianteID uuid.UUID `gorm:"column:variante_id"`
		SucursalID uuid.UUID `gorm:"column:sucursal_id"`
		Sucursal   string    `gorm:"column:sucursal"`
		Stock      float64   `gorm:"column:stock"`
	}
	if err := r.db.Table("inventario_sucursal AS i").
		Select("pn.producto_variante_id AS variante_id, s.id AS sucursal_id, s.nombre AS sucursal, i.stock_actual AS stock").
		Joins("JOIN producto_negocio pn ON pn.id = i.producto_negocio_id AND pn.negocio_id = ? AND pn.activo = TRUE AND pn.producto_variante_id IS NOT NULL", businessID).
		Joins("JOIN sucursales s ON s.id = i.sucursal_id").Where("i.sucursal_id = ?", branchID).Scan(&variantBranches).Error; err != nil {
		return nil, err
	}
	for _, branch := range variantBranches {
		key := branch.VarianteID.String()
		variantStockByID[key] = append(variantStockByID[key], domain.ProductBranchStock{SucursalID: branch.SucursalID.String(), Sucursal: branch.Sucursal, Stock: branch.Stock})
	}
	for _, variant := range variants {
		key := variant.ProductoID.String()
		variantsByProduct[key] = append(variantsByProduct[key], domain.ProductVariantRow{ID: variant.ID, SKU: variant.SKU, Precio: variant.Precio, Stock: variant.Stock, CodigoBarras: variant.CodigoBarras, Atributos: attributesByVariant[variant.ID.String()], Inventario: variantStockByID[variant.ID.String()]})
	}
	for index := range products {
		products[index].Inventario = stockByProduct[products[index].ID.String()]
		if products[index].Inventario == nil {
			products[index].Inventario = make([]domain.ProductBranchStock, 0)
		}
		if variants, exists := variantsByProduct[products[index].ID.String()]; exists {
			products[index].Variantes = variants
		} else {
			products[index].Variantes = make([]domain.ProductVariantRow, 0)
		}
	}
	return products, nil
}

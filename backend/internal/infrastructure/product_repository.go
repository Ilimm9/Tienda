package infrastructure

import (
	"errors"
	"strings"
	"tienda/backend/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

	return r.db.Transaction(func(tx *gorm.DB) error {
		var branch domain.Sucursal
		if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", input.SucursalID, businessID).First(&branch).Error; err != nil {
			return errors.New("la sucursal no pertenece al negocio o está inactiva")
		}
		var category domain.Categoria
		if err := tx.Where("id = ? AND activo = TRUE", input.CategoriaID).First(&category).Error; err != nil {
			return errors.New("la categoría no existe o está inactiva")
		}
		var duplicate int64
		if err := tx.Model(&domain.ProductoNegocio{}).Where("negocio_id = ? AND sku_interno = ?", businessID, input.SKUInterno).Count(&duplicate).Error; err != nil {
			return err
		}
		if duplicate > 0 {
			return errors.New("el SKU ya existe en este negocio")
		}

		product := domain.Producto{
			Nombre: input.Nombre, Descripcion: input.Descripcion, MarcaID: input.MarcaID,
			Contenido: input.Contenido, UnidadContenido: input.UnidadContenido,
			Presentacion: input.Presentacion, Activo: true,
		}
		if err := tx.Create(&product).Error; err != nil {
			return err
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

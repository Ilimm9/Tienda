package infrastructure

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"tienda/backend/internal/domain"
	preciodomain "tienda/backend/internal/domain/precio"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Surtido por sucursal (capability 28): a product belongs to a branch when one of
// its commercial offers has a row in inventario_sucursal for that branch.

// branchOffers selects active business offers with their parent product id.
const branchOffers = `FROM producto_negocio pn
	LEFT JOIN producto_variantes pv ON pv.id = pn.producto_variante_id
	JOIN productos p ON p.id = COALESCE(pn.producto_id, pv.producto_id)
	WHERE pn.negocio_id = ? AND pn.activo = TRUE AND p.activo = TRUE AND (pv.id IS NULL OR pv.activo = TRUE)
		AND NOT EXISTS (SELECT 1 FROM inventario_sucursal i WHERE i.producto_negocio_id = pn.id AND i.sucursal_id = ?)`

func (r *ProductRepository) ListCatalogNotInBranch(businessID, branchID uuid.UUID) ([]domain.BranchCatalogProduct, error) {
	items := make([]domain.BranchCatalogProduct, 0)
	if err := ensureActiveBranch(r.db, businessID, branchID); err != nil {
		return items, err
	}
	err := r.db.Raw(`SELECT p.id, p.nombre, p.presentacion,
			MIN(pn.sku_interno) FILTER (WHERE pn.producto_id IS NOT NULL) AS sku,
			(SELECT pc.codigo FROM producto_codigos pc WHERE pc.producto_id = p.id ORDER BY pc.es_principal DESC LIMIT 1) AS codigo_barras,
			(SELECT c.nombre FROM producto_categorias pcat JOIN categorias c ON c.id = pcat.categoria_id WHERE pcat.producto_id = p.id ORDER BY pcat.es_principal DESC, c.nombre ASC LIMIT 1) AS categoria,
			COUNT(*) FILTER (WHERE pn.producto_variante_id IS NOT NULL) AS variantes_faltantes
		`+branchOffers+`
		GROUP BY p.id, p.nombre, p.presentacion
		ORDER BY p.nombre ASC`, businessID, branchID).Scan(&items).Error
	return items, err
}

// AddToBranch adds the chosen products (every active variant of a family) to the
// branch with stock 0 and no price. Shared catalog rows are never modified.
func (r *ProductRepository) AddToBranch(businessID uuid.UUID, input domain.AddToBranchInput) (int, error) {
	if !input.Todos && len(input.ProductoIDs) == 0 {
		return 0, errors.New("elige al menos un producto")
	}
	added := 0
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := ensureActiveBranch(tx, businessID, input.SucursalID); err != nil {
			return err
		}
		query := `SELECT pn.id ` + branchOffers
		args := []any{businessID, input.SucursalID}
		if !input.Todos {
			query += ` AND p.id IN ?`
			args = append(args, input.ProductoIDs)
		}
		var offerIDs []uuid.UUID
		if err := tx.Raw(query, args...).Scan(&offerIDs).Error; err != nil {
			return err
		}
		for _, offerID := range offerIDs {
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&domain.InventarioSucursal{SucursalID: input.SucursalID, ProductoNegocioID: offerID})
			if result.Error != nil {
				return result.Error
			}
			added += int(result.RowsAffected)
		}
		return nil
	})
	return added, err
}

// AddProductToBranch adds one existing simple product to the branch with the
// amounts captured in the product form.
func (r *ProductRepository) AddProductToBranch(businessID, productID uuid.UUID, input domain.AddProductToBranchInput) error {
	for _, value := range []float64{input.PrecioVenta, input.Costo, input.StockInicial} {
		if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
			return errors.New("el precio, costo y stock deben ser válidos")
		}
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := ensureActiveBranch(tx, businessID, input.SucursalID); err != nil {
			return err
		}
		var offer domain.ProductoNegocio
		if err := tx.Where("negocio_id = ? AND producto_id = ? AND activo = TRUE", businessID, productID).First(&offer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("el producto no existe o tiene variantes; agrégalo con «Agregar del catálogo»")
			}
			return err
		}
		return addOfferToBranch(tx, businessID, input.SucursalID, offer.ID, input.UsuarioID, input.Costo, input.PrecioVenta, input.StockInicial)
	})
}

// addOfferToBranch creates the branch inventory (with an entry movement when
// stock > 0) and, when amounts are given, the branch price with its history.
func addOfferToBranch(tx *gorm.DB, businessID, branchID, offerID, userID uuid.UUID, cost, price, stock float64) error {
	inventory := domain.InventarioSucursal{SucursalID: branchID, ProductoNegocioID: offerID, StockActual: stock}
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&inventory)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrAlreadyInBranch
	}
	if stock > 0 {
		movement := domain.MovimientoInventario{SucursalID: branchID, ProductoNegocioID: offerID, Tipo: "AJUSTE_ENTRADA", Cantidad: stock, StockAnterior: 0, StockNuevo: stock}
		if userID != uuid.Nil {
			movement.UsuarioID = &userID
		}
		if err := tx.Create(&movement).Error; err != nil {
			return err
		}
	}
	if price <= 0 && cost <= 0 {
		return nil
	}
	var existing int64
	if err := tx.Table("precios_sucursal").Where("sucursal_id = ? AND producto_negocio_id = ?", branchID, offerID).Count(&existing).Error; err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}
	return createInitialBranchPrice(tx, businessID, branchID, offerID, userID, cost, price)
}

// existingSimpleProduct finds a simple product of the business with the same
// barcode, or the same name and presentation (the import identity rule).
func existingSimpleProduct(tx *gorm.DB, businessID uuid.UUID, name, presentation, barcode string) (*domain.Producto, error) {
	if barcode != "" {
		var product domain.Producto
		err := tx.Table("productos p").Select("p.*").
			Joins("JOIN producto_codigos pc ON pc.producto_id = p.id AND pc.negocio_id = ? AND pc.codigo = ?", businessID, barcode).
			Where("p.negocio_id = ? AND p.activo = TRUE", businessID).Take(&product).Error
		if err == nil {
			return &product, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	// La identidad ignora mayúsculas y acentos (catalogImportKey), así que se compara en Go.
	var candidates []domain.Producto
	if err := tx.Where("negocio_id = ? AND activo = TRUE AND familia_producto_id IS NULL", businessID).Find(&candidates).Error; err != nil {
		return nil, err
	}
	key := productImportNamePresentationKey(name, presentation)
	for index := range candidates {
		if productImportNamePresentationKey(candidates[index].Nombre, trimmedPresentation(candidates[index].Presentacion)) == key {
			return &candidates[index], nil
		}
	}
	return nil, nil
}

func ensureActiveBranch(tx *gorm.DB, businessID, branchID uuid.UUID) error {
	var branch domain.Sucursal
	if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", branchID, businessID).First(&branch).Error; err != nil {
		return errors.New("la sucursal no pertenece al negocio o está inactiva")
	}
	return nil
}

func trimmedPresentation(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

// importBranchState is the active-branch stock and price of an existing offer.
type importBranchState struct {
	stock float64
	price *float64
	cost  *float64
}

// importExistingAction classifies an import row whose product already exists in
// the business: added to the branch, or updated when stock, price or cost change.
// Changes are reported as warnings so the preview shows them.
func importExistingAction(state map[uuid.UUID]importBranchState, offerID uuid.UUID, row domain.ProductImportRow, warn func(string, string)) (string, bool) {
	current, inBranch := state[offerID]
	if !inBranch {
		return domain.ImportAccionAgregarSucursal, true
	}
	changed := false
	if current.stock != row.StockInicial {
		warn("Stock", "se reemplazará: de "+formatImportNumber(current.stock)+" a "+formatImportNumber(row.StockInicial))
		changed = true
	}
	price := math.Round(row.PrecioVenta*100) / 100
	if current.price == nil {
		if price > 0 || (row.CostoCapturado && row.CostoProveedor > 0) {
			warn("Precio de venta", "se asignará "+formatImportNumber(price))
			changed = true
		}
		return domain.ImportAccionActualizar, changed
	}
	if *current.price != price {
		warn("Precio de venta", "de "+formatImportNumber(*current.price)+" a "+formatImportNumber(price))
		changed = true
	}
	if row.CostoCapturado && current.cost != nil && *current.cost != row.CostoProveedor {
		warn("Costo", "de "+formatImportNumber(*current.cost)+" a "+formatImportNumber(row.CostoProveedor))
		changed = true
	}
	return domain.ImportAccionActualizar, changed
}

func formatImportNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

const importReason = "Carga masiva"

// applyImportToBranch adds an existing catalog offer to the branch, or updates
// its branch stock, price and cost. It never touches shared catalog rows.
func (r *ProductRepository) applyImportToBranch(businessID uuid.UUID, row domain.ValidatedProductImportRow) (bool, error) {
	input := row.Input
	for _, value := range []float64{input.PrecioVenta, input.CostoProveedor, input.StockInicial} {
		if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
			return false, errors.New("el precio, costo y stock deben ser válidos")
		}
	}
	price := math.Round(input.PrecioVenta*100) / 100
	changed := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := ensureActiveBranch(tx, businessID, input.SucursalID); err != nil {
			return err
		}
		var offer domain.ProductoNegocio
		if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", row.ProductoNegocioID, businessID).First(&offer).Error; err != nil {
			return errors.New("el producto ya no existe o está inactivo")
		}
		if row.Accion == domain.ImportAccionAgregarSucursal {
			changed = true
			return addOfferToBranch(tx, businessID, input.SucursalID, offer.ID, input.UsuarioID, input.CostoProveedor, price, input.StockInicial)
		}
		stockChanged, err := replaceImportedStock(tx, input.SucursalID, offer.ID, input.UsuarioID, input.StockInicial)
		if err != nil {
			return err
		}
		priceChanged, err := updateImportedPrice(tx, businessID, input.SucursalID, offer.ID, input.UsuarioID, price, input.CostoProveedor, input.CostoCapturado)
		if err != nil {
			return err
		}
		changed = stockChanged || priceChanged
		return nil
	})
	return changed, err
}

// replaceImportedStock sets the branch stock to the imported value and records
// the difference as a manual adjustment.
func replaceImportedStock(tx *gorm.DB, branchID, offerID, userID uuid.UUID, stock float64) (bool, error) {
	var inventory domain.InventarioSucursal
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sucursal_id = ? AND producto_negocio_id = ?", branchID, offerID).First(&inventory).Error; err != nil {
		return false, errors.New("el producto ya no está en esta sucursal")
	}
	previous := inventory.StockActual
	if previous == stock {
		return false, nil
	}
	if err := tx.Model(&inventory).Update("stock_actual", stock).Error; err != nil {
		return false, err
	}
	reason := importReason
	movement := domain.MovimientoInventario{SucursalID: branchID, ProductoNegocioID: offerID, Tipo: "AJUSTE_MANUAL", Cantidad: stock - previous, StockAnterior: previous, StockNuevo: stock, Motivo: &reason}
	if userID != uuid.Nil {
		movement.UsuarioID = &userID
	}
	return true, tx.Create(&movement).Error
}

// updateImportedPrice updates the branch price and, when captured, the cost,
// recording a CARGA_MASIVA history event with previous and new values.
func updateImportedPrice(tx *gorm.DB, businessID, branchID, offerID, userID uuid.UUID, price, cost float64, costCaptured bool) (bool, error) {
	var current preciodomain.PrecioSucursal
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sucursal_id = ? AND producto_negocio_id = ?", branchID, offerID).First(&current).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if !costCaptured {
			cost = 0
		}
		if price <= 0 && cost <= 0 {
			return false, nil
		}
		return true, createInitialBranchPrice(tx, businessID, branchID, offerID, userID, cost, price)
	}
	if err != nil {
		return false, err
	}
	newCost := current.CostoVigente
	if costCaptured {
		newCost = cost
	}
	if current.PrecioVenta == price && current.CostoVigente == newCost {
		return false, nil
	}
	values := map[string]any{"precio_venta": price, "costo_vigente": newCost}
	if newCost > 0 {
		values["margen"] = (price - newCost) / newCost
	}
	previousPrice, previousCost := current.PrecioVenta, current.CostoVigente
	if err := tx.Model(&current).Updates(values).Error; err != nil {
		return false, err
	}
	reason := importReason
	event := preciodomain.HistorialCostoPrecioSucursal{NegocioID: businessID, SucursalID: branchID, ProductoNegocioID: offerID, Origen: preciodomain.OrigenCargaMasiva,
		CostoAnterior: &previousCost, CostoNuevo: &newCost, PrecioAnterior: &previousPrice, PrecioNuevo: &price, Motivo: &reason}
	if userID != uuid.Nil {
		event.UsuarioID = &userID
	}
	return true, tx.Create(&event).Error
}

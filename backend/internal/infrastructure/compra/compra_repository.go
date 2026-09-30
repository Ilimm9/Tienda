package compra

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	compraapp "tienda/backend/internal/application/compra"
	"tienda/backend/internal/domain"
	compradomain "tienda/backend/internal/domain/compra"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) List(businessID uuid.UUID) ([]compradomain.CompraProveedor, error) {
	items := make([]compradomain.CompraProveedor, 0)
	err := r.db.Where("negocio_id = ?", businessID).Order("fecha_recepcion DESC, creado_en DESC").Find(&items).Error
	return items, err
}
func (r *Repository) Get(businessID, id uuid.UUID) (compradomain.CompraProveedor, error) {
	var item compradomain.CompraProveedor
	err := r.db.Preload("Detalles").Where("id = ? AND negocio_id = ?", id, businessID).First(&item).Error
	return item, err
}
func (r *Repository) ListProductos(businessID, providerID uuid.UUID) ([]compraapp.ProductoComprable, error) {
	items := make([]compraapp.ProductoComprable, 0)
	err := r.db.Table("producto_negocio pn").Select(`pn.id, COALESCE(p.nombre, vp.nombre) || COALESCE(' · ' || pv.clave, '') AS nombre, pn.sku_interno AS sku,
 (SELECT pc.codigo FROM producto_codigos pc WHERE pc.producto_id = p.id OR pc.producto_variante_id = pv.id ORDER BY pc.es_principal DESC LIMIT 1) AS codigo_barras`).
		Joins("LEFT JOIN productos p ON p.id = pn.producto_id").Joins("LEFT JOIN producto_variantes pv ON pv.id = pn.producto_variante_id").
		Joins("LEFT JOIN productos vp ON vp.id = pv.producto_id").
		Where("pn.negocio_id = ? AND pn.activo = TRUE AND EXISTS (SELECT 1 FROM producto_proveedor pp WHERE pp.producto_negocio_id = pn.id AND pp.proveedor_id = ?)", businessID, providerID).
		Order("nombre ASC").Scan(&items).Error
	return items, err
}

func (r *Repository) Create(businessID, userID uuid.UUID, input compraapp.CrearCompraInput) (compradomain.CompraProveedor, error) {
	providerID, err := uuid.Parse(input.ProveedorID)
	if err != nil {
		return compradomain.CompraProveedor{}, errors.New("el proveedor no es válido")
	}
	branchID, err := uuid.Parse(input.SucursalID)
	if err != nil {
		return compradomain.CompraProveedor{}, errors.New("la sucursal no es válida")
	}
	if input.TipoDocumento != compradomain.DocumentoTicket && input.TipoDocumento != compradomain.DocumentoFactura {
		return compradomain.CompraProveedor{}, errors.New("el tipo de documento no es válido")
	}
	if strings.TrimSpace(input.FolioDocumento) == "" || len(input.Detalles) == 0 {
		return compradomain.CompraProveedor{}, errors.New("el folio y al menos un producto son obligatorios")
	}
	if !validPayment(input.FormaPago) {
		return compradomain.CompraProveedor{}, errors.New("la forma de pago no es válida")
	}
	dateDocument, err := compraapp.ParseDate(input.FechaDocumento)
	if err != nil {
		return compradomain.CompraProveedor{}, errors.New("la fecha del documento no es válida")
	}
	dateReception, err := compraapp.ParseDate(input.FechaRecepcion)
	if err != nil {
		return compradomain.CompraProveedor{}, errors.New("la fecha de recepción no es válida")
	}
	var created compradomain.CompraProveedor
	err = r.db.Transaction(func(tx *gorm.DB) error {
		var provider domain.Proveedor
		if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", providerID, businessID).First(&provider).Error; err != nil {
			return errors.New("el proveedor no existe o está inactivo")
		}
		var branch domain.Sucursal
		if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", branchID, businessID).First(&branch).Error; err != nil {
			return errors.New("la sucursal no existe o está inactiva")
		}
		var count int64
		tx.Model(&compradomain.CompraProveedor{}).Where("negocio_id = ? AND proveedor_id = ? AND tipo_documento = ? AND folio_documento = ?", businessID, providerID, input.TipoDocumento, strings.TrimSpace(input.FolioDocumento)).Count(&count)
		if count > 0 {
			return errors.New("el documento ya está registrado para este proveedor")
		}
		details := make([]compradomain.CompraProveedorDetalle, 0, len(input.Detalles))
		subtotal := 0.0
		for _, row := range input.Detalles {
			detail, err := r.buildDetail(tx, businessID, providerID, row)
			if err != nil {
				return err
			}
			subtotal += detail.Subtotal
			details = append(details, detail)
		}
		subtotal = money(subtotal)
		if input.IEPS < 0 || math.IsNaN(input.IEPS) || math.IsInf(input.IEPS, 0) {
			return errors.New("el importe de IEPS no es válido")
		}
		ieps := 0.0
		if input.IncluyeIEPS {
			ieps = money(input.IEPS)
		}
		iva := 0.0
		if input.IncluyeIVA {
			iva = money((subtotal + ieps) * 0.16)
		}
		taxes := money(ieps + iva)
		total := money(subtotal + taxes)
		created = compradomain.CompraProveedor{NegocioID: businessID, ProveedorID: providerID, SucursalID: branchID, TipoDocumento: input.TipoDocumento, FolioDocumento: strings.TrimSpace(input.FolioDocumento), NumeroCuenta: trimPtr(input.NumeroCuenta), FechaDocumento: dateDocument, FechaRecepcion: dateReception, FormaPago: input.FormaPago, ImportePagado: total, Subtotal: subtotal, IEPS: ieps, IVA: iva, Impuestos: taxes, Total: total, Observaciones: trimPtr(input.Observaciones), CapturadoPorUsuarioID: userID, RecibidoPorUsuarioID: userID, Detalles: details}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		for _, detail := range details {
			if err := r.receiveDetail(tx, created, detail); err != nil {
				return err
			}
		}
		return nil
	})
	return created, err
}

func (r *Repository) buildDetail(tx *gorm.DB, businessID, providerID uuid.UUID, row compraapp.CrearDetalleInput) (compradomain.CompraProveedorDetalle, error) {
	if row.CantidadEmpaques <= 0 || row.CostoEmpaque <= 0 || row.Descuento < 0 || row.PiezasDanadas < 0 || row.PiezasFaltantes < 0 {
		return compradomain.CompraProveedorDetalle{}, errors.New("las cantidades y costos del detalle no son válidos")
	}
	unitID, err := uuid.Parse(row.UnidadMedidaID)
	if err != nil {
		return compradomain.CompraProveedorDetalle{}, errors.New("la unidad de empaque no es válida")
	}
	var unit domain.UnidadMedida
	if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", unitID, businessID).First(&unit).Error; err != nil {
		return compradomain.CompraProveedorDetalle{}, errors.New("la unidad de empaque no existe o está inactiva")
	}
	if unit.Tipo != "EMPAQUE" || unit.FactorABase < 1 || math.Trunc(unit.FactorABase) != unit.FactorABase {
		return compradomain.CompraProveedorDetalle{}, errors.New("la unidad debe ser un empaque con una conversión entera a piezas")
	}
	productID, err := uuid.Parse(row.ProductoNegocioID)
	if row.ProductoNuevo {
		if strings.TrimSpace(row.NombreProductoNuevo) == "" || strings.TrimSpace(row.CodigoBarrasNuevo) == "" {
			return compradomain.CompraProveedorDetalle{}, errors.New("el producto nuevo requiere nombre y código de barras")
		}
		productID, err = r.createNewProduct(tx, businessID, providerID, row.NombreProductoNuevo, row.CodigoBarrasNuevo)
		if err != nil {
			return compradomain.CompraProveedorDetalle{}, err
		}
	} else if err != nil {
		return compradomain.CompraProveedorDetalle{}, errors.New("el producto no es válido")
	}
	var product domain.ProductoNegocio
	if err := tx.Where("id = ? AND negocio_id = ? AND activo = TRUE", productID, businessID).First(&product).Error; err != nil {
		return compradomain.CompraProveedorDetalle{}, errors.New("el producto no existe o está inactivo")
	}
	var linked int64
	tx.Model(&domain.ProductoProveedor{}).Where("producto_negocio_id = ? AND proveedor_id = ?", productID, providerID).Count(&linked)
	if linked == 0 {
		return compradomain.CompraProveedorDetalle{}, errors.New("el producto no está asociado al proveedor")
	}
	piecesPerPackage := int(unit.FactorABase)
	totalPieces := row.CantidadEmpaques * piecesPerPackage
	good := totalPieces - row.PiezasDanadas - row.PiezasFaltantes
	if good < 0 {
		return compradomain.CompraProveedorDetalle{}, errors.New("las piezas dañadas y faltantes superan el total")
	}
	gross := float64(row.CantidadEmpaques) * row.CostoEmpaque
	if row.Descuento > gross {
		return compradomain.CompraProveedorDetalle{}, errors.New("el descuento supera el costo del detalle")
	}
	var expiry *time.Time
	if row.FechaCaducidad != nil && strings.TrimSpace(*row.FechaCaducidad) != "" {
		parsed, err := compraapp.ParseDate(*row.FechaCaducidad)
		if err != nil {
			return compradomain.CompraProveedorDetalle{}, errors.New("la caducidad no es válida")
		}
		expiry = &parsed
	}
	barcode := trimPtr(&row.CodigoBarrasNuevo)
	if barcode == nil {
		var existingBarcode string
		if err := tx.Table("producto_codigos").Select("codigo").Where("negocio_id = ? AND (producto_id = ? OR producto_variante_id = (SELECT producto_variante_id FROM producto_negocio WHERE id = ?))", businessID, product.ProductoID, productID).Order("es_principal DESC").Limit(1).Scan(&existingBarcode).Error; err == nil && existingBarcode != "" {
			barcode = &existingBarcode
		}
	}
	return compradomain.CompraProveedorDetalle{ProductoNegocioID: productID, UnidadMedidaID: unitID, CodigoBarras: barcode, CantidadEmpaques: row.CantidadEmpaques, PiezasPorEmpaque: piecesPerPackage, TotalPiezas: totalPieces, PiezasDanadas: row.PiezasDanadas, PiezasFaltantes: row.PiezasFaltantes, PiezasBuenas: good, CostoEmpaque: money(row.CostoEmpaque), CostoPieza: money4((gross - row.Descuento) / float64(totalPieces)), Descuento: money(row.Descuento), Subtotal: money(gross - row.Descuento), NumeroLote: trimPtr(row.NumeroLote), FechaCaducidad: expiry, ProductoNuevo: row.ProductoNuevo}, nil
}

func (r *Repository) receiveDetail(tx *gorm.DB, purchase compradomain.CompraProveedor, detail compradomain.CompraProveedorDetalle) error {
	var inventory domain.InventarioSucursal
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sucursal_id = ? AND producto_negocio_id = ?", purchase.SucursalID, detail.ProductoNegocioID).First(&inventory).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		inventory = domain.InventarioSucursal{SucursalID: purchase.SucursalID, ProductoNegocioID: detail.ProductoNegocioID}
		if err := tx.Create(&inventory).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	old := inventory.StockActual
	next := old + float64(detail.PiezasBuenas)
	if err := tx.Model(&inventory).Update("stock_actual", next).Error; err != nil {
		return err
	}
	var lotID *uuid.UUID
	if detail.NumeroLote != nil || detail.FechaCaducidad != nil {
		lot := domain.Lote{SucursalID: purchase.SucursalID, ProductoNegocioID: detail.ProductoNegocioID, ProveedorID: &purchase.ProveedorID, NumeroLote: detail.NumeroLote, CantidadInicial: float64(detail.PiezasBuenas), CantidadActual: float64(detail.PiezasBuenas), CostoUnitario: &detail.CostoPieza, FechaCaducidad: detail.FechaCaducidad, FechaRecepcion: purchase.FechaRecepcion, Activo: true}
		if err := tx.Create(&lot).Error; err != nil {
			return err
		}
		lotID = &lot.ID
	}
	return tx.Create(&domain.MovimientoInventario{SucursalID: purchase.SucursalID, ProductoNegocioID: detail.ProductoNegocioID, LoteID: lotID, Tipo: "COMPRA_ENTRADA", Cantidad: float64(detail.PiezasBuenas), StockAnterior: old, StockNuevo: next, CostoUnitario: &detail.CostoPieza, Referencia: &purchase.FolioDocumento, UsuarioID: &purchase.RecibidoPorUsuarioID}).Error
}

func (r *Repository) createNewProduct(tx *gorm.DB, businessID, providerID uuid.UUID, name, barcode string) (uuid.UUID, error) {
	barcode = strings.TrimSpace(barcode)
	var existing int64
	if err := tx.Model(&domain.ProductoCodigo{}).Where("negocio_id = ? AND codigo = ?", businessID, barcode).Count(&existing).Error; err != nil {
		return uuid.Nil, err
	}
	if existing > 0 {
		return uuid.Nil, errors.New("el código de barras ya está registrado")
	}
	p := domain.Producto{NegocioID: businessID, Nombre: strings.TrimSpace(name), Activo: true}
	if err := tx.Create(&p).Error; err != nil {
		return uuid.Nil, err
	}
	sku := fmt.Sprintf("COMP-%s", strings.ToUpper(uuid.NewString()[:8]))
	pn := domain.ProductoNegocio{NegocioID: businessID, ProductoID: &p.ID, SKUInterno: &sku, PrecioVenta: 0, PrecioIncluyeImpuestos: true, Activo: true}
	if err := tx.Create(&pn).Error; err != nil {
		return uuid.Nil, err
	}
	if err := tx.Create(&domain.ProductoCodigo{NegocioID: businessID, ProductoID: &p.ID, Tipo: "GTIN", Codigo: barcode, EsPrincipal: true}).Error; err != nil {
		return uuid.Nil, err
	}
	if err := tx.Create(&domain.ProductoProveedor{ProductoNegocioID: pn.ID, ProveedorID: providerID, EsPrincipal: true}).Error; err != nil {
		return uuid.Nil, err
	}
	return pn.ID, nil
}
func validPayment(value string) bool {
	for _, item := range []string{"EFECTIVO", "TRANSFERENCIA", "TARJETA", "CREDITO", "OTRO"} {
		if value == item {
			return true
		}
	}
	return false
}
func trimPtr(value *string) *string {
	if value == nil {
		return nil
	}
	text := strings.TrimSpace(*value)
	if text == "" {
		return nil
	}
	return &text
}
func money(value float64) float64  { return math.Round(value*100) / 100 }
func money4(value float64) float64 { return math.Round(value*10000) / 10000 }

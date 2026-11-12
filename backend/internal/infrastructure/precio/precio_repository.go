package precio

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	precioapp "tienda/backend/internal/application/precio"
	preciodomain "tienda/backend/internal/domain/precio"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }
func (r *Repository) ListPending(businessID, branchID uuid.UUID, filtro precioapp.FiltroPendientes) ([]precioapp.PropuestaResumen, error) {
	items := make([]precioapp.PropuestaResumen, 0)
	query := r.db.Table("propuestas_costo_precio p").Select(`
		p.id, p.negocio_id, p.sucursal_id, p.compra_detalle_id, p.producto_negocio_id,
		d.piezas_buenas, ROUND((d.subtotal + CASE WHEN d.impuestos_por_renglon THEN d.ieps + d.iva WHEN c.subtotal > 0 THEN c.impuestos * d.subtotal / c.subtotal ELSE 0 END) / NULLIF(d.piezas_buenas, 0), 4) AS costo_capturado,
		ps.costo_vigente AS costo_anterior, COALESCE(ps.precio_venta, 0) AS precio_anterior,
		COALESCE(ps.margen, p.margen_sugerido) AS margen_sugerido, p.precio_sugerido, p.estado,
		p.costo_autorizado, p.precio_autorizado, p.margen_autorizado, p.autorizado_por_usuario_id, p.autorizado_en, p.creado_en,
		COALESCE(pr.nombre, vp.nombre) || COALESCE(' · ' || pv.clave, '') AS producto,
		c.id AS compra_id, c.tipo_documento, c.folio_documento, c.proveedor_id, pro.nombre AS proveedor,
		c.fecha_documento, c.fecha_recepcion, c.total AS total_compra, c.ieps AS compra_ieps, c.iva AS compra_iva,
		d.codigo_barras, COALESCE(u.nombre, '') AS unidad, d.cantidad_empaques, d.piezas_danadas, d.piezas_faltantes, d.descuento,
		d.impuestos_por_renglon, d.tasa_ieps, d.tasa_iva, d.producto_nuevo`).
		Joins("JOIN compras_proveedor_detalles d ON d.id = p.compra_detalle_id").Joins("JOIN compras_proveedor c ON c.id = d.compra_id").
		Joins("JOIN proveedores pro ON pro.id = c.proveedor_id").Joins("LEFT JOIN unidades_medida u ON u.id = d.unidad_medida_id").
		Joins("LEFT JOIN precios_sucursal ps ON ps.sucursal_id = p.sucursal_id AND ps.producto_negocio_id = p.producto_negocio_id").
		Joins("JOIN producto_negocio pn ON pn.id = p.producto_negocio_id").Joins("LEFT JOIN productos pr ON pr.id = pn.producto_id").Joins("LEFT JOIN producto_variantes pv ON pv.id = pn.producto_variante_id").Joins("LEFT JOIN productos vp ON vp.id = pv.producto_id").
		Where("p.negocio_id = ? AND p.sucursal_id = ? AND p.estado = ?", businessID, branchID, preciodomain.EstadoPendiente)
	if filtro.ProveedorID != nil {
		query = query.Where("c.proveedor_id = ?", *filtro.ProveedorID)
	}
	if filtro.Desde != nil {
		query = query.Where("c.fecha_recepcion >= ?", filtro.Desde.Format("2006-01-02"))
	}
	if filtro.Hasta != nil {
		query = query.Where("c.fecha_recepcion <= ?", filtro.Hasta.Format("2006-01-02"))
	}
	if texto := strings.TrimSpace(filtro.Texto); texto != "" {
		like := "%" + strings.ToLower(texto) + "%"
		query = query.Where(`(lower(c.folio_documento) LIKE ? OR lower(pro.nombre) LIKE ? OR lower(COALESCE(pr.nombre, vp.nombre) || COALESCE(' ' || pv.clave, '')) LIKE ? OR COALESCE(d.codigo_barras, '') LIKE ?)`, like, like, like, like)
	}
	err := query.Order("c.fecha_recepcion DESC, c.creado_en DESC, d.creado_en ASC, p.creado_en ASC").Scan(&items).Error
	// El margen mostrado puede ser el vigente de la sucursal; el sugerido sigue la misma regla con ese margen.
	for i := range items {
		items[i].PrecioSugerido = preciodomain.PrecioSugerido(items[i].CostoCapturado, items[i].MargenSugerido)
		items[i].Nota = notaRenglon(items[i])
	}
	return items, err
}

// notaRenglon resume el renglón: unidad × empaques, dañadas, faltantes, descuento e impuestos.
func notaRenglon(item precioapp.PropuestaResumen) string {
	partes := []string{}
	if item.Unidad != "" {
		partes = append(partes, fmt.Sprintf("%d × %s", item.CantidadEmpaques, item.Unidad))
	}
	if item.PiezasDanadas > 0 {
		partes = append(partes, fmt.Sprintf("%d dañadas", item.PiezasDanadas))
	}
	if item.PiezasFaltantes > 0 {
		partes = append(partes, fmt.Sprintf("%d faltantes", item.PiezasFaltantes))
	}
	if item.Descuento > 0 {
		partes = append(partes, fmt.Sprintf("descuento $%.2f", item.Descuento))
	}
	if item.ImpuestosPorRenglon {
		if item.TasaIEPS > 0 {
			partes = append(partes, "IEPS "+porcentaje(item.TasaIEPS))
		}
		partes = append(partes, "IVA "+porcentaje(item.TasaIVA))
	} else if item.CompraIEPS > 0 || item.CompraIVA > 0 {
		partes = append(partes, "impuestos de la compra prorrateados")
	}
	return strings.Join(partes, " · ")
}

func porcentaje(tasa float64) string {
	return strings.TrimSuffix(strings.TrimRight(strconv.FormatFloat(tasa*100, 'f', 2, 64), "0"), ".") + " %"
}

func validAmounts(input precioapp.AutorizarInput) bool {
	return !(input.Costo < 0 || input.Margen < 0 || input.PrecioVenta < 0 || math.IsNaN(input.Costo) || math.IsNaN(input.Margen) || math.IsNaN(input.PrecioVenta) || math.IsInf(input.Costo, 0) || math.IsInf(input.Margen, 0) || math.IsInf(input.PrecioVenta, 0))
}

func (r *Repository) Authorize(businessID, userID, proposalID uuid.UUID, input precioapp.AutorizarInput) (preciodomain.Propuesta, error) {
	if !validAmounts(input) {
		return preciodomain.Propuesta{}, errors.New("los importes no son válidos")
	}
	var proposal preciodomain.Propuesta
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var err error
		proposal, err = r.authorizeTx(tx, businessID, userID, proposalID, uuid.Nil, input)
		return err
	})
	return proposal, err
}

// AuthorizePurchase valida varios renglones del mismo ticket en una transacción.
func (r *Repository) AuthorizePurchase(businessID, userID, purchaseID uuid.UUID, input precioapp.AutorizarCompraInput) ([]preciodomain.Propuesta, error) {
	if len(input.Renglones) == 0 {
		return nil, errors.New("elige al menos un producto para validar")
	}
	for _, row := range input.Renglones {
		if !validAmounts(row.AutorizarInput) {
			return nil, errors.New("los importes no son válidos")
		}
	}
	items := make([]preciodomain.Propuesta, 0, len(input.Renglones))
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, row := range input.Renglones {
			proposal, err := r.authorizeTx(tx, businessID, userID, row.PropuestaID, purchaseID, row.AutorizarInput)
			if err != nil {
				return err
			}
			items = append(items, proposal)
		}
		return nil
	})
	return items, err
}

// authorizeTx autoriza una propuesta pendiente; si `purchaseID` no es nulo exige que sea de esa compra.
func (r *Repository) authorizeTx(tx *gorm.DB, businessID, userID, proposalID, purchaseID uuid.UUID, input precioapp.AutorizarInput) (preciodomain.Propuesta, error) {
	var proposal preciodomain.Propuesta
	query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND negocio_id = ? AND estado = ?", proposalID, businessID, preciodomain.EstadoPendiente)
	if purchaseID != uuid.Nil {
		query = query.Where("compra_detalle_id IN (SELECT id FROM compras_proveedor_detalles WHERE compra_id = ?)", purchaseID)
	}
	if err := query.First(&proposal).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return proposal, errors.New("la propuesta no está pendiente")
		}
		return proposal, err
	}
	cost, margin, price := round4(input.Costo), round4(input.Margen), round2(input.PrecioVenta)
	now := r.db.NowFunc()
	if err := tx.Model(&proposal).Updates(map[string]any{"estado": preciodomain.EstadoAutorizada, "costo_autorizado": cost, "margen_autorizado": margin, "precio_autorizado": price, "autorizado_por_usuario_id": userID, "autorizado_en": now}).Error; err != nil {
		return proposal, err
	}
	current := preciodomain.PrecioSucursal{NegocioID: businessID, SucursalID: proposal.SucursalID, ProductoNegocioID: proposal.ProductoNegocioID, CostoVigente: cost, PrecioVenta: price, Margen: &margin}
	if err := tx.Where("sucursal_id = ? AND producto_negocio_id = ?", proposal.SucursalID, proposal.ProductoNegocioID).Assign(current).FirstOrCreate(&current).Error; err != nil {
		return proposal, err
	}
	return proposal, tx.Create(&preciodomain.HistorialCostoPrecioSucursal{NegocioID: businessID, SucursalID: proposal.SucursalID, ProductoNegocioID: proposal.ProductoNegocioID, Origen: preciodomain.OrigenCompraAutorizada, CostoAnterior: proposal.CostoAnterior, CostoNuevo: &cost, PrecioAnterior: &proposal.PrecioAnterior, PrecioNuevo: &price, PropuestaID: &proposal.ID, UsuarioID: &userID}).Error
}

func (r *Repository) ListHistory(businessID, branchID, productID uuid.UUID, limit int) ([]preciodomain.HistorialCostoPrecioSucursal, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	items := make([]preciodomain.HistorialCostoPrecioSucursal, 0)
	err := r.db.Where("negocio_id = ? AND sucursal_id = ? AND producto_negocio_id = ?", businessID, branchID, productID).Order("creado_en DESC").Limit(limit).Find(&items).Error
	return items, err
}

func (r *Repository) AdjustPrice(businessID, userID, productID uuid.UUID, input precioapp.AjustarPrecioInput) (preciodomain.HistorialCostoPrecioSucursal, error) {
	if input.SucursalID == uuid.Nil || input.PrecioVenta < 0 || math.IsNaN(input.PrecioVenta) {
		return preciodomain.HistorialCostoPrecioSucursal{}, errors.New("el precio no es válido")
	}
	motivo := strings.TrimSpace(input.Motivo)
	if motivo == "" {
		return preciodomain.HistorialCostoPrecioSucursal{}, errors.New("el motivo es obligatorio")
	}
	var event preciodomain.HistorialCostoPrecioSucursal
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current preciodomain.PrecioSucursal
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("negocio_id = ? AND sucursal_id = ? AND producto_negocio_id = ?", businessID, input.SucursalID, productID).First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("el producto no tiene precio configurado en esta sucursal")
			}
			return err
		}
		price := round2(input.PrecioVenta)
		previous := current.PrecioVenta
		if err := tx.Model(&current).Update("precio_venta", price).Error; err != nil {
			return err
		}
		event = preciodomain.HistorialCostoPrecioSucursal{NegocioID: businessID, SucursalID: input.SucursalID, ProductoNegocioID: productID, Origen: preciodomain.OrigenAjustePrecioManual, PrecioAnterior: &previous, PrecioNuevo: &price, UsuarioID: &userID, Motivo: &motivo}
		return tx.Create(&event).Error
	})
	return event, err
}
func round2(v float64) float64 { return math.Round(v*100) / 100 }
func round4(v float64) float64 { return math.Round(v*10000) / 10000 }

func (r *Repository) BranchOfProposal(businessID, proposalID uuid.UUID) (uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Table("propuestas_costo_precio").
		Where("id = ? AND negocio_id = ?", proposalID, businessID).Limit(1).Pluck("sucursal_id", &ids).Error
	if err != nil {
		return uuid.Nil, err
	}
	if len(ids) == 0 {
		return uuid.Nil, gorm.ErrRecordNotFound
	}
	return ids[0], nil
}

func (r *Repository) BranchOfPurchase(businessID, purchaseID uuid.UUID) (uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Table("compras_proveedor").Where("id = ? AND negocio_id = ?", purchaseID, businessID).Limit(1).Pluck("sucursal_id", &ids).Error
	if err != nil {
		return uuid.Nil, err
	}
	if len(ids) == 0 {
		return uuid.Nil, gorm.ErrRecordNotFound
	}
	return ids[0], nil
}

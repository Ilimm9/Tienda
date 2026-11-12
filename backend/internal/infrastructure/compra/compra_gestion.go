package compra

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	compraapp "tienda/backend/internal/application/compra"
	"tienda/backend/internal/domain"
	compradomain "tienda/backend/internal/domain/compra"
	preciodomain "tienda/backend/internal/domain/precio"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Edit cambia el documento y, si ningún costo está validado, los productos. Todo en una transacción;
// cada diferencia de piezas agrega un movimiento COMPRA_AJUSTE y el COMPRA_ENTRADA original no se toca.
func (r *Repository) Edit(businessID, userID, purchaseID uuid.UUID, input compraapp.EditarCompraInput) (compradomain.CompraProveedor, error) {
	if input.TipoDocumento != compradomain.DocumentoTicket && input.TipoDocumento != compradomain.DocumentoFactura {
		return compradomain.CompraProveedor{}, errors.New("el tipo de documento no es válido")
	}
	folio := strings.TrimSpace(input.FolioDocumento)
	if folio == "" {
		return compradomain.CompraProveedor{}, errors.New("el folio es obligatorio")
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
	var result compradomain.CompraProveedor
	err = r.db.Transaction(func(tx *gorm.DB) error {
		purchase, err := lockPurchase(tx, businessID, purchaseID)
		if err != nil {
			return err
		}
		if purchase.Estado != compradomain.EstadoActiva {
			return errors.New("solo se pueden editar compras activas")
		}
		if err := ensureLots(tx, &purchase); err != nil {
			return err
		}
		before := snapshot(purchase)
		var duplicated int64
		if err := tx.Model(&compradomain.CompraProveedor{}).Where("negocio_id = ? AND proveedor_id = ? AND tipo_documento = ? AND folio_documento = ? AND id <> ?", businessID, purchase.ProveedorID, input.TipoDocumento, folio, purchase.ID).Count(&duplicated).Error; err != nil {
			return err
		}
		if duplicated > 0 {
			return errors.New("el documento ya está registrado para este proveedor")
		}
		if input.Detalles != nil {
			if err := r.editDetails(tx, businessID, userID, &purchase, input); err != nil {
				return err
			}
		}
		now := tx.NowFunc()
		updates := map[string]any{"tipo_documento": input.TipoDocumento, "folio_documento": folio, "numero_cuenta": trimPtr(input.NumeroCuenta), "fecha_documento": dateDocument,
			"fecha_recepcion": dateReception, "forma_pago": input.FormaPago, "observaciones": trimPtr(input.Observaciones), "actualizado_en": now, "editado_por_usuario_id": userID}
		if err := tx.Model(&compradomain.CompraProveedor{}).Where("id = ?", purchase.ID).Updates(updates).Error; err != nil {
			return err
		}
		result, err = lockPurchase(tx, businessID, purchaseID)
		if err != nil {
			return err
		}
		return tx.Create(&compradomain.CompraProveedorEvento{CompraID: purchase.ID, UsuarioID: userID, Tipo: compradomain.EventoEdicion, Motivo: trimPtr(input.Motivo), Antes: before, Despues: snapshot(result)}).Error
	})
	return result, err
}

// cambioRenglon guarda lo que pasa con cada renglón al editar.
type cambioRenglon struct {
	anterior *compradomain.CompraProveedorDetalle
	nuevo    compradomain.CompraProveedorDetalle
	quitar   bool
	fila     compraapp.CrearDetalleInput
}

func (r *Repository) editDetails(tx *gorm.DB, businessID, userID uuid.UUID, purchase *compradomain.CompraProveedor, input compraapp.EditarCompraInput) error {
	var authorized int64
	if err := tx.Table("propuestas_costo_precio p").Joins("JOIN compras_proveedor_detalles d ON d.id = p.compra_detalle_id").
		Where("d.compra_id = ? AND p.estado = ?", purchase.ID, preciodomain.EstadoAutorizada).Count(&authorized).Error; err != nil {
		return err
	}
	if authorized > 0 {
		return errors.New("hay costos de esta compra ya validados; los productos no se pueden cambiar, solo el documento")
	}
	existing := map[uuid.UUID]*compradomain.CompraProveedorDetalle{}
	for i := range purchase.Detalles {
		existing[purchase.Detalles[i].ID] = &purchase.Detalles[i]
	}
	listed := map[uuid.UUID]bool{}
	cambios := make([]cambioRenglon, 0, len(input.Detalles))
	for _, row := range input.Detalles {
		cambio := cambioRenglon{fila: row.CrearDetalleInput}
		if row.ID != nil {
			id, err := uuid.Parse(*row.ID)
			if err != nil || existing[id] == nil || listed[id] {
				return errors.New("un producto de la edición no pertenece a la compra")
			}
			listed[id] = true
			cambio.anterior = existing[id]
			if row.Quitar {
				cambio.quitar = true
				cambios = append(cambios, cambio)
				continue
			}
			if row.ProductoNegocioID != cambio.anterior.ProductoNegocioID.String() {
				return errors.New("no se puede cambiar el producto de un renglón; quítalo y agrega otro")
			}
			row.ProductoNuevo = false
		}
		detail, err := r.buildDetail(tx, businessID, purchase.ProveedorID, row.CrearDetalleInput)
		if err != nil {
			return err
		}
		if err := applyRates(&detail, row.CrearDetalleInput); err != nil {
			return err
		}
		detail.CompraID = purchase.ID
		if cambio.anterior != nil {
			detail.ID, detail.CreadoEn, detail.LoteID, detail.ProductoNuevo = cambio.anterior.ID, cambio.anterior.CreadoEn, cambio.anterior.LoteID, cambio.anterior.ProductoNuevo
		}
		cambio.nuevo = detail
		cambios = append(cambios, cambio)
	}
	if len(listed) != len(existing) {
		return errors.New("la edición debe incluir todos los productos de la compra")
	}
	activos := []compradomain.CompraProveedorDetalle{}
	for _, cambio := range cambios {
		if !cambio.quitar {
			activos = append(activos, cambio.nuevo)
		}
	}
	if len(activos) == 0 {
		return errors.New("la compra debe conservar al menos un producto; para anularla, suspéndela")
	}
	if err := compradomain.CalcularImpuestosRenglon(activos, input.IEPSAjustado); err != nil {
		return err
	}
	subtotal, ieps, iva := 0.0, 0.0, 0.0
	for _, detail := range activos {
		subtotal, ieps, iva = subtotal+detail.Subtotal, ieps+detail.IEPS, iva+detail.IVA
	}
	purchase.Subtotal, purchase.IEPS, purchase.IVA = money(subtotal), money(ieps), money(iva)
	purchase.Impuestos = money(purchase.IEPS + purchase.IVA)
	purchase.Total = money(purchase.Subtotal + purchase.Impuestos)
	if err := tx.Model(&compradomain.CompraProveedor{}).Where("id = ?", purchase.ID).Updates(map[string]any{"subtotal": purchase.Subtotal, "ieps": purchase.IEPS, "iva": purchase.IVA,
		"impuestos": purchase.Impuestos, "total": purchase.Total, "importe_pagado": purchase.Total}).Error; err != nil {
		return err
	}
	// Los renglones nuevos entran a nombre de quien edita.
	receiving := *purchase
	receiving.RecibidoPorUsuarioID = userID
	activo := 0
	for _, cambio := range cambios {
		switch {
		case cambio.quitar:
			if err := r.removeDetail(tx, *purchase, *cambio.anterior, userID); err != nil {
				return err
			}
		case cambio.anterior == nil:
			detail := activos[activo]
			activo++
			if err := tx.Create(&detail).Error; err != nil {
				return err
			}
			if err := r.receiveDetail(tx, receiving, &detail, compradomain.MovimientoCompraAjuste); err != nil {
				return err
			}
			if err := r.createPriceProposal(tx, *purchase, detail); err != nil {
				return err
			}
			if err := rememberRates(tx, detail); err != nil {
				return err
			}
		default:
			detail := activos[activo]
			activo++
			if err := r.updateDetail(tx, *purchase, *cambio.anterior, detail, userID); err != nil {
				return err
			}
			if err := rememberRates(tx, detail); err != nil {
				return err
			}
		}
	}
	return nil
}

// updateDetail guarda el renglón editado, ajusta el stock por la diferencia y actualiza lote y propuesta.
func (r *Repository) updateDetail(tx *gorm.DB, purchase compradomain.CompraProveedor, before, after compradomain.CompraProveedorDetalle, userID uuid.UUID) error {
	cost := compradomain.CostoFinal(purchase, after)
	delta := float64(after.PiezasBuenas - before.PiezasBuenas)
	if delta != 0 {
		if err := moveStock(tx, purchase, after, delta, compradomain.MovimientoCompraAjuste, &cost, userID, nil); err != nil {
			return err
		}
	}
	if after.LoteID != nil {
		if err := tx.Model(&domain.Lote{}).Where("id = ?", *after.LoteID).Updates(map[string]any{"cantidad_inicial": float64(after.PiezasBuenas),
			"cantidad_actual": gorm.Expr("GREATEST(cantidad_actual + ?, 0)", delta), "costo_unitario": cost, "numero_lote": after.NumeroLote, "fecha_caducidad": after.FechaCaducidad}).Error; err != nil {
			return err
		}
	}
	if err := tx.Select("*").Omit("creado_en").Save(&after).Error; err != nil {
		return err
	}
	var proposal preciodomain.Propuesta
	if err := tx.Where("compra_detalle_id = ?", after.ID).First(&proposal).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return tx.Model(&proposal).Updates(map[string]any{"costo_capturado": cost, "piezas_buenas": after.PiezasBuenas,
		"precio_sugerido": preciodomain.PrecioSugerido(cost, proposal.MargenSugerido)}).Error
}

// removeDetail cancela un renglón: salen sus piezas, baja su lote y su propuesta pasa a CANCELADA.
func (r *Repository) removeDetail(tx *gorm.DB, purchase compradomain.CompraProveedor, detail compradomain.CompraProveedorDetalle, userID uuid.UUID) error {
	cost := compradomain.CostoFinal(purchase, detail)
	if err := moveStock(tx, purchase, detail, -float64(detail.PiezasBuenas), compradomain.MovimientoCompraAjuste, &cost, userID, nil); err != nil {
		return err
	}
	if err := changeLot(tx, detail, -float64(detail.PiezasBuenas)); err != nil {
		return err
	}
	if err := tx.Model(&compradomain.CompraProveedorDetalle{}).Where("id = ?", detail.ID).Update("cancelado", true).Error; err != nil {
		return err
	}
	return tx.Model(&preciodomain.Propuesta{}).Where("compra_detalle_id = ? AND estado = ?", detail.ID, preciodomain.EstadoPendiente).Update("estado", preciodomain.EstadoCancelada).Error
}

// Suspend devuelve las piezas de la compra y cancela sus costos pendientes. Los validados no se tocan.
func (r *Repository) Suspend(businessID, userID, purchaseID uuid.UUID, motivo string) (compradomain.CompraProveedor, error) {
	motivo = strings.TrimSpace(motivo)
	if motivo == "" {
		return compradomain.CompraProveedor{}, errors.New("el motivo es obligatorio")
	}
	var result compradomain.CompraProveedor
	err := r.db.Transaction(func(tx *gorm.DB) error {
		purchase, err := lockPurchase(tx, businessID, purchaseID)
		if err != nil {
			return err
		}
		if purchase.Estado != compradomain.EstadoActiva {
			return errors.New("la compra ya está suspendida")
		}
		if err := ensureLots(tx, &purchase); err != nil {
			return err
		}
		if err := checkStock(tx, purchase); err != nil {
			return err
		}
		before := snapshot(purchase)
		for _, detail := range purchase.Detalles {
			cost := compradomain.CostoFinal(purchase, detail)
			if err := moveStock(tx, purchase, detail, -float64(detail.PiezasBuenas), compradomain.MovimientoCompraSuspension, &cost, userID, &motivo); err != nil {
				return err
			}
			if err := changeLot(tx, detail, -float64(detail.PiezasBuenas)); err != nil {
				return err
			}
		}
		var canceled []uuid.UUID
		if err := tx.Model(&preciodomain.Propuesta{}).Where("compra_detalle_id IN (SELECT id FROM compras_proveedor_detalles WHERE compra_id = ?) AND estado = ?", purchase.ID, preciodomain.EstadoPendiente).
			Pluck("id", &canceled).Error; err != nil {
			return err
		}
		if len(canceled) > 0 {
			if err := tx.Model(&preciodomain.Propuesta{}).Where("id IN ?", canceled).Update("estado", preciodomain.EstadoCancelada).Error; err != nil {
				return err
			}
		}
		now := tx.NowFunc()
		if err := tx.Model(&compradomain.CompraProveedor{}).Where("id = ?", purchase.ID).Updates(map[string]any{"estado": compradomain.EstadoSuspendida, "motivo_suspension": motivo,
			"suspendido_por_usuario_id": userID, "suspendido_en": now, "actualizado_en": now}).Error; err != nil {
			return err
		}
		after, _ := json.Marshal(map[string]any{"estado": compradomain.EstadoSuspendida, "propuestas_canceladas": canceled})
		if err := tx.Create(&compradomain.CompraProveedorEvento{CompraID: purchase.ID, UsuarioID: userID, Tipo: compradomain.EventoSuspension, Motivo: &motivo, Antes: before, Despues: after}).Error; err != nil {
			return err
		}
		result, err = lockPurchase(tx, businessID, purchaseID)
		return err
	})
	return result, err
}

// Reactivate vuelve a sumar las piezas y regresa a pendientes las propuestas que canceló la última suspensión.
func (r *Repository) Reactivate(businessID, userID, purchaseID uuid.UUID, motivo string) (compradomain.CompraProveedor, error) {
	var result compradomain.CompraProveedor
	err := r.db.Transaction(func(tx *gorm.DB) error {
		purchase, err := lockPurchase(tx, businessID, purchaseID)
		if err != nil {
			return err
		}
		if purchase.Estado != compradomain.EstadoSuspendida {
			return errors.New("solo se pueden reactivar compras suspendidas")
		}
		if err := ensureLots(tx, &purchase); err != nil {
			return err
		}
		before := snapshot(purchase)
		reason := trimPtr(&motivo)
		for _, detail := range purchase.Detalles {
			cost := compradomain.CostoFinal(purchase, detail)
			if err := moveStock(tx, purchase, detail, float64(detail.PiezasBuenas), compradomain.MovimientoCompraReactivacion, &cost, userID, reason); err != nil {
				return err
			}
			if err := changeLot(tx, detail, float64(detail.PiezasBuenas)); err != nil {
				return err
			}
		}
		var suspension compradomain.CompraProveedorEvento
		restored := []uuid.UUID{}
		err = tx.Where("compra_id = ? AND tipo = ?", purchase.ID, compradomain.EventoSuspension).Order("creado_en DESC").First(&suspension).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			var payload struct {
				Canceladas []uuid.UUID `json:"propuestas_canceladas"`
			}
			if len(suspension.Despues) > 0 {
				if err := json.Unmarshal(suspension.Despues, &payload); err != nil {
					return err
				}
			}
			if len(payload.Canceladas) > 0 {
				// Solo regresan las de esa suspensión y que sigan canceladas; la foto anterior se conserva.
				if err := tx.Model(&preciodomain.Propuesta{}).Where("id IN ? AND estado = ? AND compra_detalle_id IN (SELECT id FROM compras_proveedor_detalles WHERE compra_id = ? AND NOT cancelado)", payload.Canceladas, preciodomain.EstadoCancelada, purchase.ID).
					Pluck("id", &restored).Error; err != nil {
					return err
				}
				if len(restored) > 0 {
					if err := tx.Model(&preciodomain.Propuesta{}).Where("id IN ?", restored).Update("estado", preciodomain.EstadoPendiente).Error; err != nil {
						return err
					}
				}
			}
		}
		now := tx.NowFunc()
		if err := tx.Model(&compradomain.CompraProveedor{}).Where("id = ?", purchase.ID).Updates(map[string]any{"estado": compradomain.EstadoActiva,
			"reactivado_por_usuario_id": userID, "reactivado_en": now, "actualizado_en": now}).Error; err != nil {
			return err
		}
		after, _ := json.Marshal(map[string]any{"estado": compradomain.EstadoActiva, "propuestas_restauradas": restored})
		if err := tx.Create(&compradomain.CompraProveedorEvento{CompraID: purchase.ID, UsuarioID: userID, Tipo: compradomain.EventoReactivacion, Motivo: reason, Antes: before, Despues: after}).Error; err != nil {
			return err
		}
		result, err = lockPurchase(tx, businessID, purchaseID)
		return err
	})
	return result, err
}

// lockPurchase bloquea la compra y carga sus renglones vigentes.
func lockPurchase(tx *gorm.DB, businessID, purchaseID uuid.UUID) (compradomain.CompraProveedor, error) {
	return loadPurchase(tx.Clauses(clause.Locking{Strength: "UPDATE"}), tx, businessID, purchaseID)
}

func loadPurchase(query, tx *gorm.DB, businessID, purchaseID uuid.UUID) (compradomain.CompraProveedor, error) {
	var purchase compradomain.CompraProveedor
	if err := query.Where("id = ? AND negocio_id = ?", purchaseID, businessID).First(&purchase).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return purchase, errors.New("la compra no existe")
		}
		return purchase, err
	}
	err := tx.Where("compra_id = ? AND cancelado = FALSE", purchase.ID).Order("creado_en ASC").Find(&purchase.Detalles).Error
	return purchase, err
}

// ensureLots liga los renglones anteriores a su lote, buscando el movimiento COMPRA_ENTRADA que lo creó.
func ensureLots(tx *gorm.DB, purchase *compradomain.CompraProveedor) error {
	claimed := map[uuid.UUID]bool{}
	for _, detail := range purchase.Detalles {
		if detail.LoteID != nil {
			claimed[*detail.LoteID] = true
		}
	}
	for i := range purchase.Detalles {
		detail := &purchase.Detalles[i]
		if detail.LoteID != nil || (detail.NumeroLote == nil && detail.FechaCaducidad == nil) {
			continue
		}
		var lots []uuid.UUID
		if err := tx.Model(&domain.MovimientoInventario{}).Where("sucursal_id = ? AND producto_negocio_id = ? AND tipo = ? AND referencia = ? AND lote_id IS NOT NULL",
			purchase.SucursalID, detail.ProductoNegocioID, compradomain.MovimientoCompraEntrada, purchase.FolioDocumento).Order("creado_en ASC").Pluck("lote_id", &lots).Error; err != nil {
			return err
		}
		for _, lot := range lots {
			if claimed[lot] {
				continue
			}
			claimed[lot] = true
			id := lot
			detail.LoteID = &id
			if err := tx.Model(&compradomain.CompraProveedorDetalle{}).Where("id = ?", detail.ID).Update("lote_id", id).Error; err != nil {
				return err
			}
			break
		}
	}
	return nil
}

// checkStock rechaza la suspensión si algún producto ya no tiene las piezas que entraron; lista todos.
func checkStock(tx *gorm.DB, purchase compradomain.CompraProveedor) error {
	need := map[uuid.UUID]int{}
	order := []uuid.UUID{}
	for _, detail := range purchase.Detalles {
		if _, ok := need[detail.ProductoNegocioID]; !ok {
			order = append(order, detail.ProductoNegocioID)
		}
		need[detail.ProductoNegocioID] += detail.PiezasBuenas
	}
	missing := []string{}
	for _, productID := range order {
		var stock float64
		if err := tx.Raw(`SELECT COALESCE((SELECT stock_actual FROM inventario_sucursal WHERE sucursal_id = ? AND producto_negocio_id = ?), 0)`, purchase.SucursalID, productID).Scan(&stock).Error; err != nil {
			return err
		}
		if stock < float64(need[productID]) {
			missing = append(missing, fmt.Sprintf("%s (hay %g, salen %d)", productName(tx, productID), stock, need[productID]))
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("no hay existencias suficientes para suspender; ya se vendieron piezas de: %s", strings.Join(missing, ", "))
	}
	return nil
}

// moveStock aplica la diferencia al inventario y agrega su movimiento; nunca deja el stock negativo.
func moveStock(tx *gorm.DB, purchase compradomain.CompraProveedor, detail compradomain.CompraProveedorDetalle, delta float64, tipo string, cost *float64, userID uuid.UUID, motivo *string) error {
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
	next := old + delta
	if next < 0 {
		return fmt.Errorf("no hay existencias suficientes de %s: hay %g y salen %g; ya se vendieron piezas de esta compra", productName(tx, detail.ProductoNegocioID), old, -delta)
	}
	if err := tx.Model(&inventory).Update("stock_actual", next).Error; err != nil {
		return err
	}
	return tx.Create(&domain.MovimientoInventario{SucursalID: purchase.SucursalID, ProductoNegocioID: detail.ProductoNegocioID, LoteID: detail.LoteID, Tipo: tipo, Cantidad: delta,
		StockAnterior: old, StockNuevo: next, CostoUnitario: cost, Motivo: motivo, Referencia: &purchase.FolioDocumento, UsuarioID: &userID}).Error
}

func changeLot(tx *gorm.DB, detail compradomain.CompraProveedorDetalle, delta float64) error {
	if detail.LoteID == nil {
		return nil
	}
	return tx.Model(&domain.Lote{}).Where("id = ?", *detail.LoteID).Update("cantidad_actual", gorm.Expr("GREATEST(cantidad_actual + ?, 0)", delta)).Error
}

func productName(tx *gorm.DB, productID uuid.UUID) string {
	var name string
	tx.Raw(`SELECT COALESCE(pr.nombre, vp.nombre) || COALESCE(' · ' || pv.clave, '') FROM producto_negocio pn
		LEFT JOIN productos pr ON pr.id = pn.producto_id LEFT JOIN producto_variantes pv ON pv.id = pn.producto_variante_id
		LEFT JOIN productos vp ON vp.id = pv.producto_id WHERE pn.id = ?`, productID).Scan(&name)
	if name == "" {
		return "un producto"
	}
	return name
}

func snapshot(purchase compradomain.CompraProveedor) []byte {
	data, _ := json.Marshal(purchase)
	return data
}

// View devuelve la compra con proveedor y sus renglones vigentes, con costo final y estado del costo.
func (r *Repository) View(businessID, purchaseID uuid.UUID) (compraapp.CompraVista, error) {
	var view compraapp.CompraVista
	purchase, err := loadPurchase(r.db, r.db, businessID, purchaseID)
	if err != nil {
		return view, err
	}
	purchase.CompletarImpuestosRenglon()
	view.CompraProveedor = purchase
	r.db.Table("proveedores").Select("nombre").Where("id = ?", purchase.ProveedorID).Scan(&view.Proveedor)
	var extras []struct {
		ID       uuid.UUID
		Producto string
		Unidad   string
		Estado   *string
	}
	if err := r.db.Table("compras_proveedor_detalles d").Select(`d.id, COALESCE(pr.nombre, vp.nombre) || COALESCE(' · ' || pv.clave, '') AS producto, COALESCE(u.nombre, '') AS unidad, p.estado`).
		Joins("JOIN producto_negocio pn ON pn.id = d.producto_negocio_id").Joins("LEFT JOIN productos pr ON pr.id = pn.producto_id").
		Joins("LEFT JOIN producto_variantes pv ON pv.id = pn.producto_variante_id").Joins("LEFT JOIN productos vp ON vp.id = pv.producto_id").
		Joins("LEFT JOIN unidades_medida u ON u.id = d.unidad_medida_id").Joins("LEFT JOIN propuestas_costo_precio p ON p.compra_detalle_id = d.id").
		Where("d.compra_id = ? AND NOT d.cancelado", purchase.ID).Scan(&extras).Error; err != nil {
		return view, err
	}
	byID := map[uuid.UUID]int{}
	for i, extra := range extras {
		byID[extra.ID] = i
	}
	for _, detail := range purchase.Detalles {
		row := compraapp.RenglonVista{CompraProveedorDetalle: detail, CostoFinal: compradomain.CostoFinal(purchase, detail)}
		if i, ok := byID[detail.ID]; ok {
			row.Producto, row.Unidad = extras[i].Producto, extras[i].Unidad
			if extras[i].Estado != nil {
				row.EstadoCosto = *extras[i].Estado
			}
		}
		view.Renglones = append(view.Renglones, row)
	}
	return view, nil
}

package precio

import (
	"errors"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"
	precioapp "tienda/backend/internal/application/precio"
	preciodomain "tienda/backend/internal/domain/precio"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }
func (r *Repository) ListPending(businessID, branchID uuid.UUID) ([]precioapp.PropuestaResumen, error) {
	items := make([]precioapp.PropuestaResumen, 0)
	err := r.db.Table("propuestas_costo_precio p").Select(`
		p.id, p.negocio_id, p.sucursal_id, p.compra_detalle_id, p.producto_negocio_id,
		d.piezas_buenas, ROUND((d.subtotal + CASE WHEN c.subtotal > 0 THEN c.impuestos * d.subtotal / c.subtotal ELSE 0 END) / NULLIF(d.piezas_buenas, 0), 4) AS costo_capturado,
		ps.costo_vigente AS costo_anterior, COALESCE(NULLIF(ps.precio_venta, 0), NULLIF(pn.precio_venta, 0), p.precio_anterior) AS precio_anterior,
		COALESCE(ps.margen, p.margen_sugerido) AS margen_sugerido, p.precio_sugerido, p.estado,
		p.costo_autorizado, p.precio_autorizado, p.margen_autorizado, p.autorizado_por_usuario_id, p.autorizado_en, p.creado_en,
		COALESCE(pr.nombre, vp.nombre) AS producto`).
		Joins("JOIN compras_proveedor_detalles d ON d.id = p.compra_detalle_id").Joins("JOIN compras_proveedor c ON c.id = d.compra_id").
		Joins("LEFT JOIN precios_sucursal ps ON ps.sucursal_id = p.sucursal_id AND ps.producto_negocio_id = p.producto_negocio_id").
		Joins("JOIN producto_negocio pn ON pn.id = p.producto_negocio_id").Joins("LEFT JOIN productos pr ON pr.id = pn.producto_id").Joins("LEFT JOIN producto_variantes pv ON pv.id = pn.producto_variante_id").Joins("LEFT JOIN productos vp ON vp.id = pv.producto_id").
		Where("p.negocio_id = ? AND p.sucursal_id = ? AND p.estado = ?", businessID, branchID, preciodomain.EstadoPendiente).Order("p.creado_en ASC").Scan(&items).Error
	return items, err
}
func (r *Repository) Authorize(businessID, userID, proposalID uuid.UUID, input precioapp.AutorizarInput) (preciodomain.Propuesta, error) {
	if input.Costo < 0 || input.Margen < 0 || input.PrecioVenta < 0 || math.IsNaN(input.Costo) || math.IsNaN(input.Margen) || math.IsNaN(input.PrecioVenta) {
		return preciodomain.Propuesta{}, errors.New("los importes no son válidos")
	}
	var proposal preciodomain.Propuesta
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND negocio_id = ? AND estado = ?", proposalID, businessID, preciodomain.EstadoPendiente).First(&proposal).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("la propuesta no está pendiente")
			}
			return err
		}
		cost, margin, price := round4(input.Costo), round4(input.Margen), round2(input.PrecioVenta)
		now := r.db.NowFunc()
		if err := tx.Model(&proposal).Updates(map[string]any{"estado": preciodomain.EstadoAutorizada, "costo_autorizado": cost, "margen_autorizado": margin, "precio_autorizado": price, "autorizado_por_usuario_id": userID, "autorizado_en": now}).Error; err != nil {
			return err
		}
		current := preciodomain.PrecioSucursal{NegocioID: businessID, SucursalID: proposal.SucursalID, ProductoNegocioID: proposal.ProductoNegocioID, CostoVigente: cost, PrecioVenta: price, Margen: &margin}
		return tx.Where("sucursal_id = ? AND producto_negocio_id = ?", proposal.SucursalID, proposal.ProductoNegocioID).Assign(current).FirstOrCreate(&current).Error
	})
	return proposal, err
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

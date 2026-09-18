package negocio

import (
	"context"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ContextoRepository struct {
	db *gorm.DB
}

func NewContextoRepository(db *gorm.DB) *ContextoRepository {
	return &ContextoRepository{db: db}
}

func (r *ContextoRepository) ListarOpcionesContexto(ctx context.Context, usuarioID uuid.UUID) ([]domain.ContextoNegocio, error) {
	type row struct {
		NegocioID         uuid.UUID
		Slug              string
		NombreComercial   string
		TipoMiembro       string
		SucursalID        *uuid.UUID
		SucursalCodigo    *string
		SucursalNombre    *string
		SucursalPrincipal *bool
	}
	rows := make([]row, 0)
	err := r.db.WithContext(ctx).Table("negocios AS n").
		Select(`n.id AS negocio_id, n.slug, n.nombre_comercial, m.tipo_miembro,
			s.id AS sucursal_id, s.codigo AS sucursal_codigo, s.nombre AS sucursal_nombre,
			s.es_principal AS sucursal_principal`).
		Joins("JOIN membresias_negocio m ON m.negocio_id = n.id AND m.usuario_id = ? AND m.estado = 'activo'", usuarioID).
		Joins("LEFT JOIN sucursales s ON s.negocio_id = n.id AND s.activo = TRUE AND s.eliminado_en IS NULL").
		Where("n.estado = 'activo'").
		Order("n.nombre_comercial ASC, n.id ASC, s.es_principal DESC, s.nombre ASC, s.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]domain.ContextoNegocio, 0)
	indexes := make(map[uuid.UUID]int)
	for _, item := range rows {
		index, exists := indexes[item.NegocioID]
		if !exists {
			index = len(items)
			indexes[item.NegocioID] = index
			items = append(items, domain.ContextoNegocio{
				ID: item.NegocioID, Slug: item.Slug, NombreComercial: item.NombreComercial,
				TipoMiembro: item.TipoMiembro, Sucursales: make([]domain.ContextoSucursal, 0),
			})
		}
		if item.SucursalID != nil {
			items[index].Sucursales = append(items[index].Sucursales, domain.ContextoSucursal{
				ID: *item.SucursalID, Codigo: derefString(item.SucursalCodigo),
				Nombre: derefString(item.SucursalNombre), EsPrincipal: derefBool(item.SucursalPrincipal),
			})
		}
	}
	return items, nil
}

func (r *ContextoRepository) NegocioActivoAccesible(ctx context.Context, usuarioID, negocioID uuid.UUID) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("negocios AS n").
		Joins("JOIN membresias_negocio m ON m.negocio_id = n.id AND m.usuario_id = ? AND m.estado = 'activo'", usuarioID).
		Where("n.id = ? AND n.estado = 'activo'", negocioID).
		Count(&total).Error
	return total == 1, err
}

func (r *ContextoRepository) SucursalActivaDelNegocio(ctx context.Context, negocioID, sucursalID uuid.UUID) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("sucursales").
		Where("id = ? AND negocio_id = ? AND activo = TRUE AND eliminado_en IS NULL", sucursalID, negocioID).
		Count(&total).Error
	return total == 1, err
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefBool(value *bool) bool {
	return value != nil && *value
}

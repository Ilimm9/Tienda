package negocio

import (
	"context"
	"errors"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AsignacionRepository struct {
	db *gorm.DB
}

func NewAsignacionRepository(db *gorm.DB) *AsignacionRepository {
	return &AsignacionRepository{db: db}
}

func (r *AsignacionRepository) ObtenerContextoNegocio(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.ContextoNegocioSucursal, error) {
	var contexto domain.ContextoNegocioSucursal
	err := r.db.WithContext(ctx).Table("negocios AS n").
		Select("n.estado AS estado_negocio, m.tipo_miembro").
		Joins("JOIN membresias_negocio m ON m.negocio_id = n.id AND m.usuario_id = ? AND m.estado = 'activo'", usuarioID).
		Where("n.id = ?", negocioID).
		Take(&contexto).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ContextoNegocioSucursal{}, application.ErrNegocioNoEncontrado
	}
	return contexto, err
}

func (r *AsignacionRepository) PermisosEfectivos(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]string, error) {
	return permisosEfectivos(ctx, r.db, usuarioID, negocioID)
}

func (r *AsignacionRepository) ObtenerEmpleado(ctx context.Context, negocioID, empleadoID uuid.UUID) (domain.EmpleadoDetalle, error) {
	return NewEmpleadoRepository(r.db).Obtener(ctx, negocioID, empleadoID)
}

func (r *AsignacionRepository) SucursalActivaDelNegocio(ctx context.Context, negocioID, sucursalID uuid.UUID) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("sucursales").
		Where("id = ? AND negocio_id = ? AND activo = TRUE AND eliminado_en IS NULL", sucursalID, negocioID).
		Count(&total).Error
	return total == 1, err
}

func (r *AsignacionRepository) Listar(ctx context.Context, negocioID, empleadoID uuid.UUID, incluirFinalizadas bool) ([]domain.AsignacionResumen, error) {
	items := make([]domain.AsignacionResumen, 0)
	query := r.db.WithContext(ctx).Table("asignaciones_empleado_sucursal AS a").
		Select(`a.id, a.negocio_id, a.empleado_id, a.sucursal_id, a.es_principal, a.activo,
			a.asignado_en, a.finalizado_en, s.codigo AS codigo_sucursal, s.nombre AS nombre_sucursal`).
		Joins("JOIN sucursales s ON s.id = a.sucursal_id").
		Where("a.negocio_id = ? AND a.empleado_id = ?", negocioID, empleadoID)
	if !incluirFinalizadas {
		query = query.Where("a.activo = TRUE")
	}
	err := query.Order("a.es_principal DESC, s.nombre ASC, a.id ASC").Scan(&items).Error
	return items, err
}

func (r *AsignacionRepository) ExisteActiva(ctx context.Context, negocioID, empleadoID, sucursalID uuid.UUID) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("asignaciones_empleado_sucursal").
		Where("negocio_id = ? AND empleado_id = ? AND sucursal_id = ? AND activo = TRUE",
			negocioID, empleadoID, sucursalID).
		Count(&total).Error
	return total > 0, err
}

// Asignar reactiva una asignación finalizada en lugar de duplicarla: el único es (empleado, sucursal).
func (r *AsignacionRepository) Asignar(ctx context.Context, negocioID, empleadoID uuid.UUID, input domain.AsignarSucursalInput) (uuid.UUID, error) {
	var asignacionID uuid.UUID
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existente domain.AsignacionEmpleadoSucursal
		err := tx.Where("empleado_id = ? AND sucursal_id = ?", empleadoID, input.SucursalID).
			Take(&existente).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			nueva := domain.AsignacionEmpleadoSucursal{
				NegocioID: negocioID, EmpleadoID: empleadoID, SucursalID: input.SucursalID,
				EsPrincipal: false, Activo: true,
			}
			if err := tx.Create(&nueva).Error; err != nil {
				return err
			}
			asignacionID = nueva.ID
		case err != nil:
			return err
		default:
			asignacionID = existente.ID
			err := tx.Table("asignaciones_empleado_sucursal").Where("id = ?", existente.ID).
				Updates(map[string]any{
					"activo": true, "finalizado_en": nil, "asignado_en": gorm.Expr("now()"),
				}).Error
			if err != nil {
				return err
			}
		}
		if input.EsPrincipal {
			return establecerPrincipal(tx, empleadoID, asignacionID)
		}
		return asegurarPrincipal(tx, empleadoID)
	})
	return asignacionID, err
}

func (r *AsignacionRepository) EstablecerPrincipal(ctx context.Context, negocioID, empleadoID, asignacionID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var total int64
		err := tx.Table("asignaciones_empleado_sucursal").
			Where("id = ? AND negocio_id = ? AND empleado_id = ? AND activo = TRUE",
				asignacionID, negocioID, empleadoID).
			Count(&total).Error
		if err != nil {
			return err
		}
		if total != 1 {
			return application.ErrAsignacionNoEncontrada
		}
		return establecerPrincipal(tx, empleadoID, asignacionID)
	})
}

func (r *AsignacionRepository) Finalizar(ctx context.Context, negocioID, empleadoID, asignacionID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		resultado := tx.Table("asignaciones_empleado_sucursal").
			Where("id = ? AND negocio_id = ? AND empleado_id = ? AND activo = TRUE",
				asignacionID, negocioID, empleadoID).
			Updates(map[string]any{
				"activo": false, "es_principal": false, "finalizado_en": gorm.Expr("now()"),
			})
		if resultado.Error != nil {
			return resultado.Error
		}
		if resultado.RowsAffected == 0 {
			return application.ErrAsignacionNoEncontrada
		}
		// Si se retiró la principal, otra asignación activa ocupa su lugar.
		return asegurarPrincipal(tx, empleadoID)
	})
}

// establecerPrincipal degrada la anterior y promueve la indicada dentro de la misma transacción.
func establecerPrincipal(tx *gorm.DB, empleadoID, asignacionID uuid.UUID) error {
	err := tx.Table("asignaciones_empleado_sucursal").
		Where("empleado_id = ? AND id <> ?", empleadoID, asignacionID).
		Update("es_principal", false).Error
	if err != nil {
		return err
	}
	return tx.Table("asignaciones_empleado_sucursal").
		Where("id = ?", asignacionID).Update("es_principal", true).Error
}

// asegurarPrincipal promueve la asignación activa más antigua cuando el empleado se quedó sin principal.
func asegurarPrincipal(tx *gorm.DB, empleadoID uuid.UUID) error {
	var total int64
	err := tx.Table("asignaciones_empleado_sucursal").
		Where("empleado_id = ? AND activo = TRUE AND es_principal = TRUE", empleadoID).
		Count(&total).Error
	if err != nil || total > 0 {
		return err
	}
	// `Pluck` sobre uuid.UUID falla al escanear: se lee como texto y se convierte explícitamente.
	var candidata string
	err = tx.Table("asignaciones_empleado_sucursal").
		Where("empleado_id = ? AND activo = TRUE", empleadoID).
		Order("asignado_en ASC, id ASC").Limit(1).Pluck("id::text", &candidata).Error
	if err != nil || candidata == "" {
		return err
	}
	return tx.Table("asignaciones_empleado_sucursal").
		Where("id = ?", candidata).Update("es_principal", true).Error
}

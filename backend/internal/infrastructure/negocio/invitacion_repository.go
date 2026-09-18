package negocio

import (
	"context"
	"errors"
	"time"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InvitacionRepository struct {
	db *gorm.DB
}

func NewInvitacionRepository(db *gorm.DB) *InvitacionRepository {
	return &InvitacionRepository{db: db}
}

func (r *InvitacionRepository) ObtenerContextoNegocio(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.ContextoNegocioSucursal, error) {
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

func (r *InvitacionRepository) PermisosEfectivos(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]string, error) {
	return permisosEfectivos(ctx, r.db, usuarioID, negocioID)
}

func (r *InvitacionRepository) ObtenerEmpleado(ctx context.Context, negocioID, empleadoID uuid.UUID) (domain.EmpleadoDetalle, error) {
	return NewEmpleadoRepository(r.db).Obtener(ctx, negocioID, empleadoID)
}

func (r *InvitacionRepository) RolPerteneceANegocio(ctx context.Context, negocioID, rolID uuid.UUID) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("roles").
		Where("id = ? AND negocio_id = ? AND activo = TRUE", rolID, negocioID).Count(&total).Error
	return total == 1, err
}

// MarcarExpiradas normaliza el estado antes de listar, para no mostrar como pendiente lo ya vencido.
func (r *InvitacionRepository) MarcarExpiradas(ctx context.Context, negocioID uuid.UUID) error {
	return r.db.WithContext(ctx).Exec(
		`UPDATE invitaciones_negocio SET estado = ?
		 WHERE negocio_id = ? AND estado = ? AND expira_en < now()`,
		domain.EstadoInvitacionExpirada, negocioID, domain.EstadoInvitacionPendiente).Error
}

func (r *InvitacionRepository) Listar(ctx context.Context, negocioID uuid.UUID, estado string) ([]domain.InvitacionResumen, error) {
	items := make([]domain.InvitacionResumen, 0)
	query := r.db.WithContext(ctx).Table("invitaciones_negocio AS i").
		Select(`i.id, i.negocio_id, i.empleado_id, i.correo, i.rol_predeterminado_id, i.estado,
			i.expira_en, i.aceptado_en, i.creado_en,
			COALESCE(e.nombre || ' ' || e.primer_apellido, '') AS nombre_empleado`).
		Joins("LEFT JOIN empleados e ON e.id = i.empleado_id").
		Where("i.negocio_id = ?", negocioID)
	if estado != "" && estado != "todos" {
		query = query.Where("i.estado = ?", estado)
	}
	err := query.Order("i.creado_en DESC, i.id ASC").Scan(&items).Error
	return items, err
}

func (r *InvitacionRepository) Crear(ctx context.Context, invitacion domain.InvitacionNegocio) (domain.InvitacionResumen, error) {
	if err := r.db.WithContext(ctx).Create(&invitacion).Error; err != nil {
		return domain.InvitacionResumen{}, err
	}
	items, err := r.Listar(ctx, invitacion.NegocioID, "")
	if err != nil {
		return domain.InvitacionResumen{}, err
	}
	for _, item := range items {
		if item.ID == invitacion.ID {
			return item, nil
		}
	}
	return domain.InvitacionResumen{}, application.ErrInvitacionNoEncontrada
}

func (r *InvitacionRepository) CancelarPendientesDeEmpleado(ctx context.Context, negocioID, empleadoID uuid.UUID) error {
	return r.db.WithContext(ctx).Exec(
		`UPDATE invitaciones_negocio SET estado = ?
		 WHERE negocio_id = ? AND empleado_id = ? AND estado = ?`,
		domain.EstadoInvitacionCancelada, negocioID, empleadoID, domain.EstadoInvitacionPendiente).Error
}

func (r *InvitacionRepository) Cancelar(ctx context.Context, negocioID, invitacionID uuid.UUID) error {
	resultado := r.db.WithContext(ctx).Exec(
		`UPDATE invitaciones_negocio SET estado = ?
		 WHERE id = ? AND negocio_id = ? AND estado = ?`,
		domain.EstadoInvitacionCancelada, invitacionID, negocioID, domain.EstadoInvitacionPendiente)
	if resultado.Error != nil {
		return resultado.Error
	}
	if resultado.RowsAffected == 0 {
		return application.ErrInvitacionNoEncontrada
	}
	return nil
}

func (r *InvitacionRepository) ObtenerPorHash(ctx context.Context, hash string) (domain.InvitacionNegocio, error) {
	var invitacion domain.InvitacionNegocio
	err := r.db.WithContext(ctx).Where("hash_token = ?", hash).Take(&invitacion).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.InvitacionNegocio{}, application.ErrInvitacionNoEncontrada
	}
	return invitacion, err
}

func (r *InvitacionRepository) DatosPublicos(ctx context.Context, invitacionID uuid.UUID) (domain.InvitacionPublica, error) {
	var publica domain.InvitacionPublica
	err := r.db.WithContext(ctx).Table("invitaciones_negocio AS i").
		Select(`i.correo, i.expira_en, n.nombre_comercial AS nombre_negocio,
			COALESCE(e.nombre || ' ' || e.primer_apellido, '') AS nombre_empleado`).
		Joins("JOIN negocios n ON n.id = i.negocio_id").
		Joins("LEFT JOIN empleados e ON e.id = i.empleado_id").
		Where("i.id = ?", invitacionID).
		Take(&publica).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.InvitacionPublica{}, application.ErrInvitacionNoEncontrada
	}
	return publica, err
}

func (r *InvitacionRepository) CorreoDeUsuario(ctx context.Context, usuarioID uuid.UUID) (string, error) {
	var correo string
	err := r.db.WithContext(ctx).Table("usuarios").
		Where("id = ?", usuarioID).Limit(1).Pluck("correo", &correo).Error
	return correo, err
}

func (r *InvitacionRepository) ExisteCuentaConCorreo(ctx context.Context, correo string) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("usuarios").
		Where("lower(correo) = lower(?)", correo).Count(&total).Error
	return total > 0, err
}

// Aceptar crea o reactiva la membresía, vincula al empleado y asigna el rol predeterminado, todo junto.
func (r *InvitacionRepository) Aceptar(ctx context.Context, invitacion domain.InvitacionNegocio, usuarioID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ahora := time.Now()
		var membresia domain.MembresiaNegocio
		err := tx.Where("negocio_id = ? AND usuario_id = ?", invitacion.NegocioID, usuarioID).
			Take(&membresia).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			membresia = domain.MembresiaNegocio{
				NegocioID: invitacion.NegocioID, UsuarioID: usuarioID,
				TipoMiembro: "miembro", Estado: "activo", SeUnioEn: &ahora,
			}
			if err := tx.Create(&membresia).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		default:
			// Reactivar una membresía suspendida o revocada no debe degradar a un propietario.
			err := tx.Table("membresias_negocio").Where("id = ?", membresia.ID).Updates(map[string]any{
				"estado": "activo", "suspendido_en": nil, "revocado_en": nil,
				"se_unio_en": gorm.Expr("COALESCE(se_unio_en, now())"), "actualizado_en": gorm.Expr("now()"),
			}).Error
			if err != nil {
				return err
			}
		}

		if invitacion.EmpleadoID != nil {
			err := tx.Table("empleados").
				Where("id = ? AND negocio_id = ?", *invitacion.EmpleadoID, invitacion.NegocioID).
				Updates(map[string]any{
					"membresia_id": membresia.ID, "estado": domain.EstadoEmpleadoActivo,
					"actualizado_en": gorm.Expr("now()"),
				}).Error
			if err != nil {
				return err
			}
		}

		if invitacion.RolPredeterminadoID != nil {
			err := tx.Exec(`INSERT INTO roles_membresia (id, membresia_negocio_id, rol_id, asignado_por_usuario_id, asignado_en)
				SELECT gen_random_uuid(), ?, r.id, ?, now() FROM roles r
				WHERE r.id = ? AND r.negocio_id = ?
				ON CONFLICT (membresia_negocio_id, rol_id) DO NOTHING`,
				membresia.ID, invitacion.InvitadoPorUsuarioID,
				*invitacion.RolPredeterminadoID, invitacion.NegocioID).Error
			if err != nil {
				return err
			}
		}

		return tx.Table("invitaciones_negocio").Where("id = ?", invitacion.ID).Updates(map[string]any{
			"estado": domain.EstadoInvitacionAceptada, "aceptado_por_usuario_id": usuarioID, "aceptado_en": ahora,
		}).Error
	})
}

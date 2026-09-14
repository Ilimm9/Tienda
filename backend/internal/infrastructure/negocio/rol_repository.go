package negocio

import (
	"context"
	"errors"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RolRepository struct {
	db *gorm.DB
}

func NewRolRepository(db *gorm.DB) *RolRepository {
	return &RolRepository{db: db}
}

func (r *RolRepository) ObtenerContextoNegocio(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.ContextoNegocioSucursal, error) {
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

func (r *RolRepository) PermisosEfectivos(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]string, error) {
	return permisosEfectivos(ctx, r.db, usuarioID, negocioID)
}

// permisosEfectivos recorre membresía activa -> roles activos -> permisos en una sola consulta.
// Vive aquí para que los repositories del dominio compartan la cadena de autorización sin duplicar SQL.
func permisosEfectivos(ctx context.Context, db *gorm.DB, usuarioID, negocioID uuid.UUID) ([]string, error) {
	codigos := make([]string, 0)
	err := db.WithContext(ctx).Table("membresias_negocio AS m").
		Distinct("p.codigo").
		Joins("JOIN roles_membresia rm ON rm.membresia_negocio_id = m.id").
		Joins("JOIN roles r ON r.id = rm.rol_id AND r.activo = TRUE AND r.negocio_id = m.negocio_id").
		Joins("JOIN permisos_rol pr ON pr.rol_id = r.id").
		Joins("JOIN permisos p ON p.id = pr.permiso_id").
		Where("m.usuario_id = ? AND m.negocio_id = ? AND m.estado = 'activo'", usuarioID, negocioID).
		Pluck("p.codigo", &codigos).Error
	return codigos, err
}

func (r *RolRepository) ListarPermisos(ctx context.Context) ([]domain.Permiso, error) {
	permisos := make([]domain.Permiso, 0)
	err := r.db.WithContext(ctx).Order("codigo_modulo ASC, codigo ASC").Find(&permisos).Error
	return permisos, err
}

func (r *RolRepository) ListarRoles(ctx context.Context, negocioID uuid.UUID, incluirInactivos bool) ([]domain.RolResumen, error) {
	items := make([]domain.RolResumen, 0)
	query := r.db.WithContext(ctx).Table("roles AS r").
		Select(`r.id, r.negocio_id, r.codigo, r.nombre, r.descripcion, r.es_rol_sistema, r.activo,
			r.creado_en, r.actualizado_en,
			(SELECT count(*) FROM permisos_rol pr WHERE pr.rol_id = r.id) AS total_permisos,
			(SELECT count(*) FROM roles_membresia rm WHERE rm.rol_id = r.id) AS total_miembros`).
		Where("r.negocio_id = ?", negocioID)
	if !incluirInactivos {
		query = query.Where("r.activo = TRUE")
	}
	err := query.Order("r.es_rol_sistema DESC, r.nombre ASC, r.id ASC").Scan(&items).Error
	return items, err
}

func (r *RolRepository) ObtenerRol(ctx context.Context, negocioID, rolID uuid.UUID) (domain.RolDetalle, error) {
	var detalle domain.RolDetalle
	err := r.db.WithContext(ctx).Table("roles AS r").
		Select(`r.id, r.negocio_id, r.codigo, r.nombre, r.descripcion, r.es_rol_sistema, r.activo,
			r.creado_en, r.actualizado_en,
			(SELECT count(*) FROM roles_membresia rm WHERE rm.rol_id = r.id) AS total_miembros`).
		Where("r.id = ? AND r.negocio_id = ?", rolID, negocioID).
		Take(&detalle).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.RolDetalle{}, application.ErrRolNoEncontrado
	}
	if err != nil {
		return domain.RolDetalle{}, err
	}
	permisos := make([]uuid.UUID, 0)
	err = r.db.WithContext(ctx).Table("permisos_rol").
		Where("rol_id = ?", rolID).Pluck("permiso_id", &permisos).Error
	detalle.Permisos = permisos
	return detalle, err
}

func (r *RolRepository) ExisteCodigoRol(ctx context.Context, negocioID uuid.UUID, codigo string) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("roles").
		Where("negocio_id = ? AND lower(codigo) = lower(?)", negocioID, codigo).
		Count(&total).Error
	return total > 0, err
}

func (r *RolRepository) CrearRol(ctx context.Context, negocioID, creadoPor uuid.UUID, input domain.CrearRolInput) (uuid.UUID, error) {
	rol := domain.Rol{
		NegocioID: negocioID, Codigo: input.Codigo, Nombre: input.Nombre,
		Descripcion: input.Descripcion, EsRolSistema: false, Activo: true,
		CreadoPorUsuarioID: &creadoPor,
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&rol).Error; err != nil {
			return err
		}
		return reemplazarPermisosDeRol(tx, rol.ID, input.Permisos)
	})
	return rol.ID, err
}

func (r *RolRepository) ActualizarRol(ctx context.Context, negocioID, rolID uuid.UUID, input domain.ActualizarRolInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cambios := map[string]any{}
		if input.Nombre.Set && input.Nombre.Value != nil {
			cambios["nombre"] = *input.Nombre.Value
		}
		if input.Descripcion.Set {
			cambios["descripcion"] = input.Descripcion.Value
		}
		if input.Activo.Set && input.Activo.Value != nil {
			cambios["activo"] = *input.Activo.Value
		}
		if len(cambios) > 0 {
			cambios["actualizado_en"] = gorm.Expr("now()")
			err := tx.Table("roles").Where("id = ? AND negocio_id = ?", rolID, negocioID).Updates(cambios).Error
			if err != nil {
				return err
			}
		}
		if input.Permisos.Set && input.Permisos.Value != nil {
			return reemplazarPermisosDeRol(tx, rolID, *input.Permisos.Value)
		}
		return nil
	})
}

func (r *RolRepository) EliminarRol(ctx context.Context, negocioID, rolID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM permisos_rol WHERE rol_id = ?`, rolID).Error; err != nil {
			return err
		}
		return tx.Exec(`DELETE FROM roles WHERE id = ? AND negocio_id = ? AND es_rol_sistema = false`, rolID, negocioID).Error
	})
}

func (r *RolRepository) ContarMiembrosConRol(ctx context.Context, negocioID, rolID uuid.UUID) (int, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("roles_membresia AS rm").
		Joins("JOIN roles r ON r.id = rm.rol_id").
		Where("rm.rol_id = ? AND r.negocio_id = ?", rolID, negocioID).
		Count(&total).Error
	return int(total), err
}

func (r *RolRepository) ListarMiembros(ctx context.Context, negocioID uuid.UUID) ([]domain.MiembroRoles, error) {
	type fila struct {
		MembresiaID uuid.UUID
		UsuarioID   uuid.UUID
		Correo      string
		TipoMiembro string
		Estado      string
		RolID       *uuid.UUID
	}
	filas := make([]fila, 0)
	err := r.db.WithContext(ctx).Table("membresias_negocio AS m").
		Select("m.id AS membresia_id, m.usuario_id, u.correo, m.tipo_miembro, m.estado, rm.rol_id").
		Joins("JOIN usuarios u ON u.id = m.usuario_id").
		Joins("LEFT JOIN roles_membresia rm ON rm.membresia_negocio_id = m.id").
		Where("m.negocio_id = ?", negocioID).
		Order("m.tipo_miembro ASC, u.correo ASC, m.id ASC").
		Scan(&filas).Error
	if err != nil {
		return nil, err
	}
	items := make([]domain.MiembroRoles, 0)
	indices := make(map[uuid.UUID]int)
	for _, actual := range filas {
		indice, existe := indices[actual.MembresiaID]
		if !existe {
			indice = len(items)
			indices[actual.MembresiaID] = indice
			items = append(items, domain.MiembroRoles{
				MembresiaID: actual.MembresiaID, UsuarioID: actual.UsuarioID, Correo: actual.Correo,
				TipoMiembro: actual.TipoMiembro, Estado: actual.Estado, Roles: make([]uuid.UUID, 0),
			})
		}
		if actual.RolID != nil {
			items[indice].Roles = append(items[indice].Roles, *actual.RolID)
		}
	}
	return items, nil
}

func (r *RolRepository) MembresiaPerteneceANegocio(ctx context.Context, negocioID, membresiaID uuid.UUID) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("membresias_negocio").
		Where("id = ? AND negocio_id = ?", membresiaID, negocioID).Count(&total).Error
	return total == 1, err
}

// QuedaPropietarioConRolSistema simula la asignación para no dejar al negocio sin propietario efectivo.
func (r *RolRepository) QuedaPropietarioConRolSistema(ctx context.Context, negocioID, membresiaID uuid.UUID, roles []uuid.UUID) (bool, error) {
	// `Pluck` sobre uuid.UUID falla al escanear: se lee como texto y se convierte explícitamente.
	var rolSistemaTexto string
	err := r.db.WithContext(ctx).Table("roles").
		Where("negocio_id = ? AND es_rol_sistema = TRUE", negocioID).
		Limit(1).Pluck("id::text", &rolSistemaTexto).Error
	if err != nil || rolSistemaTexto == "" {
		return true, err
	}
	rolSistemaID, err := uuid.Parse(rolSistemaTexto)
	if err != nil {
		return true, err
	}
	// Si la membresía conserva el rol de sistema, la invariante no puede romperse aquí.
	for _, rol := range roles {
		if rol == rolSistemaID {
			return true, nil
		}
	}
	var otros int64
	err = r.db.WithContext(ctx).Table("roles_membresia AS rm").
		Joins("JOIN membresias_negocio m ON m.id = rm.membresia_negocio_id").
		Where("rm.rol_id = ? AND m.negocio_id = ? AND m.estado = 'activo' AND m.id <> ?",
			rolSistemaID, negocioID, membresiaID).
		Count(&otros).Error
	return otros > 0, err
}

func (r *RolRepository) ReemplazarRolesDeMembresia(ctx context.Context, negocioID, membresiaID uuid.UUID, roles []uuid.UUID, asignadoPor uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM roles_membresia WHERE membresia_negocio_id = ?`, membresiaID).Error; err != nil {
			return err
		}
		for _, rolID := range roles {
			// Un rol de otro negocio nunca puede asignarse: el INSERT filtra por negocio.
			err := tx.Exec(`INSERT INTO roles_membresia (id, membresia_negocio_id, rol_id, asignado_por_usuario_id, asignado_en)
				SELECT gen_random_uuid(), ?, r.id, ?, now() FROM roles r
				WHERE r.id = ? AND r.negocio_id = ?
				ON CONFLICT (membresia_negocio_id, rol_id) DO NOTHING`,
				membresiaID, asignadoPor, rolID, negocioID).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func reemplazarPermisosDeRol(tx *gorm.DB, rolID uuid.UUID, permisos []uuid.UUID) error {
	if err := tx.Exec(`DELETE FROM permisos_rol WHERE rol_id = ?`, rolID).Error; err != nil {
		return err
	}
	for _, permisoID := range permisos {
		err := tx.Exec(`INSERT INTO permisos_rol (id, rol_id, permiso_id, creado_en)
			SELECT gen_random_uuid(), ?, p.id, now() FROM permisos p WHERE p.id = ?
			ON CONFLICT (rol_id, permiso_id) DO NOTHING`, rolID, permisoID).Error
		if err != nil {
			return err
		}
	}
	return nil
}

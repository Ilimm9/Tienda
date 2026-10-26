package plataforma

import (
	"context"
	"errors"
	"strings"
	"time"

	application "tienda/backend/internal/application/plataforma"
	cuentadomain "tienda/backend/internal/domain/cuenta"
	negociodomain "tienda/backend/internal/domain/negocio"
	domain "tienda/backend/internal/domain/plataforma"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) RolesUsuario(ctx context.Context, usuarioID uuid.UUID) ([]string, error) {
	roles := make([]string, 0)
	err := r.db.WithContext(ctx).Table("roles_plataforma AS rp").
		Joins("JOIN roles_usuario_plataforma rup ON rup.rol_id = rp.id").
		Where("rup.usuario_id = ? AND rp.activo = TRUE", usuarioID).
		Order("rp.codigo ASC").Pluck("rp.codigo", &roles).Error
	return roles, err
}

func (r *Repository) Resumen(ctx context.Context, desde, hasta time.Time) (domain.Resumen, error) {
	var result domain.Resumen
	if err := r.db.WithContext(ctx).Table("negocios").Count(&result.NegociosTotal).Error; err != nil {
		return result, err
	}
	if err := r.db.WithContext(ctx).Table("negocios").Where("estado = 'activo'").Count(&result.NegociosActivos).Error; err != nil {
		return result, err
	}
	if err := r.db.WithContext(ctx).Table("negocios").Where("estado = 'archivado'").Count(&result.NegociosArchivados).Error; err != nil {
		return result, err
	}
	if err := r.db.WithContext(ctx).Table("sucursales").Where("activo = TRUE AND eliminado_en IS NULL").Count(&result.SucursalesActivas).Error; err != nil {
		return result, err
	}
	if err := r.db.WithContext(ctx).Table("usuarios").Count(&result.CuentasRegistradas).Error; err != nil {
		return result, err
	}
	now := time.Now().UTC()
	if err := r.db.WithContext(ctx).Table("actividad_usuario_diaria").Where("fecha >= ?", now.AddDate(0, 0, -6)).Distinct("usuario_id").Count(&result.UsuariosActivos7d).Error; err != nil {
		return result, err
	}
	if err := r.db.WithContext(ctx).Table("actividad_usuario_diaria").Where("fecha >= ?", now.AddDate(0, 0, -29)).Distinct("usuario_id").Count(&result.UsuariosActivos30d).Error; err != nil {
		return result, err
	}
	result.Actividad = make([]domain.ActividadDiaria, 0)
	err := r.db.WithContext(ctx).Table("actividad_usuario_diaria").
		Select("fecha, count(DISTINCT usuario_id) AS usuarios, sum(total_solicitudes) AS solicitudes").
		Where("fecha BETWEEN ? AND ?", desde, hasta).Group("fecha").Order("fecha ASC").Scan(&result.Actividad).Error
	return result, err
}

func (r *Repository) ListarNegocios(ctx context.Context, query, estado string, page, pageSize int) (domain.PaginaNegocios, error) {
	base := r.db.WithContext(ctx).Table("negocios AS n")
	if estado != "" {
		base = base.Where("n.estado = ?", estado)
	}
	if query != "" {
		term := "%" + strings.ToLower(query) + "%"
		base = base.Where("lower(n.nombre_comercial) LIKE ? OR lower(n.slug) LIKE ? OR lower(COALESCE(n.rfc, '')) LIKE ?", term, term, term)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return domain.PaginaNegocios{}, err
	}
	items := make([]domain.NegocioResumen, 0)
	err := base.Select(`n.id, n.slug, n.nombre_comercial, n.rfc, n.estado, n.creado_en,
		(SELECT count(*) FROM sucursales s WHERE s.negocio_id = n.id AND s.activo = TRUE AND s.eliminado_en IS NULL) AS total_sucursales,
		(SELECT count(*) FROM empleados e WHERE e.negocio_id = n.id AND e.estado = 'activo') AS total_personal,
		(SELECT count(*) FROM productos p WHERE p.negocio_id = n.id AND p.activo = TRUE) AS total_productos,
		(SELECT count(*) FROM empleados e JOIN membresias_negocio m ON m.id = e.membresia_id AND m.estado = 'activo' JOIN usuarios u ON u.id = m.usuario_id AND u.estado = 'activo' AND u.deshabilitado_en IS NULL WHERE e.negocio_id = n.id AND e.estado = 'activo') AS cuentas_vinculadas,
		(SELECT max(a.ultima_actividad_en) FROM actividad_usuario_diaria a WHERE a.negocio_id = n.id) AS ultima_actividad_en`).
		Order("n.creado_en DESC, n.id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&items).Error
	return domain.PaginaNegocios{Items: items, Total: total, Page: page, PageSize: pageSize}, err
}

func (r *Repository) ObtenerNegocio(ctx context.Context, negocioID uuid.UUID, desde, hasta time.Time) (domain.NegocioDetalle, error) {
	var item domain.NegocioDetalle
	err := r.db.WithContext(ctx).Table("negocios AS n").Select(`n.id, n.slug, n.nombre_comercial, n.rfc, n.estado, n.creado_en,
		n.razon_social, n.correo, n.zona_horaria,
		(SELECT count(*) FROM sucursales s WHERE s.negocio_id = n.id AND s.activo = TRUE AND s.eliminado_en IS NULL) AS total_sucursales,
		(SELECT count(*) FROM empleados e WHERE e.negocio_id = n.id AND e.estado = 'activo') AS total_personal,
		(SELECT count(*) FROM productos p WHERE p.negocio_id = n.id AND p.activo = TRUE) AS total_productos,
		(SELECT count(*) FROM empleados e JOIN membresias_negocio m ON m.id = e.membresia_id AND m.estado = 'activo' JOIN usuarios u ON u.id = m.usuario_id AND u.estado = 'activo' AND u.deshabilitado_en IS NULL WHERE e.negocio_id = n.id AND e.estado = 'activo') AS cuentas_vinculadas,
		(SELECT max(a.ultima_actividad_en) FROM actividad_usuario_diaria a WHERE a.negocio_id = n.id) AS ultima_actividad_en,
		(SELECT count(*) FROM membresias_negocio m WHERE m.negocio_id = n.id AND m.estado = 'activo' AND m.tipo_miembro = 'propietario') AS total_propietarios`).
		Where("n.id = ?", negocioID).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return item, application.ErrNegocioNoEncontrado
	}
	if err != nil {
		return item, err
	}
	item.Sucursales = make([]domain.SucursalResumen, 0)
	err = r.db.WithContext(ctx).Table("sucursales AS s").Select(`s.id, s.codigo, s.nombre, s.es_principal,
		CASE WHEN s.activo = TRUE AND s.eliminado_en IS NULL THEN 'activo' ELSE 'archivado' END AS estado,
		(SELECT count(*) FROM asignaciones_empleado_sucursal a JOIN empleados e ON e.id = a.empleado_id AND e.estado = 'activo' WHERE a.sucursal_id = s.id AND a.activo = TRUE AND a.finalizado_en IS NULL) AS personal_asignado,
		(SELECT count(*) FROM asignaciones_empleado_sucursal a JOIN empleados e ON e.id = a.empleado_id AND e.estado = 'activo' JOIN membresias_negocio m ON m.id = e.membresia_id AND m.estado = 'activo' JOIN usuarios u ON u.id = m.usuario_id AND u.estado = 'activo' AND u.deshabilitado_en IS NULL WHERE a.sucursal_id = s.id AND a.activo = TRUE AND a.finalizado_en IS NULL) AS con_cuenta_activa`).
		Where("s.negocio_id = ?", negocioID).Order("s.es_principal DESC, s.nombre ASC").Scan(&item.Sucursales).Error
	if err != nil {
		return item, err
	}
	item.Actividad = make([]domain.ActividadDiaria, 0)
	err = r.db.WithContext(ctx).Table("actividad_usuario_diaria").
		Select("fecha, count(DISTINCT usuario_id) AS usuarios, sum(total_solicitudes) AS solicitudes").
		Where("negocio_id = ? AND fecha BETWEEN ? AND ?", negocioID, desde, hasta).
		Group("fecha").Order("fecha ASC").Scan(&item.Actividad).Error
	return item, err
}

func (r *Repository) ListarUsuarios(ctx context.Context, negocioID uuid.UUID) ([]domain.UsuarioResumen, error) {
	type row struct {
		ID                                                               uuid.UUID
		Nombre, Correo, Estado, TipoMiembro, RolesTexto, SucursalesTexto string
		UltimoInicioSesionEn                                             *time.Time
	}
	rows := make([]row, 0)
	err := r.db.WithContext(ctx).Table("membresias_negocio AS m").
		Select(`u.id, trim(concat_ws(' ', p.nombres, p.primer_apellido, p.segundo_apellido)) AS nombre,
			u.correo, u.estado, m.tipo_miembro, u.ultimo_inicio_sesion_en,
			COALESCE(string_agg(DISTINCT r.nombre, E'\x1f'), '') AS roles_texto,
			COALESCE(string_agg(DISTINCT s.nombre, E'\x1f'), '') AS sucursales_texto`).
		Joins("JOIN usuarios u ON u.id = m.usuario_id").
		Joins("LEFT JOIN perfil_usuarios p ON p.usuario_id = u.id").
		Joins("LEFT JOIN roles_membresia rm ON rm.membresia_negocio_id = m.id").
		Joins("LEFT JOIN roles r ON r.id = rm.rol_id AND r.activo = TRUE").
		Joins("LEFT JOIN empleados e ON e.membresia_id = m.id AND e.estado = 'activo'").
		Joins("LEFT JOIN asignaciones_empleado_sucursal a ON a.empleado_id = e.id AND a.activo = TRUE AND a.finalizado_en IS NULL").
		Joins("LEFT JOIN sucursales s ON s.id = a.sucursal_id").
		Where("m.negocio_id = ? AND m.estado = 'activo'", negocioID).
		Group("u.id, p.nombres, p.primer_apellido, p.segundo_apellido, m.tipo_miembro").
		Order("nombre ASC, u.correo ASC").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]domain.UsuarioResumen, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.UsuarioResumen{ID: row.ID, Nombre: row.Nombre, Correo: row.Correo,
			Estado: row.Estado, TipoMiembro: row.TipoMiembro, Roles: splitAggregate(row.RolesTexto),
			Sucursales: splitAggregate(row.SucursalesTexto), UltimoInicioSesionEn: row.UltimoInicioSesionEn})
	}
	return items, nil
}

func (r *Repository) ListarCuentas(ctx context.Context, query, estado string, page, pageSize int) (domain.PaginaCuentas, error) {
	base := r.db.WithContext(ctx).Table("usuarios AS u").
		Joins("LEFT JOIN perfil_usuarios p ON p.usuario_id = u.id").
		Where("NOT EXISTS (SELECT 1 FROM roles_usuario_plataforma rup WHERE rup.usuario_id = u.id)")
	switch estado {
	case "activo":
		base = base.Where("u.estado = 'activo' AND u.correo_verificado_en IS NOT NULL AND u.deshabilitado_en IS NULL")
	case "sin_verificar":
		base = base.Where("u.correo_verificado_en IS NULL")
	case "deshabilitado":
		base = base.Where("u.deshabilitado_en IS NOT NULL OR u.estado <> 'activo'")
	}
	if query != "" {
		term := "%" + strings.ToLower(query) + "%"
		base = base.Where("lower(u.correo) LIKE ? OR lower(concat_ws(' ', p.nombres, p.primer_apellido, p.segundo_apellido)) LIKE ?", term, term)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return domain.PaginaCuentas{}, err
	}
	type row struct {
		domain.CuentaResumen
		NegociosTexto string
	}
	rows := make([]row, 0)
	err := base.Select(`u.id, trim(concat_ws(' ', p.nombres, p.primer_apellido, p.segundo_apellido)) AS nombre,
		u.correo, u.estado, u.correo_verificado_en IS NOT NULL AS correo_verificado, u.creado_en, u.ultimo_inicio_sesion_en,
		(SELECT count(*) FROM membresias_negocio m WHERE m.usuario_id = u.id AND m.estado = 'activo') AS total_negocios,
		(SELECT COALESCE(string_agg(n.nombre_comercial, E'\x1f' ORDER BY n.nombre_comercial), '') FROM membresias_negocio m JOIN negocios n ON n.id = m.negocio_id WHERE m.usuario_id = u.id AND m.estado = 'activo') AS negocios_texto`).
		Order("u.creado_en DESC, u.id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&rows).Error
	if err != nil {
		return domain.PaginaCuentas{}, err
	}
	items := make([]domain.CuentaResumen, 0, len(rows))
	for _, row := range rows {
		item := row.CuentaResumen
		item.Negocios = splitAggregate(row.NegociosTexto)
		items = append(items, item)
	}
	return domain.PaginaCuentas{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func splitAggregate(value string) []string {
	if value == "" {
		return []string{}
	}
	return strings.Split(value, "\x1f")
}

func (r *Repository) RegistrarActividad(ctx context.Context, usuarioID, negocioID uuid.UUID, at time.Time) error {
	fecha := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, time.UTC)
	item := domain.ActividadUsuarioDiaria{Fecha: fecha, UsuarioID: usuarioID, NegocioID: negocioID,
		PrimeraActividadEn: at, UltimaActividadEn: at, TotalSolicitudes: 1}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "fecha"}, {Name: "usuario_id"}, {Name: "negocio_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"ultima_actividad_en": at,
			"total_solicitudes":   gorm.Expr("actividad_usuario_diaria.total_solicitudes + 1"),
		}),
	}).Create(&item).Error
}

func (r *Repository) LimpiarActividad(ctx context.Context, antes time.Time) error {
	return r.db.WithContext(ctx).Where("fecha < ?", antes).Delete(&domain.ActividadUsuarioDiaria{}).Error
}

func (r *Repository) AsignarAdminLectura(ctx context.Context, correo, motivo string) (application.AdminAsignado, error) {
	result := application.AdminAsignado{}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user, err := cuentaPorCorreo(tx, correo)
		if err != nil {
			return err
		}
		if user.Estado != "activo" || user.CorreoVerificadoEn == nil || user.DeshabilitadoEn != nil {
			return application.ErrCuentaNoDisponible
		}
		var memberships int64
		if err := tx.Model(&negociodomain.MembresiaNegocio{}).Where("usuario_id = ?", user.ID).Count(&memberships).Error; err != nil {
			return err
		}
		if memberships > 0 {
			return application.ErrCuentaConNegocios
		}
		var role domain.RolPlataforma
		if err := tx.Where("codigo = ? AND activo = TRUE", domain.RolAdminLectura).Take(&role).Error; err != nil {
			return err
		}
		assignment := domain.RolUsuarioPlataforma{UsuarioID: user.ID, RolID: role.ID, Origen: "cli"}
		created := tx.Where("usuario_id = ? AND rol_id = ?", user.ID, role.ID).FirstOrCreate(&assignment)
		if created.Error != nil {
			return created.Error
		}
		if created.RowsAffected > 0 {
			if err := tx.Create(&domain.AuditoriaPlataforma{UsuarioID: user.ID, RolID: &role.ID, Accion: "rol_asignado", Motivo: motivo, Origen: "cli"}).Error; err != nil {
				return err
			}
		}
		result = application.AdminAsignado{UsuarioID: user.ID, Correo: user.Correo, Asignado: created.RowsAffected > 0}
		return nil
	})
	return result, err
}

func (r *Repository) RetirarAdminLectura(ctx context.Context, correo, motivo string) (application.AdminAsignado, error) {
	result := application.AdminAsignado{}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user, err := cuentaPorCorreo(tx, correo)
		if err != nil {
			return err
		}
		var role domain.RolPlataforma
		if err := tx.Where("codigo = ?", domain.RolAdminLectura).Take(&role).Error; err != nil {
			return err
		}
		deleted := tx.Where("usuario_id = ? AND rol_id = ?", user.ID, role.ID).Delete(&domain.RolUsuarioPlataforma{})
		if deleted.Error != nil {
			return deleted.Error
		}
		if deleted.RowsAffected > 0 {
			if err := tx.Create(&domain.AuditoriaPlataforma{UsuarioID: user.ID, RolID: &role.ID, Accion: "rol_retirado", Motivo: motivo, Origen: "cli"}).Error; err != nil {
				return err
			}
		}
		result = application.AdminAsignado{UsuarioID: user.ID, Correo: user.Correo, Asignado: deleted.RowsAffected > 0}
		return nil
	})
	return result, err
}

func (r *Repository) ListarAdmins(ctx context.Context) ([]application.AdminAsignado, error) {
	items := make([]application.AdminAsignado, 0)
	err := r.db.WithContext(ctx).Table("roles_usuario_plataforma AS rup").
		Select("u.id AS usuario_id, u.correo, TRUE AS asignado").
		Joins("JOIN usuarios u ON u.id = rup.usuario_id").Joins("JOIN roles_plataforma rp ON rp.id = rup.rol_id").
		Where("rp.codigo = ?", domain.RolAdminLectura).Order("u.correo ASC").Scan(&items).Error
	return items, err
}

func cuentaPorCorreo(db *gorm.DB, correo string) (cuentadomain.Usuario, error) {
	var user cuentadomain.Usuario
	err := db.Where("lower(correo) = lower(?)", correo).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user, application.ErrCuentaNoEncontrada
	}
	return user, err
}

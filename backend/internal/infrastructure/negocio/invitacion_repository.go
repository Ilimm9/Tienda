package negocio

import (
	"context"
	"errors"
	"strings"
	"time"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// CodigoRolActivo devuelve el código del rol solo si está activo y pertenece al negocio.
func (r *InvitacionRepository) CodigoRolActivo(ctx context.Context, negocioID, rolID uuid.UUID) (string, bool, error) {
	return codigoRolActivo(r.db.WithContext(ctx), negocioID, rolID)
}

func codigoRolActivo(db *gorm.DB, negocioID, rolID uuid.UUID) (string, bool, error) {
	var codigos []string
	err := db.Table("roles").
		Where("id = ? AND negocio_id = ? AND activo = TRUE", rolID, negocioID).
		Limit(1).Pluck("codigo", &codigos).Error
	if err != nil || len(codigos) == 0 {
		return "", false, err
	}
	return codigos[0], true, nil
}

func (r *InvitacionRepository) SucursalActivaDelNegocio(ctx context.Context, negocioID, sucursalID uuid.UUID) (bool, error) {
	return sucursalActiva(r.db.WithContext(ctx), negocioID, sucursalID)
}

func sucursalActiva(db *gorm.DB, negocioID, sucursalID uuid.UUID) (bool, error) {
	var total int64
	err := db.Table("sucursales").
		Where("id = ? AND negocio_id = ? AND activo = TRUE AND eliminado_en IS NULL", sucursalID, negocioID).
		Count(&total).Error
	return total == 1, err
}

// MarcarExpiradas normaliza el estado antes de listar, para no mostrar como pendiente lo ya vencido.
func (r *InvitacionRepository) MarcarExpiradas(ctx context.Context, negocioID uuid.UUID) error {
	return r.db.WithContext(ctx).Exec(
		`UPDATE invitaciones_negocio SET estado = ?
		 WHERE negocio_id = ? AND estado = ? AND expira_en < ?`,
		domain.EstadoInvitacionExpirada, negocioID, domain.EstadoInvitacionPendiente, time.Now().UTC()).Error
}

// consultaResumen calcula el estado efectivo con el reloj del servidor aunque la fila diga pendiente.
func consultaResumen(db *gorm.DB, ahora time.Time) *gorm.DB {
	return db.Table("invitaciones_negocio AS i").
		Select(`i.id, i.negocio_id, i.empleado_id, i.correo, i.sucursal_id, i.rol_predeterminado_id,
			CASE WHEN i.estado = 'pendiente' AND i.expira_en < ? THEN 'expirada' ELSE i.estado END AS estado,
			i.expira_en, i.aceptado_en, i.creado_en,
			COALESCE(e.nombre || ' ' || e.primer_apellido, '') AS nombre_empleado,
			COALESCE(s.nombre, '') AS nombre_sucursal, COALESCE(ro.nombre, '') AS nombre_rol`, ahora).
		Joins("LEFT JOIN empleados e ON e.id = i.empleado_id").
		Joins("LEFT JOIN sucursales s ON s.id = i.sucursal_id").
		Joins("LEFT JOIN roles ro ON ro.id = i.rol_predeterminado_id")
}

func (r *InvitacionRepository) Listar(ctx context.Context, negocioID uuid.UUID, filtro domain.FiltroInvitaciones) ([]domain.InvitacionResumen, error) {
	items := make([]domain.InvitacionResumen, 0)
	ahora := time.Now().UTC()
	query := consultaResumen(r.db.WithContext(ctx), ahora).Where("i.negocio_id = ?", negocioID)
	switch {
	case filtro.SinAceptar:
		query = query.Where("i.estado IN ?", []string{domain.EstadoInvitacionPendiente, domain.EstadoInvitacionExpirada})
	case filtro.Estado != "" && filtro.Estado != "todos":
		query = query.Where("i.estado = ?", filtro.Estado)
	}
	if filtro.SucursalID != nil {
		query = query.Where("i.sucursal_id = ?", *filtro.SucursalID)
	}
	err := query.Order("i.creado_en DESC, i.id ASC").Scan(&items).Error
	return items, err
}

func obtenerResumen(tx *gorm.DB, invitacionID uuid.UUID) (domain.InvitacionResumen, error) {
	var resumen domain.InvitacionResumen
	err := consultaResumen(tx, time.Now().UTC()).Where("i.id = ?", invitacionID).Take(&resumen).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.InvitacionResumen{}, application.ErrInvitacionNoEncontrada
	}
	return resumen, err
}

// bloquearEmpleado serializa emisión, reenvío y aceptación del mismo empleado.
func bloquearEmpleado(tx *gorm.DB, negocioID, empleadoID uuid.UUID) (domain.Empleado, error) {
	var empleado domain.Empleado
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND negocio_id = ?", empleadoID, negocioID).Take(&empleado).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Empleado{}, application.ErrEmpleadoNoEncontrado
	}
	return empleado, err
}

// retirarPendientes deja libre el índice único de pendiente por empleado antes de insertar otra.
func retirarPendientes(tx *gorm.DB, negocioID, empleadoID uuid.UUID, ahora time.Time) error {
	err := tx.Exec(`UPDATE invitaciones_negocio SET estado = ?
		WHERE negocio_id = ? AND empleado_id = ? AND estado = ? AND expira_en < ?`,
		domain.EstadoInvitacionExpirada, negocioID, empleadoID, domain.EstadoInvitacionPendiente, ahora).Error
	if err != nil {
		return err
	}
	return tx.Exec(`UPDATE invitaciones_negocio SET estado = ?
		WHERE negocio_id = ? AND empleado_id = ? AND estado = ?`,
		domain.EstadoInvitacionCancelada, negocioID, empleadoID, domain.EstadoInvitacionPendiente).Error
}

// Emitir cancela pendientes previas del empleado e inserta la nueva en una sola transacción.
func (r *InvitacionRepository) Emitir(ctx context.Context, invitacion domain.InvitacionNegocio) (domain.InvitacionResumen, error) {
	var resumen domain.InvitacionResumen
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if invitacion.EmpleadoID == nil {
			return application.ErrEmpleadoNoEncontrado
		}
		empleado, err := bloquearEmpleado(tx, invitacion.NegocioID, *invitacion.EmpleadoID)
		if err != nil {
			return err
		}
		if empleado.MembresiaID != nil {
			return application.ErrInvitacionYaVinculado
		}
		if err := retirarPendientes(tx, invitacion.NegocioID, empleado.ID, time.Now().UTC()); err != nil {
			return err
		}
		if err := tx.Create(&invitacion).Error; err != nil {
			return err
		}
		resumen, err = obtenerResumen(tx, invitacion.ID)
		return err
	})
	return resumen, err
}

// Reemitir sustituye la última invitación del empleado; si llega correo, lo corrige en el empleado.
//
// La fila anterior se conserva como historial con su correo original.
func (r *InvitacionRepository) Reemitir(ctx context.Context, negocioID, invitacionID, usuarioID uuid.UUID, correo *string, hash string, expiraEn time.Time) (domain.InvitacionResumen, error) {
	var resumen domain.InvitacionResumen
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var referencia domain.InvitacionNegocio
		err := tx.Where("id = ? AND negocio_id = ?", invitacionID, negocioID).Take(&referencia).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return application.ErrInvitacionNoEncontrada
		}
		if err != nil {
			return err
		}
		if referencia.EmpleadoID == nil {
			return application.ErrInvitacionSinSucursal
		}
		// El empleado se bloquea antes que la invitación, igual que en la aceptación.
		empleado, err := bloquearEmpleado(tx, negocioID, *referencia.EmpleadoID)
		if err != nil {
			return err
		}
		var anterior domain.InvitacionNegocio
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", invitacionID).Take(&anterior).Error; err != nil {
			return err
		}
		if anterior.Estado == domain.EstadoInvitacionAceptada {
			return application.ErrInvitacionYaAceptada
		}
		if empleado.MembresiaID != nil {
			return application.ErrInvitacionYaVinculado
		}
		if empleado.Estado == domain.EstadoEmpleadoSuspendido || empleado.Estado == domain.EstadoEmpleadoTerminado {
			return application.ErrInvitacionEmpleadoNoElegible
		}
		var ultimas []string
		if err := tx.Table("invitaciones_negocio").
			Where("negocio_id = ? AND empleado_id = ?", negocioID, empleado.ID).
			Order("creado_en DESC, id DESC").Limit(1).Pluck("id", &ultimas).Error; err != nil {
			return err
		}
		if len(ultimas) == 0 || ultimas[0] != anterior.ID.String() {
			return application.ErrInvitacionReemplazada
		}
		if anterior.SucursalID == nil || anterior.RolPredeterminadoID == nil {
			return application.ErrInvitacionSinSucursal
		}
		if activa, err := sucursalActiva(tx, negocioID, *anterior.SucursalID); err != nil {
			return err
		} else if !activa {
			return application.ErrInvitacionSucursalNoDisponible
		}
		codigo, existe, err := codigoRolActivo(tx, negocioID, *anterior.RolPredeterminadoID)
		if err != nil {
			return err
		}
		if !existe || strings.EqualFold(codigo, domain.CodigoRolPropietario) {
			return application.ErrInvitacionRolNoDisponible
		}

		ahora := time.Now().UTC()
		destinatario := anterior.Correo
		if correo != nil {
			destinatario = *correo
			if err := tx.Table("empleados").Where("id = ?", empleado.ID).
				Updates(map[string]any{"correo": destinatario, "actualizado_en": ahora}).Error; err != nil {
				return err
			}
		}
		if err := retirarPendientes(tx, negocioID, empleado.ID, ahora); err != nil {
			return err
		}
		nueva := domain.InvitacionNegocio{
			NegocioID: negocioID, EmpleadoID: &empleado.ID, SucursalID: anterior.SucursalID,
			Correo: destinatario, RolPredeterminadoID: anterior.RolPredeterminadoID,
			HashToken: hash, Estado: domain.EstadoInvitacionPendiente,
			InvitadoPorUsuarioID: usuarioID, ExpiraEn: expiraEn,
		}
		if err := tx.Create(&nueva).Error; err != nil {
			return err
		}
		resumen, err = obtenerResumen(tx, nueva.ID)
		return err
	})
	return resumen, err
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
			COALESCE(e.nombre || ' ' || e.primer_apellido, '') AS nombre_empleado,
			COALESCE(s.nombre, '') AS nombre_sucursal, COALESCE(ro.nombre, '') AS nombre_rol`).
		Joins("JOIN negocios n ON n.id = i.negocio_id").
		Joins("LEFT JOIN empleados e ON e.id = i.empleado_id").
		Joins("LEFT JOIN sucursales s ON s.id = i.sucursal_id").
		Joins("LEFT JOIN roles ro ON ro.id = i.rol_predeterminado_id").
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

// ExisteCuentaVerificada ignora cuentas pendientes: esas todavía pueden completarse desde el enlace.
func (r *InvitacionRepository) ExisteCuentaVerificada(ctx context.Context, correo string) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("usuarios").
		Where("lower(correo) = lower(?) AND correo_verificado_en IS NOT NULL", correo).Count(&total).Error
	return total > 0, err
}

// Aceptar bloquea empleado e invitación y revalida todo dentro de la transacción: de dos
// aceptaciones simultáneas, o de una aceptación contra un reenvío, solo una produce efectos.
func (r *InvitacionRepository) Aceptar(ctx context.Context, invitacionID, usuarioID uuid.UUID) (domain.InvitacionAceptada, error) {
	var aceptada domain.InvitacionAceptada
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ahora := time.Now().UTC()
		var referencia domain.InvitacionNegocio
		err := tx.Where("id = ?", invitacionID).Take(&referencia).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return application.ErrInvitacionNoEncontrada
		}
		if err != nil {
			return err
		}
		if referencia.EmpleadoID == nil || referencia.SucursalID == nil {
			return application.ErrInvitacionNoVigente
		}
		empleado, err := bloquearEmpleado(tx, referencia.NegocioID, *referencia.EmpleadoID)
		if errors.Is(err, application.ErrEmpleadoNoEncontrado) {
			return application.ErrInvitacionEmpleadoNoElegible
		}
		if err != nil {
			return err
		}
		var invitacion domain.InvitacionNegocio
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", invitacionID).Take(&invitacion).Error; err != nil {
			return err
		}
		if invitacion.Estado == domain.EstadoInvitacionAceptada {
			return application.ErrInvitacionYaAceptada
		}
		if invitacion.Estado != domain.EstadoInvitacionPendiente || !invitacion.ExpiraEn.After(ahora) {
			return application.ErrInvitacionNoVigente
		}
		if empleado.MembresiaID != nil {
			return application.ErrInvitacionYaVinculado
		}
		if empleado.Estado == domain.EstadoEmpleadoSuspendido || empleado.Estado == domain.EstadoEmpleadoTerminado {
			return application.ErrInvitacionEmpleadoNoElegible
		}
		var estadoNegocio string
		if err := tx.Table("negocios").Where("id = ?", invitacion.NegocioID).
			Limit(1).Pluck("estado", &estadoNegocio).Error; err != nil {
			return err
		}
		if estadoNegocio != "activo" {
			return application.ErrEstadoNegocio
		}
		if activa, err := sucursalActiva(tx, invitacion.NegocioID, *invitacion.SucursalID); err != nil {
			return err
		} else if !activa {
			return application.ErrInvitacionSucursalNoDisponible
		}
		if invitacion.RolPredeterminadoID != nil {
			codigo, existe, err := codigoRolActivo(tx, invitacion.NegocioID, *invitacion.RolPredeterminadoID)
			if err != nil {
				return err
			}
			if !existe || strings.EqualFold(codigo, domain.CodigoRolPropietario) {
				return application.ErrInvitacionRolNoDisponible
			}
		}

		membresia, err := membresiaParaAceptar(tx, invitacion.NegocioID, usuarioID, ahora)
		if err != nil {
			return err
		}
		// La membresía ya puede pertenecer a otro empleado del negocio: no se reemplaza.
		var vinculados int64
		if err := tx.Table("empleados").Where("membresia_id = ?", membresia.ID).Count(&vinculados).Error; err != nil {
			return err
		}
		if vinculados > 0 {
			return application.ErrInvitacionYaVinculado
		}

		cambios := map[string]any{
			"membresia_id": membresia.ID, "estado": domain.EstadoEmpleadoActivo, "actualizado_en": ahora,
		}
		if err := datosPersonalesConfirmados(tx, usuarioID, cambios); err != nil {
			return err
		}
		if err := tx.Table("empleados").Where("id = ?", empleado.ID).Updates(cambios).Error; err != nil {
			return err
		}

		if invitacion.RolPredeterminadoID != nil {
			err := tx.Exec(`INSERT INTO roles_membresia (id, membresia_negocio_id, rol_id, asignado_por_usuario_id, asignado_en)
				VALUES (gen_random_uuid(), ?, ?, ?, ?)
				ON CONFLICT (membresia_negocio_id, rol_id) DO NOTHING`,
				membresia.ID, *invitacion.RolPredeterminadoID, invitacion.InvitadoPorUsuarioID, ahora).Error
			if err != nil {
				return err
			}
		}
		if err := asignarSucursalInvitada(tx, invitacion.NegocioID, empleado.ID, *invitacion.SucursalID); err != nil {
			return err
		}
		err = tx.Table("invitaciones_negocio").Where("id = ?", invitacion.ID).Updates(map[string]any{
			"estado": domain.EstadoInvitacionAceptada, "aceptado_por_usuario_id": usuarioID, "aceptado_en": ahora,
		}).Error
		if err != nil {
			return err
		}
		aceptada = domain.InvitacionAceptada{Aceptada: true, NegocioID: invitacion.NegocioID, SucursalID: invitacion.SucursalID}
		return nil
	})
	return aceptada, err
}

// membresiaParaAceptar crea la membresía o reutiliza la activa; una suspendida o revocada no se
// reactiva por un enlace y un propietario existente nunca se degrada.
func membresiaParaAceptar(tx *gorm.DB, negocioID, usuarioID uuid.UUID, ahora time.Time) (domain.MembresiaNegocio, error) {
	var membresia domain.MembresiaNegocio
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("negocio_id = ? AND usuario_id = ?", negocioID, usuarioID).Take(&membresia).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		membresia = domain.MembresiaNegocio{
			NegocioID: negocioID, UsuarioID: usuarioID,
			TipoMiembro: "miembro", Estado: "activo", SeUnioEn: &ahora,
		}
		return membresia, tx.Create(&membresia).Error
	}
	if err != nil {
		return domain.MembresiaNegocio{}, err
	}
	if membresia.Estado != "activo" {
		return domain.MembresiaNegocio{}, application.ErrInvitacionMembresiaInactiva
	}
	return membresia, nil
}

// datosPersonalesConfirmados copia al empleado lo que el invitado confirmó en su perfil.
// Número, puesto, fechas laborales y roles siguen bajo control administrativo.
func datosPersonalesConfirmados(tx *gorm.DB, usuarioID uuid.UUID, cambios map[string]any) error {
	var perfil struct {
		Nombres   string
		Apellidos string
		Telefono  *string
	}
	err := tx.Table("perfil_usuarios").Select("nombres, apellidos, telefono").
		Where("usuario_id = ?", usuarioID).Take(&perfil).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	apellidos := strings.Fields(perfil.Apellidos)
	if strings.TrimSpace(perfil.Nombres) == "" || len(apellidos) == 0 {
		return nil
	}
	cambios["nombre"] = recortar(strings.TrimSpace(perfil.Nombres), 100)
	cambios["segundo_nombre"] = nil
	cambios["primer_apellido"] = recortar(apellidos[0], 100)
	cambios["segundo_apellido"] = nil
	if len(apellidos) > 1 {
		cambios["segundo_apellido"] = recortar(strings.Join(apellidos[1:], " "), 100)
	}
	if perfil.Telefono != nil && strings.TrimSpace(*perfil.Telefono) != "" {
		cambios["telefono"] = strings.TrimSpace(*perfil.Telefono)
	}
	return nil
}

func recortar(valor string, maximo int) string {
	runas := []rune(valor)
	if len(runas) > maximo {
		return string(runas[:maximo])
	}
	return valor
}

// asignarSucursalInvitada crea o reactiva la asignación; solo es principal si el empleado no tiene otra.
func asignarSucursalInvitada(tx *gorm.DB, negocioID, empleadoID, sucursalID uuid.UUID) error {
	var principales int64
	if err := tx.Table("asignaciones_empleado_sucursal").
		Where("empleado_id = ? AND activo = TRUE AND es_principal = TRUE AND sucursal_id <> ?", empleadoID, sucursalID).
		Count(&principales).Error; err != nil {
		return err
	}
	var existente domain.AsignacionEmpleadoSucursal
	err := tx.Where("empleado_id = ? AND sucursal_id = ?", empleadoID, sucursalID).Take(&existente).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tx.Create(&domain.AsignacionEmpleadoSucursal{
			NegocioID: negocioID, EmpleadoID: empleadoID, SucursalID: sucursalID,
			EsPrincipal: principales == 0, Activo: true,
		}).Error
	}
	if err != nil {
		return err
	}
	if existente.Activo {
		return nil
	}
	return tx.Table("asignaciones_empleado_sucursal").Where("id = ?", existente.ID).Updates(map[string]any{
		"activo": true, "finalizado_en": nil, "es_principal": principales == 0,
	}).Error
}

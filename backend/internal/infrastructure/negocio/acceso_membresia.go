package negocio

import (
	"context"
	"strings"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Espera el alias m. Conserva cuentas legacy sin empleado y propietarios sin ficha laboral.
const membresiaHabilitada = `NOT EXISTS (
 SELECT 1 FROM empleados acceso_empleado
 WHERE acceso_empleado.membresia_id = m.id AND acceso_empleado.estado <> 'activo')`

// Serializa cambios de acceso y aceptación para no autorizar con un rol que está cambiando.
func bloquearAccesoNegocio(tx *gorm.DB, negocioID uuid.UUID) error {
	return tx.Exec("SELECT id FROM negocios WHERE id = ? FOR UPDATE", negocioID).Error
}

func validarDelegacionInvitacion(ctx context.Context, tx *gorm.DB, usuarioID, negocioID, rolID uuid.UUID) error {
	var cuentas int64
	if err := tx.Table("usuarios").Where("id = ? AND estado = 'activo'", usuarioID).Count(&cuentas).Error; err != nil {
		return err
	}
	if cuentas != 1 {
		return application.ErrInvitacionProhibida
	}
	permisos, err := permisosEfectivos(ctx, tx, usuarioID, negocioID)
	if err != nil {
		return err
	}
	permitido := false
	for _, codigo := range permisos {
		if codigo == domain.PermisoInvitacionEnviar {
			permitido = true
		}
	}
	if !permitido {
		return application.ErrInvitacionProhibida
	}
	codigo, activo, err := codigoRolActivo(tx, negocioID, rolID)
	if err != nil {
		return err
	}
	if !activo || strings.EqualFold(codigo, domain.CodigoRolPropietario) {
		return application.ErrInvitacionRolNoDisponible
	}
	delegable, err := rolesDelegables(ctx, tx, usuarioID, negocioID, []uuid.UUID{rolID}, nil)
	if err != nil {
		return err
	}
	if !delegable {
		return application.ErrRolNoDelegable
	}
	return nil
}

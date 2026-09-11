package negocio

import (
	"context"
	"errors"
	"strings"
	"time"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SucursalRepository struct {
	db *gorm.DB
}

func NewSucursalRepository(db *gorm.DB) *SucursalRepository {
	return &SucursalRepository{db: db}
}

func (r *SucursalRepository) ObtenerContextoNegocio(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.ContextoNegocioSucursal, error) {
	var access domain.ContextoNegocioSucursal
	err := r.db.WithContext(ctx).Table("negocios AS n").
		Select("n.estado AS estado_negocio, m.tipo_miembro").
		Joins("JOIN membresias_negocio m ON m.negocio_id = n.id AND m.usuario_id = ? AND m.estado = 'activo'", usuarioID).
		Where("n.id = ?", negocioID).
		Take(&access).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ContextoNegocioSucursal{}, application.ErrNegocioNoEncontrado
	}
	return access, err
}

func (r *SucursalRepository) Listar(ctx context.Context, negocioID uuid.UUID, estado, buscar string) ([]domain.SucursalResumen, error) {
	items := make([]domain.SucursalResumen, 0)
	query := r.db.WithContext(ctx).Table("sucursales AS s").
		Select(`s.id, s.negocio_id, s.codigo, s.nombre, s.telefono, s.es_principal,
			CASE WHEN s.activo THEN 'activo' ELSE 'archivado' END AS estado,
			s.creado_en, s.actualizado_en,
			NULLIF(COALESCE(NULLIF(concat_ws(', ', d.ciudad, d.estado), ''), d.referencias), '') AS direccion_resumida`).
		Joins("LEFT JOIN direcciones d ON d.id = s.direccion_id").
		Where("s.negocio_id = ? AND s.activo = ?", negocioID, estado == "activo")
	if buscar != "" {
		pattern := "%" + strings.ToLower(buscar) + "%"
		query = query.Where("lower(s.nombre) LIKE ? OR lower(s.codigo) LIKE ?", pattern, pattern)
	}
	err := query.Order("s.es_principal DESC, s.nombre ASC, s.id ASC").Scan(&items).Error
	return items, err
}

func (r *SucursalRepository) Obtener(ctx context.Context, negocioID, sucursalID uuid.UUID) (domain.SucursalDetalle, error) {
	var branch domain.Sucursal
	if err := r.db.WithContext(ctx).Preload("Direccion").
		Where("id = ? AND negocio_id = ?", sucursalID, negocioID).
		Take(&branch).Error; err != nil {
		return domain.SucursalDetalle{}, sucursalNoEncontrada(err)
	}
	return detalleSucursal(branch), nil
}

func (r *SucursalRepository) Crear(ctx context.Context, negocioID uuid.UUID, input domain.CrearSucursalInput) (uuid.UUID, error) {
	branchID := uuid.New()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var active []domain.Sucursal
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("negocio_id = ? AND activo = TRUE", negocioID).Find(&active).Error; err != nil {
			return err
		}
		principal := input.EsPrincipal || len(active) == 0
		if principal && len(active) > 0 {
			if err := tx.Model(&domain.Sucursal{}).
				Where("negocio_id = ? AND activo = TRUE AND es_principal = TRUE", negocioID).
				Update("es_principal", false).Error; err != nil {
				return err
			}
		}

		var direccionID *uuid.UUID
		if input.Direccion != nil {
			direccion := direccionDesdeInput(*input.Direccion)
			if err := tx.Create(&direccion).Error; err != nil {
				return err
			}
			direccionID = &direccion.ID
		}

		branch := domain.Sucursal{
			ID: branchID, NegocioID: negocioID, Codigo: input.Codigo, Nombre: input.Nombre,
			Telefono: input.Telefono, DireccionID: direccionID, EsPrincipal: principal, Activo: true,
		}
		return normalizarErrorSucursal(tx.Create(&branch).Error)
	})
	return branchID, err
}

func (r *SucursalRepository) Actualizar(ctx context.Context, negocioID, sucursalID uuid.UUID, input domain.ActualizarSucursalInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var branch domain.Sucursal
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND negocio_id = ?", sucursalID, negocioID).Take(&branch).Error; err != nil {
			return sucursalNoEncontrada(err)
		}

		if input.EsPrincipal.Set {
			if input.EsPrincipal.Value == nil {
				return &application.ErrorValidacion{Campos: map[string]string{"es_principal": "no puede ser null"}}
			}
			if !*input.EsPrincipal.Value && branch.EsPrincipal {
				return application.ErrSucursalPrincipalRequerida
			}
			if *input.EsPrincipal.Value {
				if !branch.Activo {
					return application.ErrEstadoSucursal
				}
				var active []domain.Sucursal
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("negocio_id = ? AND activo = TRUE", negocioID).Find(&active).Error; err != nil {
					return err
				}
				if err := tx.Model(&domain.Sucursal{}).
					Where("negocio_id = ? AND activo = TRUE AND id <> ?", negocioID, sucursalID).
					Update("es_principal", false).Error; err != nil {
					return err
				}
			}
		}

		values := map[string]interface{}{"actualizado_en": time.Now()}
		if input.Nombre.Set {
			values["nombre"] = *input.Nombre.Value
		}
		if input.Telefono.Set {
			values["telefono"] = input.Telefono.Value
		}
		if input.EsPrincipal.Set && input.EsPrincipal.Value != nil {
			values["es_principal"] = *input.EsPrincipal.Value
		}
		if err := tx.Model(&domain.Sucursal{}).Where("id = ? AND negocio_id = ?", sucursalID, negocioID).
			Updates(values).Error; err != nil {
			return normalizarErrorSucursal(err)
		}
		if input.Direccion.Set {
			return actualizarDireccionSucursal(tx, &branch, input.Direccion.Value)
		}
		return nil
	})
}

func (r *SucursalRepository) Archivar(ctx context.Context, negocioID, sucursalID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var active []domain.Sucursal
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("negocio_id = ? AND activo = TRUE", negocioID).Find(&active).Error; err != nil {
			return err
		}
		var branch domain.Sucursal
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND negocio_id = ?", sucursalID, negocioID).Take(&branch).Error; err != nil {
			return sucursalNoEncontrada(err)
		}
		if !branch.Activo {
			return application.ErrEstadoSucursal
		}
		if branch.EsPrincipal && len(active) > 1 {
			return application.ErrSucursalPrincipalRequerida
		}
		now := time.Now()
		return tx.Model(&domain.Sucursal{}).Where("id = ? AND negocio_id = ?", sucursalID, negocioID).
			Updates(map[string]interface{}{
				"activo": false, "es_principal": false, "eliminado_en": now, "actualizado_en": now,
			}).Error
	})
}

func (r *SucursalRepository) Restaurar(ctx context.Context, negocioID, sucursalID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var active []domain.Sucursal
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("negocio_id = ? AND activo = TRUE", negocioID).Find(&active).Error; err != nil {
			return err
		}
		var branch domain.Sucursal
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND negocio_id = ?", sucursalID, negocioID).Take(&branch).Error; err != nil {
			return sucursalNoEncontrada(err)
		}
		if branch.Activo {
			return application.ErrEstadoSucursal
		}
		hasPrincipal := false
		for _, item := range active {
			if item.EsPrincipal {
				hasPrincipal = true
				break
			}
		}
		return tx.Model(&domain.Sucursal{}).Where("id = ? AND negocio_id = ?", sucursalID, negocioID).
			Updates(map[string]interface{}{
				"activo": true, "es_principal": !hasPrincipal, "eliminado_en": nil, "actualizado_en": time.Now(),
			}).Error
	})
}

func actualizarDireccionSucursal(tx *gorm.DB, branch *domain.Sucursal, input *domain.ActualizarDireccionInput) error {
	if input == nil {
		if branch.DireccionID == nil {
			return nil
		}
		oldID := *branch.DireccionID
		if err := tx.Model(&domain.Sucursal{}).Where("id = ?", branch.ID).Update("direccion_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.Direccion{}, "id = ?", oldID).Error
	}

	if branch.DireccionID == nil {
		if !direccionUpdateTieneContenido(*input) {
			return nil
		}
		direccion := domain.Direccion{CodigoPais: "MX"}
		aplicarDireccionUpdates(&direccion, *input)
		if err := tx.Create(&direccion).Error; err != nil {
			return err
		}
		return tx.Model(&domain.Sucursal{}).Where("id = ?", branch.ID).Update("direccion_id", direccion.ID).Error
	}

	values := direccionUpdates(*input)
	values["actualizado_en"] = time.Now()
	if err := tx.Model(&domain.Direccion{}).Where("id = ?", *branch.DireccionID).Updates(values).Error; err != nil {
		return err
	}
	var address domain.Direccion
	if err := tx.First(&address, "id = ?", *branch.DireccionID).Error; err != nil {
		return err
	}
	if direccionTieneContenido(address) {
		return nil
	}
	oldID := *branch.DireccionID
	if err := tx.Model(&domain.Sucursal{}).Where("id = ?", branch.ID).Update("direccion_id", nil).Error; err != nil {
		return err
	}
	return tx.Delete(&domain.Direccion{}, "id = ?", oldID).Error
}

func detalleSucursal(branch domain.Sucursal) domain.SucursalDetalle {
	estado := "activo"
	if !branch.Activo {
		estado = "archivado"
	}
	return domain.SucursalDetalle{
		ID: branch.ID, NegocioID: branch.NegocioID, Codigo: branch.Codigo,
		Nombre: branch.Nombre, Telefono: branch.Telefono, EsPrincipal: branch.EsPrincipal,
		Estado: estado, CreadoEn: branch.CreadoEn, ActualizadoEn: branch.ActualizadoEn,
		EliminadoEn: branch.EliminadoEn, Direccion: branch.Direccion,
	}
}

func normalizarErrorSucursal(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return application.ErrSucursalConflicto
	}
	return err
}

func sucursalNoEncontrada(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return application.ErrSucursalNoEncontrada
	}
	return err
}

package negocio

import (
	"context"
	"errors"
	"time"

	application "tienda/backend/internal/application/negocio"
	catalogodomain "tienda/backend/internal/domain"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type NegocioRepository struct {
	db *gorm.DB
}

func NewNegocioRepository(db *gorm.DB) *NegocioRepository {
	return &NegocioRepository{db: db}
}

func (r *NegocioRepository) Listar(ctx context.Context, usuarioID uuid.UUID, estado string) ([]domain.NegocioResumen, error) {
	items := make([]domain.NegocioResumen, 0)
	err := r.db.WithContext(ctx).Table("negocios AS n").
		Select(`n.id, n.slug, n.nombre_comercial, n.rfc, m.tipo_miembro, n.estado, n.creado_en,
			EXISTS (SELECT 1 FROM sucursales s WHERE s.negocio_id = n.id AND s.activo = TRUE) AS tiene_sucursales,
			(SELECT COUNT(*) FROM sucursales s WHERE s.negocio_id = n.id AND s.activo = TRUE) AS total_sucursales`).
		Joins("JOIN membresias_negocio m ON m.negocio_id = n.id AND m.usuario_id = ? AND m.estado = 'activo'", usuarioID).
		Where("n.estado = ?", estado).
		Order("n.nombre_comercial ASC").
		Scan(&items).Error
	return items, err
}

func (r *NegocioRepository) ObtenerAccesible(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.NegocioDetalle, error) {
	type accessRow struct {
		TipoMiembro string
	}
	var access accessRow
	if err := r.db.WithContext(ctx).Table("membresias_negocio").
		Select("tipo_miembro").
		Where("negocio_id = ? AND usuario_id = ? AND estado = 'activo'", negocioID, usuarioID).
		Take(&access).Error; err != nil {
		return domain.NegocioDetalle{}, negocioNoEncontrado(err)
	}

	var negocio domain.Negocio
	if err := r.db.WithContext(ctx).Preload("Direccion").First(&negocio, "id = ?", negocioID).Error; err != nil {
		return domain.NegocioDetalle{}, err
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&catalogodomain.Sucursal{}).
		Where("negocio_id = ? AND activo = TRUE", negocioID).Count(&total).Error; err != nil {
		return domain.NegocioDetalle{}, err
	}
	return detalleNegocio(negocio, access.TipoMiembro, total), nil
}

func (r *NegocioRepository) ExisteSlug(ctx context.Context, slug string) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&domain.Negocio{}).
		Where("lower(slug) = lower(?)", slug).Count(&total).Error
	return total > 0, err
}

func (r *NegocioRepository) Crear(ctx context.Context, usuarioID uuid.UUID, slug string, input domain.CrearNegocioInput) (domain.NegocioDetalle, error) {
	negocioID := uuid.New()
	now := time.Now()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var direccionID *uuid.UUID
		if input.Direccion != nil {
			direccion := direccionDesdeInput(*input.Direccion)
			if err := tx.Create(&direccion).Error; err != nil {
				return err
			}
			direccionID = &direccion.ID
		}

		negocio := domain.Negocio{
			ID: negocioID, Slug: slug, NombreComercial: input.NombreComercial,
			RazonSocial: input.RazonSocial, RFC: input.RFC, Telefono: input.Telefono,
			Correo: input.Correo, DireccionID: direccionID, CodigoMoneda: input.CodigoMoneda,
			ZonaHoraria: input.ZonaHoraria, Estado: "activo", CreadoPorUsuarioID: &usuarioID,
		}
		if err := tx.Create(&negocio).Error; err != nil {
			return normalizarErrorPostgres(err)
		}

		configuracion := domain.ConfiguracionNegocio{
			NegocioID: negocioID, MetodoCosteo: "promedio_ponderado",
			RequerirAutorizacionCambioPrecio:  true,
			PermitirCambioCajaConAutorizacion: true,
			PreciosMayoreoHabilitados:         true,
		}
		if err := tx.Create(&configuracion).Error; err != nil {
			return err
		}
		membresia := domain.MembresiaNegocio{
			NegocioID: negocioID, UsuarioID: usuarioID, TipoMiembro: "propietario",
			Estado: "activo", SeUnioEn: &now,
		}
		return tx.Create(&membresia).Error
	})
	if err != nil {
		return domain.NegocioDetalle{}, err
	}
	return r.ObtenerAccesible(ctx, usuarioID, negocioID)
}

func (r *NegocioRepository) Actualizar(ctx context.Context, usuarioID, negocioID uuid.UUID, input domain.ActualizarNegocioInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var negocio domain.Negocio
		if err := tx.Table("negocios AS n").Select("n.*").
			Joins("JOIN membresias_negocio m ON m.negocio_id = n.id").
			Where("n.id = ? AND m.usuario_id = ? AND m.estado = 'activo' AND m.tipo_miembro = 'propietario'", negocioID, usuarioID).
			Take(&negocio).Error; err != nil {
			return negocioNoEncontrado(err)
		}

		values := negocioUpdates(input)
		if len(values) > 0 {
			if err := tx.Model(&domain.Negocio{}).Where("id = ?", negocioID).Updates(values).Error; err != nil {
				return normalizarErrorPostgres(err)
			}
		}
		if input.Direccion.Set {
			if err := actualizarDireccion(tx, &negocio, input.Direccion.Value); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *NegocioRepository) Archivar(ctx context.Context, usuarioID, negocioID uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&domain.Negocio{}).
		Where(`id = ? AND estado <> 'archivado' AND EXISTS (
			SELECT 1 FROM membresias_negocio m
			WHERE m.negocio_id = negocios.id AND m.usuario_id = ?
			AND m.estado = 'activo' AND m.tipo_miembro = 'propietario'
		)`, negocioID, usuarioID).
		Updates(map[string]interface{}{"estado": "archivado", "archivado_en": &now, "actualizado_en": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return application.ErrEstadoNegocio
	}
	return nil
}

func (r *NegocioRepository) Restaurar(ctx context.Context, usuarioID, negocioID uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&domain.Negocio{}).
		Where(`id = ? AND estado = 'archivado' AND EXISTS (
			SELECT 1 FROM membresias_negocio m
			WHERE m.negocio_id = negocios.id AND m.usuario_id = ?
			AND m.estado = 'activo' AND m.tipo_miembro = 'propietario'
		)`, negocioID, usuarioID).
		Updates(map[string]interface{}{"estado": "activo", "archivado_en": nil, "actualizado_en": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return application.ErrEstadoNegocio
	}
	return nil
}

func detalleNegocio(negocio domain.Negocio, tipoMiembro string, totalSucursales int64) domain.NegocioDetalle {
	return domain.NegocioDetalle{
		ID: negocio.ID, Slug: negocio.Slug, NombreComercial: negocio.NombreComercial,
		RazonSocial: negocio.RazonSocial, RFC: negocio.RFC, Telefono: negocio.Telefono,
		Correo: negocio.Correo, CodigoMoneda: negocio.CodigoMoneda,
		ZonaHoraria: negocio.ZonaHoraria, Estado: negocio.Estado, TipoMiembro: tipoMiembro,
		TieneSucursales: totalSucursales > 0, TotalSucursales: totalSucursales,
		CreadoEn: negocio.CreadoEn, ActualizadoEn: negocio.ActualizadoEn,
		ArchivadoEn: negocio.ArchivadoEn, Direccion: negocio.Direccion,
	}
}

func direccionDesdeInput(input domain.DireccionInput) domain.Direccion {
	return domain.Direccion{
		CodigoPais: input.CodigoPais, Estado: input.Estado, Municipio: input.Municipio,
		Ciudad: input.Ciudad, Colonia: input.Colonia, CodigoPostal: input.CodigoPostal,
		Calle: input.Calle, NumeroExterior: input.NumeroExterior,
		NumeroInterior: input.NumeroInterior, Referencias: input.Referencias,
	}
}

func negocioUpdates(input domain.ActualizarNegocioInput) map[string]interface{} {
	values := map[string]interface{}{"actualizado_en": time.Now()}
	if input.NombreComercial.Set {
		values["nombre_comercial"] = *input.NombreComercial.Value
		values["nombre"] = *input.NombreComercial.Value
	}
	agregarOptional(values, "razon_social", input.RazonSocial)
	agregarOptional(values, "rfc", input.RFC)
	agregarOptional(values, "telefono", input.Telefono)
	if input.Correo.Set {
		values["correo"] = input.Correo.Value
		values["email"] = input.Correo.Value
	}
	agregarOptional(values, "codigo_moneda", input.CodigoMoneda)
	agregarOptional(values, "zona_horaria", input.ZonaHoraria)
	return values
}

func agregarOptional(values map[string]interface{}, key string, optional domain.Optional[string]) {
	if optional.Set {
		values[key] = optional.Value
	}
}

func actualizarDireccion(tx *gorm.DB, negocio *domain.Negocio, input *domain.ActualizarDireccionInput) error {
	if input == nil {
		if negocio.DireccionID == nil {
			return nil
		}
		oldID := *negocio.DireccionID
		if err := tx.Model(&domain.Negocio{}).Where("id = ?", negocio.ID).Update("direccion_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.Direccion{}, "id = ?", oldID).Error
	}

	values := direccionUpdates(*input)
	if negocio.DireccionID == nil {
		if !direccionUpdateTieneContenido(*input) {
			return nil
		}
		direccion := domain.Direccion{CodigoPais: "MX"}
		aplicarDireccionUpdates(&direccion, *input)
		if err := tx.Create(&direccion).Error; err != nil {
			return err
		}
		return tx.Model(&domain.Negocio{}).Where("id = ?", negocio.ID).Update("direccion_id", direccion.ID).Error
	}

	values["actualizado_en"] = time.Now()
	if err := tx.Model(&domain.Direccion{}).Where("id = ?", *negocio.DireccionID).Updates(values).Error; err != nil {
		return err
	}
	var direccion domain.Direccion
	if err := tx.First(&direccion, "id = ?", *negocio.DireccionID).Error; err != nil {
		return err
	}
	if direccionTieneContenido(direccion) {
		return nil
	}
	oldID := *negocio.DireccionID
	if err := tx.Model(&domain.Negocio{}).Where("id = ?", negocio.ID).Update("direccion_id", nil).Error; err != nil {
		return err
	}
	return tx.Delete(&domain.Direccion{}, "id = ?", oldID).Error
}

func direccionUpdates(input domain.ActualizarDireccionInput) map[string]interface{} {
	values := map[string]interface{}{}
	agregarOptional(values, "codigo_pais", input.CodigoPais)
	agregarOptional(values, "estado", input.Estado)
	agregarOptional(values, "municipio", input.Municipio)
	agregarOptional(values, "ciudad", input.Ciudad)
	agregarOptional(values, "colonia", input.Colonia)
	agregarOptional(values, "codigo_postal", input.CodigoPostal)
	agregarOptional(values, "calle", input.Calle)
	agregarOptional(values, "numero_exterior", input.NumeroExterior)
	agregarOptional(values, "numero_interior", input.NumeroInterior)
	agregarOptional(values, "referencias", input.Referencias)
	return values
}

func aplicarDireccionUpdates(direccion *domain.Direccion, input domain.ActualizarDireccionInput) {
	if input.CodigoPais.Set && input.CodigoPais.Value != nil {
		direccion.CodigoPais = *input.CodigoPais.Value
	}
	if input.Estado.Set {
		direccion.Estado = input.Estado.Value
	}
	if input.Municipio.Set {
		direccion.Municipio = input.Municipio.Value
	}
	if input.Ciudad.Set {
		direccion.Ciudad = input.Ciudad.Value
	}
	if input.Colonia.Set {
		direccion.Colonia = input.Colonia.Value
	}
	if input.CodigoPostal.Set {
		direccion.CodigoPostal = input.CodigoPostal.Value
	}
	if input.Calle.Set {
		direccion.Calle = input.Calle.Value
	}
	if input.NumeroExterior.Set {
		direccion.NumeroExterior = input.NumeroExterior.Value
	}
	if input.NumeroInterior.Set {
		direccion.NumeroInterior = input.NumeroInterior.Value
	}
	if input.Referencias.Set {
		direccion.Referencias = input.Referencias.Value
	}
}

func direccionUpdateTieneContenido(input domain.ActualizarDireccionInput) bool {
	return input.Estado.Value != nil || input.Municipio.Value != nil || input.Ciudad.Value != nil ||
		input.Colonia.Value != nil || input.CodigoPostal.Value != nil || input.Calle.Value != nil ||
		input.NumeroExterior.Value != nil || input.NumeroInterior.Value != nil || input.Referencias.Value != nil
}

func direccionTieneContenido(input domain.Direccion) bool {
	return input.Estado != nil || input.Municipio != nil || input.Ciudad != nil || input.Colonia != nil ||
		input.CodigoPostal != nil || input.Calle != nil || input.NumeroExterior != nil ||
		input.NumeroInterior != nil || input.Referencias != nil
}

func normalizarErrorPostgres(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return application.ErrNegocioConflicto
	}
	return err
}

func negocioNoEncontrado(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return application.ErrNegocioNoEncontrado
	}
	return err
}

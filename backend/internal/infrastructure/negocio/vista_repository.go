package negocio

import (
	"context"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VistaRepository struct {
	db *gorm.DB
}

func NewVistaRepository(db *gorm.DB) *VistaRepository {
	return &VistaRepository{db: db}
}

// SeccionesOcultas devuelve el valor guardado; un negocio sin configuración ve todo.
func (r *VistaRepository) SeccionesOcultas(ctx context.Context, negocioID uuid.UUID) (string, error) {
	var valores []string
	if err := r.db.WithContext(ctx).Model(&domain.ConfiguracionNegocio{}).
		Where("negocio_id = ?", negocioID).Limit(1).Pluck("secciones_ocultas", &valores).Error; err != nil {
		return "", err
	}
	if len(valores) == 0 {
		return "", nil
	}
	return valores[0], nil
}

// GuardarSeccionesOcultas crea la configuración del negocio si todavía no existe.
func (r *VistaRepository) GuardarSeccionesOcultas(ctx context.Context, negocioID uuid.UUID, valor string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		resultado := tx.Model(&domain.ConfiguracionNegocio{}).Where("negocio_id = ?", negocioID).Update("secciones_ocultas", valor)
		if resultado.Error != nil || resultado.RowsAffected > 0 {
			return resultado.Error
		}
		return tx.Create(&domain.ConfiguracionNegocio{
			NegocioID: negocioID, MetodoCosteo: "promedio_ponderado",
			RequerirAutorizacionCambioPrecio:  true,
			PermitirCambioCajaConAutorizacion: true,
			PreciosMayoreoHabilitados:         true,
			SeccionesOcultas:                  valor,
		}).Error
	})
}

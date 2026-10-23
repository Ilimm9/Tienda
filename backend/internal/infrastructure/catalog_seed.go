package infrastructure

import (
	"context"

	"tienda/backend/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NombreProveedorGeneral es el proveedor que se siembra para compras sin proveedor fijo; se puede desactivar.
const NombreProveedorGeneral = "Proveedor general"

// unidadesEmpaqueIniciales son las formas de compra listas en Registrar compra (capability 34).
var unidadesEmpaqueIniciales = []struct {
	nombre, simbolo string
	factor          float64
}{
	{"Pieza", "pz", 1},
	{"Caja 12", "cj12", 12},
	{"Caja 24", "cj24", 24},
}

// CatalogSeeder siembra los catálogos básicos de un negocio recién creado.
type CatalogSeeder struct{ db *gorm.DB }

func NewCatalogSeeder(db *gorm.DB) *CatalogSeeder { return &CatalogSeeder{db: db} }

// SembrarCatalogoInicial crea las unidades de empaque y el proveedor general; una unidad cuyo nombre
// o símbolo ya exista se omite gracias a los índices únicos del negocio.
func (s *CatalogSeeder) SembrarCatalogoInicial(ctx context.Context, negocioID uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, unidad := range unidadesEmpaqueIniciales {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&domain.UnidadMedida{
				NegocioID: negocioID, Nombre: unidad.nombre, Simbolo: unidad.simbolo, Tipo: "EMPAQUE",
				FactorABase: unidad.factor, PermiteFraccion: false, Decimales: 0, Activo: true,
			}).Error; err != nil {
				return err
			}
		}
		return tx.Create(&domain.Proveedor{NegocioID: negocioID, Nombre: NombreProveedorGeneral, Activo: true}).Error
	})
}

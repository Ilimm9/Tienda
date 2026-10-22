package negocio

import (
	"context"
	"errors"
	"strings"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const indicePuestoNombre = "uq_puestos_negocio_nombre"

type PuestoRepository struct {
	db *gorm.DB
}

func NewPuestoRepository(db *gorm.DB) *PuestoRepository {
	return &PuestoRepository{db: db}
}

func (r *PuestoRepository) Listar(ctx context.Context, negocioID uuid.UUID) ([]domain.PuestoNegocio, error) {
	puestos := []domain.PuestoNegocio{}
	err := r.db.WithContext(ctx).Where("negocio_id = ?", negocioID).Order("lower(nombre)").Find(&puestos).Error
	return puestos, err
}

func (r *PuestoRepository) Obtener(ctx context.Context, negocioID, puestoID uuid.UUID) (domain.PuestoNegocio, error) {
	var puesto domain.PuestoNegocio
	err := r.db.WithContext(ctx).Where("id = ? AND negocio_id = ?", puestoID, negocioID).Take(&puesto).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return puesto, application.ErrPuestoNoEncontrado
	}
	return puesto, err
}

func (r *PuestoRepository) ExisteNombre(ctx context.Context, negocioID uuid.UUID, nombre string, excluir *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Table("puestos_negocio").Where("negocio_id = ? AND lower(nombre) = lower(?)", negocioID, nombre)
	if excluir != nil {
		query = query.Where("id <> ?", *excluir)
	}
	var total int64
	err := query.Count(&total).Error
	return total > 0, err
}

func (r *PuestoRepository) RolDelNegocio(ctx context.Context, negocioID, rolID uuid.UUID) (bool, error) {
	_, existe, err := codigoRolActivo(r.db.WithContext(ctx), negocioID, rolID)
	return existe, err
}

func (r *PuestoRepository) Crear(ctx context.Context, puesto *domain.PuestoNegocio) error {
	return errorPuestoDuplicado(r.db.WithContext(ctx).Create(puesto).Error)
}

func (r *PuestoRepository) Guardar(ctx context.Context, puesto *domain.PuestoNegocio) error {
	return errorPuestoDuplicado(r.db.WithContext(ctx).Save(puesto).Error)
}

// errorPuestoDuplicado traduce la violación del índice cuando dos altas simultáneas pasan la consulta previa.
func errorPuestoDuplicado(err error) error {
	if err != nil && strings.Contains(err.Error(), indicePuestoNombre) {
		return application.ErrPuestoDuplicado
	}
	return err
}

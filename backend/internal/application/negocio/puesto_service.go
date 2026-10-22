package negocio

import (
	"context"
	"errors"
	"strings"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

var (
	ErrPuestoNoEncontrado = errors.New("puesto no encontrado")
	ErrPuestoDuplicado    = errors.New("ya existe un puesto con ese nombre")
)

type PuestoRepository interface {
	Listar(ctx context.Context, negocioID uuid.UUID) ([]domain.PuestoNegocio, error)
	Obtener(ctx context.Context, negocioID, puestoID uuid.UUID) (domain.PuestoNegocio, error)
	// ExisteNombre compara sin distinguir mayúsculas; `excluir` es el puesto que se edita.
	ExisteNombre(ctx context.Context, negocioID uuid.UUID, nombre string, excluir *uuid.UUID) (bool, error)
	RolDelNegocio(ctx context.Context, negocioID, rolID uuid.UUID) (bool, error)
	Crear(ctx context.Context, puesto *domain.PuestoNegocio) error
	Guardar(ctx context.Context, puesto *domain.PuestoNegocio) error
}

// PuestoService reutiliza la autorización de empleados: el catálogo es parte de administrar el equipo.
type PuestoService struct {
	puestos   PuestoRepository
	empleados *EmpleadoService
}

func NewPuestoService(puestos PuestoRepository, empleados *EmpleadoService) *PuestoService {
	return &PuestoService{puestos: puestos, empleados: empleados}
}

func (s *PuestoService) Listar(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]domain.PuestoNegocio, error) {
	if err := s.empleados.autorizar(ctx, usuarioID, negocioID, domain.PermisoEmpleadoVer, false); err != nil {
		return nil, err
	}
	return s.puestos.Listar(ctx, negocioID)
}

func (s *PuestoService) Crear(ctx context.Context, usuarioID, negocioID uuid.UUID, input domain.CrearPuestoInput) (domain.PuestoNegocio, error) {
	if err := s.empleados.autorizar(ctx, usuarioID, negocioID, domain.PermisoEmpleadoGestionar, true); err != nil {
		return domain.PuestoNegocio{}, err
	}
	puesto := domain.PuestoNegocio{NegocioID: negocioID, Nombre: strings.Join(strings.Fields(input.Nombre), " "), RolSugeridoID: input.RolSugeridoID, Activo: true}
	if err := s.validar(ctx, &puesto); err != nil {
		return domain.PuestoNegocio{}, err
	}
	return puesto, s.puestos.Crear(ctx, &puesto)
}

func (s *PuestoService) Actualizar(ctx context.Context, usuarioID, negocioID, puestoID uuid.UUID, input domain.ActualizarPuestoInput) (domain.PuestoNegocio, error) {
	if err := s.empleados.autorizar(ctx, usuarioID, negocioID, domain.PermisoEmpleadoGestionar, true); err != nil {
		return domain.PuestoNegocio{}, err
	}
	puesto, err := s.puestos.Obtener(ctx, negocioID, puestoID)
	if err != nil {
		return domain.PuestoNegocio{}, err
	}
	if !input.Nombre.Set && !input.RolSugeridoID.Set && input.Activo == nil {
		return domain.PuestoNegocio{}, ErrEmpleadoSinCambios
	}
	if input.Nombre.Set {
		puesto.Nombre = ""
		if input.Nombre.Value != nil {
			puesto.Nombre = strings.Join(strings.Fields(*input.Nombre.Value), " ")
		}
	}
	if input.RolSugeridoID.Set {
		puesto.RolSugeridoID = input.RolSugeridoID.Value
	}
	if input.Activo != nil {
		puesto.Activo = *input.Activo
	}
	if err := s.validar(ctx, &puesto); err != nil {
		return domain.PuestoNegocio{}, err
	}
	return puesto, s.puestos.Guardar(ctx, &puesto)
}

func (s *PuestoService) validar(ctx context.Context, puesto *domain.PuestoNegocio) error {
	if longitud := len([]rune(puesto.Nombre)); longitud < 2 || longitud > 120 {
		return &ErrorValidacion{Campos: map[string]string{"nombre": "usa entre 2 y 120 caracteres"}}
	}
	if puesto.RolSugeridoID != nil {
		existe, err := s.puestos.RolDelNegocio(ctx, puesto.NegocioID, *puesto.RolSugeridoID)
		if err != nil {
			return err
		}
		if !existe {
			return &ErrorValidacion{Campos: map[string]string{"rol_sugerido_id": "no es un rol activo del negocio"}}
		}
	}
	var excluir *uuid.UUID
	if puesto.ID != uuid.Nil {
		excluir = &puesto.ID
	}
	duplicado, err := s.puestos.ExisteNombre(ctx, puesto.NegocioID, puesto.Nombre, excluir)
	if err != nil {
		return err
	}
	if duplicado {
		return ErrPuestoDuplicado
	}
	return nil
}

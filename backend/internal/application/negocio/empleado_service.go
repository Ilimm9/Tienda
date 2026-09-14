package negocio

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

var (
	ErrEmpleadoNoEncontrado = errors.New("empleado no encontrado")
	ErrEmpleadoProhibido    = errors.New("no tienes permiso para realizar esta acción")
	ErrEmpleadoSinCambios   = errors.New("no se recibieron campos para actualizar")
	ErrEmpleadoConflicto    = errors.New("ya existe un empleado con ese número")
)

var (
	numeroEmpleadoPattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{0,39}$`)
	correoPattern         = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
)

var estadosEmpleado = map[string]struct{}{
	domain.EstadoEmpleadoPendiente:  {},
	domain.EstadoEmpleadoActivo:     {},
	domain.EstadoEmpleadoSuspendido: {},
	domain.EstadoEmpleadoTerminado:  {},
}

type EmpleadoRepository interface {
	ObtenerContextoNegocio(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.ContextoNegocioSucursal, error)
	PermisosEfectivos(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]string, error)
	Listar(ctx context.Context, negocioID uuid.UUID, estado, buscar string) ([]domain.EmpleadoResumen, error)
	Obtener(ctx context.Context, negocioID, empleadoID uuid.UUID) (domain.EmpleadoDetalle, error)
	ExisteNumero(ctx context.Context, negocioID uuid.UUID, numero string, excluir *uuid.UUID) (bool, error)
	Crear(ctx context.Context, negocioID, creadoPor uuid.UUID, input domain.CrearEmpleadoInput) (uuid.UUID, error)
	Actualizar(ctx context.Context, negocioID, empleadoID uuid.UUID, input domain.ActualizarEmpleadoInput) error
}

type EmpleadoService struct {
	empleados EmpleadoRepository
}

func NewEmpleadoService(empleados EmpleadoRepository) *EmpleadoService {
	return &EmpleadoService{empleados: empleados}
}

func (s *EmpleadoService) Listar(ctx context.Context, usuarioID, negocioID uuid.UUID, estado, buscar string) ([]domain.EmpleadoResumen, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoEmpleadoVer, false); err != nil {
		return nil, err
	}
	estado = strings.TrimSpace(estado)
	if estado != "" && estado != "todos" {
		if _, ok := estadosEmpleado[estado]; !ok {
			return nil, &ErrorValidacion{Campos: map[string]string{"estado": "no es un estado válido"}}
		}
	}
	return s.empleados.Listar(ctx, negocioID, estado, strings.TrimSpace(buscar))
}

func (s *EmpleadoService) Obtener(ctx context.Context, usuarioID, negocioID, empleadoID uuid.UUID) (domain.EmpleadoDetalle, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoEmpleadoVer, false); err != nil {
		return domain.EmpleadoDetalle{}, err
	}
	return s.empleados.Obtener(ctx, negocioID, empleadoID)
}

func (s *EmpleadoService) Crear(ctx context.Context, usuarioID, negocioID uuid.UUID, input domain.CrearEmpleadoInput) (domain.EmpleadoDetalle, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoEmpleadoGestionar, true); err != nil {
		return domain.EmpleadoDetalle{}, err
	}
	normalizarCrearEmpleado(&input)
	if err := validarCrearEmpleado(input); err != nil {
		return domain.EmpleadoDetalle{}, err
	}
	if input.NumeroEmpleado != nil {
		existe, err := s.empleados.ExisteNumero(ctx, negocioID, *input.NumeroEmpleado, nil)
		if err != nil {
			return domain.EmpleadoDetalle{}, err
		}
		if existe {
			return domain.EmpleadoDetalle{}, ErrEmpleadoConflicto
		}
	}
	id, err := s.empleados.Crear(ctx, negocioID, usuarioID, input)
	if err != nil {
		return domain.EmpleadoDetalle{}, err
	}
	return s.empleados.Obtener(ctx, negocioID, id)
}

func (s *EmpleadoService) Actualizar(ctx context.Context, usuarioID, negocioID, empleadoID uuid.UUID, input domain.ActualizarEmpleadoInput) (domain.EmpleadoDetalle, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoEmpleadoGestionar, true); err != nil {
		return domain.EmpleadoDetalle{}, err
	}
	if _, err := s.empleados.Obtener(ctx, negocioID, empleadoID); err != nil {
		return domain.EmpleadoDetalle{}, err
	}
	if !actualizacionEmpleadoTieneCampos(input) {
		return domain.EmpleadoDetalle{}, ErrEmpleadoSinCambios
	}
	if err := normalizarYValidarActualizarEmpleado(&input); err != nil {
		return domain.EmpleadoDetalle{}, err
	}
	if input.NumeroEmpleado.Set && input.NumeroEmpleado.Value != nil {
		existe, err := s.empleados.ExisteNumero(ctx, negocioID, *input.NumeroEmpleado.Value, &empleadoID)
		if err != nil {
			return domain.EmpleadoDetalle{}, err
		}
		if existe {
			return domain.EmpleadoDetalle{}, ErrEmpleadoConflicto
		}
	}
	if err := s.empleados.Actualizar(ctx, negocioID, empleadoID, input); err != nil {
		return domain.EmpleadoDetalle{}, err
	}
	return s.empleados.Obtener(ctx, negocioID, empleadoID)
}

func (s *EmpleadoService) autorizar(ctx context.Context, usuarioID, negocioID uuid.UUID, permiso string, escritura bool) error {
	access, err := s.empleados.ObtenerContextoNegocio(ctx, usuarioID, negocioID)
	if err != nil {
		return err
	}
	if escritura && access.EstadoNegocio != "activo" {
		return ErrEstadoNegocio
	}
	permisos, err := s.empleados.PermisosEfectivos(ctx, usuarioID, negocioID)
	if err != nil {
		return err
	}
	for _, actual := range permisos {
		if actual == permiso {
			return nil
		}
	}
	return ErrEmpleadoProhibido
}

func normalizarCrearEmpleado(input *domain.CrearEmpleadoInput) {
	input.Nombre = strings.TrimSpace(input.Nombre)
	input.PrimerApellido = strings.TrimSpace(input.PrimerApellido)
	input.SegundoNombre = limpiarOpcional(input.SegundoNombre, false)
	input.SegundoApellido = limpiarOpcional(input.SegundoApellido, false)
	input.Telefono = limpiarOpcional(input.Telefono, false)
	input.Puesto = limpiarOpcional(input.Puesto, false)
	input.Correo = limpiarOpcional(input.Correo, true)
	input.ContratadoEn = limpiarOpcional(input.ContratadoEn, false)
	if input.NumeroEmpleado != nil {
		numero := strings.ToUpper(strings.TrimSpace(*input.NumeroEmpleado))
		if numero == "" {
			input.NumeroEmpleado = nil
		} else {
			input.NumeroEmpleado = &numero
		}
	}
}

func validarCrearEmpleado(input domain.CrearEmpleadoInput) error {
	campos := map[string]string{}
	validarNombreEmpleado(campos, "nombre", input.Nombre)
	validarNombreEmpleado(campos, "primer_apellido", input.PrimerApellido)
	if input.NumeroEmpleado != nil && !numeroEmpleadoPattern.MatchString(*input.NumeroEmpleado) {
		campos["numero_empleado"] = "usa hasta 40 caracteres: letras, números, guion o guion bajo"
	}
	if input.Correo != nil && !correoPattern.MatchString(*input.Correo) {
		campos["correo"] = "no es un correo válido"
	}
	if input.ContratadoEn != nil && !fechaValida(*input.ContratadoEn) {
		campos["contratado_en"] = "usa el formato AAAA-MM-DD"
	}
	if len(campos) > 0 {
		return &ErrorValidacion{Campos: campos}
	}
	return nil
}

func normalizarYValidarActualizarEmpleado(input *domain.ActualizarEmpleadoInput) error {
	campos := map[string]string{}
	recortarOpcional(&input.Nombre)
	recortarOpcional(&input.PrimerApellido)
	recortarOpcional(&input.SegundoNombre)
	recortarOpcional(&input.SegundoApellido)
	recortarOpcional(&input.Telefono)
	recortarOpcional(&input.Puesto)
	recortarOpcional(&input.Correo)

	if input.Nombre.Set {
		if input.Nombre.Value == nil {
			campos["nombre"] = "no puede quedar vacío"
		} else {
			validarNombreEmpleado(campos, "nombre", *input.Nombre.Value)
		}
	}
	if input.PrimerApellido.Set {
		if input.PrimerApellido.Value == nil {
			campos["primer_apellido"] = "no puede quedar vacío"
		} else {
			validarNombreEmpleado(campos, "primer_apellido", *input.PrimerApellido.Value)
		}
	}
	if input.NumeroEmpleado.Set && input.NumeroEmpleado.Value != nil {
		numero := strings.ToUpper(strings.TrimSpace(*input.NumeroEmpleado.Value))
		if numero == "" {
			input.NumeroEmpleado.Value = nil
		} else if !numeroEmpleadoPattern.MatchString(numero) {
			campos["numero_empleado"] = "usa hasta 40 caracteres: letras, números, guion o guion bajo"
		} else {
			input.NumeroEmpleado.Value = &numero
		}
	}
	if input.Correo.Set && input.Correo.Value != nil && !correoPattern.MatchString(*input.Correo.Value) {
		campos["correo"] = "no es un correo válido"
	}
	if input.Estado.Set {
		if input.Estado.Value == nil {
			campos["estado"] = "no puede quedar vacío"
		} else if _, ok := estadosEmpleado[*input.Estado.Value]; !ok {
			campos["estado"] = "no es un estado válido"
		}
	}
	for nombre, campo := range map[string]domain.Optional[string]{
		"contratado_en": input.ContratadoEn, "terminado_en": input.TerminadoEn,
	} {
		if campo.Set && campo.Value != nil && !fechaValida(*campo.Value) {
			campos[nombre] = "usa el formato AAAA-MM-DD"
		}
	}
	if len(campos) > 0 {
		return &ErrorValidacion{Campos: campos}
	}
	return nil
}

func validarNombreEmpleado(campos map[string]string, campo, valor string) {
	if longitud := len([]rune(valor)); longitud < 2 || longitud > 100 {
		campos[campo] = "usa entre 2 y 100 caracteres"
	}
}

func actualizacionEmpleadoTieneCampos(input domain.ActualizarEmpleadoInput) bool {
	return input.NumeroEmpleado.Set || input.Nombre.Set || input.SegundoNombre.Set ||
		input.PrimerApellido.Set || input.SegundoApellido.Set || input.Correo.Set ||
		input.Telefono.Set || input.Puesto.Set || input.Estado.Set ||
		input.ContratadoEn.Set || input.TerminadoEn.Set
}

func recortarOpcional(campo *domain.Optional[string]) {
	if !campo.Set || campo.Value == nil {
		return
	}
	valor := strings.TrimSpace(*campo.Value)
	if valor == "" {
		campo.Value = nil
		return
	}
	campo.Value = &valor
}

func fechaValida(valor string) bool {
	_, err := time.Parse("2006-01-02", valor)
	return err == nil
}

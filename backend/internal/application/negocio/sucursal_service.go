package negocio

import (
	"context"
	"errors"
	"regexp"
	"strings"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

var (
	ErrSucursalNoEncontrada       = errors.New("sucursal no encontrada")
	ErrSucursalProhibida          = errors.New("no tienes permiso para modificar esta sucursal")
	ErrSucursalSinCambios         = errors.New("no se recibieron campos para actualizar")
	ErrEstadoSucursal             = errors.New("la sucursal ya se encuentra en ese estado")
	ErrSucursalConflicto          = errors.New("ya existe una sucursal con ese código")
	ErrSucursalPrincipalRequerida = errors.New("asigna otra sucursal principal antes de archivar esta")
)

var codigoSucursalPattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{1,39}$`)

type SucursalRepository interface {
	ObtenerContextoNegocio(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.ContextoNegocioSucursal, error)
	PermisosEfectivos(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]string, error)
	Listar(ctx context.Context, negocioID uuid.UUID, estado, buscar string) ([]domain.SucursalResumen, error)
	Obtener(ctx context.Context, negocioID, sucursalID uuid.UUID) (domain.SucursalDetalle, error)
	Crear(ctx context.Context, negocioID uuid.UUID, input domain.CrearSucursalInput) (uuid.UUID, error)
	Actualizar(ctx context.Context, negocioID, sucursalID uuid.UUID, input domain.ActualizarSucursalInput) error
	Archivar(ctx context.Context, negocioID, sucursalID uuid.UUID) error
	Restaurar(ctx context.Context, negocioID, sucursalID uuid.UUID) error
}

type SucursalService struct {
	sucursales SucursalRepository
}

func NewSucursalService(sucursales SucursalRepository) *SucursalService {
	return &SucursalService{sucursales: sucursales}
}

func (s *SucursalService) Listar(ctx context.Context, usuarioID, negocioID uuid.UUID, estado, buscar string) ([]domain.SucursalResumen, error) {
	access, err := s.acceso(ctx, usuarioID, negocioID, "")
	if err != nil {
		return nil, err
	}
	if estado == "" {
		estado = "activo"
	}
	if estado != "activo" && estado != "archivado" {
		return nil, &ErrorValidacion{Campos: map[string]string{"estado": "debe ser activo o archivado"}}
	}
	items, err := s.sucursales.Listar(ctx, negocioID, estado, strings.TrimSpace(buscar))
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index].TipoMiembro = access.TipoMiembro
	}
	return items, nil
}

func (s *SucursalService) Obtener(ctx context.Context, usuarioID, negocioID, sucursalID uuid.UUID) (domain.SucursalDetalle, error) {
	access, err := s.acceso(ctx, usuarioID, negocioID, "")
	if err != nil {
		return domain.SucursalDetalle{}, err
	}
	detalle, err := s.sucursales.Obtener(ctx, negocioID, sucursalID)
	if err != nil {
		return domain.SucursalDetalle{}, err
	}
	detalle.TipoMiembro = access.TipoMiembro
	return detalle, nil
}

func (s *SucursalService) Crear(ctx context.Context, usuarioID, negocioID uuid.UUID, input domain.CrearSucursalInput) (domain.SucursalDetalle, error) {
	if _, err := s.acceso(ctx, usuarioID, negocioID, domain.PermisoSucursalCrear); err != nil {
		return domain.SucursalDetalle{}, err
	}
	normalizarCrearSucursal(&input)
	if campos := validarSucursal(input.Codigo, input.Nombre, input.Telefono, input.Direccion); len(campos) > 0 {
		return domain.SucursalDetalle{}, &ErrorValidacion{Campos: campos}
	}
	id, err := s.sucursales.Crear(ctx, negocioID, input)
	if err != nil {
		return domain.SucursalDetalle{}, err
	}
	return s.Obtener(ctx, usuarioID, negocioID, id)
}

func (s *SucursalService) Actualizar(ctx context.Context, usuarioID, negocioID, sucursalID uuid.UUID, input domain.ActualizarSucursalInput) (domain.SucursalDetalle, error) {
	if _, err := s.acceso(ctx, usuarioID, negocioID, domain.PermisoSucursalEditar); err != nil {
		return domain.SucursalDetalle{}, err
	}
	if !actualizacionSucursalTieneCampos(input) {
		return domain.SucursalDetalle{}, ErrSucursalSinCambios
	}
	normalizarActualizacionSucursal(&input)
	if campos := validarActualizacionSucursal(input); len(campos) > 0 {
		return domain.SucursalDetalle{}, &ErrorValidacion{Campos: campos}
	}
	if err := s.sucursales.Actualizar(ctx, negocioID, sucursalID, input); err != nil {
		return domain.SucursalDetalle{}, err
	}
	return s.Obtener(ctx, usuarioID, negocioID, sucursalID)
}

func (s *SucursalService) Archivar(ctx context.Context, usuarioID, negocioID, sucursalID uuid.UUID) error {
	if _, err := s.acceso(ctx, usuarioID, negocioID, domain.PermisoSucursalArchivar); err != nil {
		return err
	}
	return s.sucursales.Archivar(ctx, negocioID, sucursalID)
}

func (s *SucursalService) Restaurar(ctx context.Context, usuarioID, negocioID, sucursalID uuid.UUID) (domain.SucursalDetalle, error) {
	if _, err := s.acceso(ctx, usuarioID, negocioID, domain.PermisoSucursalArchivar); err != nil {
		return domain.SucursalDetalle{}, err
	}
	if err := s.sucursales.Restaurar(ctx, negocioID, sucursalID); err != nil {
		return domain.SucursalDetalle{}, err
	}
	return s.Obtener(ctx, usuarioID, negocioID, sucursalID)
}

// acceso exige membresía activa para leer y, para escribir, negocio activo más el permiso indicado.
//
// Desde fase 4 la autorización ya no depende de `tipo_miembro`: el rol de sistema del propietario
// concentra todos los permisos, y un miembro puede recibirlos mediante roles del negocio.
func (s *SucursalService) acceso(ctx context.Context, usuarioID, negocioID uuid.UUID, permiso string) (domain.ContextoNegocioSucursal, error) {
	access, err := s.sucursales.ObtenerContextoNegocio(ctx, usuarioID, negocioID)
	if err != nil {
		return domain.ContextoNegocioSucursal{}, err
	}
	if permiso == "" {
		return access, nil
	}
	if access.EstadoNegocio != "activo" {
		return domain.ContextoNegocioSucursal{}, ErrEstadoNegocio
	}
	permisos, err := s.sucursales.PermisosEfectivos(ctx, usuarioID, negocioID)
	if err != nil {
		return domain.ContextoNegocioSucursal{}, err
	}
	for _, actual := range permisos {
		if actual == permiso {
			return access, nil
		}
	}
	return domain.ContextoNegocioSucursal{}, ErrSucursalProhibida
}

func normalizarCrearSucursal(input *domain.CrearSucursalInput) {
	input.Codigo = strings.ToUpper(strings.TrimSpace(input.Codigo))
	input.Nombre = strings.TrimSpace(input.Nombre)
	input.Telefono = limpiarOpcional(input.Telefono, false)
	if input.Direccion != nil {
		normalizarDireccion(input.Direccion)
		if !direccionTieneDatos(*input.Direccion) {
			input.Direccion = nil
		}
	}
}

func validarSucursal(codigo, nombre string, telefono *string, direccion *domain.DireccionInput) map[string]string {
	errores := map[string]string{}
	if !codigoSucursalPattern.MatchString(codigo) {
		errores["codigo"] = "usa de 2 a 40 caracteres: letras, números, guion o guion bajo"
	}
	validarLongitud(nombre, 2, 160, "nombre", errores)
	validarPunteroLongitud(telefono, 30, "telefono", errores)
	if direccion != nil {
		validarDireccion(*direccion, errores)
	}
	return errores
}

func actualizacionSucursalTieneCampos(input domain.ActualizarSucursalInput) bool {
	return input.Nombre.Set || input.Telefono.Set || input.EsPrincipal.Set || input.Direccion.Set
}

func normalizarActualizacionSucursal(input *domain.ActualizarSucursalInput) {
	normalizarOptional(&input.Nombre, false)
	normalizarOptional(&input.Telefono, false)
	if input.Direccion.Set && input.Direccion.Value != nil {
		normalizarActualizacionDireccion(input.Direccion.Value)
	}
}

func validarActualizacionSucursal(input domain.ActualizarSucursalInput) map[string]string {
	errores := map[string]string{}
	if input.Nombre.Set {
		if input.Nombre.Value == nil {
			errores["nombre"] = "es obligatorio"
		} else {
			validarLongitud(*input.Nombre.Value, 2, 160, "nombre", errores)
		}
	}
	if input.Telefono.Set {
		validarPunteroLongitud(input.Telefono.Value, 30, "telefono", errores)
	}
	if input.EsPrincipal.Set && input.EsPrincipal.Value == nil {
		errores["es_principal"] = "no puede ser null"
	}
	if input.Direccion.Set && input.Direccion.Value != nil {
		validarDireccionActualizada(*input.Direccion.Value, errores)
	}
	return errores
}

func validarDireccionActualizada(input domain.ActualizarDireccionInput, errores map[string]string) {
	if input.CodigoPais.Set {
		if input.CodigoPais.Value == nil || !regexp.MustCompile(`^[A-Z]{2}$`).MatchString(*input.CodigoPais.Value) {
			errores["direccion.codigo_pais"] = "debe contener dos letras mayúsculas"
		}
	}
	validarOptionalLongitud(input.Estado, 120, "direccion.estado", errores)
	validarOptionalLongitud(input.Municipio, 120, "direccion.municipio", errores)
	validarOptionalLongitud(input.Ciudad, 120, "direccion.ciudad", errores)
	validarOptionalLongitud(input.Colonia, 150, "direccion.colonia", errores)
	validarOptionalLongitud(input.CodigoPostal, 12, "direccion.codigo_postal", errores)
	validarOptionalLongitud(input.Calle, 180, "direccion.calle", errores)
	validarOptionalLongitud(input.NumeroExterior, 30, "direccion.numero_exterior", errores)
	validarOptionalLongitud(input.NumeroInterior, 30, "direccion.numero_interior", errores)
}

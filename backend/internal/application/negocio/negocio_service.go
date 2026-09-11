package negocio

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

var (
	ErrNegocioNoEncontrado = errors.New("negocio no encontrado")
	ErrNegocioProhibido    = errors.New("no tienes permiso para modificar este negocio")
	ErrNegocioSinCambios   = errors.New("no se recibieron campos para actualizar")
	ErrEstadoNegocio       = errors.New("el negocio ya se encuentra en ese estado")
	ErrNegocioConflicto    = errors.New("ya existe un negocio con esos datos")
)

type ErrorValidacion struct {
	Campos map[string]string
}

func (e *ErrorValidacion) Error() string { return "los datos del negocio no son válidos" }

type NegocioRepository interface {
	Listar(ctx context.Context, usuarioID uuid.UUID, estado string) ([]domain.NegocioResumen, error)
	ObtenerAccesible(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.NegocioDetalle, error)
	ExisteSlug(ctx context.Context, slug string) (bool, error)
	Crear(ctx context.Context, usuarioID uuid.UUID, slug string, input domain.CrearNegocioInput) (domain.NegocioDetalle, error)
	Actualizar(ctx context.Context, usuarioID, negocioID uuid.UUID, input domain.ActualizarNegocioInput) error
	Archivar(ctx context.Context, usuarioID, negocioID uuid.UUID) error
	Restaurar(ctx context.Context, usuarioID, negocioID uuid.UUID) error
}

type NegocioService struct {
	negocios NegocioRepository
}

func NewNegocioService(negocios NegocioRepository) *NegocioService {
	return &NegocioService{negocios: negocios}
}

func (s *NegocioService) Listar(ctx context.Context, usuarioID uuid.UUID, estado string) ([]domain.NegocioResumen, error) {
	if estado == "" {
		estado = "activo"
	}
	if estado != "activo" && estado != "archivado" {
		return nil, &ErrorValidacion{Campos: map[string]string{"estado": "debe ser activo o archivado"}}
	}
	return s.negocios.Listar(ctx, usuarioID, estado)
}

func (s *NegocioService) Obtener(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.NegocioDetalle, error) {
	return s.negocios.ObtenerAccesible(ctx, usuarioID, negocioID)
}

func (s *NegocioService) Crear(ctx context.Context, usuarioID uuid.UUID, input domain.CrearNegocioInput) (domain.NegocioDetalle, error) {
	validarCrearNegocio(&input)
	if campos := validarNegocio(input.NombreComercial, input.RazonSocial, input.RFC, input.Telefono, input.Correo, input.CodigoMoneda, input.ZonaHoraria, input.Direccion); len(campos) > 0 {
		return domain.NegocioDetalle{}, &ErrorValidacion{Campos: campos}
	}

	base := slugNegocio(input.NombreComercial)
	slug := base
	existe, err := s.negocios.ExisteSlug(ctx, slug)
	if err != nil {
		return domain.NegocioDetalle{}, err
	}
	if existe {
		slug = fmt.Sprintf("%s-%s", base, strings.ReplaceAll(uuid.NewString(), "-", "")[:8])
	}
	return s.negocios.Crear(ctx, usuarioID, slug, input)
}

func (s *NegocioService) Actualizar(ctx context.Context, usuarioID, negocioID uuid.UUID, input domain.ActualizarNegocioInput) (domain.NegocioDetalle, error) {
	detalle, err := s.Obtener(ctx, usuarioID, negocioID)
	if err != nil {
		return domain.NegocioDetalle{}, err
	}
	if detalle.TipoMiembro != "propietario" {
		return domain.NegocioDetalle{}, ErrNegocioProhibido
	}
	if !actualizacionTieneCampos(input) {
		return domain.NegocioDetalle{}, ErrNegocioSinCambios
	}
	normalizarActualizacion(&input)
	if campos := validarActualizacion(input); len(campos) > 0 {
		return domain.NegocioDetalle{}, &ErrorValidacion{Campos: campos}
	}
	if err := s.negocios.Actualizar(ctx, usuarioID, negocioID, input); err != nil {
		return domain.NegocioDetalle{}, err
	}
	return s.Obtener(ctx, usuarioID, negocioID)
}

func (s *NegocioService) Archivar(ctx context.Context, usuarioID, negocioID uuid.UUID) error {
	detalle, err := s.Obtener(ctx, usuarioID, negocioID)
	if err != nil {
		return err
	}
	if detalle.TipoMiembro != "propietario" {
		return ErrNegocioProhibido
	}
	if detalle.Estado == "archivado" {
		return ErrEstadoNegocio
	}
	return s.negocios.Archivar(ctx, usuarioID, negocioID)
}

func (s *NegocioService) Restaurar(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.NegocioDetalle, error) {
	detalle, err := s.Obtener(ctx, usuarioID, negocioID)
	if err != nil {
		return domain.NegocioDetalle{}, err
	}
	if detalle.TipoMiembro != "propietario" {
		return domain.NegocioDetalle{}, ErrNegocioProhibido
	}
	if detalle.Estado == "activo" {
		return domain.NegocioDetalle{}, ErrEstadoNegocio
	}
	if err := s.negocios.Restaurar(ctx, usuarioID, negocioID); err != nil {
		return domain.NegocioDetalle{}, err
	}
	return s.Obtener(ctx, usuarioID, negocioID)
}

func validarCrearNegocio(input *domain.CrearNegocioInput) {
	input.NombreComercial = strings.TrimSpace(input.NombreComercial)
	input.RazonSocial = limpiarOpcional(input.RazonSocial, false)
	input.RFC = limpiarOpcional(input.RFC, true)
	input.Telefono = limpiarOpcional(input.Telefono, false)
	input.Correo = limpiarOpcional(input.Correo, true)
	input.CodigoMoneda = strings.ToUpper(strings.TrimSpace(input.CodigoMoneda))
	if input.CodigoMoneda == "" {
		input.CodigoMoneda = "MXN"
	}
	input.ZonaHoraria = strings.TrimSpace(input.ZonaHoraria)
	if input.ZonaHoraria == "" {
		input.ZonaHoraria = "America/Mexico_City"
	}
	if input.Direccion != nil {
		normalizarDireccion(input.Direccion)
		if !direccionTieneDatos(*input.Direccion) {
			input.Direccion = nil
		}
	}
}

func validarNegocio(nombre string, razon, rfc, telefono, correo *string, moneda, zona string, direccion *domain.DireccionInput) map[string]string {
	errores := map[string]string{}
	validarLongitud(nombre, 2, 180, "nombre_comercial", errores)
	validarPunteroLongitud(razon, 220, "razon_social", errores)
	validarPunteroLongitud(telefono, 30, "telefono", errores)
	if rfc != nil && !regexp.MustCompile(`^[A-ZÑ&]{3,4}[0-9]{6}[A-Z0-9]{3}$`).MatchString(*rfc) {
		errores["rfc"] = "debe contener 12 o 13 caracteres con formato válido"
	}
	if correo != nil {
		if len(*correo) > 254 {
			errores["correo"] = "no puede superar 254 caracteres"
		} else if _, err := mail.ParseAddress(*correo); err != nil {
			errores["correo"] = "debe ser un correo válido"
		}
	}
	if !regexp.MustCompile(`^[A-Z]{3}$`).MatchString(moneda) {
		errores["codigo_moneda"] = "debe contener tres letras mayúsculas"
	}
	if _, err := time.LoadLocation(zona); err != nil {
		errores["zona_horaria"] = "debe ser una zona horaria IANA válida"
	}
	if direccion != nil {
		validarDireccion(*direccion, errores)
	}
	return errores
}

func validarActualizacion(input domain.ActualizarNegocioInput) map[string]string {
	errores := map[string]string{}
	if input.NombreComercial.Set {
		if input.NombreComercial.Value == nil {
			errores["nombre_comercial"] = "no puede ser null"
		} else {
			validarLongitud(*input.NombreComercial.Value, 2, 180, "nombre_comercial", errores)
		}
	}
	validarOptionalLongitud(input.RazonSocial, 220, "razon_social", errores)
	validarOptionalLongitud(input.Telefono, 30, "telefono", errores)
	if input.RFC.Set && input.RFC.Value != nil && !regexp.MustCompile(`^[A-ZÑ&]{3,4}[0-9]{6}[A-Z0-9]{3}$`).MatchString(*input.RFC.Value) {
		errores["rfc"] = "debe contener 12 o 13 caracteres con formato válido"
	}
	if input.Correo.Set && input.Correo.Value != nil {
		if len(*input.Correo.Value) > 254 {
			errores["correo"] = "no puede superar 254 caracteres"
		} else if _, err := mail.ParseAddress(*input.Correo.Value); err != nil {
			errores["correo"] = "debe ser un correo válido"
		}
	}
	if input.CodigoMoneda.Set {
		if input.CodigoMoneda.Value == nil || !regexp.MustCompile(`^[A-Z]{3}$`).MatchString(*input.CodigoMoneda.Value) {
			errores["codigo_moneda"] = "debe contener tres letras mayúsculas"
		}
	}
	if input.ZonaHoraria.Set {
		if input.ZonaHoraria.Value == nil {
			errores["zona_horaria"] = "no puede ser null"
		} else if _, err := time.LoadLocation(*input.ZonaHoraria.Value); err != nil {
			errores["zona_horaria"] = "debe ser una zona horaria IANA válida"
		}
	}
	if input.Direccion.Set && input.Direccion.Value != nil {
		direccion := input.Direccion.Value
		if direccion.CodigoPais.Set {
			if direccion.CodigoPais.Value == nil || !regexp.MustCompile(`^[A-Z]{2}$`).MatchString(*direccion.CodigoPais.Value) {
				errores["direccion.codigo_pais"] = "debe contener dos letras mayúsculas"
			}
		}
		validarOptionalLongitud(direccion.Estado, 120, "direccion.estado", errores)
		validarOptionalLongitud(direccion.Municipio, 120, "direccion.municipio", errores)
		validarOptionalLongitud(direccion.Ciudad, 120, "direccion.ciudad", errores)
		validarOptionalLongitud(direccion.Colonia, 150, "direccion.colonia", errores)
		validarOptionalLongitud(direccion.CodigoPostal, 12, "direccion.codigo_postal", errores)
		validarOptionalLongitud(direccion.Calle, 180, "direccion.calle", errores)
		validarOptionalLongitud(direccion.NumeroExterior, 30, "direccion.numero_exterior", errores)
		validarOptionalLongitud(direccion.NumeroInterior, 30, "direccion.numero_interior", errores)
	}
	return errores
}

func normalizarActualizacion(input *domain.ActualizarNegocioInput) {
	normalizarOptional(&input.NombreComercial, false)
	normalizarOptional(&input.RazonSocial, false)
	normalizarOptional(&input.RFC, true)
	normalizarOptional(&input.Telefono, false)
	normalizarOptional(&input.Correo, true)
	normalizarOptional(&input.CodigoMoneda, true)
	normalizarOptional(&input.ZonaHoraria, false)
	if input.Direccion.Set && input.Direccion.Value != nil {
		normalizarActualizacionDireccion(input.Direccion.Value)
	}
}

func normalizarDireccion(input *domain.DireccionInput) {
	input.CodigoPais = strings.ToUpper(strings.TrimSpace(input.CodigoPais))
	if input.CodigoPais == "" {
		input.CodigoPais = "MX"
	}
	input.Estado = limpiarOpcional(input.Estado, false)
	input.Municipio = limpiarOpcional(input.Municipio, false)
	input.Ciudad = limpiarOpcional(input.Ciudad, false)
	input.Colonia = limpiarOpcional(input.Colonia, false)
	input.CodigoPostal = limpiarOpcional(input.CodigoPostal, false)
	input.Calle = limpiarOpcional(input.Calle, false)
	input.NumeroExterior = limpiarOpcional(input.NumeroExterior, false)
	input.NumeroInterior = limpiarOpcional(input.NumeroInterior, false)
	input.Referencias = limpiarOpcional(input.Referencias, false)
}

func validarDireccion(input domain.DireccionInput, errores map[string]string) {
	if !regexp.MustCompile(`^[A-Z]{2}$`).MatchString(input.CodigoPais) {
		errores["direccion.codigo_pais"] = "debe contener dos letras mayúsculas"
	}
	validarPunteroLongitud(input.Estado, 120, "direccion.estado", errores)
	validarPunteroLongitud(input.Municipio, 120, "direccion.municipio", errores)
	validarPunteroLongitud(input.Ciudad, 120, "direccion.ciudad", errores)
	validarPunteroLongitud(input.Colonia, 150, "direccion.colonia", errores)
	validarPunteroLongitud(input.CodigoPostal, 12, "direccion.codigo_postal", errores)
	validarPunteroLongitud(input.Calle, 180, "direccion.calle", errores)
	validarPunteroLongitud(input.NumeroExterior, 30, "direccion.numero_exterior", errores)
	validarPunteroLongitud(input.NumeroInterior, 30, "direccion.numero_interior", errores)
}

func direccionTieneDatos(input domain.DireccionInput) bool {
	return input.Estado != nil || input.Municipio != nil || input.Ciudad != nil || input.Colonia != nil ||
		input.CodigoPostal != nil || input.Calle != nil || input.NumeroExterior != nil ||
		input.NumeroInterior != nil || input.Referencias != nil
}

func actualizacionTieneCampos(input domain.ActualizarNegocioInput) bool {
	return input.NombreComercial.Set || input.RazonSocial.Set || input.RFC.Set || input.Telefono.Set ||
		input.Correo.Set || input.CodigoMoneda.Set || input.ZonaHoraria.Set || input.Direccion.Set
}

func limpiarOpcional(value *string, mayusculasOMinusculas bool) *string {
	if value == nil {
		return nil
	}
	clean := strings.TrimSpace(*value)
	if clean == "" {
		return nil
	}
	if mayusculasOMinusculas {
		if strings.Contains(clean, "@") {
			clean = strings.ToLower(clean)
		} else {
			clean = strings.ToUpper(clean)
		}
	}
	return &clean
}

func normalizarOptional(value *domain.Optional[string], mayusculasOMinusculas bool) {
	if !value.Set || value.Value == nil {
		return
	}
	value.Value = limpiarOpcional(value.Value, mayusculasOMinusculas)
}

func normalizarActualizacionDireccion(input *domain.ActualizarDireccionInput) {
	values := []*domain.Optional[string]{&input.Estado, &input.Municipio, &input.Ciudad, &input.Colonia, &input.CodigoPostal, &input.Calle, &input.NumeroExterior, &input.NumeroInterior, &input.Referencias}
	for _, value := range values {
		normalizarOptional(value, false)
	}
	normalizarOptional(&input.CodigoPais, true)
}

func validarLongitud(value string, min, max int, field string, errores map[string]string) {
	length := len([]rune(value))
	if length < min || length > max {
		errores[field] = fmt.Sprintf("debe contener entre %d y %d caracteres", min, max)
	}
}

func validarPunteroLongitud(value *string, max int, field string, errores map[string]string) {
	if value != nil && len([]rune(*value)) > max {
		errores[field] = fmt.Sprintf("no puede superar %d caracteres", max)
	}
}

func validarOptionalLongitud(value domain.Optional[string], max int, field string, errores map[string]string) {
	if value.Set {
		validarPunteroLongitud(value.Value, max, field, errores)
	}
}

func slugNegocio(value string) string {
	normalized := norm.NFD.String(strings.ToLower(strings.TrimSpace(value)))
	var builder strings.Builder
	lastDash := false
	for _, current := range normalized {
		if unicode.Is(unicode.Mn, current) {
			continue
		}
		if current >= 'a' && current <= 'z' || current >= '0' && current <= '9' {
			builder.WriteRune(current)
			lastDash = false
		} else if !lastDash && builder.Len() > 0 {
			builder.WriteByte('-')
			lastDash = true
		}
	}
	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		return "negocio"
	}
	if len(slug) > 110 {
		slug = strings.TrimRight(slug[:110], "-")
	}
	return slug
}

package negocio

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

var (
	ErrInvitacionNoEncontrada   = errors.New("invitación no encontrada")
	ErrInvitacionProhibida      = errors.New("no tienes permiso para realizar esta acción")
	ErrInvitacionNoVigente      = errors.New("la invitación ya no está vigente")
	ErrInvitacionCorreoRequerido = errors.New("el empleado necesita un correo para poder ser invitado")
	ErrInvitacionYaVinculado    = errors.New("el empleado ya tiene una cuenta vinculada")
	ErrInvitacionCorreoDistinto = errors.New("la invitación fue emitida para otro correo")
	ErrInvitacionRolAjeno       = errors.New("el rol no pertenece a este negocio")
)

type InvitacionRepository interface {
	ObtenerContextoNegocio(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.ContextoNegocioSucursal, error)
	PermisosEfectivos(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]string, error)
	ObtenerEmpleado(ctx context.Context, negocioID, empleadoID uuid.UUID) (domain.EmpleadoDetalle, error)
	RolPerteneceANegocio(ctx context.Context, negocioID, rolID uuid.UUID) (bool, error)
	Listar(ctx context.Context, negocioID uuid.UUID, estado string) ([]domain.InvitacionResumen, error)
	Crear(ctx context.Context, invitacion domain.InvitacionNegocio) (domain.InvitacionResumen, error)
	CancelarPendientesDeEmpleado(ctx context.Context, negocioID, empleadoID uuid.UUID) error
	Cancelar(ctx context.Context, negocioID, invitacionID uuid.UUID) error
	ObtenerPorHash(ctx context.Context, hash string) (domain.InvitacionNegocio, error)
	DatosPublicos(ctx context.Context, invitacionID uuid.UUID) (domain.InvitacionPublica, error)
	CorreoDeUsuario(ctx context.Context, usuarioID uuid.UUID) (string, error)
	ExisteCuentaConCorreo(ctx context.Context, correo string) (bool, error)
	Aceptar(ctx context.Context, invitacion domain.InvitacionNegocio, usuarioID uuid.UUID) error
	MarcarExpiradas(ctx context.Context, negocioID uuid.UUID) error
}

type InvitacionService struct {
	invitaciones InvitacionRepository
}

func NewInvitacionService(invitaciones InvitacionRepository) *InvitacionService {
	return &InvitacionService{invitaciones: invitaciones}
}

func (s *InvitacionService) Listar(ctx context.Context, usuarioID, negocioID uuid.UUID, estado string) ([]domain.InvitacionResumen, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoInvitacionVer, false); err != nil {
		return nil, err
	}
	if err := s.invitaciones.MarcarExpiradas(ctx, negocioID); err != nil {
		return nil, err
	}
	return s.invitaciones.Listar(ctx, negocioID, strings.TrimSpace(estado))
}

// Crear emite un token nuevo e invalida cualquier invitación pendiente previa del mismo empleado.
func (s *InvitacionService) Crear(ctx context.Context, usuarioID, negocioID uuid.UUID, input domain.CrearInvitacionInput) (domain.InvitacionCreada, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoInvitacionEnviar, true); err != nil {
		return domain.InvitacionCreada{}, err
	}
	empleado, err := s.invitaciones.ObtenerEmpleado(ctx, negocioID, input.EmpleadoID)
	if err != nil {
		return domain.InvitacionCreada{}, err
	}
	if empleado.MembresiaID != nil {
		return domain.InvitacionCreada{}, ErrInvitacionYaVinculado
	}
	if empleado.Correo == nil || strings.TrimSpace(*empleado.Correo) == "" {
		return domain.InvitacionCreada{}, ErrInvitacionCorreoRequerido
	}
	if input.RolPredeterminadoID != nil {
		pertenece, err := s.invitaciones.RolPerteneceANegocio(ctx, negocioID, *input.RolPredeterminadoID)
		if err != nil {
			return domain.InvitacionCreada{}, err
		}
		if !pertenece {
			return domain.InvitacionCreada{}, ErrInvitacionRolAjeno
		}
	}

	token, hash, err := generarTokenInvitacion()
	if err != nil {
		return domain.InvitacionCreada{}, err
	}
	if err := s.invitaciones.CancelarPendientesDeEmpleado(ctx, negocioID, empleado.ID); err != nil {
		return domain.InvitacionCreada{}, err
	}
	resumen, err := s.invitaciones.Crear(ctx, domain.InvitacionNegocio{
		NegocioID: negocioID, EmpleadoID: &empleado.ID,
		Correo:              strings.ToLower(strings.TrimSpace(*empleado.Correo)),
		RolPredeterminadoID: input.RolPredeterminadoID,
		HashToken:           hash, Estado: domain.EstadoInvitacionPendiente,
		InvitadoPorUsuarioID: usuarioID,
		ExpiraEn:             time.Now().Add(domain.DuracionInvitacion),
	})
	if err != nil {
		return domain.InvitacionCreada{}, err
	}
	return domain.InvitacionCreada{Invitacion: resumen, Token: token}, nil
}

func (s *InvitacionService) Cancelar(ctx context.Context, usuarioID, negocioID, invitacionID uuid.UUID) error {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoInvitacionEnviar, true); err != nil {
		return err
	}
	return s.invitaciones.Cancelar(ctx, negocioID, invitacionID)
}

// Consultar resuelve el token público sin exigir sesión, para que el invitado sepa qué debe hacer.
func (s *InvitacionService) Consultar(ctx context.Context, token string) (domain.InvitacionPublica, error) {
	invitacion, err := s.invitacionVigente(ctx, token)
	if err != nil {
		return domain.InvitacionPublica{}, err
	}
	publica, err := s.invitaciones.DatosPublicos(ctx, invitacion.ID)
	if err != nil {
		return domain.InvitacionPublica{}, err
	}
	existe, err := s.invitaciones.ExisteCuentaConCorreo(ctx, invitacion.Correo)
	if err != nil {
		return domain.InvitacionPublica{}, err
	}
	publica.RequiereCuenta = !existe
	return publica, nil
}

// Aceptar exige sesión iniciada con el mismo correo invitado y crea membresía y roles de forma transaccional.
func (s *InvitacionService) Aceptar(ctx context.Context, usuarioID uuid.UUID, token string) error {
	invitacion, err := s.invitacionVigente(ctx, token)
	if err != nil {
		return err
	}
	correo, err := s.invitaciones.CorreoDeUsuario(ctx, usuarioID)
	if err != nil {
		return err
	}
	if !strings.EqualFold(strings.TrimSpace(correo), invitacion.Correo) {
		return ErrInvitacionCorreoDistinto
	}
	return s.invitaciones.Aceptar(ctx, invitacion, usuarioID)
}

func (s *InvitacionService) invitacionVigente(ctx context.Context, token string) (domain.InvitacionNegocio, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.InvitacionNegocio{}, ErrInvitacionNoEncontrada
	}
	invitacion, err := s.invitaciones.ObtenerPorHash(ctx, hashDeToken(token))
	if err != nil {
		return domain.InvitacionNegocio{}, err
	}
	if invitacion.Estado != domain.EstadoInvitacionPendiente {
		return domain.InvitacionNegocio{}, ErrInvitacionNoVigente
	}
	if invitacion.ExpiraEn.Before(time.Now()) {
		return domain.InvitacionNegocio{}, ErrInvitacionNoVigente
	}
	return invitacion, nil
}

func (s *InvitacionService) autorizar(ctx context.Context, usuarioID, negocioID uuid.UUID, permiso string, escritura bool) error {
	access, err := s.invitaciones.ObtenerContextoNegocio(ctx, usuarioID, negocioID)
	if err != nil {
		return err
	}
	if escritura && access.EstadoNegocio != "activo" {
		return ErrEstadoNegocio
	}
	permisos, err := s.invitaciones.PermisosEfectivos(ctx, usuarioID, negocioID)
	if err != nil {
		return err
	}
	for _, actual := range permisos {
		if actual == permiso {
			return nil
		}
	}
	return ErrInvitacionProhibida
}

// generarTokenInvitacion produce el valor en claro y el hash que sí se persiste.
func generarTokenInvitacion() (string, string, error) {
	crudo := make([]byte, 32)
	if _, err := rand.Read(crudo); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(crudo)
	return token, hashDeToken(token), nil
}

func hashDeToken(token string) string {
	suma := sha256.Sum256([]byte(token))
	return hex.EncodeToString(suma[:])
}

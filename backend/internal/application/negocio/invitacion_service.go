package negocio

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/mail"
	"strings"
	"time"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

var (
	ErrInvitacionNoEncontrada         = errors.New("invitación no encontrada")
	ErrInvitacionProhibida            = errors.New("no tienes permiso para realizar esta acción")
	ErrInvitacionNoVigente            = errors.New("la invitación ya no está vigente")
	ErrInvitacionYaAceptada           = errors.New("la invitación ya fue aceptada")
	ErrInvitacionReemplazada          = errors.New("la invitación ya fue reemplazada por una más reciente")
	ErrInvitacionSinSucursal          = errors.New("la invitación no tiene sucursal; emite una nueva")
	ErrInvitacionCorreoRequerido      = errors.New("el empleado necesita un correo para poder ser invitado")
	ErrInvitacionYaVinculado          = errors.New("el empleado ya tiene una cuenta vinculada")
	ErrInvitacionCorreoDistinto       = errors.New("la invitación fue emitida para otro correo")
	ErrInvitacionRolNoDisponible      = errors.New("el rol no está disponible para invitaciones de este negocio")
	ErrInvitacionSucursalNoDisponible = errors.New("la sucursal no está activa en este negocio")
	ErrInvitacionEmpleadoNoElegible   = errors.New("el empleado ya no puede recibir acceso")
	ErrInvitacionMembresiaInactiva    = errors.New("tu acceso a este negocio está suspendido; contacta al administrador")
	ErrInvitacionCuentaExistente      = errors.New("ya existe una cuenta con el correo invitado; inicia sesión")
	ErrInvitacionLimite               = errors.New("demasiados intentos; espera un momento")
	// ErrInvitacionEnvioCodigo conserva la cuenta pendiente: el invitado puede pedir otro código.
	ErrInvitacionEnvioCodigo = errors.New("no fue posible enviar el código de verificación")
)

type InvitacionRepository interface {
	ObtenerContextoNegocio(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.ContextoNegocioSucursal, error)
	PermisosEfectivos(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]string, error)
	ObtenerEmpleado(ctx context.Context, negocioID, empleadoID uuid.UUID) (domain.EmpleadoDetalle, error)
	CodigoRolActivo(ctx context.Context, negocioID, rolID uuid.UUID) (string, bool, error)
	RolesDelegables(ctx context.Context, usuarioID, negocioID uuid.UUID, roles []uuid.UUID, membresiaDestino *uuid.UUID) (bool, error)
	SucursalActivaDelNegocio(ctx context.Context, negocioID, sucursalID uuid.UUID) (bool, error)
	Listar(ctx context.Context, negocioID uuid.UUID, filtro domain.FiltroInvitaciones) ([]domain.InvitacionResumen, error)
	Emitir(ctx context.Context, invitacion domain.InvitacionNegocio) (domain.InvitacionResumen, error)
	Reemitir(ctx context.Context, negocioID, invitacionID, usuarioID uuid.UUID, correo *string, hash string, expiraEn time.Time) (domain.InvitacionResumen, error)
	Cancelar(ctx context.Context, negocioID, invitacionID uuid.UUID) error
	ObtenerPorHash(ctx context.Context, hash string) (domain.InvitacionNegocio, error)
	DatosPublicos(ctx context.Context, invitacionID uuid.UUID) (domain.InvitacionPublica, error)
	CorreoDeUsuario(ctx context.Context, usuarioID uuid.UUID) (string, error)
	ExisteCuentaVerificada(ctx context.Context, correo string) (bool, error)
	Aceptar(ctx context.Context, invitacionID, usuarioID uuid.UUID) (domain.InvitacionAceptada, error)
	MarcarExpiradas(ctx context.Context, negocioID uuid.UUID) error
}

// InvitacionMailer entrega el enlace de invitación; su implementación vive fuera de `negocio`.
type InvitacionMailer interface {
	EnviarInvitacion(ctx context.Context, para, nombreNegocio, nombreInvitado, token string, expiraEn time.Time) error
}

// DesafioRegistro es lo que el invitado necesita para verificar su correo con el OTP existente.
type DesafioRegistro struct {
	DesafioID          uuid.UUID `json:"desafio_id"`
	CorreoEnmascarado  string    `json:"correo_enmascarado"`
	ReenviarEnSegundos int       `json:"reenviar_en_segundos"`
}

// RegistroCuentaInvitacion es el contrato hacia `cuenta`: crea la cuenta pendiente y envía el OTP.
type RegistroCuentaInvitacion interface {
	RegistrarCuentaPendiente(ctx context.Context, nombres, primerApellido, segundoApellido, correo, telefono, contrasena, ip string) (DesafioRegistro, error)
}

type InvitacionService struct {
	invitaciones InvitacionRepository
	mailer       InvitacionMailer
	registro     RegistroCuentaInvitacion
	now          func() time.Time
}

// NewInvitacionService acepta mailer nil: la invitación sigue disponible como enlace copiable.
func NewInvitacionService(invitaciones InvitacionRepository, mailer InvitacionMailer, registro RegistroCuentaInvitacion) *InvitacionService {
	return &InvitacionService{invitaciones: invitaciones, mailer: mailer, registro: registro, now: time.Now}
}

func (s *InvitacionService) Listar(ctx context.Context, usuarioID, negocioID uuid.UUID, filtro domain.FiltroInvitaciones) ([]domain.InvitacionResumen, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, false, domain.PermisoInvitacionVer); err != nil {
		return nil, err
	}
	if err := s.invitaciones.MarcarExpiradas(ctx, negocioID); err != nil {
		return nil, err
	}
	filtro.Estado = strings.TrimSpace(filtro.Estado)
	return s.invitaciones.Listar(ctx, negocioID, filtro)
}

// Crear emite un token nuevo e invalida cualquier invitación pendiente previa del mismo empleado.
func (s *InvitacionService) Crear(ctx context.Context, usuarioID, negocioID uuid.UUID, input domain.CrearInvitacionInput) (domain.InvitacionCreada, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, true, domain.PermisoInvitacionEnviar); err != nil {
		return domain.InvitacionCreada{}, err
	}
	campos := map[string]string{}
	if input.EmpleadoID == uuid.Nil {
		campos["empleado_id"] = "es obligatorio"
	}
	if input.SucursalID == uuid.Nil {
		campos["sucursal_id"] = "es obligatoria"
	}
	if input.RolPredeterminadoID == uuid.Nil {
		campos["rol_predeterminado_id"] = "es obligatorio"
	}
	if len(campos) > 0 {
		return domain.InvitacionCreada{}, &ErrorValidacion{Campos: campos}
	}
	empleado, err := s.invitaciones.ObtenerEmpleado(ctx, negocioID, input.EmpleadoID)
	if err != nil {
		return domain.InvitacionCreada{}, err
	}
	if empleado.MembresiaID != nil {
		return domain.InvitacionCreada{}, ErrInvitacionYaVinculado
	}
	if empleado.Estado == domain.EstadoEmpleadoSuspendido || empleado.Estado == domain.EstadoEmpleadoTerminado {
		return domain.InvitacionCreada{}, ErrInvitacionEmpleadoNoElegible
	}
	if empleado.Correo == nil || strings.TrimSpace(*empleado.Correo) == "" {
		return domain.InvitacionCreada{}, ErrInvitacionCorreoRequerido
	}
	activa, err := s.invitaciones.SucursalActivaDelNegocio(ctx, negocioID, input.SucursalID)
	if err != nil {
		return domain.InvitacionCreada{}, err
	}
	if !activa {
		return domain.InvitacionCreada{}, ErrInvitacionSucursalNoDisponible
	}
	// El rol propietario nunca se delega mediante una invitación de empleado.
	codigoRol, existe, err := s.invitaciones.CodigoRolActivo(ctx, negocioID, input.RolPredeterminadoID)
	if err != nil {
		return domain.InvitacionCreada{}, err
	}
	if !existe || strings.EqualFold(codigoRol, domain.CodigoRolPropietario) {
		return domain.InvitacionCreada{}, ErrInvitacionRolNoDisponible
	}
	delegable, err := s.invitaciones.RolesDelegables(ctx, usuarioID, negocioID, []uuid.UUID{input.RolPredeterminadoID}, nil)
	if err != nil {
		return domain.InvitacionCreada{}, err
	}
	if !delegable {
		return domain.InvitacionCreada{}, ErrRolNoDelegable
	}

	token, hash, err := generarTokenInvitacion()
	if err != nil {
		return domain.InvitacionCreada{}, err
	}
	resumen, err := s.invitaciones.Emitir(ctx, domain.InvitacionNegocio{
		NegocioID: negocioID, EmpleadoID: &empleado.ID, SucursalID: &input.SucursalID,
		Correo:              strings.ToLower(strings.TrimSpace(*empleado.Correo)),
		RolPredeterminadoID: &input.RolPredeterminadoID,
		HashToken:           hash, Estado: domain.EstadoInvitacionPendiente,
		InvitadoPorUsuarioID: usuarioID,
		ExpiraEn:             s.now().UTC().Add(domain.DuracionInvitacion),
	})
	if err != nil {
		return domain.InvitacionCreada{}, err
	}
	return domain.InvitacionCreada{Invitacion: resumen, Token: token, CorreoEnviado: s.enviarCorreo(ctx, resumen, token)}, nil
}

// Reenviar invalida el enlace anterior y emite otro con 72 horas nuevas, conservando sucursal y rol.
//
// Con `Correo` presente también corrige el correo del empleado; exige además permiso de gestión.
func (s *InvitacionService) Reenviar(ctx context.Context, usuarioID, negocioID, invitacionID uuid.UUID, input domain.ReenviarInvitacionInput) (domain.InvitacionCreada, error) {
	permisos := []string{domain.PermisoInvitacionEnviar}
	var correo *string
	if input.Correo != nil {
		normalizado := strings.ToLower(strings.TrimSpace(*input.Correo))
		direccion, err := mail.ParseAddress(normalizado)
		if err != nil || direccion.Address != normalizado || len(normalizado) > 254 {
			return domain.InvitacionCreada{}, &ErrorValidacion{Campos: map[string]string{"correo": "no es un correo válido"}}
		}
		correo = &normalizado
		permisos = append(permisos, domain.PermisoEmpleadoGestionar)
	}
	if err := s.autorizar(ctx, usuarioID, negocioID, true, permisos...); err != nil {
		return domain.InvitacionCreada{}, err
	}
	token, hash, err := generarTokenInvitacion()
	if err != nil {
		return domain.InvitacionCreada{}, err
	}
	resumen, err := s.invitaciones.Reemitir(ctx, negocioID, invitacionID, usuarioID, correo, hash, s.now().UTC().Add(domain.DuracionInvitacion))
	if err != nil {
		return domain.InvitacionCreada{}, err
	}
	return domain.InvitacionCreada{Invitacion: resumen, Token: token, CorreoEnviado: s.enviarCorreo(ctx, resumen, token)}, nil
}

// enviarCorreo nunca revierte la invitación: si falla, el enlace copiable sigue siendo válido.
func (s *InvitacionService) enviarCorreo(ctx context.Context, resumen domain.InvitacionResumen, token string) bool {
	if s.mailer == nil {
		return false
	}
	publica, err := s.invitaciones.DatosPublicos(ctx, resumen.ID)
	if err != nil {
		return false
	}
	return s.mailer.EnviarInvitacion(ctx, resumen.Correo, publica.NombreNegocio, publica.NombreEmpleado, token, resumen.ExpiraEn) == nil
}

func (s *InvitacionService) Cancelar(ctx context.Context, usuarioID, negocioID, invitacionID uuid.UUID) error {
	if err := s.autorizar(ctx, usuarioID, negocioID, true, domain.PermisoInvitacionEnviar); err != nil {
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
	existe, err := s.invitaciones.ExisteCuentaVerificada(ctx, invitacion.Correo)
	if err != nil {
		return domain.InvitacionPublica{}, err
	}
	publica.CorreoEnmascarado = domain.EnmascararCorreo(invitacion.Correo)
	publica.RequiereCuenta = !existe
	return publica, nil
}

// Registrar crea la cuenta pendiente del invitado con el correo de la invitación; no la consume.
func (s *InvitacionService) Registrar(ctx context.Context, token string, input domain.RegistroInvitacionInput, ip string) (DesafioRegistro, error) {
	invitacion, err := s.invitacionVigente(ctx, token)
	if err != nil {
		return DesafioRegistro{}, err
	}
	nombres := strings.TrimSpace(input.Nombres)
	primerApellido := strings.TrimSpace(input.PrimerApellido)
	segundoApellido := strings.TrimSpace(input.SegundoApellido)
	campos := map[string]string{}
	if nombres == "" || len([]rune(nombres)) > 120 {
		campos["nombres"] = "es obligatorio y admite hasta 120 caracteres"
	}
	if primerApellido == "" || len([]rune(primerApellido)) > 120 {
		campos["primer_apellido"] = "es obligatorio y admite hasta 120 caracteres"
	}
	if len([]rune(segundoApellido)) > 120 {
		campos["segundo_apellido"] = "admite hasta 120 caracteres"
	}
	if len(input.Contrasena) < 8 || len(input.Contrasena) > 72 {
		campos["contrasena"] = "debe tener entre 8 y 72 caracteres"
	}
	if len([]rune(strings.TrimSpace(input.Telefono))) > 30 {
		campos["telefono"] = "admite hasta 30 caracteres"
	}
	if len(campos) > 0 {
		return DesafioRegistro{}, &ErrorValidacion{Campos: campos}
	}
	existe, err := s.invitaciones.ExisteCuentaVerificada(ctx, invitacion.Correo)
	if err != nil {
		return DesafioRegistro{}, err
	}
	if existe {
		return DesafioRegistro{}, ErrInvitacionCuentaExistente
	}
	if s.registro == nil {
		return DesafioRegistro{}, errors.New("registro de cuenta no configurado")
	}
	return s.registro.RegistrarCuentaPendiente(ctx, nombres, primerApellido, segundoApellido, invitacion.Correo, strings.TrimSpace(input.Telefono), input.Contrasena, ip)
}

// Aceptar exige sesión iniciada con el mismo correo invitado; el repositorio revalida bajo bloqueo.
func (s *InvitacionService) Aceptar(ctx context.Context, usuarioID uuid.UUID, token string) (domain.InvitacionAceptada, error) {
	invitacion, err := s.invitacionVigente(ctx, token)
	if err != nil {
		return domain.InvitacionAceptada{}, err
	}
	correo, err := s.invitaciones.CorreoDeUsuario(ctx, usuarioID)
	if err != nil {
		return domain.InvitacionAceptada{}, err
	}
	if !strings.EqualFold(strings.TrimSpace(correo), invitacion.Correo) {
		return domain.InvitacionAceptada{}, ErrInvitacionCorreoDistinto
	}
	return s.invitaciones.Aceptar(ctx, invitacion.ID, usuarioID)
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
	if invitacion.Estado == domain.EstadoInvitacionAceptada {
		return domain.InvitacionNegocio{}, ErrInvitacionYaAceptada
	}
	if invitacion.Estado != domain.EstadoInvitacionPendiente || !invitacion.ExpiraEn.After(s.now()) {
		return domain.InvitacionNegocio{}, ErrInvitacionNoVigente
	}
	return invitacion, nil
}

// autorizar exige todos los permisos indicados; corregir un correo requiere dos.
func (s *InvitacionService) autorizar(ctx context.Context, usuarioID, negocioID uuid.UUID, escritura bool, requeridos ...string) error {
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
	otorgados := make(map[string]struct{}, len(permisos))
	for _, actual := range permisos {
		otorgados[actual] = struct{}{}
	}
	for _, requerido := range requeridos {
		if _, ok := otorgados[requerido]; !ok {
			return ErrInvitacionProhibida
		}
	}
	return nil
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

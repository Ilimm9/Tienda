package negocio

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

type invitacionRepositoryStub struct {
	access         domain.ContextoNegocioSucursal
	permisos       []string
	empleado       domain.EmpleadoDetalle
	rolPertenece   bool
	invitacion     domain.InvitacionNegocio
	correoUsuario  string
	existeCuenta   bool
	creada         domain.InvitacionNegocio
	canceloPrevias bool
	aceptoLlamado  bool
	errorEmpleado  error
	errorHash      error
}

func (r *invitacionRepositoryStub) ObtenerContextoNegocio(context.Context, uuid.UUID, uuid.UUID) (domain.ContextoNegocioSucursal, error) {
	return r.access, nil
}
func (r *invitacionRepositoryStub) PermisosEfectivos(context.Context, uuid.UUID, uuid.UUID) ([]string, error) {
	return r.permisos, nil
}
func (r *invitacionRepositoryStub) ObtenerEmpleado(context.Context, uuid.UUID, uuid.UUID) (domain.EmpleadoDetalle, error) {
	return r.empleado, r.errorEmpleado
}
func (r *invitacionRepositoryStub) RolPerteneceANegocio(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return r.rolPertenece, nil
}
func (r *invitacionRepositoryStub) Listar(context.Context, uuid.UUID, string) ([]domain.InvitacionResumen, error) {
	return []domain.InvitacionResumen{}, nil
}
func (r *invitacionRepositoryStub) Crear(_ context.Context, invitacion domain.InvitacionNegocio) (domain.InvitacionResumen, error) {
	r.creada = invitacion
	return domain.InvitacionResumen{ID: uuid.New(), Correo: invitacion.Correo}, nil
}
func (r *invitacionRepositoryStub) CancelarPendientesDeEmpleado(context.Context, uuid.UUID, uuid.UUID) error {
	r.canceloPrevias = true
	return nil
}
func (r *invitacionRepositoryStub) Cancelar(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (r *invitacionRepositoryStub) ObtenerPorHash(context.Context, string) (domain.InvitacionNegocio, error) {
	if r.errorHash != nil {
		return domain.InvitacionNegocio{}, r.errorHash
	}
	return r.invitacion, nil
}
func (r *invitacionRepositoryStub) DatosPublicos(context.Context, uuid.UUID) (domain.InvitacionPublica, error) {
	return domain.InvitacionPublica{Correo: r.invitacion.Correo}, nil
}
func (r *invitacionRepositoryStub) CorreoDeUsuario(context.Context, uuid.UUID) (string, error) {
	return r.correoUsuario, nil
}
func (r *invitacionRepositoryStub) ExisteCuentaConCorreo(context.Context, string) (bool, error) {
	return r.existeCuenta, nil
}
func (r *invitacionRepositoryStub) Aceptar(context.Context, domain.InvitacionNegocio, uuid.UUID) error {
	r.aceptoLlamado = true
	return nil
}
func (r *invitacionRepositoryStub) MarcarExpiradas(context.Context, uuid.UUID) error { return nil }

func invitacionStub(permisos ...string) *invitacionRepositoryStub {
	correo := "ana@tienda.mx"
	return &invitacionRepositoryStub{
		access:   domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "miembro"},
		permisos: permisos,
		empleado: domain.EmpleadoDetalle{ID: uuid.New(), Correo: &correo},
	}
}

func TestInvitacionServiceCrearExigePermiso(t *testing.T) {
	service := NewInvitacionService(invitacionStub(domain.PermisoInvitacionVer))

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(),
		domain.CrearInvitacionInput{EmpleadoID: uuid.New()})

	if !errors.Is(err, ErrInvitacionProhibida) {
		t.Fatalf("se requiere permiso de envío: %v", err)
	}
}

func TestInvitacionServiceCrearGuardaSoloElHash(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	service := NewInvitacionService(repository)

	creada, err := service.Crear(context.Background(), uuid.New(), uuid.New(),
		domain.CrearInvitacionInput{EmpleadoID: repository.empleado.ID})

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if creada.Token == "" {
		t.Fatal("el token en claro debe entregarse una sola vez")
	}
	if repository.creada.HashToken == creada.Token {
		t.Fatal("nunca debe persistirse el token en claro")
	}
	if len(repository.creada.HashToken) != 64 {
		t.Fatalf("se esperaba un hash SHA-256 hexadecimal: %q", repository.creada.HashToken)
	}
	if !repository.canceloPrevias {
		t.Fatal("emitir un token nuevo debe invalidar las invitaciones pendientes previas")
	}
	if repository.creada.ExpiraEn.Before(time.Now()) {
		t.Fatal("la invitación debe nacer vigente")
	}
}

func TestInvitacionServiceCrearExigeCorreoDelEmpleado(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	repository.empleado.Correo = nil
	service := NewInvitacionService(repository)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(),
		domain.CrearInvitacionInput{EmpleadoID: repository.empleado.ID})

	if !errors.Is(err, ErrInvitacionCorreoRequerido) {
		t.Fatalf("sin correo no puede entregarse el enlace: %v", err)
	}
}

func TestInvitacionServiceCrearRechazaEmpleadoYaVinculado(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	membresia := uuid.New()
	repository.empleado.MembresiaID = &membresia
	service := NewInvitacionService(repository)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(),
		domain.CrearInvitacionInput{EmpleadoID: repository.empleado.ID})

	if !errors.Is(err, ErrInvitacionYaVinculado) {
		t.Fatalf("un empleado con cuenta no vuelve a invitarse: %v", err)
	}
}

func TestInvitacionServiceCrearRechazaRolDeOtroNegocio(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	repository.rolPertenece = false
	service := NewInvitacionService(repository)
	rolAjeno := uuid.New()

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(),
		domain.CrearInvitacionInput{EmpleadoID: repository.empleado.ID, RolPredeterminadoID: &rolAjeno})

	if !errors.Is(err, ErrInvitacionRolAjeno) {
		t.Fatalf("el rol debe pertenecer al negocio: %v", err)
	}
}

func TestInvitacionServiceCrearRechazaNegocioArchivado(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	repository.access.EstadoNegocio = "archivado"
	service := NewInvitacionService(repository)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(),
		domain.CrearInvitacionInput{EmpleadoID: repository.empleado.ID})

	if !errors.Is(err, ErrEstadoNegocio) {
		t.Fatalf("un negocio archivado no invita: %v", err)
	}
}

func invitacionVigente(correo string) domain.InvitacionNegocio {
	return domain.InvitacionNegocio{
		ID: uuid.New(), NegocioID: uuid.New(), Correo: correo,
		Estado: domain.EstadoInvitacionPendiente, ExpiraEn: time.Now().Add(time.Hour),
	}
}

func TestInvitacionServiceAceptarExigeCorreoCoincidente(t *testing.T) {
	repository := invitacionStub()
	repository.invitacion = invitacionVigente("ana@tienda.mx")
	repository.correoUsuario = "otro@tienda.mx"
	service := NewInvitacionService(repository)

	err := service.Aceptar(context.Background(), uuid.New(), "token-en-claro")

	if !errors.Is(err, ErrInvitacionCorreoDistinto) || repository.aceptoLlamado {
		t.Fatalf("la aceptación exige el mismo correo invitado: %v", err)
	}
}

func TestInvitacionServiceAceptarConCorreoCoincidenteIgnoraMayusculas(t *testing.T) {
	repository := invitacionStub()
	repository.invitacion = invitacionVigente("ana@tienda.mx")
	repository.correoUsuario = "  ANA@Tienda.MX "
	service := NewInvitacionService(repository)

	if err := service.Aceptar(context.Background(), uuid.New(), "token-en-claro"); err != nil {
		t.Fatalf("el correo debe compararse normalizado: %v", err)
	}
	if !repository.aceptoLlamado {
		t.Fatal("la aceptación debió ejecutarse")
	}
}

func TestInvitacionServiceRechazaTokenExpirado(t *testing.T) {
	repository := invitacionStub()
	repository.invitacion = invitacionVigente("ana@tienda.mx")
	repository.invitacion.ExpiraEn = time.Now().Add(-time.Minute)
	repository.correoUsuario = "ana@tienda.mx"
	service := NewInvitacionService(repository)

	err := service.Aceptar(context.Background(), uuid.New(), "token-en-claro")

	if !errors.Is(err, ErrInvitacionNoVigente) {
		t.Fatalf("un token vencido no debe aceptarse: %v", err)
	}
}

func TestInvitacionServiceRechazaTokenYaUsado(t *testing.T) {
	repository := invitacionStub()
	repository.invitacion = invitacionVigente("ana@tienda.mx")
	repository.invitacion.Estado = domain.EstadoInvitacionAceptada
	repository.correoUsuario = "ana@tienda.mx"
	service := NewInvitacionService(repository)

	err := service.Aceptar(context.Background(), uuid.New(), "token-en-claro")

	if !errors.Is(err, ErrInvitacionNoVigente) {
		t.Fatalf("un token ya aceptado no debe reutilizarse: %v", err)
	}
}

func TestInvitacionServiceRechazaTokenVacio(t *testing.T) {
	service := NewInvitacionService(invitacionStub())

	_, err := service.Consultar(context.Background(), "   ")

	if !errors.Is(err, ErrInvitacionNoEncontrada) {
		t.Fatalf("un token vacío no debe consultar la base: %v", err)
	}
}

func TestInvitacionServiceConsultarIndicaSiFaltaCuenta(t *testing.T) {
	repository := invitacionStub()
	repository.invitacion = invitacionVigente("ana@tienda.mx")
	repository.existeCuenta = false
	service := NewInvitacionService(repository)

	publica, err := service.Consultar(context.Background(), "token-en-claro")

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !publica.RequiereCuenta {
		t.Fatal("si el correo no tiene cuenta, la pantalla debe pedir registro")
	}
}

func TestGenerarTokenInvitacionProduceValoresDistintos(t *testing.T) {
	primero, hashPrimero, err := generarTokenInvitacion()
	if err != nil {
		t.Fatal(err)
	}
	segundo, hashSegundo, err := generarTokenInvitacion()
	if err != nil {
		t.Fatal(err)
	}

	if primero == segundo || hashPrimero == hashSegundo {
		t.Fatal("cada invitación debe recibir un token único")
	}
	if hashDeToken(primero) != hashPrimero {
		t.Fatal("el hash debe ser reproducible a partir del token")
	}
}

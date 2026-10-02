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
	access          domain.ContextoNegocioSucursal
	permisos        []string
	empleado        domain.EmpleadoDetalle
	codigoRol       string
	rolActivo       bool
	sucursalActiva  bool
	invitacion      domain.InvitacionNegocio
	correoUsuario   string
	existeCuenta    bool
	creada          domain.InvitacionNegocio
	aceptoLlamado   bool
	reemitida       bool
	correoReemitido *string
	expiraReemitida time.Time
	errorReemitir   error
	errorEmpleado   error
	errorHash       error
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
func (r *invitacionRepositoryStub) CodigoRolActivo(context.Context, uuid.UUID, uuid.UUID) (string, bool, error) {
	return r.codigoRol, r.rolActivo, nil
}
func (r *invitacionRepositoryStub) SucursalActivaDelNegocio(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return r.sucursalActiva, nil
}
func (r *invitacionRepositoryStub) Listar(context.Context, uuid.UUID, domain.FiltroInvitaciones) ([]domain.InvitacionResumen, error) {
	return []domain.InvitacionResumen{}, nil
}
func (r *invitacionRepositoryStub) Emitir(_ context.Context, invitacion domain.InvitacionNegocio) (domain.InvitacionResumen, error) {
	r.creada = invitacion
	return domain.InvitacionResumen{ID: uuid.New(), Correo: invitacion.Correo, ExpiraEn: invitacion.ExpiraEn}, nil
}
func (r *invitacionRepositoryStub) Reemitir(_ context.Context, _, _, _ uuid.UUID, correo *string, hash string, expiraEn time.Time) (domain.InvitacionResumen, error) {
	if r.errorReemitir != nil {
		return domain.InvitacionResumen{}, r.errorReemitir
	}
	r.reemitida, r.correoReemitido, r.expiraReemitida = true, correo, expiraEn
	r.creada.HashToken = hash
	destinatario := "ana@tienda.mx"
	if correo != nil {
		destinatario = *correo
	}
	return domain.InvitacionResumen{ID: uuid.New(), Correo: destinatario, ExpiraEn: expiraEn}, nil
}
func (r *invitacionRepositoryStub) Cancelar(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (r *invitacionRepositoryStub) ObtenerPorHash(context.Context, string) (domain.InvitacionNegocio, error) {
	if r.errorHash != nil {
		return domain.InvitacionNegocio{}, r.errorHash
	}
	return r.invitacion, nil
}
func (r *invitacionRepositoryStub) DatosPublicos(context.Context, uuid.UUID) (domain.InvitacionPublica, error) {
	return domain.InvitacionPublica{
		Correo: r.invitacion.Correo, NombreNegocio: "Abarrotes Luna", NombreEmpleado: "Ana Ruiz",
		NombreSucursal: "Centro", NombreRol: "Cajero",
	}, nil
}
func (r *invitacionRepositoryStub) CorreoDeUsuario(context.Context, uuid.UUID) (string, error) {
	return r.correoUsuario, nil
}
func (r *invitacionRepositoryStub) ExisteCuentaVerificada(context.Context, string) (bool, error) {
	return r.existeCuenta, nil
}
func (r *invitacionRepositoryStub) Aceptar(context.Context, uuid.UUID, uuid.UUID) (domain.InvitacionAceptada, error) {
	r.aceptoLlamado = true
	return domain.InvitacionAceptada{Aceptada: true, NegocioID: r.invitacion.NegocioID, SucursalID: r.invitacion.SucursalID}, nil
}
func (r *invitacionRepositoryStub) MarcarExpiradas(context.Context, uuid.UUID) error { return nil }

func invitacionStub(permisos ...string) *invitacionRepositoryStub {
	correo := "ana@tienda.mx"
	return &invitacionRepositoryStub{
		access:    domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "miembro"},
		permisos:  permisos,
		empleado:  domain.EmpleadoDetalle{ID: uuid.New(), Correo: &correo, Estado: domain.EstadoEmpleadoPendiente},
		codigoRol: "CAJERO", rolActivo: true, sucursalActiva: true,
	}
}

// entradaInvitacion arma una solicitud completa: empleado, sucursal y rol son obligatorios.
func entradaInvitacion(repository *invitacionRepositoryStub) domain.CrearInvitacionInput {
	return domain.CrearInvitacionInput{
		EmpleadoID: repository.empleado.ID, SucursalID: uuid.New(), RolPredeterminadoID: uuid.New(),
	}
}

type registroCuentaStub struct {
	nombres, apellidos, correo string
	llamadas                   int
	err                        error
}

func (r *registroCuentaStub) RegistrarCuentaPendiente(_ context.Context, nombres, apellidos, correo, _, _, _ string) (DesafioRegistro, error) {
	r.llamadas++
	r.nombres, r.apellidos, r.correo = nombres, apellidos, correo
	return DesafioRegistro{DesafioID: uuid.New(), CorreoEnmascarado: "a***@tienda.mx", ReenviarEnSegundos: 60}, r.err
}

func TestInvitacionServiceCrearExigePermiso(t *testing.T) {
	service := NewInvitacionService(invitacionStub(domain.PermisoInvitacionVer), nil, nil)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(),
		domain.CrearInvitacionInput{EmpleadoID: uuid.New(), SucursalID: uuid.New(), RolPredeterminadoID: uuid.New()})

	if !errors.Is(err, ErrInvitacionProhibida) {
		t.Fatalf("se requiere permiso de envío: %v", err)
	}
}

func TestInvitacionServiceCrearGuardaSoloElHash(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	service := NewInvitacionService(repository, nil, nil)

	creada, err := service.Crear(context.Background(), uuid.New(), uuid.New(),
		entradaInvitacion(repository))

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
	if repository.creada.SucursalID == nil || repository.creada.RolPredeterminadoID == nil {
		t.Fatal("la invitación debe fijar sucursal y rol")
	}
	vigencia := time.Until(repository.creada.ExpiraEn)
	if vigencia < 71*time.Hour+59*time.Minute || vigencia > 72*time.Hour {
		t.Fatalf("la invitación debe vencer a las 72 horas: %v", vigencia)
	}
}

func TestInvitacionServiceCrearExigeCorreoDelEmpleado(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	repository.empleado.Correo = nil
	service := NewInvitacionService(repository, nil, nil)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(),
		entradaInvitacion(repository))

	if !errors.Is(err, ErrInvitacionCorreoRequerido) {
		t.Fatalf("sin correo no puede entregarse el enlace: %v", err)
	}
}

func TestInvitacionServiceCrearRechazaEmpleadoYaVinculado(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	membresia := uuid.New()
	repository.empleado.MembresiaID = &membresia
	service := NewInvitacionService(repository, nil, nil)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(),
		entradaInvitacion(repository))

	if !errors.Is(err, ErrInvitacionYaVinculado) {
		t.Fatalf("un empleado con cuenta no vuelve a invitarse: %v", err)
	}
}

func TestInvitacionServiceCrearRechazaRolDeOtroNegocio(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	repository.rolActivo = false
	service := NewInvitacionService(repository, nil, nil)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), entradaInvitacion(repository))

	if !errors.Is(err, ErrInvitacionRolNoDisponible) {
		t.Fatalf("el rol debe pertenecer al negocio: %v", err)
	}
}

func TestInvitacionServiceCrearRechazaNegocioArchivado(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	repository.access.EstadoNegocio = "archivado"
	service := NewInvitacionService(repository, nil, nil)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(),
		entradaInvitacion(repository))

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
	service := NewInvitacionService(repository, nil, nil)

	_, err := service.Aceptar(context.Background(), uuid.New(), "token-en-claro")

	if !errors.Is(err, ErrInvitacionCorreoDistinto) || repository.aceptoLlamado {
		t.Fatalf("la aceptación exige el mismo correo invitado: %v", err)
	}
}

func TestInvitacionServiceAceptarConCorreoCoincidenteIgnoraMayusculas(t *testing.T) {
	repository := invitacionStub()
	repository.invitacion = invitacionVigente("ana@tienda.mx")
	repository.correoUsuario = "  ANA@Tienda.MX "
	service := NewInvitacionService(repository, nil, nil)

	if _, err := service.Aceptar(context.Background(), uuid.New(), "token-en-claro"); err != nil {
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
	service := NewInvitacionService(repository, nil, nil)

	_, err := service.Aceptar(context.Background(), uuid.New(), "token-en-claro")

	if !errors.Is(err, ErrInvitacionNoVigente) {
		t.Fatalf("un token vencido no debe aceptarse: %v", err)
	}
}

func TestInvitacionServiceRechazaTokenYaUsado(t *testing.T) {
	repository := invitacionStub()
	repository.invitacion = invitacionVigente("ana@tienda.mx")
	repository.invitacion.Estado = domain.EstadoInvitacionAceptada
	repository.correoUsuario = "ana@tienda.mx"
	service := NewInvitacionService(repository, nil, nil)

	_, err := service.Aceptar(context.Background(), uuid.New(), "token-en-claro")

	if !errors.Is(err, ErrInvitacionYaAceptada) || repository.aceptoLlamado {
		t.Fatalf("un token ya aceptado no debe reutilizarse: %v", err)
	}
}

func TestInvitacionServiceRechazaTokenVacio(t *testing.T) {
	service := NewInvitacionService(invitacionStub(), nil, nil)

	_, err := service.Consultar(context.Background(), "   ")

	if !errors.Is(err, ErrInvitacionNoEncontrada) {
		t.Fatalf("un token vacío no debe consultar la base: %v", err)
	}
}

func TestInvitacionServiceConsultarIndicaSiFaltaCuenta(t *testing.T) {
	repository := invitacionStub()
	repository.invitacion = invitacionVigente("ana@tienda.mx")
	repository.existeCuenta = false
	service := NewInvitacionService(repository, nil, nil)

	publica, err := service.Consultar(context.Background(), "token-en-claro")

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !publica.RequiereCuenta {
		t.Fatal("si el correo no tiene cuenta, la pantalla debe pedir registro")
	}
	if publica.CorreoEnmascarado != "a•••a@tienda.mx" || publica.NombreSucursal != "Centro" || publica.NombreRol != "Cajero" {
		t.Fatalf("la consulta pública debe enmascarar el correo y nombrar sucursal y rol: %#v", publica)
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

type invitacionMailerStub struct {
	para, negocio, invitado, token string
	llamadas                       int
	err                            error
}

func (m *invitacionMailerStub) EnviarInvitacion(_ context.Context, para, negocio, invitado, token string, _ time.Time) error {
	m.llamadas++
	m.para, m.negocio, m.invitado, m.token = para, negocio, invitado, token
	return m.err
}

func TestInvitacionServiceCrearEnviaElEnlacePorCorreo(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	mailer := &invitacionMailerStub{}
	service := NewInvitacionService(repository, mailer, nil)

	creada, err := service.Crear(context.Background(), uuid.New(), uuid.New(),
		entradaInvitacion(repository))

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !creada.CorreoEnviado || mailer.llamadas != 1 {
		t.Fatalf("se esperaba un envío exitoso: %#v", mailer)
	}
	if mailer.para != "ana@tienda.mx" || mailer.token != creada.Token || mailer.negocio != "Abarrotes Luna" || mailer.invitado != "Ana Ruiz" {
		t.Fatalf("datos del correo inesperados: %#v", mailer)
	}
}

func TestInvitacionServiceCrearConservaLaInvitacionSiFallaElCorreo(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	service := NewInvitacionService(repository, &invitacionMailerStub{err: errors.New("smtp caído")}, nil)

	creada, err := service.Crear(context.Background(), uuid.New(), uuid.New(),
		entradaInvitacion(repository))

	if err != nil {
		t.Fatalf("un fallo de correo no debe revertir la invitación: %v", err)
	}
	if creada.CorreoEnviado || creada.Token == "" || repository.creada.HashToken == "" {
		t.Fatalf("la invitación debía quedar creada y marcada sin correo: %#v", creada)
	}
}

func TestInvitacionServiceCrearExigeSucursalYRol(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	service := NewInvitacionService(repository, nil, nil)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(),
		domain.CrearInvitacionInput{EmpleadoID: repository.empleado.ID})

	var validacion *ErrorValidacion
	if !errors.As(err, &validacion) || validacion.Campos["sucursal_id"] == "" || validacion.Campos["rol_predeterminado_id"] == "" {
		t.Fatalf("sucursal y rol son obligatorios: %v", err)
	}
}

func TestInvitacionServiceCrearRechazaSucursalInactiva(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	repository.sucursalActiva = false
	service := NewInvitacionService(repository, nil, nil)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), entradaInvitacion(repository))

	if !errors.Is(err, ErrInvitacionSucursalNoDisponible) {
		t.Fatalf("la sucursal debe estar activa en el negocio: %v", err)
	}
}

func TestInvitacionServiceCrearNuncaDelegaRolPropietario(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	repository.codigoRol = "propietario"
	service := NewInvitacionService(repository, nil, nil)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), entradaInvitacion(repository))

	if !errors.Is(err, ErrInvitacionRolNoDisponible) {
		t.Fatalf("el rol propietario no se delega por invitación: %v", err)
	}
}

func TestInvitacionServiceReenviarRenuevaLas72Horas(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	mailer := &invitacionMailerStub{}
	service := NewInvitacionService(repository, mailer, nil)

	creada, err := service.Reenviar(context.Background(), uuid.New(), uuid.New(), uuid.New(), domain.ReenviarInvitacionInput{})

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !repository.reemitida || repository.correoReemitido != nil {
		t.Fatal("reenviar sin correo debe conservar al destinatario")
	}
	vigencia := time.Until(repository.expiraReemitida)
	if vigencia < 71*time.Hour+59*time.Minute || vigencia > 72*time.Hour {
		t.Fatalf("el reenvío debe renovar 72 horas: %v", vigencia)
	}
	if creada.Token == "" || repository.creada.HashToken != hashDeToken(creada.Token) || mailer.token != creada.Token {
		t.Fatal("el reenvío debe entregar un token nuevo cuyo hash es el persistido")
	}
}

func TestInvitacionServiceCorregirCorreoExigePermisoDeEmpleados(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	service := NewInvitacionService(repository, nil, nil)
	correo := "nuevo@tienda.mx"

	_, err := service.Reenviar(context.Background(), uuid.New(), uuid.New(), uuid.New(), domain.ReenviarInvitacionInput{Correo: &correo})

	if !errors.Is(err, ErrInvitacionProhibida) || repository.reemitida {
		t.Fatalf("corregir el correo exige gestionar empleados: %v", err)
	}
}

func TestInvitacionServiceCorregirCorreoNormalizaYValida(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar, domain.PermisoEmpleadoGestionar)
	service := NewInvitacionService(repository, nil, nil)
	invalido := "no-es-correo"

	_, err := service.Reenviar(context.Background(), uuid.New(), uuid.New(), uuid.New(), domain.ReenviarInvitacionInput{Correo: &invalido})
	var validacion *ErrorValidacion
	if !errors.As(err, &validacion) || repository.reemitida {
		t.Fatalf("un correo inválido no debe reemitir: %v", err)
	}

	valido := "  Nuevo@Tienda.MX "
	creada, err := service.Reenviar(context.Background(), uuid.New(), uuid.New(), uuid.New(), domain.ReenviarInvitacionInput{Correo: &valido})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if repository.correoReemitido == nil || *repository.correoReemitido != "nuevo@tienda.mx" || creada.Invitacion.Correo != "nuevo@tienda.mx" {
		t.Fatalf("el correo nuevo debe guardarse normalizado: %#v", creada.Invitacion)
	}
}

func TestInvitacionServiceReenviarPropagaConflictoDeAceptada(t *testing.T) {
	repository := invitacionStub(domain.PermisoInvitacionEnviar)
	repository.errorReemitir = ErrInvitacionYaAceptada
	mailer := &invitacionMailerStub{}
	service := NewInvitacionService(repository, mailer, nil)

	_, err := service.Reenviar(context.Background(), uuid.New(), uuid.New(), uuid.New(), domain.ReenviarInvitacionInput{})

	if !errors.Is(err, ErrInvitacionYaAceptada) || mailer.llamadas != 0 {
		t.Fatalf("una invitación aceptada no se reenvía ni manda correo: %v", err)
	}
}

func TestInvitacionServiceRegistrarUsaElCorreoDeLaInvitacion(t *testing.T) {
	repository := invitacionStub()
	repository.invitacion = invitacionVigente("ana@tienda.mx")
	registro := &registroCuentaStub{}
	service := NewInvitacionService(repository, nil, registro)

	desafio, err := service.Registrar(context.Background(), "token-en-claro",
		domain.RegistroInvitacionInput{Nombres: " María José ", Apellidos: "Ruiz Luna", Contrasena: "contrasena-segura"}, "127.0.0.1")

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if registro.correo != "ana@tienda.mx" || registro.nombres != "María José" || registro.apellidos != "Ruiz Luna" {
		t.Fatalf("el registro debe tomar el correo de la invitación: %#v", registro)
	}
	if desafio.DesafioID == uuid.Nil || repository.aceptoLlamado {
		t.Fatal("registrar devuelve un desafío OTP y no consume la invitación")
	}
}

func TestInvitacionServiceRegistrarValidaDatos(t *testing.T) {
	repository := invitacionStub()
	repository.invitacion = invitacionVigente("ana@tienda.mx")
	registro := &registroCuentaStub{}
	service := NewInvitacionService(repository, nil, registro)

	_, err := service.Registrar(context.Background(), "token-en-claro",
		domain.RegistroInvitacionInput{Nombres: "Ana", Contrasena: "corta"}, "127.0.0.1")

	var validacion *ErrorValidacion
	if !errors.As(err, &validacion) || validacion.Campos["apellidos"] == "" || validacion.Campos["contrasena"] == "" || registro.llamadas != 0 {
		t.Fatalf("apellidos y contraseña deben validarse antes de crear la cuenta: %v", err)
	}
}

func TestInvitacionServiceRegistrarRechazaCuentaExistente(t *testing.T) {
	repository := invitacionStub()
	repository.invitacion = invitacionVigente("ana@tienda.mx")
	repository.existeCuenta = true
	registro := &registroCuentaStub{}
	service := NewInvitacionService(repository, nil, registro)

	_, err := service.Registrar(context.Background(), "token-en-claro",
		domain.RegistroInvitacionInput{Nombres: "Ana", Apellidos: "Ruiz", Contrasena: "contrasena-segura"}, "127.0.0.1")

	if !errors.Is(err, ErrInvitacionCuentaExistente) || registro.llamadas != 0 {
		t.Fatalf("una cuenta verificada inicia sesión, no se registra de nuevo: %v", err)
	}
}

func TestInvitacionServiceRegistrarRechazaTokenVencido(t *testing.T) {
	repository := invitacionStub()
	repository.invitacion = invitacionVigente("ana@tienda.mx")
	repository.invitacion.ExpiraEn = time.Now().Add(-time.Minute)
	registro := &registroCuentaStub{}
	service := NewInvitacionService(repository, nil, registro)

	_, err := service.Registrar(context.Background(), "token-en-claro",
		domain.RegistroInvitacionInput{Nombres: "Ana", Apellidos: "Ruiz", Contrasena: "contrasena-segura"}, "127.0.0.1")

	if !errors.Is(err, ErrInvitacionNoVigente) || registro.llamadas != 0 {
		t.Fatalf("un enlace vencido no crea cuentas: %v", err)
	}
}

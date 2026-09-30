package correo

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

type transporteFalso struct {
	enviados []Mensaje
	err      error
}

func (t *transporteFalso) Enviar(_ context.Context, m Mensaje) (string, error) {
	if t.err != nil {
		return "", t.err
	}
	t.enviados = append(t.enviados, m)
	return "<id@prueba>", nil
}

type renderizadorFalso struct {
	plantilla string
	datos     any
	err       error
}

func (r *renderizadorFalso) Render(plantilla string, datos any) (Contenido, error) {
	r.plantilla, r.datos = plantilla, datos
	if r.err != nil {
		return Contenido{}, r.err
	}
	return Contenido{Asunto: "Asunto", HTML: "<p>html</p>", Texto: "texto"}, nil
}

func nuevoServicio(t *transporteFalso, r *renderizadorFalso) *Servicio {
	s := NewServicio(t, r, Config{NombreApp: "Tienda", FrontendURL: "https://tienda.mergemakers.com/"})
	s.now = func() time.Time { return time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC) }
	return s
}

func TestSendVerificationOTPRendersCodeAndMinutes(t *testing.T) {
	transporte, plantillas := &transporteFalso{}, &renderizadorFalso{}
	s := nuevoServicio(transporte, plantillas)

	err := s.SendVerificationOTP(context.Background(), "ana@ejemplo.com", "123456", s.now().Add(10*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	datos, ok := plantillas.datos.(DatosVerificacion)
	if plantillas.plantilla != PlantillaVerificacion || !ok || datos.Codigo != "123456" || datos.Minutos != 10 || datos.NombreApp != "Tienda" || datos.Anio != 2026 {
		t.Fatalf("datos de verificación inesperados: %s %#v", plantillas.plantilla, plantillas.datos)
	}
	if len(transporte.enviados) != 1 || transporte.enviados[0].Para != "ana@ejemplo.com" || transporte.enviados[0].Texto != "texto" {
		t.Fatalf("mensaje inesperado: %#v", transporte.enviados)
	}
}

func TestPasswordResetLinkUsesFragmentAndFrontendURL(t *testing.T) {
	plantillas := &renderizadorFalso{}
	s := nuevoServicio(&transporteFalso{}, plantillas)

	if err := s.SendPasswordReset(context.Background(), "ana@ejemplo.com", "desafio-1", "tok/en+", s.now().Add(30*time.Minute)); err != nil {
		t.Fatal(err)
	}
	datos := plantillas.datos.(DatosRecuperacion)
	esperado := "https://tienda.mergemakers.com/restablecer-contrasena#desafio=desafio-1&token=tok%2Fen%2B"
	if datos.Enlace != esperado || datos.Minutos != 30 {
		t.Fatalf("enlace=%q minutos=%d", datos.Enlace, datos.Minutos)
	}
}

func TestInvitationLinkAndDays(t *testing.T) {
	plantillas := &renderizadorFalso{}
	s := nuevoServicio(&transporteFalso{}, plantillas)

	if err := s.EnviarInvitacion(context.Background(), "luis@ejemplo.com", "Abarrotes Luna", "Luis Pérez", "abc_DEF-123", s.now().Add(7*24*time.Hour-time.Second)); err != nil {
		t.Fatal(err)
	}
	datos := plantillas.datos.(DatosInvitacion)
	if datos.Enlace != "https://tienda.mergemakers.com/invitacion/abc_DEF-123" || datos.Dias != 7 || datos.NombreNegocio != "Abarrotes Luna" || datos.NombreInvitado != "Luis Pérez" {
		t.Fatalf("datos de invitación inesperados: %#v", datos)
	}
}

func TestPasswordChangedLinksToRecovery(t *testing.T) {
	plantillas := &renderizadorFalso{}
	s := nuevoServicio(&transporteFalso{}, plantillas)

	if err := s.SendPasswordChanged(context.Background(), "ana@ejemplo.com"); err != nil {
		t.Fatal(err)
	}
	if datos := plantillas.datos.(DatosContrasenaCambiada); datos.EnlaceRecuperacion != "https://tienda.mergemakers.com/recuperar-contrasena" {
		t.Fatalf("enlace inesperado: %q", datos.EnlaceRecuperacion)
	}
}

func TestFailuresReturnGenericErrorAndNeverLogSecrets(t *testing.T) {
	var salida bytes.Buffer
	log.SetOutput(&salida)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	transporte := &transporteFalso{err: errors.New("smtp: rechazo permanente código=550")}
	s := nuevoServicio(transporte, &renderizadorFalso{})
	err := s.SendVerificationOTP(context.Background(), "ana@ejemplo.com", "987654", s.now().Add(time.Minute))
	if !errors.Is(err, ErrEnvio) {
		t.Fatalf("error = %v, se esperaba ErrEnvio", err)
	}
	if err := s.SendPasswordReset(context.Background(), "ana@ejemplo.com", "d", "token-secreto", s.now()); !errors.Is(err, ErrEnvio) {
		t.Fatalf("error = %v, se esperaba ErrEnvio", err)
	}
	registro := salida.String()
	for _, secreto := range []string{"987654", "token-secreto", "ana@ejemplo.com"} {
		if strings.Contains(registro, secreto) {
			t.Fatalf("el log contiene %q: %s", secreto, registro)
		}
	}
	if !strings.Contains(registro, "a***@ejemplo.com") || !strings.Contains(registro, "evento=correo_fallido") {
		t.Fatalf("el log no registra el fallo enmascarado: %s", registro)
	}

	if err := nuevoServicio(&transporteFalso{}, &renderizadorFalso{err: errors.New("roto")}).SendPasswordChanged(context.Background(), "x@y.com"); !errors.Is(err, ErrEnvio) {
		t.Fatalf("un error de plantilla debía devolver ErrEnvio; error=%v", err)
	}
}

func TestEnmascarar(t *testing.T) {
	for entrada, esperado := range map[string]string{"ana@ejemplo.com": "a***@ejemplo.com", "ñu@x.mx": "ñ***@x.mx", "invalido": "***", "@x.com": "***"} {
		if got := Enmascarar(entrada); got != esperado {
			t.Errorf("Enmascarar(%q) = %q, se esperaba %q", entrada, got, esperado)
		}
	}
}

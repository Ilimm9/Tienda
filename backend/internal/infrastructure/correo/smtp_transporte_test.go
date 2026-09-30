package correo

import (
	"bufio"
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	application "tienda/backend/internal/application/correo"
)

// servidorSMTP es un servidor mínimo en memoria para probar el transporte sin red externa.
type servidorSMTP struct {
	listener net.Listener
	rcptCode string
	silencio bool
	mu       sync.Mutex
	comandos []string
	datos    string
}

func nuevoServidorSMTP(t *testing.T) *servidorSMTP {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &servidorSMTP{listener: listener, rcptCode: "250 OK"}
	t.Cleanup(func() { _ = listener.Close() })
	go s.aceptar()
	return s
}

func (s *servidorSMTP) puerto() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

func (s *servidorSMTP) aceptar() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.atender(conn)
	}
}

func (s *servidorSMTP) atender(conn net.Conn) {
	defer conn.Close()
	if s.silencio {
		time.Sleep(2 * time.Second)
		return
	}
	lector := bufio.NewReader(conn)
	escribir := func(linea string) { _, _ = conn.Write([]byte(linea + "\r\n")) }
	escribir("220 prueba ESMTP")
	for {
		linea, err := lector.ReadString('\n')
		if err != nil {
			return
		}
		comando := strings.TrimSpace(linea)
		s.mu.Lock()
		s.comandos = append(s.comandos, comando)
		s.mu.Unlock()
		switch verbo := strings.ToUpper(strings.SplitN(comando, " ", 2)[0]); {
		case verbo == "EHLO" || verbo == "HELO":
			escribir("250-prueba")
			escribir("250 8BITMIME")
		case strings.HasPrefix(strings.ToUpper(comando), "RCPT"):
			escribir(s.rcptCode)
		case verbo == "DATA":
			escribir("354 continúa")
			var cuerpo strings.Builder
			for {
				l, err := lector.ReadString('\n')
				if err != nil {
					return
				}
				if l == ".\r\n" {
					break
				}
				cuerpo.WriteString(l)
			}
			s.mu.Lock()
			s.datos = cuerpo.String()
			s.mu.Unlock()
			escribir("250 OK id-prueba")
		case verbo == "QUIT":
			escribir("221 adiós")
			return
		default:
			escribir("250 OK")
		}
	}
}

func (s *servidorSMTP) capturado() (string, []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.datos, append([]string(nil), s.comandos...)
}

func transporteDePrueba(t *testing.T, puerto int, ajustes func(*SMTPConfig)) *SMTPTransporte {
	t.Helper()
	cfg := SMTPConfig{
		Host: "127.0.0.1", Port: puerto, TLS: TLSNinguno, Timeout: time.Second,
		Remitente: "no-reply@mergemakers.com", NombreRemitente: "Tienda",
	}
	if ajustes != nil {
		ajustes(&cfg)
	}
	logo, err := Logo()
	if err != nil {
		t.Fatal(err)
	}
	transporte, err := NewSMTPTransporte(cfg, logo)
	if err != nil {
		t.Fatal(err)
	}
	return transporte
}

func mensajeDePrueba() application.Mensaje {
	return application.Mensaje{
		Para: "ana@ejemplo.com", Asunto: "Tu código de verificación de Tienda",
		HTML: `<img src="cid:` + LogoCID + `"><p>Código 123456</p>`, Texto: "Código 123456\n",
	}
}

func TestSMTPTransportSendsMultipartMessageWithInlineLogo(t *testing.T) {
	servidor := nuevoServidorSMTP(t)
	transporte := transporteDePrueba(t, servidor.puerto(), func(c *SMTPConfig) {
		c.ResponderA = "soporte@mergemakers.com"
		c.ConfigurationSetSES = "tienda-transaccional"
	})

	id, err := transporte.Enviar(context.Background(), mensajeDePrueba())
	if err != nil {
		t.Fatal(err)
	}
	datos, comandos := servidor.capturado()
	if !strings.HasSuffix(id, "@mergemakers.com>") {
		t.Fatalf("Message-ID inesperado: %q", id)
	}
	unidos := strings.Join(comandos, "\n")
	if !strings.Contains(unidos, "MAIL FROM:<no-reply@mergemakers.com>") || !strings.Contains(unidos, "RCPT TO:<ana@ejemplo.com>") {
		t.Fatalf("sobre SMTP inesperado:\n%s", unidos)
	}
	for _, esperado := range []string{
		`From: "Tienda" <no-reply@mergemakers.com>`,
		"To: <ana@ejemplo.com>",
		"Reply-To: <soporte@mergemakers.com>",
		"X-SES-CONFIGURATION-SET: tienda-transaccional",
		"@mergemakers.com>",
		"multipart/related",
		"multipart/alternative",
		"text/plain",
		"text/html",
		"Content-Id: <" + LogoCID + ">",
		"image/png",
	} {
		if !strings.Contains(datos, esperado) {
			t.Errorf("el mensaje no contiene %q", esperado)
		}
	}
	if strings.Contains(unidos, "AUTH") {
		t.Fatal("no debe autenticarse sin usuario configurado")
	}
}

func TestSMTPTransportClassifiesRecipientRejection(t *testing.T) {
	servidor := nuevoServidorSMTP(t)
	servidor.rcptCode = "550 5.1.1 buzón inexistente ana@ejemplo.com"
	transporte := transporteDePrueba(t, servidor.puerto(), nil)

	_, err := transporte.Enviar(context.Background(), mensajeDePrueba())
	if err == nil {
		t.Fatal("se esperaba rechazo")
	}
	if strings.Contains(err.Error(), "ana@ejemplo.com") || !strings.Contains(err.Error(), "smtp:") {
		t.Fatalf("error sin clasificar o con datos personales: %v", err)
	}
}

func TestSMTPTransportTimesOut(t *testing.T) {
	servidor := nuevoServidorSMTP(t)
	servidor.silencio = true
	transporte := transporteDePrueba(t, servidor.puerto(), func(c *SMTPConfig) { c.Timeout = 200 * time.Millisecond })

	inicio := time.Now()
	_, err := transporte.Enviar(context.Background(), mensajeDePrueba())
	if err == nil || time.Since(inicio) > 1500*time.Millisecond {
		t.Fatalf("se esperaba timeout rápido; error=%v duración=%s", err, time.Since(inicio))
	}
}

func TestSMTPTransportRejectsHeaderInjectionAndConnectionErrors(t *testing.T) {
	servidor := nuevoServidorSMTP(t)
	transporte := transporteDePrueba(t, servidor.puerto(), nil)
	mensaje := mensajeDePrueba()
	mensaje.Para = "ana@ejemplo.com\r\nBcc: otro@ejemplo.com"
	if _, err := transporte.Enviar(context.Background(), mensaje); err == nil {
		t.Fatal("se esperaba rechazo del destinatario con salto de línea")
	}

	libre, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	puerto := libre.Addr().(*net.TCPAddr).Port
	_ = libre.Close()
	_, err = transporteDePrueba(t, puerto, nil).Enviar(context.Background(), mensajeDePrueba())
	if err == nil || !strings.HasPrefix(err.Error(), "smtp:") {
		t.Fatalf("error de conexión sin clasificar: %v", err)
	}
}

func TestNewSMTPTransportValidatesConfiguration(t *testing.T) {
	base := SMTPConfig{Host: "localhost", Port: 1025, TLS: TLSNinguno, Timeout: time.Second, Remitente: "a@b.com"}
	for nombre, ajuste := range map[string]func(*SMTPConfig){
		"tls":       func(c *SMTPConfig) { c.TLS = "ssl" },
		"host":      func(c *SMTPConfig) { c.Host = "" },
		"remitente": func(c *SMTPConfig) { c.Remitente = " " },
	} {
		cfg := base
		ajuste(&cfg)
		if _, err := NewSMTPTransporte(cfg, nil); err == nil {
			t.Errorf("%s: se esperaba error", nombre)
		}
	}
	for _, politica := range []string{TLSNinguno, TLSStartTLS, TLSImplicito} {
		cfg := base
		cfg.TLS = politica
		if _, err := NewSMTPTransporte(cfg, nil); err != nil {
			t.Errorf("%s: %v", politica, err)
		}
	}
}

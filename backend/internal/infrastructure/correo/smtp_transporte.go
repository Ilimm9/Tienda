package correo

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	application "tienda/backend/internal/application/correo"

	mail "github.com/wneessen/go-mail"
)

// Políticas TLS admitidas en SMTP_TLS.
const (
	TLSNinguno   = "none"
	TLSStartTLS  = "starttls"
	TLSImplicito = "tls"
)

type SMTPConfig struct {
	Host                string
	Port                int
	TLS                 string
	Usuario             string
	Contrasena          string
	Timeout             time.Duration
	Remitente           string
	NombreRemitente     string
	ResponderA          string
	ConfigurationSetSES string
}

// SMTPTransporte envía por SMTP: Mailpit en desarrollo y Amazon SES en producción.
type SMTPTransporte struct {
	config SMTPConfig
	logo   []byte
}

func NewSMTPTransporte(cfg SMTPConfig, logo []byte) (*SMTPTransporte, error) {
	switch cfg.TLS {
	case TLSNinguno, TLSStartTLS, TLSImplicito:
	default:
		return nil, fmt.Errorf("SMTP_TLS debe ser %q, %q o %q", TLSNinguno, TLSStartTLS, TLSImplicito)
	}
	if strings.TrimSpace(cfg.Host) == "" || strings.TrimSpace(cfg.Remitente) == "" {
		return nil, errors.New("SMTP_HOST y MAIL_FROM son obligatorios")
	}
	return &SMTPTransporte{config: cfg, logo: logo}, nil
}

func (t *SMTPTransporte) Enviar(ctx context.Context, mensaje application.Mensaje) (string, error) {
	msg, err := t.construir(mensaje)
	if err != nil {
		return "", err
	}
	client, err := mail.NewClient(t.config.Host, t.opciones()...)
	if err != nil {
		return "", fmt.Errorf("configuración SMTP inválida: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, t.config.Timeout)
	defer cancel()
	if err := client.DialAndSendWithContext(ctx, msg); err != nil {
		return "", clasificar(err)
	}
	return msg.GetMessageID(), nil
}

func (t *SMTPTransporte) construir(mensaje application.Mensaje) (*mail.Msg, error) {
	msg := mail.NewMsg(mail.WithNoDefaultUserAgent())
	if err := msg.FromFormat(t.config.NombreRemitente, t.config.Remitente); err != nil {
		return nil, fmt.Errorf("remitente inválido: %w", err)
	}
	// go-mail valida la dirección y rechaza saltos de línea en cabeceras.
	if err := msg.To(mensaje.Para); err != nil {
		return nil, errors.New("destinatario inválido")
	}
	if t.config.ResponderA != "" {
		if err := msg.ReplyTo(t.config.ResponderA); err != nil {
			return nil, fmt.Errorf("MAIL_REPLY_TO inválido: %w", err)
		}
	}
	if t.config.ConfigurationSetSES != "" {
		msg.SetGenHeader("X-SES-CONFIGURATION-SET", t.config.ConfigurationSetSES)
	}
	msg.Subject(mensaje.Asunto)
	// El Message-ID usa el dominio del remitente para no exponer el hostname del servidor.
	msg.SetMessageIDWithValue(rand.Text() + "@" + dominio(t.config.Remitente))
	msg.SetDate()
	msg.SetBodyString(mail.TypeTextPlain, mensaje.Texto)
	if mensaje.HTML != "" {
		msg.AddAlternativeString(mail.TypeTextHTML, mensaje.HTML)
		if len(t.logo) > 0 && strings.Contains(mensaje.HTML, "cid:"+LogoCID) {
			if err := msg.EmbedReader("logo.png", bytes.NewReader(t.logo),
				mail.WithFileContentID("<"+LogoCID+">"), mail.WithFileContentType(mail.ContentType("image/png"))); err != nil {
				return nil, fmt.Errorf("no fue posible adjuntar el logo: %w", err)
			}
		}
	}
	return msg, nil
}

func dominio(direccion string) string {
	if i := strings.LastIndex(direccion, "@"); i >= 0 && i < len(direccion)-1 {
		return direccion[i+1:]
	}
	return "localhost"
}

func (t *SMTPTransporte) opciones() []mail.Option {
	opciones := []mail.Option{mail.WithPort(t.config.Port), mail.WithTimeout(t.config.Timeout)}
	switch t.config.TLS {
	case TLSNinguno:
		opciones = append(opciones, mail.WithTLSPolicy(mail.NoTLS))
	case TLSStartTLS:
		opciones = append(opciones, mail.WithTLSPolicy(mail.TLSMandatory))
	case TLSImplicito:
		opciones = append(opciones, mail.WithSSL())
	}
	if t.config.Usuario != "" {
		opciones = append(opciones,
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(t.config.Usuario),
			mail.WithPassword(t.config.Contrasena),
		)
	}
	return opciones
}

// clasificar reduce el error a una categoría útil para operar sin exponer
// direcciones, cuerpo ni respuestas completas del servidor en los logs.
func clasificar(err error) error {
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return errors.New("smtp: tiempo de espera agotado")
	}
	var sendErr *mail.SendError
	if errors.As(err, &sendErr) {
		tipo := "permanente"
		if sendErr.IsTemp() {
			tipo = "temporal"
		}
		return fmt.Errorf("smtp: rechazo %s código=%d estado=%s", tipo, sendErr.ErrorCode(), sendErr.EnhancedStatusCode())
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return fmt.Errorf("smtp: no fue posible conectar (%s)", opErr.Op)
	}
	if strings.Contains(strings.ToLower(err.Error()), "auth") {
		return errors.New("smtp: autenticación rechazada")
	}
	return errors.New("smtp: error de envío")
}

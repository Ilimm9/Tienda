package cuenta

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type DevelopmentSMTPMailer struct {
	host    string
	port    int
	from    string
	timeout time.Duration
}

func NewDevelopmentSMTPMailer(host string, port int, from string, timeout time.Duration) *DevelopmentSMTPMailer {
	return &DevelopmentSMTPMailer{host: host, port: port, from: from, timeout: timeout}
}

func (m *DevelopmentSMTPMailer) SendVerificationOTP(ctx context.Context, recipient, code string, expiresAt time.Time) error {
	if containsNewline(recipient) || containsNewline(m.from) {
		return fmt.Errorf("dirección de correo inválida")
	}
	address := net.JoinHostPort(m.host, fmt.Sprintf("%d", m.port))
	connection, err := (&net.Dialer{Timeout: m.timeout}).DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	defer connection.Close()
	_ = connection.SetDeadline(time.Now().Add(m.timeout))

	client, err := smtp.NewClient(connection, m.host)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Mail(m.from); err != nil {
		return err
	}
	if err := client.Rcpt(recipient); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	buffer := bufio.NewWriter(writer)
	minutes := int(time.Until(expiresAt).Minutes())
	if minutes < 1 {
		minutes = 1
	}
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Verifica tu correo en Tienda\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nTu código de verificación es: %s\r\n\r\nExpira en aproximadamente %d minutos. Si no solicitaste este código, ignora este correo.\r\n", m.from, recipient, code, minutes)
	if _, err := buffer.WriteString(message); err != nil {
		_ = writer.Close()
		return err
	}
	if err := buffer.Flush(); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func containsNewline(value string) bool {
	return strings.ContainsAny(value, "\r\n")
}

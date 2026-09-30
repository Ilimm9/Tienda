package config

import (
	"strings"
	"testing"
)

func TestProductionRejectsDefaultOTPSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("OTP_HMAC_SECRET", "development-only-change-otp-secret")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "OTP_HMAC_SECRET") {
		t.Fatalf("Load() debía rechazar el secreto local en producción; error=%v", err)
	}
}

func TestProductionAcceptsStrongOTPSecret(t *testing.T) {
	setProductionMail(t)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OTPHMACSecret == "" || cfg.OTPMaxAttempts != 5 || cfg.SMTPPort != 587 || cfg.SMTPTLS != "starttls" {
		t.Fatalf("configuración OTP inválida: %#v", cfg)
	}
}

func TestDevelopmentMailDefaultsTargetMailpit(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SMTPHost != "localhost" || cfg.SMTPPort != 1025 || cfg.SMTPTLS != "none" || cfg.MailFrom != "no-reply@tienda.local" || cfg.PasswordResetTTL.Minutes() != 30 {
		t.Fatalf("valores de correo de desarrollo inesperados: %#v", cfg)
	}
}

func TestSMTPFromRemainsAliasOfMailFrom(t *testing.T) {
	t.Setenv("SMTP_FROM", "legacy@tienda.local")
	cfg, err := Load()
	if err != nil || cfg.MailFrom != "legacy@tienda.local" {
		t.Fatalf("SMTP_FROM debía usarse como alias; cfg=%q err=%v", cfg.MailFrom, err)
	}
	t.Setenv("MAIL_FROM", "nuevo@tienda.local")
	cfg, err = Load()
	if err != nil || cfg.MailFrom != "nuevo@tienda.local" {
		t.Fatalf("MAIL_FROM debía tener prioridad; cfg=%q err=%v", cfg.MailFrom, err)
	}
}

func TestInvalidMailSettingsAreRejected(t *testing.T) {
	cases := map[string][2]string{
		"tls desconocido": {"SMTP_TLS", "ssl3"},
		"remitente":       {"MAIL_FROM", "no es correo"},
		"responder a":     {"MAIL_REPLY_TO", "tampoco"},
	}
	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			t.Setenv(env[0], env[1])
			if _, err := Load(); err == nil || !strings.Contains(err.Error(), env[0]) {
				t.Fatalf("Load() debía rechazar %s; error=%v", env[0], err)
			}
		})
	}
}

func TestProductionRejectsInsecureMailSettings(t *testing.T) {
	cases := map[string][2]string{
		"host local":         {"SMTP_HOST", "mailpit"},
		"sin tls":            {"SMTP_TLS", "none"},
		"sin usuario":        {"SMTP_USERNAME", ""},
		"sin contraseña":     {"SMTP_PASSWORD", ""},
		"remitente local":    {"MAIL_FROM", "no-reply@tienda.local"},
		"frontend sin https": {"FRONTEND_URL", "http://tienda.mergemakers.com"},
	}
	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			setProductionMail(t)
			t.Setenv(env[0], env[1])
			if _, err := Load(); err == nil || !strings.Contains(err.Error(), env[0]) {
				t.Fatalf("Load() debía rechazar %s en producción; error=%v", env[0], err)
			}
		})
	}
}

func setProductionMail(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "production")
	t.Setenv("OTP_HMAC_SECRET", "a-unique-production-secret-with-32-characters")
	t.Setenv("SMTP_HOST", "email-smtp.us-east-1.amazonaws.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_TLS", "starttls")
	t.Setenv("SMTP_USERNAME", "AKIAEXAMPLE")
	t.Setenv("SMTP_PASSWORD", "secreto-de-prueba")
	t.Setenv("MAIL_FROM", "no-reply@mergemakers.com")
	t.Setenv("FRONTEND_URL", "https://tienda.mergemakers.com")
}

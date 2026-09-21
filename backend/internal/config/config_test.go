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
	t.Setenv("APP_ENV", "production")
	t.Setenv("OTP_HMAC_SECRET", "a-unique-production-secret-with-32-characters")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OTPHMACSecret == "" || cfg.OTPMaxAttempts != 5 || cfg.SMTPPort != 1025 {
		t.Fatalf("configuración OTP inválida: %#v", cfg)
	}
}

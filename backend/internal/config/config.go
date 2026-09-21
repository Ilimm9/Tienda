package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv               string
	AppPort              string
	DatabaseURL          string
	SessionDuration      time.Duration
	RememberDuration     time.Duration
	RememberIdle         time.Duration
	SessionTouchInterval time.Duration
	SessionRetention     time.Duration
	SessionCleanup       time.Duration
	OTPHMACSecret        string
	OTPTTL               time.Duration
	OTPMaxAttempts       int
	OTPResendWait        time.Duration
	OTPHourlySendMax     int
	SMTPHost             string
	SMTPPort             int
	SMTPFrom             string
	SMTPTimeout          time.Duration
	LoginHeaderImageURL  string
	FrontendURL          string
	PrecioCheckAPIKey    string
	PrecioCheckBaseURL   string
	UPCItemDBBaseURL     string
}

func Load() (Config, error) {
	_ = godotenv.Load()
	appEnv := get("APP_ENV", "development")
	sessionDuration, err := requiredDuration("SESSION_DURATION", 24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	rememberDuration, err := requiredDuration("SESSION_REMEMBER_DURATION", 30*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	rememberIdle, err := requiredDuration("SESSION_REMEMBER_IDLE", 7*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	touchInterval, err := requiredDuration("SESSION_TOUCH_INTERVAL", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}
	if rememberIdle > rememberDuration {
		return Config{}, fmt.Errorf("SESSION_REMEMBER_IDLE no puede ser mayor que SESSION_REMEMBER_DURATION")
	}
	retention, err := requiredDuration("SESSION_RETENTION", 7*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	cleanup, err := requiredDuration("SESSION_CLEANUP_INTERVAL", time.Hour)
	if err != nil {
		return Config{}, err
	}
	otpTTL, err := requiredDuration("OTP_TTL", 10*time.Minute)
	if err != nil {
		return Config{}, err
	}
	otpResendWait, err := requiredDuration("OTP_RESEND_WAIT", time.Minute)
	if err != nil {
		return Config{}, err
	}
	smtpTimeout, err := requiredDuration("SMTP_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	otpMaxAttempts, err := requiredPositiveInt("OTP_MAX_ATTEMPTS", 5)
	if err != nil {
		return Config{}, err
	}
	otpHourlySendMax, err := requiredPositiveInt("OTP_HOURLY_SEND_MAX", 5)
	if err != nil {
		return Config{}, err
	}
	smtpPort, err := requiredPositiveInt("SMTP_PORT", 1025)
	if err != nil || smtpPort > 65535 {
		return Config{}, fmt.Errorf("SMTP_PORT debe ser un puerto válido")
	}
	otpSecret := get("OTP_HMAC_SECRET", "development-only-change-otp-secret")
	if appEnv == "production" && (len(otpSecret) < 32 || otpSecret == "development-only-change-otp-secret") {
		return Config{}, fmt.Errorf("OTP_HMAC_SECRET debe configurarse en producción con al menos 32 caracteres")
	}
	return Config{
		AppEnv:               appEnv,
		AppPort:              get("APP_PORT", "8080"),
		DatabaseURL:          databaseURL(),
		SessionDuration:      sessionDuration,
		RememberDuration:     rememberDuration,
		RememberIdle:         rememberIdle,
		SessionTouchInterval: touchInterval,
		SessionRetention:     retention,
		SessionCleanup:       cleanup,
		OTPHMACSecret:        otpSecret,
		OTPTTL:               otpTTL,
		OTPMaxAttempts:       otpMaxAttempts,
		OTPResendWait:        otpResendWait,
		OTPHourlySendMax:     otpHourlySendMax,
		SMTPHost:             get("SMTP_HOST", "localhost"),
		SMTPPort:             smtpPort,
		SMTPFrom:             get("SMTP_FROM", "no-reply@tienda.local"),
		SMTPTimeout:          smtpTimeout,
		LoginHeaderImageURL:  get("LOGIN_HEADER_IMAGE_URL", ""),
		FrontendURL:          get("FRONTEND_URL", "http://localhost:4200"),
		PrecioCheckAPIKey:    get("PRECIOCHECK_API_KEY", ""),
		PrecioCheckBaseURL:   get("PRECIOCHECK_BASE_URL", "https://preciocheck.com/api/v1"),
		UPCItemDBBaseURL:     get("UPCITEMDB_BASE_URL", "https://api.upcitemdb.com/prod/trial"),
	}, nil
}

func requiredPositiveInt(key string, fallback int) (int, error) {
	value, err := strconv.Atoi(strings.TrimSpace(get(key, strconv.Itoa(fallback))))
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s debe ser un entero positivo válido", key)
	}
	return value, nil
}

func (c Config) SessionCookieName() string {
	if c.AppEnv == "production" {
		return "__Host-tienda_session"
	}
	return "tienda_session"
}

func get(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func requiredDuration(key string, fallback time.Duration) (time.Duration, error) {
	value, err := time.ParseDuration(get(key, fallback.String()))
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s debe ser una duración positiva válida", key)
	}
	return value, nil
}

func databaseURL() string {
	return "host=" + get("DB_HOST", "localhost") +
		" port=" + get("DB_PORT", "5432") +
		" user=" + get("DB_USER", "postgres") +
		" password=" + get("DB_PASSWORD", "postgres") +
		" dbname=" + get("DB_NAME", "tienda") +
		" sslmode=" + get("DB_SSLMODE", "disable")
}

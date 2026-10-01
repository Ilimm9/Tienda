package config

import (
	"fmt"
	"log"
	"net/mail"
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
	PasswordResetTTL     time.Duration
	SMTPHost             string
	SMTPPort             int
	SMTPTLS              string
	SMTPUsername         string
	SMTPPassword         string
	SMTPTimeout          time.Duration
	MailFrom             string
	MailFromName         string
	MailReplyTo          string
	SESConfigurationSet  string
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
	passwordResetTTL, err := requiredDuration("PASSWORD_RESET_TTL", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}
	smtpTimeout, err := requiredDuration("SMTP_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	otpMaxAttempts, err := requiredPositiveInt("OTP_MAX_ATTEMPTS", 10)
	if err != nil {
		return Config{}, err
	}
	otpHourlySendMax, err := requiredPositiveInt("OTP_HOURLY_SEND_MAX", 10)
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
	mailCfg := mailConfig{
		host:     get("SMTP_HOST", "localhost"),
		tls:      strings.ToLower(get("SMTP_TLS", "none")),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     mailFrom(),
		fromName: get("MAIL_FROM_NAME", "Tienda"),
		replyTo:  os.Getenv("MAIL_REPLY_TO"),
	}
	frontendURL := get("FRONTEND_URL", "http://localhost:4200")
	if err := mailCfg.validate(appEnv, frontendURL); err != nil {
		return Config{}, err
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
		PasswordResetTTL:     passwordResetTTL,
		SMTPHost:             mailCfg.host,
		SMTPPort:             smtpPort,
		SMTPTLS:              mailCfg.tls,
		SMTPUsername:         mailCfg.username,
		SMTPPassword:         mailCfg.password,
		SMTPTimeout:          smtpTimeout,
		MailFrom:             mailCfg.from,
		MailFromName:         mailCfg.fromName,
		MailReplyTo:          mailCfg.replyTo,
		SESConfigurationSet:  os.Getenv("SES_CONFIGURATION_SET"),
		LoginHeaderImageURL:  get("LOGIN_HEADER_IMAGE_URL", ""),
		FrontendURL:          frontendURL,
		PrecioCheckAPIKey:    get("PRECIOCHECK_API_KEY", ""),
		PrecioCheckBaseURL:   get("PRECIOCHECK_BASE_URL", "https://preciocheck.com/api/v1"),
		UPCItemDBBaseURL:     get("UPCITEMDB_BASE_URL", "https://api.upcitemdb.com/prod/trial"),
	}, nil
}

type mailConfig struct {
	host, tls, username, password, from, fromName, replyTo string
}

func (m mailConfig) validate(appEnv, frontendURL string) error {
	switch m.tls {
	case "none", "starttls", "tls":
	default:
		return fmt.Errorf("SMTP_TLS debe ser none, starttls o tls")
	}
	if _, err := mail.ParseAddress(m.from); err != nil {
		return fmt.Errorf("MAIL_FROM debe ser una dirección de correo válida")
	}
	if m.replyTo != "" {
		if _, err := mail.ParseAddress(m.replyTo); err != nil {
			return fmt.Errorf("MAIL_REPLY_TO debe ser una dirección de correo válida")
		}
	}
	if appEnv != "production" {
		return nil
	}
	switch strings.ToLower(m.host) {
	case "localhost", "127.0.0.1", "::1", "mailpit":
		return fmt.Errorf("SMTP_HOST no puede apuntar a un servidor local en producción")
	}
	if m.tls == "none" {
		return fmt.Errorf("SMTP_TLS debe ser starttls o tls en producción")
	}
	if m.username == "" || m.password == "" {
		return fmt.Errorf("SMTP_USERNAME y SMTP_PASSWORD son obligatorios en producción")
	}
	if strings.HasSuffix(strings.ToLower(m.from), ".local") {
		return fmt.Errorf("MAIL_FROM debe pertenecer a un dominio verificado en producción")
	}
	if !strings.HasPrefix(frontendURL, "https://") {
		return fmt.Errorf("FRONTEND_URL debe usar https en producción")
	}
	return nil
}

// mailFrom acepta SMTP_FROM como alias temporal de MAIL_FROM.
func mailFrom() string {
	if value := os.Getenv("MAIL_FROM"); value != "" {
		return value
	}
	if value := os.Getenv("SMTP_FROM"); value != "" {
		log.Printf("SMTP_FROM está obsoleta; usa MAIL_FROM")
		return value
	}
	return "no-reply@tienda.local"
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

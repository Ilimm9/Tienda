package config

import (
	"fmt"
	"os"
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
	LoginHeaderImageURL  string
	FrontendURL          string
	PrecioCheckAPIKey    string
	PrecioCheckBaseURL   string
	UPCItemDBBaseURL     string
}

func Load() (Config, error) {
	_ = godotenv.Load()
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
	return Config{
		AppEnv:               get("APP_ENV", "development"),
		AppPort:              get("APP_PORT", "8080"),
		DatabaseURL:          databaseURL(),
		SessionDuration:      sessionDuration,
		RememberDuration:     rememberDuration,
		RememberIdle:         rememberIdle,
		SessionTouchInterval: touchInterval,
		SessionRetention:     retention,
		SessionCleanup:       cleanup,
		LoginHeaderImageURL:  get("LOGIN_HEADER_IMAGE_URL", ""),
		FrontendURL:          get("FRONTEND_URL", "http://localhost:4200"),
		PrecioCheckAPIKey:    get("PRECIOCHECK_API_KEY", ""),
		PrecioCheckBaseURL:   get("PRECIOCHECK_BASE_URL", "https://preciocheck.com/api/v1"),
		UPCItemDBBaseURL:     get("UPCITEMDB_BASE_URL", "https://api.upcitemdb.com/prod/trial"),
	}, nil
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

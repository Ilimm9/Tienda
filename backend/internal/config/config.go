package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv              string
	AppPort             string
	DatabaseURL         string
	JWTSecret           string
	JWTExpiration       time.Duration
	LoginHeaderImageURL string
	FrontendURL         string
}

func Load() Config {
	_ = godotenv.Load()
	return Config{
		AppEnv:              get("APP_ENV", "development"),
		AppPort:             get("APP_PORT", "8080"),
		DatabaseURL:         databaseURL(),
		JWTSecret:           get("JWT_SECRET", "change-this-secret-in-development"),
		JWTExpiration:       duration("JWT_EXPIRATION", 24*time.Hour),
		LoginHeaderImageURL: get("LOGIN_HEADER_IMAGE_URL", ""),
		FrontendURL:         get("FRONTEND_URL", "http://localhost:4200"),
	}
}

func get(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func duration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(get(key, fallback.String()))
	if err != nil {
		return fallback
	}
	return value
}

func databaseURL() string {
	return "host=" + get("DB_HOST", "localhost") +
		" port=" + get("DB_PORT", "5432") +
		" user=" + get("DB_USER", "postgres") +
		" password=" + get("DB_PASSWORD", "postgres") +
		" dbname=" + get("DB_NAME", "tienda") +
		" sslmode=" + get("DB_SSLMODE", "disable")
}

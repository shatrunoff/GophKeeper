// Package config содержит конфигурацию приложения.
package config

import (
	"os"
	"time"
)

// Config содержит настройки приложения.
type Config struct {
	DBDSN             string
	JWTSecret         string
	JWTTTL            time.Duration
	DataEncryptionKey string
	ServerPort        string
}

// Load загружает конфигурацию из ENV.
func Load() *Config {
	return &Config{
		DBDSN:             getEnv("DB_DSN", "postgres://localhost:5432/gopherpass"),
		JWTSecret:         getEnv("JWT_SECRET", "secret"),
		JWTTTL:            24 * time.Hour,
		DataEncryptionKey: getEnv("DATA_ENCRYPTION_KEY", "12345678901234567890123456789012"),
		ServerPort:        getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

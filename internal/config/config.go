// Package config содержит конфигурацию приложения.
package config

import (
	"time"

	"gopherpass/internal/helpers"
)

// Config содержит настройки приложения.
type Config struct {
	DBDSN             string
	JWTSecret         string
	JWTTTL            time.Duration
	DataEncryptionKey string
	ServerPort        string
	ShutdownTimeout   time.Duration
}

// Load загружает конфигурацию из ENV.
func Load() *Config {
	return &Config{
		DBDSN:             helpers.GetEnv("DB_DSN", "postgres://localhost:5432/gopherpass"),
		JWTSecret:         helpers.GetEnv("JWT_SECRET", "secret"),
		JWTTTL:            24 * time.Hour,
		DataEncryptionKey: helpers.GetEnv("DATA_ENCRYPTION_KEY", "12345678901234567890123456789012"),
		ServerPort:        helpers.GetEnv("SERVER_PORT", "8080"),
		ShutdownTimeout:   helpers.GetDuration("SHUTDOWN_TIMEOUT", 5*time.Second),
	}
}

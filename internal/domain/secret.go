// Package domain содержит бизнес-модели приложения.
package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SecretType определяет тип секрета.
type SecretType string

const (
	SecretTypeCredentials SecretType = "credentials" // логин/пароль
	SecretTypeText        SecretType = "text"        // произвольный текст
	SecretTypeBinary      SecretType = "binary"      // бинарные данные
	SecretTypeCard        SecretType = "card"        // банковская карта
)

// Secret представляет зашифрованный секрет пользователя.
type Secret struct {
	ID               uuid.UUID       `json:"id"`
	UserID           uuid.UUID       `json:"user_id"`
	Type             SecretType      `json:"type"`
	EncryptedPayload []byte          `json:"-"`
	Meta             json.RawMessage `json:"meta,omitempty"`
	Version          int64           `json:"version"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

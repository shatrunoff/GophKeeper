// Package domain содержит бизнес-модели приложения.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя системы.
type User struct {
	ID           uuid.UUID `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Package repository содержит интерфейсы репозиториев.
package repository

import (
	"context"
	"time"

	"gopherpass/internal/domain"

	"github.com/google/uuid"
)

// UserRepository определяет методы работы с пользователями.
type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	GetByLogin(ctx context.Context, login string) (*domain.User, error)
}

// SecretRepository определяет методы работы с секретами.
type SecretRepository interface {
	Upsert(ctx context.Context, secret domain.Secret) error
	GetAll(ctx context.Context, userID uuid.UUID) ([]domain.Secret, error)
	GetUpdatedAfter(ctx context.Context, userID uuid.UUID, t time.Time) ([]domain.Secret, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

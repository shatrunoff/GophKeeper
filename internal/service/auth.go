// Package service содержит бизнес-логику приложения.
package service

import (
	"context"
	"time"

	"gopherpass/internal/domain"
	apperrors "gopherpass/internal/errors"
	"gopherpass/internal/repository"
	"gopherpass/internal/security"

	"github.com/google/uuid"
)

// AuthService реализует логику аутентификации.
type AuthService struct {
	userRepo repository.UserRepository
	jwt      *security.JWTManager
}

// NewAuthService создаёт AuthService.
func NewAuthService(userRepo repository.UserRepository, jwt *security.JWTManager) *AuthService {
	return &AuthService{userRepo: userRepo, jwt: jwt}
}

// Register регистрирует нового пользователя.
func (s *AuthService) Register(ctx context.Context, login, password string) (*domain.User, error) {
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := domain.User{
		ID:           uuid.New(),
		Login:        login,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Login выполняет вход и возвращает JWT токен.
func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		return "", apperrors.ErrNotFound
	}
	if err := security.ComparePassword(user.PasswordHash, password); err != nil {
		return "", apperrors.ErrNotFound
	}
	return s.jwt.Generate(user.ID)
}

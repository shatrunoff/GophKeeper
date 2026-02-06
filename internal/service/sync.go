// Package service содержит бизнес-логику приложения.
package service

import (
	"context"
	"time"

	"gopherpass/internal/domain"
	"gopherpass/internal/repository"
	"gopherpass/internal/security"

	"github.com/google/uuid"
)

// SyncService реализует синхронизацию секретов.
type SyncService struct {
	secretRepo repository.SecretRepository
	crypto     *security.Crypto
}

// NewSyncService создаёт SyncService.
func NewSyncService(secretRepo repository.SecretRepository, crypto *security.Crypto) *SyncService {
	return &SyncService{secretRepo: secretRepo, crypto: crypto}
}

// GetUpdates возвращает секреты, обновлённые после указанного времени (Last Write Wins).
func (s *SyncService) GetUpdates(ctx context.Context, userID uuid.UUID, since time.Time) ([]domain.Secret, error) {
	secrets, err := s.secretRepo.GetUpdatedAfter(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	for i := range secrets {
		decrypted, err := s.crypto.Decrypt(secrets[i].EncryptedPayload)
		if err != nil {
			return nil, err
		}
		secrets[i].EncryptedPayload = decrypted
	}
	return secrets, nil
}

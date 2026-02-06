// Package service содержит бизнес-логику приложения.
package service

import (
	"context"
	"encoding/json"
	"time"

	"gopherpass/internal/domain"
	"gopherpass/internal/repository"
	"gopherpass/internal/security"

	"github.com/google/uuid"
)

// DataService управляет секретами пользователя.
type DataService struct {
	secretRepo repository.SecretRepository
	crypto     *security.Crypto
}

// NewDataService создаёт DataService.
func NewDataService(secretRepo repository.SecretRepository, crypto *security.Crypto) *DataService {
	return &DataService{secretRepo: secretRepo, crypto: crypto}
}

// SaveSecret сохраняет секрет с шифрованием и инкрементом версии.
func (s *DataService) SaveSecret(ctx context.Context, userID uuid.UUID, id uuid.UUID, secretType domain.SecretType, payload []byte, meta json.RawMessage, version int64) error {
	encrypted, err := s.crypto.Encrypt(payload)
	if err != nil {
		return err
	}
	secret := domain.Secret{
		ID:               id,
		UserID:           userID,
		Type:             secretType,
		EncryptedPayload: encrypted,
		Meta:             meta,
		Version:          version + 1,
		UpdatedAt:        time.Now(),
	}
	return s.secretRepo.Upsert(ctx, secret)
}

// GetSecrets возвращает все секреты пользователя (расшифрованные).
func (s *DataService) GetSecrets(ctx context.Context, userID uuid.UUID) ([]domain.Secret, error) {
	secrets, err := s.secretRepo.GetAll(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.decryptSecrets(secrets)
}

// DeleteSecret удаляет секрет пользователя.
func (s *DataService) DeleteSecret(ctx context.Context, id, userID uuid.UUID) error {
	return s.secretRepo.Delete(ctx, id, userID)
}

func (s *DataService) decryptSecrets(secrets []domain.Secret) ([]domain.Secret, error) {
	for i := range secrets {
		decrypted, err := s.crypto.Decrypt(secrets[i].EncryptedPayload)
		if err != nil {
			return nil, err
		}
		secrets[i].EncryptedPayload = decrypted
	}
	return secrets, nil
}

// Package postgres содержит PostgreSQL реализацию репозиториев.
package postgres

import (
	"context"
	"time"

	"gopherpass/internal/domain"
	apperrors "gopherpass/internal/errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SecretRepo реализует repository.SecretRepository.
type SecretRepo struct {
	pool *pgxpool.Pool
}

// NewSecretRepo создаёт SecretRepo.
func NewSecretRepo(pool *pgxpool.Pool) *SecretRepo {
	return &SecretRepo{pool: pool}
}

// Upsert создаёт или обновляет секрет.
func (r *SecretRepo) Upsert(ctx context.Context, s domain.Secret) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO secrets (id, user_id, type, encrypted_payload, meta, version, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			type = EXCLUDED.type,
			encrypted_payload = EXCLUDED.encrypted_payload,
			meta = EXCLUDED.meta,
			version = EXCLUDED.version,
			updated_at = EXCLUDED.updated_at
		WHERE secrets.user_id = EXCLUDED.user_id`,
		s.ID, s.UserID, s.Type, s.EncryptedPayload, s.Meta, s.Version, s.UpdatedAt)
	return err
}

// GetAll возвращает все секреты пользователя.
func (r *SecretRepo) GetAll(ctx context.Context, userID uuid.UUID) ([]domain.Secret, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, type, encrypted_payload, meta, version, updated_at 
		 FROM secrets WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSecrets(rows)
}

// GetUpdatedAfter возвращает секреты, обновлённые после указанного времени.
func (r *SecretRepo) GetUpdatedAfter(ctx context.Context, userID uuid.UUID, t time.Time) ([]domain.Secret, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, type, encrypted_payload, meta, version, updated_at 
		 FROM secrets WHERE user_id = $1 AND updated_at > $2`, userID, t)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSecrets(rows)
}

// Delete удаляет секрет.
func (r *SecretRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	res, err := r.pool.Exec(ctx,
		`DELETE FROM secrets WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func scanSecrets(rows pgx.Rows) ([]domain.Secret, error) {
	var secrets []domain.Secret
	for rows.Next() {
		var s domain.Secret
		if err := rows.Scan(&s.ID, &s.UserID, &s.Type, &s.EncryptedPayload, &s.Meta, &s.Version, &s.UpdatedAt); err != nil {
			return nil, err
		}
		secrets = append(secrets, s)
	}
	return secrets, rows.Err()
}

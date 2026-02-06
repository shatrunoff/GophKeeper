// Package postgres содержит PostgreSQL реализацию репозиториев.
package postgres

import (
	"context"
	"errors"

	"gopherpass/internal/domain"
	apperrors "gopherpass/internal/errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepo реализует repository.UserRepository.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo создаёт UserRepo.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// Create создаёт пользователя.
func (r *UserRepo) Create(ctx context.Context, user domain.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, login, password_hash, created_at) VALUES ($1, $2, $3, $4)`,
		user.ID, user.Login, user.PasswordHash, user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrAlreadyExists
		}
		return err
	}
	return nil
}

// GetByLogin возвращает пользователя по логину.
func (r *UserRepo) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, login, password_hash, created_at FROM users WHERE login = $1`, login).
		Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Package postgres содержит базовую реализацию generic repository.
package postgres

import (
	"context"
	"errors"

	apperrors "gopherpass/internal/errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BaseRepo предоставляет базовые CRUD операции.
type BaseRepo[T any] struct {
	pool      *pgxpool.Pool
	tableName string
	scanFunc  func(row pgx.Row) (*T, error)
}

// NewBaseRepo создаёт базовый репозиторий.
func NewBaseRepo[T any](pool *pgxpool.Pool, tableName string, scanFunc func(row pgx.Row) (*T, error)) *BaseRepo[T] {
	return &BaseRepo[T]{
		pool:      pool,
		tableName: tableName,
		scanFunc:  scanFunc,
	}
}

// GetByID возвращает сущность по ID.
func (r *BaseRepo[T]) GetByID(ctx context.Context, id uuid.UUID) (*T, error) {
	row := r.pool.QueryRow(ctx, `SELECT * FROM `+r.tableName+` WHERE id = $1`, id)
	entity, err := r.scanFunc(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	return entity, err
}

// Delete удаляет сущность по ID.
func (r *BaseRepo[T]) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM `+r.tableName+` WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

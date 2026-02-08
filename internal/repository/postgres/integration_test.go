package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"gopherpass/internal/domain"
	apperrors "gopherpass/internal/errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "postgres://test:test@localhost:5433/gopherpass_test"
	}
	var err error
	testPool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(err)
	}
	code := m.Run()
	testPool.Close()
	os.Exit(code)
}

func cleanupUsers(t *testing.T) {
	t.Helper()
	testPool.Exec(context.Background(), "DELETE FROM secrets")
	testPool.Exec(context.Background(), "DELETE FROM users")
}

func TestUserRepo_Create(t *testing.T) {
	cleanupUsers(t)
	repo := NewUserRepo(testPool)

	user := domain.User{
		ID:           uuid.New(),
		Login:        "testuser",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
	}
	err := repo.Create(context.Background(), user)
	if err != nil {
		t.Fatal(err)
	}

	// Duplicate
	err = repo.Create(context.Background(), user)
	if err != apperrors.ErrAlreadyExists {
		t.Error("expected ErrAlreadyExists")
	}
}

func TestUserRepo_GetByLogin(t *testing.T) {
	cleanupUsers(t)
	repo := NewUserRepo(testPool)

	user := domain.User{
		ID:           uuid.New(),
		Login:        "findme",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
	}
	repo.Create(context.Background(), user)

	found, err := repo.GetByLogin(context.Background(), "findme")
	if err != nil {
		t.Fatal(err)
	}
	if found.ID != user.ID {
		t.Error("ID mismatch")
	}

	_, err = repo.GetByLogin(context.Background(), "notexist")
	if err != apperrors.ErrNotFound {
		t.Error("expected ErrNotFound")
	}
}

func TestSecretRepo_Upsert(t *testing.T) {
	cleanupUsers(t)
	userRepo := NewUserRepo(testPool)
	secretRepo := NewSecretRepo(testPool)

	user := domain.User{ID: uuid.New(), Login: "secretuser", PasswordHash: "h", CreatedAt: time.Now()}
	userRepo.Create(context.Background(), user)

	secret := domain.Secret{
		ID:               uuid.New(),
		UserID:           user.ID,
		Type:             domain.SecretTypeText,
		EncryptedPayload: []byte("data"),
		Version:          1,
		UpdatedAt:        time.Now(),
	}
	err := secretRepo.Upsert(context.Background(), secret)
	if err != nil {
		t.Fatal(err)
	}

	// Update
	secret.Version = 2
	err = secretRepo.Upsert(context.Background(), secret)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSecretRepo_GetAll(t *testing.T) {
	cleanupUsers(t)
	userRepo := NewUserRepo(testPool)
	secretRepo := NewSecretRepo(testPool)

	user := domain.User{ID: uuid.New(), Login: "getalluser", PasswordHash: "h", CreatedAt: time.Now()}
	userRepo.Create(context.Background(), user)

	secretRepo.Upsert(context.Background(), domain.Secret{
		ID: uuid.New(), UserID: user.ID, Type: domain.SecretTypeText,
		EncryptedPayload: []byte("1"), Version: 1, UpdatedAt: time.Now(),
	})
	secretRepo.Upsert(context.Background(), domain.Secret{
		ID: uuid.New(), UserID: user.ID, Type: domain.SecretTypeText,
		EncryptedPayload: []byte("2"), Version: 1, UpdatedAt: time.Now(),
	})

	secrets, err := secretRepo.GetAll(context.Background(), user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(secrets) != 2 {
		t.Errorf("expected 2, got %d", len(secrets))
	}
}

func TestSecretRepo_Delete(t *testing.T) {
	cleanupUsers(t)
	userRepo := NewUserRepo(testPool)
	secretRepo := NewSecretRepo(testPool)

	user := domain.User{ID: uuid.New(), Login: "deluser", PasswordHash: "h", CreatedAt: time.Now()}
	userRepo.Create(context.Background(), user)

	secretID := uuid.New()
	secretRepo.Upsert(context.Background(), domain.Secret{
		ID: secretID, UserID: user.ID, Type: domain.SecretTypeText,
		EncryptedPayload: []byte("del"), Version: 1, UpdatedAt: time.Now(),
	})

	err := secretRepo.Delete(context.Background(), secretID, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = secretRepo.Delete(context.Background(), secretID, user.ID)
	if err != apperrors.ErrNotFound {
		t.Error("expected ErrNotFound")
	}
}

// Package security содержит функции для работы с паролями.
package security

import "golang.org/x/crypto/bcrypt"

// HashPassword хеширует пароль с использованием bcrypt.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// ComparePassword сравнивает пароль с хешем.
func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

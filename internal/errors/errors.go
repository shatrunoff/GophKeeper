// Package errors содержит доменные ошибки приложения.
package errors

import "errors"

var (
	// ErrNotFound возвращается когда запрашиваемый ресурс не найден.
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists возвращается при попытке создать дубликат.
	ErrAlreadyExists = errors.New("already exists")
)

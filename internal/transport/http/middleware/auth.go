// Package middleware содержит HTTP middleware.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"gopherpass/internal/security"

	"github.com/google/uuid"
)

type ctxKey string

const UserIDKey ctxKey = "user_id"

// Auth возвращает middleware для проверки JWT.
func Auth(jwtMgr *security.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			claims, err := jwtMgr.Validate(strings.TrimPrefix(auth, "Bearer "))
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext извлекает user_id из контекста.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

// Package http содержит HTTP транспортный слой приложения.
package http

import (
	"net/http"

	"gopherpass/internal/security"
	"gopherpass/internal/transport/http/handlers"
	"gopherpass/internal/transport/http/middleware"
)

// NewRouter создаёт HTTP роутер.
func NewRouter(authH *handlers.AuthHandler, dataH *handlers.DataHandler, syncH *handlers.SyncHandler, jwtMgr *security.JWTManager) http.Handler {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("POST /api/register", authH.Register)
	mux.HandleFunc("POST /api/login", authH.Login)

	// Protected routes
	auth := middleware.Auth(jwtMgr)
	mux.Handle("POST /api/secrets", auth(http.HandlerFunc(dataH.Create)))
	mux.Handle("GET /api/secrets", auth(http.HandlerFunc(dataH.GetAll)))
	mux.Handle("DELETE /api/secrets/{id}", auth(http.HandlerFunc(dataH.Delete)))
	mux.Handle("GET /api/sync", auth(http.HandlerFunc(syncH.GetUpdates)))

	// Apply global middleware
	return middleware.Recovery(middleware.Logger(mux))
}

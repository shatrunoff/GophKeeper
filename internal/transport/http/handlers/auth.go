// Package handlers содержит HTTP обработчики для API.
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	apperrors "gopherpass/internal/errors"
	"gopherpass/internal/service"
)

// AuthHandler обрабатывает запросы аутентификации.
type AuthHandler struct {
	authSvc *service.AuthService
}

// NewAuthHandler создаёт AuthHandler.
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID string `json:"id"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// Register обрабатывает POST /api/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	user, err := h.authSvc.Register(r.Context(), req.Login, req.Password)
	if errors.Is(err, apperrors.ErrAlreadyExists) {
		http.Error(w, "user already exists", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(registerResponse{ID: user.ID.String()})
}

// Login обрабатывает POST /api/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	token, err := h.authSvc.Login(r.Context(), req.Login, req.Password)
	if errors.Is(err, apperrors.ErrInvalidCredentials) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(loginResponse{Token: token})
}

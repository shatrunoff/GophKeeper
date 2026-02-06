// Package handlers содержит HTTP обработчики.
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"gopherpass/internal/domain"
	apperrors "gopherpass/internal/errors"
	"gopherpass/internal/service"
	"gopherpass/internal/transport/http/middleware"

	"github.com/google/uuid"
)

// DataHandler обрабатывает запросы секретов.
type DataHandler struct {
	dataSvc *service.DataService
}

// NewDataHandler создаёт DataHandler.
func NewDataHandler(dataSvc *service.DataService) *DataHandler {
	return &DataHandler{dataSvc: dataSvc}
}

type secretRequest struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Meta    json.RawMessage `json:"meta,omitempty"`
	Version int64           `json:"version"`
}

type secretResponse struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Meta      json.RawMessage `json:"meta,omitempty"`
	Version   int64           `json:"version"`
	UpdatedAt string          `json:"updated_at"`
}

// Create обрабатывает POST /api/secrets.
func (h *DataHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req secretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	id, _ := uuid.Parse(req.ID)
	if id == uuid.Nil {
		id = uuid.New()
	}
	if err := h.dataSvc.SaveSecret(r.Context(), userID, id, domain.SecretType(req.Type), req.Payload, req.Meta, req.Version); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id.String()})
}

// GetAll обрабатывает GET /api/secrets.
func (h *DataHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	secrets, err := h.dataSvc.GetSecrets(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	resp := make([]secretResponse, len(secrets))
	for i, s := range secrets {
		resp[i] = secretResponse{
			ID:        s.ID.String(),
			Type:      string(s.Type),
			Payload:   s.EncryptedPayload,
			Meta:      s.Meta,
			Version:   s.Version,
			UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	json.NewEncoder(w).Encode(resp)
}

// Delete обрабатывает DELETE /api/secrets/{id}.
func (h *DataHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.dataSvc.DeleteSecret(r.Context(), id, userID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

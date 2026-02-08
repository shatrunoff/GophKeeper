// Package handlers содержит HTTP обработчики для API.
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"gopherpass/internal/service"
	"gopherpass/internal/transport/http/middleware"
)

// SyncHandler обрабатывает запросы синхронизации.
type SyncHandler struct {
	syncSvc *service.SyncService
}

// NewSyncHandler создаёт SyncHandler.
func NewSyncHandler(syncSvc *service.SyncService) *SyncHandler {
	return &SyncHandler{syncSvc: syncSvc}
}

// GetUpdates обрабатывает GET /api/sync?since=<timestamp>.
func (h *SyncHandler) GetUpdates(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	sinceStr := r.URL.Query().Get("since")
	var since time.Time
	if sinceStr != "" {
		var err error
		since, err = time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			http.Error(w, "invalid since format", http.StatusBadRequest)
			return
		}
	}
	secrets, err := h.syncSvc.GetUpdates(r.Context(), userID, since)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	type resp struct {
		ID        string          `json:"id"`
		Type      string          `json:"type"`
		Payload   json.RawMessage `json:"payload"`
		Meta      json.RawMessage `json:"meta,omitempty"`
		Version   int64           `json:"version"`
		UpdatedAt string          `json:"updated_at"`
	}
	result := make([]resp, len(secrets))
	for i, s := range secrets {
		result[i] = resp{
			ID:        s.ID.String(),
			Type:      string(s.Type),
			Payload:   s.EncryptedPayload,
			Meta:      s.Meta,
			Version:   s.Version,
			UpdatedAt: s.UpdatedAt.Format(time.RFC3339),
		}
	}
	json.NewEncoder(w).Encode(result)
}

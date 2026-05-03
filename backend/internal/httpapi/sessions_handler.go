package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/phucvd2512/agentic-in-production/backend/internal/sessions"
)

type SessionsHandler struct{ Sessions *sessions.Store }

func (h SessionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AgentName string `json:"agent_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	user := UserFromCtx(r.Context())
	sess, err := h.Sessions.Create(r.Context(), user, req.AgentName)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":         sess.ID,
		"agent_name": sess.AgentName,
		"created_at": sess.CreatedAt,
	})
}

func (h SessionsHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	msgs, err := h.Sessions.ListMessages(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(msgs)
}

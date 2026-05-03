package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/phucvd2512/agentic-in-production/backend/internal/audit"
)

type AuditHandler struct{ Log *audit.Log }

func (h AuditHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entries, err := h.Log.List(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(entries)
}

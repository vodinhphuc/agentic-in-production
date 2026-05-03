package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/phucvd2512/agentic-in-production/backend/internal/agentregistry"
)

type AgentsHandler struct{ Registry *agentregistry.Store }

func (h AgentsHandler) List(w http.ResponseWriter, r *http.Request) {
	agents, err := h.Registry.ListEnabled(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(agents)
}

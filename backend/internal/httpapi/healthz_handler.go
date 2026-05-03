package httpapi

import (
	"encoding/json"
	"net/http"
)

type HealthzHandler struct {
	Version string
}

func (h HealthzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"version": h.Version,
	})
}

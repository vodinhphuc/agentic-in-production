package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Deps struct {
	Version string
	// More deps wired in later tasks (sessions store, adapter registry, audit, auth)
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Get("/api/healthz", HealthzHandler{Version: d.Version}.ServeHTTP)
	return r
}

package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/phucvd2512/agentic-in-production/backend/internal/agentregistry"
	"github.com/phucvd2512/agentic-in-production/backend/internal/audit"
	"github.com/phucvd2512/agentic-in-production/backend/internal/auth"
	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
	"github.com/phucvd2512/agentic-in-production/backend/internal/sessions"
)

type Deps struct {
	Version  string
	DB       *db.DB
	Auth     *auth.Authenticator
	Sessions *sessions.Store
	Registry *agentregistry.Store
	Audit    *audit.Log
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/api/healthz", HealthzHandler{Version: d.Version}.ServeHTTP)
	r.Post("/api/auth/login", AuthHandler{Auth: d.Auth}.Login)

	r.Group(func(r chi.Router) {
		r.Use(RequireAuth(d.Auth))
		r.Get("/api/agents", AgentsHandler{Registry: d.Registry}.List)
		r.Post("/api/sessions", SessionsHandler{Sessions: d.Sessions}.Create)
		r.Get("/api/sessions/{id}/messages", SessionsHandler{Sessions: d.Sessions}.ListMessages)
		r.Post("/api/sessions/{id}/messages",
			MessagesHandler{Sessions: d.Sessions, Registry: d.Registry, Audit: d.Audit}.Send)
		r.Get("/api/sessions/{id}/audit", AuditHandler{Log: d.Audit}.Get)
	})
	return r
}

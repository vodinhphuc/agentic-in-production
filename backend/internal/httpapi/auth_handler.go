package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/phucvd2512/agentic-in-production/backend/internal/auth"
)

const cookieName = "aip_session"

type AuthHandler struct {
	Auth *auth.Authenticator
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username, Password string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	tok, err := h.Auth.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Hour),
	})
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"token": tok})
}

type ctxKey int

const userKey ctxKey = 1

func RequireAuth(a *auth.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(cookieName)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			claims, err := a.ParseToken(c.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromCtx(ctx context.Context) string {
	v, _ := ctx.Value(userKey).(string)
	return v
}

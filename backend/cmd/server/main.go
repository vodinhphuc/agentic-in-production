package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/phucvd2512/agentic-in-production/backend/internal/auth"
	"github.com/phucvd2512/agentic-in-production/backend/internal/db"
	"github.com/phucvd2512/agentic-in-production/backend/internal/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()
	d, err := db.Open(ctx)
	if err != nil {
		slog.Error("db.Open", "err", err)
		os.Exit(1)
	}
	defer d.Close()

	a := auth.NewAuthenticator(
		[]byte(mustEnv("JWT_SIGNING_KEY")),
		envOr("ADMIN_USERNAME", "admin"),
		os.Getenv("ADMIN_PASSWORD_HASH"),
		time.Hour,
	)

	port := envOr("BACKEND_PORT", "8080")
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewRouter(httpapi.Deps{Version: "0.0.0-dev", DB: d, Auth: a}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		slog.Error("missing required env", "key", k)
		os.Exit(1)
	}
	return v
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

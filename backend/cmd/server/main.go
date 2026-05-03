package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

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

	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewRouter(httpapi.Deps{Version: "0.0.0-dev", DB: d}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}

package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// buildRouter constructs the chi router for the application.
//
// No migrated package in this project currently exposes an exported
// RegisterRoutes(chi.Router) function: internal/api provides error-handling
// helpers (ErrorHandler, WriteError) but no concrete route handlers, and
// internal/service/internal/model define no HTTP endpoints. There is
// therefore nothing to mount beyond the health check, and no placeholder
// routes are added. When a route-owning package gains a RegisterRoutes
// function, it should be invoked here.
func buildRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	return r
}
package main

import (
	"net/http"

	"migrated-app/internal/api"
	"migrated-app/internal/service"
	"migrated-app/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// buildRouter constructs the chi router for the application.
func buildRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	dsn := store.MySQLDSNFromEnv()
	st, err := store.New(dsn)
	if err != nil {
		panic(err)
	}
	svc, err := service.NewUserService(st)
	if err != nil {
		panic(err)
	}
	h := &userHandler{svc: svc}

	r.Post("/save_user_data", api.ErrorHandler(h.SaveUserData).ServeHTTP)
	r.Get("/get_user_data", api.ErrorHandler(h.GetUserData).ServeHTTP)
	r.Get("/get_user_data/{id}", api.ErrorHandler(h.GetUserDataById).ServeHTTP)
	r.Get("/get_user_name/name/{name}", api.ErrorHandler(h.GetUserNameByName).ServeHTTP)
	r.Put("/update_user_data/{id}", api.ErrorHandler(h.UpdateUserData).ServeHTTP)
	r.Delete("/delete_user_data/{id}", api.ErrorHandler(h.DeleteUserData).ServeHTTP)

	return r
}
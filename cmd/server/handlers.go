package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"migrated-app/internal/api"
	"migrated-app/internal/service"
)

// userHandler serves the user-related HTTP routes of the application,
// replacing the Spring @RestController com.smartContact.controller controller
// class.
type userHandler struct {
	svc service.UserService
}

// GetUserData handles GET /get_user_data, returning every user as JSON.
func (h *userHandler) GetUserData(w http.ResponseWriter, r *http.Request) error {
	users, err := h.svc.FetchUserList(r.Context())
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, users)
}

// GetUserDataById handles GET /get_user_data/{id}, returning one user.
func (h *userHandler) GetUserDataById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return api.NewHTTPStatusError(http.StatusBadRequest, errors.New("invalid user id"))
	}
	user, err := h.svc.FetchUserById(r.Context(), id)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, user)
}

// GetUserNameByName handles GET /get_user_name/name/{name}.
func (h *userHandler) GetUserNameByName(w http.ResponseWriter, r *http.Request) error {
	name := r.PathValue("name")
	user, err := h.svc.GetUserNameByName(r.Context(), name)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, user)
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

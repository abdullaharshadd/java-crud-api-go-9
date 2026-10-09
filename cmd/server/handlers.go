package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"migrated-app/internal/api"
	"migrated-app/internal/model"
	"migrated-app/internal/service"

	"github.com/go-chi/chi/v5"
)

// userHandler serves the user-related HTTP routes of the application,
// replacing the Spring @RestController com.smartContact.controller controller
// class.
type userHandler struct {
	svc service.UserService
}

// SaveUserData handles POST /save_user_data, persisting the user in the
// request body. The original Spring controller returned a plain-text
// confirmation string, so the response is plain text too.
func (h *userHandler) SaveUserData(w http.ResponseWriter, r *http.Request) error {
	var user model.User
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return api.NewHTTPStatusError(http.StatusBadRequest, errors.New("invalid request body"))
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return api.NewHTTPStatusError(http.StatusBadRequest, errors.New("invalid user payload"))
	}
	if err := user.Validate(); err != nil {
		return api.NewHTTPStatusError(http.StatusBadRequest, err)
	}
	if _, err := h.svc.SaveUser(r.Context(), &user); err != nil {
		return err
	}
	return writeRaw(w, http.StatusOK, "User data saved successfully!")
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
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return api.NewHTTPStatusError(http.StatusBadRequest, errors.New("invalid user id"))
	}
	user, err := h.svc.FetchUserById(r.Context(), id)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, user)
}

// GetUserNameByName handles GET /get_user_name/name/{name}. The original
// Spring controller returned the matched user's name-or-null body with a
// 200 status even when nothing matched, so a missing user yields a 200 with
// a JSON null body rather than a 404.
func (h *userHandler) GetUserNameByName(w http.ResponseWriter, r *http.Request) error {
	name := chi.URLParam(r, "name")
	user, err := h.svc.GetUserNameByName(r.Context(), name)
	if err != nil {
		if errors.Is(err, api.ErrUserNotFound) {
			return writeJSON(w, http.StatusOK, nil)
		}
		return err
	}
	return writeJSON(w, http.StatusOK, user.Name)
}

// UpdateUserData handles PUT /update_user_data/{id}, replacing the user's
// data and returning the updated entity as JSON.
func (h *userHandler) UpdateUserData(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return api.NewHTTPStatusError(http.StatusBadRequest, errors.New("invalid user id"))
	}
	var user model.User
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return api.NewHTTPStatusError(http.StatusBadRequest, errors.New("invalid request body"))
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return api.NewHTTPStatusError(http.StatusBadRequest, errors.New("invalid user payload"))
	}
	if err := h.svc.UpdateUser(r.Context(), id, &user); err != nil {
		return err
	}
	saved, err := h.svc.FetchUserById(r.Context(), id)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, saved)
}

// DeleteUserData handles DELETE /delete_user_data/{id}. The original Spring
// controller returned a plain-text confirmation string.
func (h *userHandler) DeleteUserData(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return api.NewHTTPStatusError(http.StatusBadRequest, errors.New("invalid user id"))
	}
	if err := h.svc.DeleteUser(r.Context(), id); err != nil {
		return err
	}
	return writeRaw(w, http.StatusOK, "user data deleted Successfully")
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// writeRaw writes a plain-text body, matching the original controller's
// @ResponseBody String return values.
func writeRaw(w http.ResponseWriter, status int, text string) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, err := w.Write([]byte(text))
	return err
}

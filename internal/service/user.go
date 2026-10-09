// Package service contains the service-layer contracts of the smartContact
// application. This file replaces the Spring service interface
// com.smartContact.service.UserService.
//
// MIGRATION_NOTE: The Java interface declared a checked exception
// (UserNotFoundException) on fetchUserById. In Go the convention is to
// return an error; implementations must return (or wrap) the sentinel
// api.ErrUserNotFound when a user cannot be found, which handlers detect
// with errors.Is to produce a 404 response.
package service

import (
	"context"

	"migrated-app/internal/model"
)

// UserService defines the business operations available for User entities.
// It replaces the Java interface UserService; a concrete implementation
// backed by the user store is injected via constructor (manual DI).
type UserService interface {
	// SaveUser persists the given user and returns the saved entity,
	// including its database-assigned ID.
	SaveUser(ctx context.Context, user *model.User) (*model.User, error)

	// FetchUserList returns all users.
	FetchUserList(ctx context.Context) ([]model.User, error)

	// FetchUserById returns the user with the given ID.
	// It returns an error wrapping api.ErrUserNotFound if no such user exists
	// (replacing the Java `throws UserNotFoundException`).
	FetchUserById(ctx context.Context, id int) (*model.User, error)

	// DeleteUser removes the user with the given ID.
	DeleteUser(ctx context.Context, id int) error

	// UpdateUser replaces the data of the user identified by id with the
	// fields of the provided user object.
	UpdateUser(ctx context.Context, id int, user *model.User) error

	// GetUserNameByName returns the user whose (unique) name matches the
	// given string.
	GetUserNameByName(ctx context.Context, name string) (*model.User, error)
}

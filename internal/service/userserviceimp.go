// Package service contains the service-layer contracts of the smartContact
// application. This file implements the Spring service class
// com.smartContact.service.UserServiceImp (the @Service concrete
// implementation); the UserService interface it implements already lives in
// internal/service/user.go and is NOT redeclared here.
//
// MIGRATION_NOTE: The Java interface declared a checked exception
// (UserNotFoundException) on fetchUserById. In Go the convention is to
// return an error; implementations return (or wrap) the sentinel
// api.ErrUserNotFound when a user cannot be found, which handlers detect
// with errors.Is to produce a 404 response.
package service

import (
	"context"
	"fmt"

	"migrated-app/internal/model"
)

// UserStore is the persistence contract UserServiceImp depends on. It
// replaces the Spring Data JPA repository interface
// com.smartContact.repository.UserDao, whose derived query methods
// (findById, findAll, save, deleteById, findByName) Spring generated from
// method names. Here the store is an explicit interface injected through the
// constructor, so a MySQL-backed implementation (internal/store) and a fake
// for table-driven tests can both satisfy it.
//
// MIGRATION_NOTE: In Java, save handled both insert and update implicitly
// (via JPA's merge semantics); Go keeps that same dual behavior on the store
// side (Save inserts when ID is zero and updates otherwise).
type UserStore interface {
	// Save persists a user, returning the stored entity with its
	// database-assigned ID populated.
	Save(ctx context.Context, user *model.User) (*model.User, error)

	// FindByID returns the user with the given ID, or an error wrapping
	// api.ErrUserNotFound when it does not exist (replacing
	// Optional<User>).
	FindByID(ctx context.Context, id int) (*model.User, error)

	// FindAll returns every user in the database.
	FindAll(ctx context.Context) ([]model.User, error)

	// DeleteByID removes the user with the given ID.
	DeleteByID(ctx context.Context, id int) error

	// FindByName returns the user whose name matches exactly, or an error
	// wrapping api.ErrUserNotFound when it does not exist. In the source
	// this was a derived query method returning a bare User (null when
	// absent); Go surfaces the null case as ErrUserNotFound.
	FindByName(ctx context.Context, name string) (*model.User, error)
}

// UserServiceImp is the concrete implementation of UserService. It replaces
// the Spring @Service class UserServiceImp.
//
// MIGRATION_NOTE: Spring's @Autowired field injection has no Go equivalent;
// the dependency is supplied via NewUserService (constructor injection),
// which keeps the dependency explicit and makes the type trivially testable
// with a fake store. The @NotNull hint on updateUser's id parameter is not
// migrated: a primitive int cannot be null, so the constraint was inert in
// the source (it was never enforced — the class lacks @Validated).
type UserServiceImp struct {
	store UserStore
}

// Compile-time assertion that UserServiceImp satisfies UserService.
var _ UserService = (*UserServiceImp)(nil)

// NewUserService constructs a UserServiceImp backed by the given store,
// replacing Spring's implicit bean wiring of the UserDao into the service.
func NewUserService(store UserStore) (*UserServiceImp, error) {
	if store == nil {
		return nil, fmt.Errorf("service: user store must not be nil")
	}
	return &UserServiceImp{store: store}, nil
}

// SaveUser persists the given user and returns the saved entity, including
// its database-assigned ID. It replaces saveUser(User).
func (s *UserServiceImp) SaveUser(ctx context.Context, user *model.User) (*model.User, error) {
	if user == nil {
		return nil, fmt.Errorf("service: user must not be nil")
	}
	saved, err := s.store.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("service: saving user: %w", err)
	}
	return saved, nil
}

// FetchUserList returns all users from the database. It replaces
// fetchUserList().
func (s *UserServiceImp) FetchUserList(ctx context.Context) ([]model.User, error) {
	users, err := s.store.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: fetching user list: %w", err)
	}
	return users, nil
}

// FetchUserById returns the user with the given ID, or an error wrapping
// api.ErrUserNotFound when no such user exists. It replaces
// fetchUserById(int) throws UserNotFoundException: the Optional.isPresent
// check and explicit throw are collapsed into the store's (T, error)
// contract, and the sentinel carries the original message
// "User are not available".
func (s *UserServiceImp) FetchUserById(ctx context.Context, id int) (*model.User, error) {
	user, err := s.store.FindByID(ctx, id)
	if err != nil {
		return nil, err // store already wraps api.ErrUserNotFound; preserve it
	}
	return user, nil
}

// DeleteUser removes the user with the given ID, delegating directly to the
// store. It replaces deleteUser(int).
func (s *UserServiceImp) DeleteUser(ctx context.Context, id int) error {
	if err := s.store.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("service: deleting user %d: %w", id, err)
	}
	return nil
}

// UpdateUser overwrites the ID of the given user with the path id and saves
// the entity, matching the source's `user.setId(id); userDao.save(user);`.
// It replaces updateUser(@NotNull int id, User user).
func (s *UserServiceImp) UpdateUser(ctx context.Context, id int, user *model.User) error {
	if user == nil {
		return fmt.Errorf("service: user must not be nil")
	}
	user.ID = id
	if _, err := s.store.Save(ctx, user); err != nil {
		return fmt.Errorf("service: updating user %d: %w", id, err)
	}
	return nil
}

// GetUserNameByName returns the single user whose name matches the given
// string exactly, or an error wrapping api.ErrUserNotFound when absent. It
// replaces getUserNameByName(String name).
func (s *UserServiceImp) GetUserNameByName(ctx context.Context, name string) (*model.User, error) {
	user, err := s.store.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return user, nil
}

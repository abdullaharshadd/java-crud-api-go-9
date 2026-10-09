package api

import (
	"errors"
	"fmt"
)

// ErrUserNotFound is the sentinel error signaling that a user could not be
// found. It replaces the Java checked exception UserNotFoundException and
// keeps the original message text from the source service layer
// ("User are not available").
//
// Store and service layers return this value (wrapped via fmt.Errorf with
// %w when additional context is needed) and HTTP handlers detect it with
// errors.Is to map it to a 404 response, replacing the Java
// @ExceptionHandler(RestResponseEntityExceptionHandling) mechanism.
var ErrUserNotFound = errors.New("User are not available")

// NewUserNotFoundError returns ErrUserNotFound wrapped with an optional
// additional detail message, mirroring the Java constructor
// UserNotFoundException(String message). If detail is empty or identical to
// the sentinel's own text, the bare sentinel is returned so that
// errors.Is(err, ErrUserNotFound) always matches.
func NewUserNotFoundError(detail string) error {
	if detail == "" || detail == ErrUserNotFound.Error() {
		return ErrUserNotFound
	}
	return fmt.Errorf("%s: %w", detail, ErrUserNotFound)
}

// NewUserNotFoundErrorWithCause returns ErrUserNotFound wrapped with an
// underlying cause, mirroring the Java constructors
// UserNotFoundException(Throwable cause) and
// UserNotFoundException(String message, Throwable cause). A nil cause
// degrades to NewUserNotFoundError.
//
// MIGRATION_NOTE: wrapping two errors in one chain (%w twice) requires
// Go 1.20+.
func NewUserNotFoundErrorWithCause(detail string, cause error) error {
	if cause == nil {
		return NewUserNotFoundError(detail)
	}
	if detail == "" {
		return fmt.Errorf("%w: %w", ErrUserNotFound, cause)
	}
	return fmt.Errorf("%s: %w: %w", detail, ErrUserNotFound, cause)
}

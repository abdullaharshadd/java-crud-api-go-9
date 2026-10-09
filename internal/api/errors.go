// Package api contains the HTTP-facing error handling of the smartContact
// application. This file replaces the Spring @ControllerAdvice class
// com.smartContact.error.RestResponseEntityExceptionHandling and the checked
// exception com.smartContact.error.UserNotFoundException.
package api

import (
	"errors"
	"fmt"
	"net/http"
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

// httpStatusError couples an arbitrary error with an HTTP status code. It is
// the Go replacement for throwing one of Spring's framework exceptions
// (e.g. HttpRequestMethodNotSupportedException) that the inherited
// ResponseEntityExceptionHandler translated into a specific status.
type httpStatusError struct {
	status int
	err    error
}

// Error implements the error interface, delegating to the wrapped error.
func (e *httpStatusError) Error() string { return e.err.Error() }

// Unwrap exposes the wrapped error so errors.Is/errors.As keep working.
func (e *httpStatusError) Unwrap() error { return e.err }

// NewHTTPStatusError wraps err with an explicit HTTP status code, letting any
// handler override the default 500 for non-"not found" errors.
func NewHTTPStatusError(status int, err error) error {
	if err == nil {
		return nil
	}
	return &httpStatusError{status: status, err: err}
}

// StatusCode maps an error returned by a handler to the HTTP status the
// error handler should respond with:
//   - ErrUserNotFound (the UserNotFoundException equivalent) -> 404
//   - an explicitly wrapped HTTPStatusError                  -> its status
//   - anything else                                          -> 500
func StatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	var se *httpStatusError
	if errors.As(err, &se) {
		return se.status
	}
	if errors.Is(err, ErrUserNotFound) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

// HandlerFunc is the signature handlers implement under this error-handling
// scheme: instead of writing error responses themselves, they return an
// error and let ErrorHandler decide how it is rendered.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ErrorHandler is the global exception handler that replaced
// RestResponseEntityExceptionHandling. Wrap every route handler with it
// (equivalent to applying @ControllerAdvice across all controllers):
//
//	mux.Handle("/api/user/{id}", api.ErrorHandler(userHandler.GetByID))
//
// On a nil error it does nothing (the handler already wrote its response).
// On ErrUserNotFound it writes a 404 with an ErrorMessage body whose message
// is the error's text, exactly as the Java advice did with
// exception.getMessage(). Any other error is written as a 500 with a generic
// ErrorMessage body, mirroring the Spring default of not leaking internal
// error details to clients.
func ErrorHandler(next HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := next(w, r)
		if err == nil {
			return
		}

		status := StatusCode(err)
		msg := err.Error()
		if status >= http.StatusInternalServerError {
			msg = http.StatusText(status)
		}
		WriteError(w, NewErrorMessage(status, msg))
	})
}

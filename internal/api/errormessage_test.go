package api

import (
	"errors"
	"fmt"
	"testing"
)

func TestNewUserNotFoundError(t *testing.T) {
	tests := []struct {
		name      string
		detail    string
		wantErr   error
		wantIdent bool // want err to be exactly the sentinel (errors.Is match)
		wantText  string
	}{
		{name: "empty detail returns sentinel", detail: "", wantErr: ErrUserNotFound, wantIdent: true, wantText: "User are not available"},
		{name: "sentinel text returns sentinel", detail: "User are not available", wantErr: ErrUserNotFound, wantIdent: true, wantText: "User are not available"},
		{name: "detail wraps sentinel", detail: "db lookup failed", wantErr: ErrUserNotFound, wantIdent: true, wantText: "db lookup failed: User are not available"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := NewUserNotFoundError(tc.detail)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("errors.Is(err, ErrUserNotFound) = false, want true (err = %v)", err)
			}
			if err.Error() != tc.wantText {
				t.Errorf("err.Error() = %q, want %q", err.Error(), tc.wantText)
			}
		})
	}
}

func TestNewUserNotFoundErrorSentinelIdentity(t *testing.T) {
	// The bare sentinel must be returned as-is (identical value) so that
	// callers can rely on == in addition to errors.Is.
	if got := NewUserNotFoundError(""); got != ErrUserNotFound {
		t.Errorf("NewUserNotFoundError(\"\") = %v, want identical sentinel %v", got, ErrUserNotFound)
	}
	if got := NewUserNotFoundError(ErrUserNotFound.Error()); got != ErrUserNotFound {
		t.Errorf("NewUserNotFoundError(sentinel text) = %v, want identical sentinel %v", got, ErrUserNotFound)
	}
}

func TestNewUserNotFoundErrorWithCause(t *testing.T) {
	cause := fmt.Errorf("connection refused")

	t.Run("nil cause degrades to NewUserNotFoundError", func(t *testing.T) {
		err := NewUserNotFoundErrorWithCause("", nil)
		if err != ErrUserNotFound {
			t.Errorf("err = %v, want identical sentinel %v", err, ErrUserNotFound)
		}
	})

	t.Run("nil cause with detail wraps sentinel", func(t *testing.T) {
		err := NewUserNotFoundErrorWithCause("db lookup failed", nil)
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("errors.Is(err, ErrUserNotFound) = false, want true (err = %v)", err)
		}
		if want := "db lookup failed: User are not available"; err.Error() != want {
			t.Errorf("err.Error() = %q, want %q", err.Error(), want)
		}
	})

	t.Run("cause only", func(t *testing.T) {
		err := NewUserNotFoundErrorWithCause("", cause)
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("errors.Is(err, ErrUserNotFound) = false, want true (err = %v)", err)
		}
		if !errors.Is(err, cause) {
			t.Errorf("errors.Is(err, cause) = false, want true (err = %v)", err)
		}
		if want := "User are not available: connection refused"; err.Error() != want {
			t.Errorf("err.Error() = %q, want %q", err.Error(), want)
		}
	})

	t.Run("detail and cause", func(t *testing.T) {
		err := NewUserNotFoundErrorWithCause("db lookup failed", cause)
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("errors.Is(err, ErrUserNotFound) = false, want true (err = %v)", err)
		}
		if !errors.Is(err, cause) {
			t.Errorf("errors.Is(err, cause) = false, want true (err = %v)", err)
		}
		if want := "db lookup failed: User are not available: connection refused"; err.Error() != want {
			t.Errorf("err.Error() = %q, want %q", err.Error(), want)
		}
	})
}
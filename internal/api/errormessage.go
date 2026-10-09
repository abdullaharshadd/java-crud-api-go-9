package api

import (
	"encoding/json"
	"net/http"
)

// ErrorMessage is the JSON body returned for error responses. It replaces
// the Lombok @Data POJO com.smartContact.model.ErrorMessage, which carried a
// Spring HttpStatus and a message string.
//
// MIGRATION_NOTE (Lombok): @Data / @NoArgsConstructor / @AllArgsConstructor
// have no Go equivalent and are unnecessary — fields are exported and set via
// struct literals or the constructors below. equals/hashCode are not
// migrated: value comparison uses ==, and the toString representation is
// effectively the JSON encoding produced by WriteError.
//
// The Status field is typed as int so it serializes as a numeric HTTP status
// code (e.g. 404), matching what Spring's ResponseEntity<HttpStatus> produced
// on the wire in the source application.
type ErrorMessage struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// NewErrorMessage is the all-args-constructor equivalent: it builds an
// ErrorMessage from an HTTP status and message. (The no-args constructor has
// no Go equivalent; the zero value of ErrorMessage is equally usable.)
func NewErrorMessage(status int, message string) ErrorMessage {
	return ErrorMessage{Status: status, Message: message}
}

// NewNotFoundErrorMessage builds the body used by the handler that replaced
// the Java @ControllerAdvice RestResponseEntityExceptionHandling#userNotFound:
// a 404 response carrying the exception's message.
func NewNotFoundErrorMessage(message string) ErrorMessage {
	return ErrorMessage{Status: http.StatusNotFound, Message: message}
}

// WriteError serializes e as a JSON error response with the corresponding
// HTTP status code. It replaces ResponseEntity.status(...).body(errorMessage).
// The Content-Type header and status line are written before the body, so a
// (theoretically impossible for this type) encoding failure cannot corrupt an
// already-committed response; net/http best-effort semantics apply.
func WriteError(w http.ResponseWriter, e ErrorMessage) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(e.Status)
	// Encoding a struct of two plain fields cannot fail; the returned error
	// is discarded deliberately (documented above).
	_ = json.NewEncoder(w).Encode(e)
}

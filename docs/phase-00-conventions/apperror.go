//go:build ignore

// Template for Phase 0 — copy to: internal/domain/apperror.go
//
// A small typed error model usecases return instead of bare errors. The delivery
// layer maps *AppError -> HTTP status via a single helper (see response.go), so we
// stop doing strings.Contains(err.Error(), "not found") all over the handlers.
//
package domain

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError carries a machine-readable code, the HTTP status to return, and a
// human message safe to surface to the client.
type AppError struct {
	Code    string // stable, machine-readable: "not_found", "unauthorized", ...
	Status  int    // HTTP status
	Message string // human-friendly, client-safe
	Err     error  // optional wrapped cause (never serialized)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Err }

// Wrap attaches a cause without changing the public code/status/message.
func (e *AppError) Wrap(cause error) *AppError {
	clone := *e
	clone.Err = cause
	return &clone
}

// Constructors. Pass a custom message or "" to use a sensible default.

func ErrNotFound(msg string) *AppError {
	return &AppError{Code: "not_found", Status: http.StatusNotFound, Message: orDefault(msg, "resource not found")}
}

func ErrUnauthorized(msg string) *AppError {
	return &AppError{Code: "unauthorized", Status: http.StatusUnauthorized, Message: orDefault(msg, "authentication required")}
}

func ErrForbidden(msg string) *AppError {
	return &AppError{Code: "forbidden", Status: http.StatusForbidden, Message: orDefault(msg, "not allowed")}
}

func ErrConflict(msg string) *AppError {
	return &AppError{Code: "conflict", Status: http.StatusConflict, Message: orDefault(msg, "already exists")}
}

func ErrValidation(msg string) *AppError {
	return &AppError{Code: "validation", Status: http.StatusBadRequest, Message: orDefault(msg, "invalid request")}
}

func ErrInternal(msg string) *AppError {
	return &AppError{Code: "internal", Status: http.StatusInternalServerError, Message: orDefault(msg, "something went wrong")}
}

// AsAppError extracts an *AppError from any error, falling back to a 500 so the
// delivery layer always has a status to return.
func AsAppError(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return ErrInternal("").Wrap(err)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

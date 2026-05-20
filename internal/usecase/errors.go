package usecase

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("conflict")
)

// ValidationError carries Laravel-shaped field-keyed validation messages.
type ValidationError struct {
	Fields map[string][]string
}

func (e *ValidationError) Error() string { return "validation failed" }

func NewValidationError(field, msg string) *ValidationError {
	return &ValidationError{Fields: map[string][]string{field: {msg}}}
}

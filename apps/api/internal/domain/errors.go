package domain

import "fmt"

var (
	ErrNotFound      = fmt.Errorf("not found")
	ErrAlreadyExists = fmt.Errorf("already exists")
	ErrInvalidInput  = fmt.Errorf("invalid input")
	ErrUnauthorized  = fmt.Errorf("unauthorized")
	ErrForbidden     = fmt.Errorf("forbidden")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field %q: %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return ErrInvalidInput
}

func NewValidationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}

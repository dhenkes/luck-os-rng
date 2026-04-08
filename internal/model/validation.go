package model

import (
	"fmt"
	"strings"
)

// FieldError represents a single validation error on a named field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors collects one or more field-level validation errors.
type ValidationErrors struct {
	errors []FieldError
}

func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{}
}

func (v *ValidationErrors) Add(field, message string) {
	v.errors = append(v.errors, FieldError{Field: field, Message: message})
}

func (v *ValidationErrors) HasErrors() bool {
	return len(v.errors) > 0
}

func (v *ValidationErrors) Fields() []FieldError {
	return v.errors
}

func (v *ValidationErrors) Error() string {
	if len(v.errors) == 0 {
		return "validation failed"
	}
	msgs := make([]string, len(v.errors))
	for i, e := range v.errors {
		msgs[i] = fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return "validation failed: " + strings.Join(msgs, "; ")
}

func (v *ValidationErrors) OrNil() error {
	if !v.HasErrors() {
		return nil
	}
	return v
}

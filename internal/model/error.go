package model

import (
	"encoding/json"
	"net/http"
)

// ErrorCode represents an AEP-inspired error code.
type ErrorCode int

const (
	ErrorCodeInvalidArgument ErrorCode = iota
	ErrorCodeNotFound
	ErrorCodeAlreadyExists
	ErrorCodeInternal
)

func (c ErrorCode) String() string {
	switch c {
	case ErrorCodeInvalidArgument:
		return "INVALID_ARGUMENT"
	case ErrorCodeNotFound:
		return "NOT_FOUND"
	case ErrorCodeAlreadyExists:
		return "ALREADY_EXISTS"
	case ErrorCodeInternal:
		return "INTERNAL"
	default:
		return "UNKNOWN"
	}
}

func (c ErrorCode) HTTPStatus() int {
	switch c {
	case ErrorCodeInvalidArgument:
		return http.StatusBadRequest
	case ErrorCodeNotFound:
		return http.StatusNotFound
	case ErrorCodeAlreadyExists:
		return http.StatusConflict
	case ErrorCodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func (c ErrorCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// AppError is the standard API error response shape.
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details any       `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewNotFound(message string) *AppError {
	return &AppError{Code: ErrorCodeNotFound, Message: message}
}

func NewAlreadyExists(message string) *AppError {
	return &AppError{Code: ErrorCodeAlreadyExists, Message: message}
}

func NewInvalidArgument(message string) *AppError {
	return &AppError{Code: ErrorCodeInvalidArgument, Message: message}
}

func NewInternal(message string) *AppError {
	return &AppError{Code: ErrorCodeInternal, Message: message}
}

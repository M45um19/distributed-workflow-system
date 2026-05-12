package apperror

import (
	"net/http"
	"runtime/debug"
)

type AppError struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	Stack      string `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(message string, statusCode int) *AppError {
	return &AppError{
		Success:    false,
		StatusCode: statusCode,
		Message:    message,
		Stack:      string(debug.Stack()),
	}
}

// BadRequest helper
func BadRequest(msg string) *AppError {
	return NewAppError(msg, http.StatusBadRequest)
}

// Unauthorized helper
func Unauthorized(msg string) *AppError {
	return NewAppError(msg, http.StatusUnauthorized)
}

func Forbidden(msg string) *AppError {
	return NewAppError(msg, http.StatusForbidden)
}

// NotFound helper
func NotFound(msg string) *AppError {
	return NewAppError(msg, http.StatusNotFound)
}

func InternalServer(msg string) *AppError {
	return NewAppError(msg, http.StatusInternalServerError)
}

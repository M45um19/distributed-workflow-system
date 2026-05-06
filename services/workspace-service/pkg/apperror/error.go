package apperror

import "net/http"

type AppError struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(message string, statusCode int) *AppError {
	return &AppError{

		Success:    false,
		StatusCode: statusCode,
		Message:    message,
	}
}

func BadRequest(msg string) *AppError {
	return NewAppError(msg, http.StatusBadRequest)
}

func Unauthorized(msg string) *AppError {
	return NewAppError(msg, http.StatusUnauthorized)
}

package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorCode represents an application error code
type ErrorCode string

const (
	// General errors
	ErrCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrCodeBadRequest   ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeConflict     ErrorCode = "CONFLICT"
	ErrCodeValidation   ErrorCode = "VALIDATION_ERROR"

	// Auth errors
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeInvalidToken       ErrorCode = "INVALID_TOKEN"
	ErrCodeExpiredToken       ErrorCode = "EXPIRED_TOKEN"

	// Resource errors
	ErrCodeUserNotFound      ErrorCode = "USER_NOT_FOUND"
	ErrCodeTenantNotFound    ErrorCode = "TENANT_NOT_FOUND"
	ErrCodeUserAlreadyExists ErrorCode = "USER_ALREADY_EXISTS"
)

// AppError represents an application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// New creates a new AppError
func New(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// Common error constructors
func Internal(message string) *AppError {
	return New(ErrCodeInternal, message, http.StatusInternalServerError)
}

func BadRequest(message string) *AppError {
	return New(ErrCodeBadRequest, message, http.StatusBadRequest)
}

func Unauthorized(message string) *AppError {
	return New(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func Forbidden(message string) *AppError {
	return New(ErrCodeForbidden, message, http.StatusForbidden)
}

func NotFound(message string) *AppError {
	return New(ErrCodeNotFound, message, http.StatusNotFound)
}

func Conflict(message string) *AppError {
	return New(ErrCodeConflict, message, http.StatusConflict)
}

func Validation(message string) *AppError {
	return New(ErrCodeValidation, message, http.StatusBadRequest)
}

// ErrorResponse represents the JSON error response
type ErrorResponse struct {
	Error AppError `json:"error"`
}

// ToJSON converts AppError to JSON
func (e *AppError) ToJSON() []byte {
	resp := ErrorResponse{Error: *e}
	data, _ := json.Marshal(resp)
	return data
}

// FromError converts a standard error to AppError
func FromError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return Internal(err.Error())
}

// NewValidationError creates a validation error (backward compatibility)
func NewValidationError(message string, err error) *AppError {
	if err != nil {
		return Validation(fmt.Sprintf("%s: %v", message, err))
	}
	return Validation(message)
}

// NewInternalError creates an internal error (backward compatibility)
func NewInternalError(message string, err error) *AppError {
	if err != nil {
		return Internal(fmt.Sprintf("%s: %v", message, err))
	}
	return Internal(message)
}

// NewNotFoundError creates a not found error (backward compatibility)
func NewNotFoundError(message string, err error) *AppError {
	if err != nil {
		return NotFound(fmt.Sprintf("%s: %v", message, err))
	}
	return NotFound(message)
}

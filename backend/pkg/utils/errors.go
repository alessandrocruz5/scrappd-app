package utils

import (
	"fmt"
	"net/http"
)

// AppError represents a custom application error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Internal   error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

// Unwrap returns the internal error for error unwrapping
func (e *AppError) Unwrap() error {
	return e.Internal
}

// Common error codes
const (
	ErrCodeBadRequest         = "BAD_REQUEST"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeConflict           = "CONFLICT"
	ErrCodeValidation         = "VALIDATION_ERROR"
	ErrCodeInternalServer     = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeInvalidImage       = "INVALID_IMAGE"
	ErrCodeImageTooLarge      = "IMAGE_TOO_LARGE"
	ErrCodeUnsupportedFormat  = "UNSUPPORTED_FORMAT"
	ErrCodeMLServiceError     = "ML_SERVICE_ERROR"
	ErrCodeStorageError       = "STORAGE_ERROR"
	ErrCodeDatabaseError      = "DATABASE_ERROR"
)

// NewAppError creates a new AppError
func NewAppError(code, message string, statusCode int, internal error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Internal:   internal,
	}
}

// Common error constructors

func ErrBadRequest(message string, internal error) *AppError {
	if message == "" {
		message = "Invalid request"
	}
	return NewAppError(ErrCodeBadRequest, message, http.StatusBadRequest, internal)
}

func ErrUnauthorized(message string, internal error) *AppError {
	if message == "" {
		message = "Unauthorized"
	}
	return NewAppError(ErrCodeUnauthorized, message, http.StatusUnauthorized, internal)
}

func ErrForbidden(message string, internal error) *AppError {
	if message == "" {
		message = "Forbidden"
	}
	return NewAppError(ErrCodeForbidden, message, http.StatusForbidden, internal)
}

func ErrNotFound(resource string) *AppError {
	message := "Resource not found"
	if resource != "" {
		message = fmt.Sprintf("%s not found", resource)
	}
	return NewAppError(ErrCodeNotFound, message, http.StatusNotFound, nil)
}

func ErrConflict(message string, internal error) *AppError {
	if message == "" {
		message = "Resource conflict"
	}
	return NewAppError(ErrCodeConflict, message, http.StatusConflict, internal)
}

func ErrValidation(message string, internal error) *AppError {
	if message == "" {
		message = "Validation failed"
	}
	return NewAppError(ErrCodeValidation, message, http.StatusBadRequest, internal)
}

func ErrInternalServer(message string, internal error) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return NewAppError(ErrCodeInternalServer, message, http.StatusInternalServerError, internal)
}

func ErrServiceUnavailable(service string, internal error) *AppError {
	message := "Service temporarily unavailable"
	if service != "" {
		message = fmt.Sprintf("%s service temporarily unavailable", service)
	}
	return NewAppError(ErrCodeServiceUnavailable, message, http.StatusServiceUnavailable, internal)
}

func ErrInvalidImage(message string, internal error) *AppError {
	if message == "" {
		message = "Invalid image file"
	}
	return NewAppError(ErrCodeInvalidImage, message, http.StatusBadRequest, internal)
}

func ErrImageTooLarge(maxSize int64) *AppError {
	message := fmt.Sprintf("Image size exceeds maximum allowed size of %d MB", maxSize/(1024*1024))
	return NewAppError(ErrCodeImageTooLarge, message, http.StatusBadRequest, nil)
}

func ErrUnsupportedFormat(format string) *AppError {
	message := "Unsupported image format"
	if format != "" {
		message = fmt.Sprintf("Unsupported image format: %s", format)
	}
	return NewAppError(ErrCodeUnsupportedFormat, message, http.StatusBadRequest, nil)
}

func ErrMLService(message string, internal error) *AppError {
	if message == "" {
		message = "ML service error"
	}
	return NewAppError(ErrCodeMLServiceError, message, http.StatusInternalServerError, internal)
}

func ErrStorage(message string, internal error) *AppError {
	if message == "" {
		message = "Storage service error"
	}
	return NewAppError(ErrCodeStorageError, message, http.StatusInternalServerError, internal)
}

func ErrDatabase(message string, internal error) *AppError {
	if message == "" {
		message = "Database error"
	}
	return NewAppError(ErrCodeDatabaseError, message, http.StatusInternalServerError, internal)
}

package utils

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		expected string
	}{
		{
			name: "error without internal error",
			appErr: &AppError{
				Code:       ErrCodeBadRequest,
				Message:    "Invalid input",
				StatusCode: http.StatusBadRequest,
			},
			expected: "Invalid input",
		},
		{
			name: "error with internal error",
			appErr: &AppError{
				Code:       ErrCodeInternalServer,
				Message:    "Processing failed",
				StatusCode: http.StatusInternalServerError,
				Internal:   errors.New("database connection failed"),
			},
			expected: "Processing failed: database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.appErr.Error())
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	internalErr := errors.New("internal error")
	appErr := &AppError{
		Code:     ErrCodeInternalServer,
		Message:  "Something went wrong",
		Internal: internalErr,
	}

	assert.Equal(t, internalErr, appErr.Unwrap())
}

func TestErrBadRequest(t *testing.T) {
	internalErr := errors.New("invalid json")
	err := ErrBadRequest("Request body is invalid", internalErr)

	assert.Equal(t, ErrCodeBadRequest, err.Code)
	assert.Equal(t, "Request body is invalid", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
	assert.Equal(t, internalErr, err.Internal)

	// Test with empty message
	err = ErrBadRequest("", nil)
	assert.Equal(t, "Invalid request", err.Message)
}

func TestErrUnauthorized(t *testing.T) {
	err := ErrUnauthorized("Invalid token", nil)

	assert.Equal(t, ErrCodeUnauthorized, err.Code)
	assert.Equal(t, "Invalid token", err.Message)
	assert.Equal(t, http.StatusUnauthorized, err.StatusCode)

	// Test with empty message
	err = ErrUnauthorized("", nil)
	assert.Equal(t, "Unauthorized", err.Message)
}

func TestErrNotFound(t *testing.T) {
	err := ErrNotFound("Image")

	assert.Equal(t, ErrCodeNotFound, err.Code)
	assert.Equal(t, "Image not found", err.Message)
	assert.Equal(t, http.StatusNotFound, err.StatusCode)
	assert.Nil(t, err.Internal)

	// Test with empty resource
	err = ErrNotFound("")
	assert.Equal(t, "Resource not found", err.Message)
}

func TestErrValidation(t *testing.T) {
	internalErr := errors.New("field required")
	err := ErrValidation("Email is required", internalErr)

	assert.Equal(t, ErrCodeValidation, err.Code)
	assert.Equal(t, "Email is required", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
	assert.Equal(t, internalErr, err.Internal)
}

func TestErrInternalServer(t *testing.T) {
	internalErr := errors.New("panic occurred")
	err := ErrInternalServer("Unexpected error", internalErr)

	assert.Equal(t, ErrCodeInternalServer, err.Code)
	assert.Equal(t, "Unexpected error", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
	assert.Equal(t, internalErr, err.Internal)
}

func TestErrServiceUnavailable(t *testing.T) {
	internalErr := errors.New("connection timeout")
	err := ErrServiceUnavailable("ML", internalErr)

	assert.Equal(t, ErrCodeServiceUnavailable, err.Code)
	assert.Equal(t, "ML service temporarily unavailable", err.Message)
	assert.Equal(t, http.StatusServiceUnavailable, err.StatusCode)
	assert.Equal(t, internalErr, err.Internal)
}

func TestErrInvalidImage(t *testing.T) {
	err := ErrInvalidImage("Corrupted image file", nil)

	assert.Equal(t, ErrCodeInvalidImage, err.Code)
	assert.Equal(t, "Corrupted image file", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
}

func TestErrImageTooLarge(t *testing.T) {
	maxSize := int64(10 * 1024 * 1024) // 10 MB
	err := ErrImageTooLarge(maxSize)

	assert.Equal(t, ErrCodeImageTooLarge, err.Code)
	assert.Equal(t, "Image size exceeds maximum allowed size of 10 MB", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
}

func TestErrUnsupportedFormat(t *testing.T) {
	err := ErrUnsupportedFormat("bmp")

	assert.Equal(t, ErrCodeUnsupportedFormat, err.Code)
	assert.Equal(t, "Unsupported image format: bmp", err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)

	// Test with empty format
	err = ErrUnsupportedFormat("")
	assert.Equal(t, "Unsupported image format", err.Message)
}

func TestErrMLService(t *testing.T) {
	internalErr := errors.New("model loading failed")
	err := ErrMLService("Background removal failed", internalErr)

	assert.Equal(t, ErrCodeMLServiceError, err.Code)
	assert.Equal(t, "Background removal failed", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
	assert.Equal(t, internalErr, err.Internal)
}

func TestErrStorage(t *testing.T) {
	internalErr := errors.New("bucket not found")
	err := ErrStorage("Failed to upload image", internalErr)

	assert.Equal(t, ErrCodeStorageError, err.Code)
	assert.Equal(t, "Failed to upload image", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
	assert.Equal(t, internalErr, err.Internal)
}

func TestErrDatabase(t *testing.T) {
	internalErr := errors.New("connection pool exhausted")
	err := ErrDatabase("Failed to save image metadata", internalErr)

	assert.Equal(t, ErrCodeDatabaseError, err.Code)
	assert.Equal(t, "Failed to save image metadata", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
	assert.Equal(t, internalErr, err.Internal)
}

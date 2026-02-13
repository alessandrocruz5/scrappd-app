package utils

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorInfo represents error information in the response
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta represents metadata for the response (pagination, etc.)
type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// RespondSuccess sends a successful JSON response
func RespondSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Data:    data,
	})
}

// RespondSuccessWithMeta sends a successful JSON response with metadata
func RespondSuccessWithMeta(c *gin.Context, statusCode int, data interface{}, meta *Meta) {
	c.JSON(statusCode, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// RespondError sends an error JSON response
func RespondError(c *gin.Context, err error) {
	var statusCode int
	var errorInfo *ErrorInfo

	// Check if it's an AppError
	if appErr, ok := err.(*AppError); ok {
		statusCode = appErr.StatusCode
		errorInfo = &ErrorInfo{
			Code:    appErr.Code,
			Message: appErr.Message,
		}

		requestID, _ := c.Get("request_id")
		if appErr.Internal != nil {
			log.Printf(
				"error request_id=%v method=%s path=%s status=%d code=%s message=%s internal=%v",
				requestID,
				c.Request.Method,
				c.Request.URL.String(),
				statusCode,
				appErr.Code,
				appErr.Message,
				appErr.Internal,
			)
		} else {
			log.Printf(
				"error request_id=%v method=%s path=%s status=%d code=%s message=%s",
				requestID,
				c.Request.Method,
				c.Request.URL.String(),
				statusCode,
				appErr.Code,
				appErr.Message,
			)
		}
	} else {
		// Default to internal server error for unknown errors
		statusCode = http.StatusInternalServerError
		errorInfo = &ErrorInfo{
			Code:    ErrCodeInternalServer,
			Message: "An unexpected error occurred",
		}

		requestID, _ := c.Get("request_id")
		log.Printf(
			"error request_id=%v method=%s path=%s status=%d code=%s err=%v",
			requestID,
			c.Request.Method,
			c.Request.URL.String(),
			statusCode,
			ErrCodeInternalServer,
			err,
		)
	}

	c.JSON(statusCode, Response{
		Success: false,
		Error:   errorInfo,
	})
}

// RespondCreated sends a 201 Created response
func RespondCreated(c *gin.Context, data interface{}) {
	RespondSuccess(c, http.StatusCreated, data)
}

// RespondNoContent sends a 204 No Content response
func RespondNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// RespondBadRequest sends a 400 Bad Request response
func RespondBadRequest(c *gin.Context, message string) {
	RespondError(c, ErrBadRequest(message, nil))
}

// RespondUnauthorized sends a 401 Unauthorized response
func RespondUnauthorized(c *gin.Context, message string) {
	RespondError(c, ErrUnauthorized(message, nil))
}

// RespondForbidden sends a 403 Forbidden response
func RespondForbidden(c *gin.Context, message string) {
	RespondError(c, ErrForbidden(message, nil))
}

// RespondNotFound sends a 404 Not Found response
func RespondNotFound(c *gin.Context, resource string) {
	RespondError(c, ErrNotFound(resource))
}

// RespondInternalError sends a 500 Internal Server Error response
func RespondInternalError(c *gin.Context, message string) {
	RespondError(c, ErrInternalServer(message, nil))
}

package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger creates a logging middleware
func Logger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)

		// Get status code
		statusCode := c.Writer.Status()

		// Build log fields
		fields := logrus.Fields{
			"status":     statusCode,
			"method":     c.Request.Method,
			"path":       path,
			"query":      query,
			"ip":         c.ClientIP(),
			"latency":    latency,
			"user_agent": c.Request.UserAgent(),
		}

		if requestID, ok := c.Get("request_id"); ok {
			fields["request_id"] = requestID
		}

		if userID, ok := c.Get("user_id"); ok {
			fields["user_id"] = userID
		}

		// Add error if exists
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// Log based on status code
		if statusCode >= 500 {
			logger.WithFields(fields).Error("Server error")
		} else if statusCode >= 400 {
			logger.WithFields(fields).Warn("Client error")
		} else {
			logger.WithFields(fields).Info("Request processed")
		}
	}
}

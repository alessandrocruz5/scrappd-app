package middleware

import (
	"fmt"

	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Recovery recovers from panics and returns a proper error response
func Recovery(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				logger.WithFields(logrus.Fields{
					"error": err,
					"path":  c.Request.URL.Path,
				}).Error("Panic recovered")

				// Return error response
				utils.RespondError(c, utils.ErrInternalServer(
					fmt.Sprintf("Internal server error: %v", err),
					nil,
				))

				c.Abort()
			}
		}()

		c.Next()
	}
}

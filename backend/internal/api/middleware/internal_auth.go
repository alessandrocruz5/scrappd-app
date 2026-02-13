package middleware

import (
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

const internalAuthHeader = "X-Internal-Secret"

// InternalAuthMiddleware validates internal task requests with a shared secret.
func InternalAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret == "" {
			utils.RespondError(c, utils.ErrUnauthorized("Internal auth not configured", nil))
			c.Abort()
			return
		}

		if c.GetHeader(internalAuthHeader) != secret {
			utils.RespondError(c, utils.ErrUnauthorized("Invalid internal auth token", nil))
			c.Abort()
			return
		}

		c.Next()
	}
}

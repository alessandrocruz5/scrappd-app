package middleware

import (
	"strings"

	"github.com/alessandrocruz5/scrappd-app/backend/pkg/auth"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(tokenManager *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.RespondError(c, utils.ErrUnauthorized("Authorization header required", nil))
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondError(c, utils.ErrUnauthorized("Invalid authorization header format", nil))
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := tokenManager.ValidateAccessToken(token)
		if err != nil {
			utils.RespondError(c, utils.ErrUnauthorized("Invalid or expired token", err))
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID.String())
		c.Set("email", claims.Email)
		c.Set("username", claims.Username)
		c.Set("subscription_tier", claims.SubscriptionTier)

		c.Next()
	}
}

// OptionalAuthMiddleware is like AuthMiddleware but doesn't fail if no token is provided
func OptionalAuthMiddleware(tokenManager *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]
		claims, err := tokenManager.ValidateAccessToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID.String())
		c.Set("email", claims.Email)
		c.Set("username", claims.Username)
		c.Set("subscription_tier", claims.SubscriptionTier)

		c.Next()
	}
}

package middleware

import (
	"github.com/alessandrocruz5/scrappd-app/backend/internal/services"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

// RateLimitHeaders adds rate limit headers when user context is available.
func RateLimitHeaders(usageService services.UsageService, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if usageService == nil {
			return
		}

		userIDRaw, ok := c.Get("user_id")
		if !ok {
			return
		}

		userID, err := uuid.Parse(userIDRaw.(string))
		if err != nil {
			return
		}

		stats, err := usageService.GetCurrentUsageStats(c.Request.Context(), userID)
		if err != nil {
			if logger != nil {
				logger.WithFields(logrus.Fields{
					"user_id": userID.String(),
					"error":   err,
				}).Warn("Failed to attach rate limit headers")
			}
			return
		}

		headers := usageService.GetRateLimitHeaders(stats)
		for key, value := range headers {
			c.Header(key, value)
		}
	}
}

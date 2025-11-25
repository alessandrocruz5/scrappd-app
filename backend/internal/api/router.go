package api

import (
	"github.com/alessandrocruz5/scrappd-app/backend/internal/api/handlers"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/api/middleware"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SetupRouter configures and returns the Gin router
func SetupRouter(mlClient services.MLClient, logger *logrus.Logger) *gin.Engine {
	// Set Gin mode based on environment
	// gin.SetMode(gin.ReleaseMode) // Set this in production

	router := gin.New()

	// Add middleware
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(mlClient)
	mlHandler := handlers.NewMLHandler(mlClient)

	// Health check endpoints
	healthGroup := router.Group("/health")
	{
		healthGroup.GET("", healthHandler.BasicHealth)
		healthGroup.GET("/deep", healthHandler.DeepHealth)
		healthGroup.GET("/ready", healthHandler.ReadinessProbe)
		healthGroup.GET("/live", healthHandler.LivenessProbe)
	}

	// API v1 endpoints
	v1 := router.Group("/api/v1")
	{
		// ML endpoints
		v1.POST("/remove-background", mlHandler.RemoveBackground)
		v1.POST("/remove-background/upload", mlHandler.RemoveBackgroundFromFile)
	}

	return router
}

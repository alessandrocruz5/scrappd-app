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
	router := gin.New()

	// Add middleware
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(mlClient)
	mlHandler := handlers.NewMLHandler(mlClient)

	// Root level health checks
	router.GET("/health", healthHandler.BasicHealth)
	router.GET("/health/deep", healthHandler.DeepHealth)
	router.GET("/health/ready", healthHandler.ReadinessProbe)
	router.GET("/health/live", healthHandler.LivenessProbe)
	router.GET("/readyz", healthHandler.ReadinessProbe)
	router.GET("/healthz", healthHandler.LivenessProbe)

	// ML endpoints - register directly
	router.POST("/api/v1/ml/process", mlHandler.RemoveBackgroundFromFile)
	router.POST("/api/v1/remove-background", mlHandler.RemoveBackground)
	router.POST("/api/v1/remove-background/upload", mlHandler.RemoveBackgroundFromFile)

	// Debug route
	router.GET("/debug/routes", func(c *gin.Context) {
		routes := router.Routes()
		routeList := make([]map[string]string, 0)
		for _, route := range routes {
			routeList = append(routeList, map[string]string{
				"method": route.Method,
				"path":   route.Path,
			})
		}
		c.JSON(200, gin.H{
			"total":  len(routes),
			"routes": routeList,
		})
	})

	return router
}

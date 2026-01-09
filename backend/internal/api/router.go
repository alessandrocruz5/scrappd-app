package api

import (
	"github.com/alessandrocruz5/scrappd-app/backend/internal/api/handlers"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/api/middleware"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/services"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SetupRouter configures and returns the Gin router
func SetupRouter(
	mlClient services.MLClient,
	authService services.AuthService,
	tokenManager *auth.TokenManager,
	logger *logrus.Logger,
) *gin.Engine {
	router := gin.New()

	// Add middleware
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(mlClient)
	mlHandler := handlers.NewMLHandler(mlClient)
	authHandler := handlers.NewAuthHandler(authService)

	// Root level health checks
	router.GET("/health", healthHandler.BasicHealth)
	router.GET("/health/deep", healthHandler.DeepHealth)
	router.GET("/health/ready", healthHandler.ReadinessProbe)
	router.GET("/health/live", healthHandler.LivenessProbe)
	router.GET("/readyz", healthHandler.ReadinessProbe)
	router.GET("/healthz", healthHandler.LivenessProbe)

	// ML endpoints - register directly
	// router.POST("/api/v1/ml/process", mlHandler.RemoveBackgroundFromFile)
	// router.POST("/api/v1/remove-background", mlHandler.RemoveBackground)
	// router.POST("/api/v1/remove-background/upload", mlHandler.RemoveBackgroundFromFile)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public auth routes
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/refresh", authHandler.RefreshToken)
			authRoutes.POST("/logout", authHandler.Logout)

			// Protected auth routes
			authRoutes.GET("/me", middleware.AuthMiddleware(tokenManager), authHandler.GetMe)
		}

		// ML endpoints - now protected
		mlRoutes := v1.Group("/ml")
		mlRoutes.Use(middleware.AuthMiddleware(tokenManager))
		{
			mlRoutes.POST("/process", mlHandler.RemoveBackgroundFromFile)
		}

		// Legacy endpoints (keep for backward compatibility, but add optional auth)
		router.POST("/api/v1/remove-background",
			middleware.OptionalAuthMiddleware(tokenManager),
			mlHandler.RemoveBackground,
		)
		router.POST("/api/v1/remove-background/upload",
			middleware.OptionalAuthMiddleware(tokenManager),
			mlHandler.RemoveBackgroundFromFile,
		)
	}

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

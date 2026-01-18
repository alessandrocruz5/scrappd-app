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
	itemsService services.ItemsService,
	projectsService services.ProjectsService,
	pagesService services.PagesService,
	pageItemsService services.PageItemsService,
	usageService services.UsageService,
	dbHealth handlers.DBHealthChecker,
	redisHealth handlers.RedisHealthChecker,
	storageHealth handlers.StorageHealthChecker,
	tokenManager *auth.TokenManager,
	logger *logrus.Logger,
) *gin.Engine {
	router := gin.New()

	// Add middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimitHeaders(usageService, logger))

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(mlClient, dbHealth, redisHealth, storageHealth)
	mlHandler := handlers.NewMLHandler(mlClient)
	authHandler := handlers.NewAuthHandler(authService)
	itemsHandler := handlers.NewItemsHandler(itemsService)
	projectsHandler := handlers.NewProjectsHandler(projectsService)
	pagesHandler := handlers.NewPagesHandler(pagesService)
	pageItemsHandler := handlers.NewPageItemsHandler(pageItemsService)

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

		itemsRoutes := v1.Group("/items")
		itemsRoutes.Use(middleware.AuthMiddleware(tokenManager))
		{
			itemsRoutes.POST("", itemsHandler.CreateItem)
			itemsRoutes.GET("", itemsHandler.ListItems)
			itemsRoutes.GET("/usage", itemsHandler.GetUsage)
			itemsRoutes.GET("/:id", itemsHandler.GetItem)
			itemsRoutes.DELETE("/:id", itemsHandler.DeleteItem)
		}

		projectsRoutes := v1.Group("/projects")
		projectsRoutes.Use(middleware.AuthMiddleware(tokenManager))
		{
			projectsRoutes.POST("", projectsHandler.CreateProject)
			projectsRoutes.GET("", projectsHandler.ListProjects)
			projectsRoutes.GET("/:id", projectsHandler.GetProject)
			projectsRoutes.PATCH("/:id", projectsHandler.UpdateProject)
			projectsRoutes.DELETE("/:id", projectsHandler.DeleteProject)
		}

		pagesRoutes := v1.Group("/pages")
		pagesRoutes.Use(middleware.AuthMiddleware(tokenManager))
		{
			pagesRoutes.POST("", pagesHandler.CreatePage)
			pagesRoutes.GET("", pagesHandler.ListPages)
			pagesRoutes.GET("/:id", pagesHandler.GetPage)
			pagesRoutes.PATCH("/:id", pagesHandler.UpdatePage)
			pagesRoutes.DELETE("/:id", pagesHandler.DeletePage)
			pagesRoutes.GET("/:id/items", pageItemsHandler.ListPageItems)
			pagesRoutes.POST("/:id/items", pageItemsHandler.CreatePageItem)
			pagesRoutes.PATCH("/:id/items/:item_id", pageItemsHandler.UpdatePageItem)
			pagesRoutes.DELETE("/:id/items/:item_id", pageItemsHandler.DeletePageItem)
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

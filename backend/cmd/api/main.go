package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/api"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/config"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/database"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/repository"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/services"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/auth"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Set log level based on environment
	if cfg.Server.Environment == "development" {
		logger.SetLevel(logrus.DebugLevel)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	logger.WithFields(logrus.Fields{
		"environment": cfg.Server.Environment,
		"port":        cfg.Server.Port,
	}).Info("Starting Scrapp'd API")

	db, err := database.NewDB(cfg.Database.DSN, logger)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	usageRepo := repository.NewUsageRepository(db.Pool)
	itemsRepo := repository.NewItemsRepository(db)
	pagesRepo := repository.NewPagesRepository(db)
	projectsRepo := repository.NewProjectsRepository(db)
	pageItemsRepo := repository.NewPageItemsRepository(db)

	tokenManager := auth.NewTokenManager(
		cfg.JWT.AccessTokenSecret,
		cfg.JWT.RefreshTokenSecret,
		cfg.JWT.VerifyTokenSecret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
		cfg.JWT.VerifyTokenExpiry,
	)

	// Initialize ML client
	mlClient := services.NewMLClient(&cfg.MLService)
	emailSender := services.NewEmailSender(cfg.Email, logger)
	authService := services.NewAuthService(userRepo, tokenManager, emailSender, cfg.App.BaseURL)
	usageService := services.NewUsageService(usageRepo)

	storage, err := services.NewR2Storage(&cfg.Storage, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize storage: %v", err)
	}

	var taskQueue services.TaskQueue
	if cfg.CloudTasks.Enabled {
		if cfg.App.InternalTaskSecret == "" {
			logger.Fatal("INTERNAL_TASK_SECRET must be set when Cloud Tasks is enabled")
		}
		taskQueue, err = services.NewCloudTasksQueue(&cfg.CloudTasks, cfg.App.InternalTaskSecret, logger)
		if err != nil {
			logger.Fatalf("Failed to initialize Cloud Tasks queue: %v", err)
		}
	}

	itemsService := services.NewItemsService(
		itemsRepo,
		usageService,
		mlClient,
		storage,
		taskQueue,
		cfg.App.BypassUsageLimits,
	)
	pagesService := services.NewPagesService(pagesRepo)
	projectsService := services.NewProjectsService(projectsRepo)
	pageItemsService := services.NewPageItemsService(pageItemsRepo)
	pageRenderService := services.NewPageRenderService(pagesRepo, pageItemsRepo, itemsRepo, storage)

	logger.Info("ML client initialized")

	// Setup router
	router := api.SetupRouter(
		mlClient,
		authService,
		itemsService,
		projectsService,
		pagesService,
		pageItemsService,
		pageRenderService,
		usageService,
		db,
		storage,
		tokenManager,
		cfg.App.InternalTaskSecret,
		logger,
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  120 * time.Second, // Allow time for large uploads
		WriteTimeout: 300 * time.Second, // Allow time for ML processing + response
		IdleTimeout:  240 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.WithField("address", server.Addr).Info("Server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}

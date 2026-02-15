package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/services"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	mlClient services.MLClient
	db       DBHealthChecker
	storage  StorageHealthChecker
}

type DBHealthChecker interface {
	Health(ctx context.Context) error
}

type StorageHealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(mlClient services.MLClient, db DBHealthChecker, storage StorageHealthChecker) *HealthHandler {
	return &HealthHandler{
		mlClient: mlClient,
		db:       db,
		storage:  storage,
	}
}

// HealthResponse represents the basic health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// DeepHealthResponse represents the deep health check response
type DeepHealthResponse struct {
	Status    string                   `json:"status"`
	Timestamp time.Time                `json:"timestamp"`
	Version   string                   `json:"version"`
	Services  map[string]ServiceHealth `json:"services"`
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
	StatusDegraded  = "degraded"
)

// BasicHealth returns basic health status of the API
// @Summary Basic health check
// @Description Returns basic health status of the API
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) BasicHealth(c *gin.Context) {
	response := HealthResponse{
		Status:    StatusHealthy,
		Timestamp: time.Now(),
		Version:   "1.0.0", // TODO: Get from build info
	}

	utils.RespondSuccess(c, http.StatusOK, response)
}

// DeepHealth performs a deep health check including external services
// @Summary Deep health check
// @Description Performs a comprehensive health check including ML service
// @Tags health
// @Produce json
// @Success 200 {object} DeepHealthResponse
// @Success 503 {object} utils.Response
// @Router /health/deep [get]
func (h *HealthHandler) DeepHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	services := make(map[string]ServiceHealth)
	overallStatus := StatusHealthy

	// Check Database
	dbHealth := h.checkDatabase(ctx)
	services["database"] = dbHealth
	if dbHealth.Status == StatusUnhealthy {
		overallStatus = StatusDegraded
	}

	// Check ML Service
	mlHealth := h.checkMLService(ctx)
	services["ml_service"] = mlHealth

	if mlHealth.Status == StatusUnhealthy {
		overallStatus = StatusDegraded
	}

	// Check Storage
	storageHealth := h.checkStorage(ctx)
	services["storage"] = storageHealth
	if storageHealth.Status == StatusUnhealthy {
		overallStatus = StatusDegraded
	}

	response := DeepHealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Services:  services,
	}

	statusCode := http.StatusOK
	if overallStatus == StatusDegraded || overallStatus == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	utils.RespondSuccess(c, statusCode, response)
}

// checkMLService checks the health of the ML service
func (h *HealthHandler) checkMLService(ctx context.Context) ServiceHealth {
	healthResp, err := h.mlClient.HealthCheck(ctx)
	if err != nil {
		return ServiceHealth{
			Status:  StatusUnhealthy,
			Message: "ML service is unreachable",
		}
	}

	if healthResp.Status != "healthy" {
		return ServiceHealth{
			Status:  StatusUnhealthy,
			Message: "ML service reported unhealthy status",
		}
	}

	return ServiceHealth{
		Status:  StatusHealthy,
		Message: "ML service is operational",
	}
}

func (h *HealthHandler) checkDatabase(ctx context.Context) ServiceHealth {
	if h.db == nil {
		return ServiceHealth{
			Status:  StatusUnhealthy,
			Message: "Database health check not configured",
		}
	}

	if err := h.db.Health(ctx); err != nil {
		return ServiceHealth{
			Status:  StatusUnhealthy,
			Message: "Database is unreachable",
		}
	}

	return ServiceHealth{
		Status:  StatusHealthy,
		Message: "Database is operational",
	}
}

func (h *HealthHandler) checkStorage(ctx context.Context) ServiceHealth {
	if h.storage == nil {
		return ServiceHealth{
			Status:  StatusUnhealthy,
			Message: "Storage health check not configured",
		}
	}

	if err := h.storage.HealthCheck(ctx); err != nil {
		return ServiceHealth{
			Status:  StatusUnhealthy,
			Message: "Storage is unreachable",
		}
	}

	return ServiceHealth{
		Status:  StatusHealthy,
		Message: "Storage is operational",
	}
}

// ReadinessProbe checks if the service is ready to accept traffic
// @Summary Readiness probe
// @Description Kubernetes readiness probe endpoint
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Success 503 {object} utils.Response
// @Router /health/ready [get]
func (h *HealthHandler) ReadinessProbe(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check critical dependencies
	if h.db == nil || h.storage == nil {
		utils.RespondError(c, utils.ErrServiceUnavailable("API", nil))
		return
	}

	if err := h.db.Health(ctx); err != nil {
		utils.RespondError(c, utils.ErrServiceUnavailable("Database", err))
		return
	}

	if err := h.storage.HealthCheck(ctx); err != nil {
		utils.RespondError(c, utils.ErrServiceUnavailable("Storage", err))
		return
	}

	// Check if ML service is accessible
	if _, err := h.mlClient.HealthCheck(ctx); err != nil {
		utils.RespondError(c, utils.ErrServiceUnavailable("ML", err))
		return
	}

	response := HealthResponse{
		Status:    StatusHealthy,
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	utils.RespondSuccess(c, http.StatusOK, response)
}

// LivenessProbe checks if the service is alive
// @Summary Liveness probe
// @Description Kubernetes liveness probe endpoint
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health/live [get]
func (h *HealthHandler) LivenessProbe(c *gin.Context) {
	response := HealthResponse{
		Status:    StatusHealthy,
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	utils.RespondSuccess(c, http.StatusOK, response)
}

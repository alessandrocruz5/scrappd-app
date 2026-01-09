package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_Success(t *testing.T) {
	tokenManager := auth.NewTokenManager("test-secret", "refresh-secret", 15*time.Minute, 7*24*time.Hour)

	router := gin.New()
	router.Use(AuthMiddleware(tokenManager))
	router.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	userID := uuid.New()
	token, _ := tokenManager.GenerateAccessToken(userID, "test@example.com", "testuser", "free")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	tokenManager := auth.NewTokenManager("test-secret", "refresh-secret", 15*time.Minute, 7*24*time.Hour)

	router := gin.New()
	router.Use(AuthMiddleware(tokenManager))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	tokenManager := auth.NewTokenManager("test-secret", "refresh-secret", 15*time.Minute, 7*24*time.Hour)

	router := gin.New()
	router.Use(AuthMiddleware(tokenManager))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	tokenManager := auth.NewTokenManager("test-secret", "refresh-secret", -1*time.Hour, 7*24*time.Hour)

	router := gin.New()
	router.Use(AuthMiddleware(tokenManager))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	userID := uuid.New()
	token, _ := tokenManager.GenerateAccessToken(userID, "test@example.com", "testuser", "free")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOptionalAuthMiddleware_WithToken(t *testing.T) {
	tokenManager := auth.NewTokenManager("test-secret", "refresh-secret", 15*time.Minute, 7*24*time.Hour)

	router := gin.New()
	router.Use(OptionalAuthMiddleware(tokenManager))
	router.GET("/public", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"authenticated": exists, "user_id": userID})
	})

	userID := uuid.New()
	token, _ := tokenManager.GenerateAccessToken(userID, "test@example.com", "testuser", "free")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/public", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOptionalAuthMiddleware_WithoutToken(t *testing.T) {
	tokenManager := auth.NewTokenManager("test-secret", "refresh-secret", 15*time.Minute, 7*24*time.Hour)

	router := gin.New()
	router.Use(OptionalAuthMiddleware(tokenManager))
	router.GET("/public", func(c *gin.Context) {
		_, exists := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"authenticated": exists})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/public", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestRespondSuccess(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		data := map[string]string{"message": "success"}
		RespondSuccess(c, http.StatusOK, data)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
	assert.Nil(t, response.Error)

	dataMap := response.Data.(map[string]interface{})
	assert.Equal(t, "success", dataMap["message"])
}

func TestRespondSuccessWithMeta(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		data := []string{"item1", "item2"}
		meta := &Meta{
			Page:       1,
			PerPage:    10,
			Total:      100,
			TotalPages: 10,
		}
		RespondSuccessWithMeta(c, http.StatusOK, data, meta)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
	assert.NotNil(t, response.Meta)
	assert.Equal(t, 1, int(response.Meta.Page))
	assert.Equal(t, 10, int(response.Meta.PerPage))
	assert.Equal(t, 100, int(response.Meta.Total))
	assert.Equal(t, 10, int(response.Meta.TotalPages))
}

func TestRespondError_WithAppError(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		err := ErrBadRequest("Invalid input", nil)
		RespondError(c, err)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Nil(t, response.Data)
	assert.NotNil(t, response.Error)
	assert.Equal(t, ErrCodeBadRequest, response.Error.Code)
	assert.Equal(t, "Invalid input", response.Error.Message)
}

func TestRespondError_WithGenericError(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		err := errors.New("generic error")
		RespondError(c, err)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, ErrCodeInternalServer, response.Error.Code)
	assert.Equal(t, "An unexpected error occurred", response.Error.Message)
}

func TestRespondCreated(t *testing.T) {
	router := setupTestRouter()

	router.POST("/test", func(c *gin.Context) {
		data := map[string]string{"id": "123"}
		RespondCreated(c, data)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
}

func TestRespondNoContent(t *testing.T) {
	router := setupTestRouter()

	router.DELETE("/test", func(c *gin.Context) {
		RespondNoContent(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestRespondBadRequest(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		RespondBadRequest(c, "Missing required field")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "Missing required field", response.Error.Message)
}

func TestRespondUnauthorized(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		RespondUnauthorized(c, "Invalid token")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "Invalid token", response.Error.Message)
}

func TestRespondNotFound(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		RespondNotFound(c, "User")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "User not found", response.Error.Message)
}

func TestRespondInternalError(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		RespondInternalError(c, "Database connection failed")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "Database connection failed", response.Error.Message)
}

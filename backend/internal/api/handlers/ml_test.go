package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRemoveBackground_Success(t *testing.T) {
	mockMLClient := new(MockMLClient)
	handler := NewMLHandler(mockMLClient)

	testImage := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}
	base64Image := base64.StdEncoding.EncodeToString(testImage)

	mockMLClient.On("RemoveBackground", mock.Anything, base64Image, "png").Return(&models.RemoveBackgroundResponse{
		ProcessedImage: "processed_base64_image",
		Metadata: models.BackgroundRemovalMeta{
			ProcessingTime: 14.5,
			Model:          "BiRefNet",
			OriginalSize:   models.Size{Width: 1920, Height: 1080},
			ProcessedSize:  models.Size{Width: 1920, Height: 1080},
		},
	}, nil)

	router := setupTestRouter()
	router.POST("/remove-background", handler.RemoveBackground)

	reqBody := RemoveBackgroundRequest{
		ImageData: base64Image,
		Format:    "png",
	}
	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/remove-background", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	data := response.Data.(map[string]interface{})
	assert.Equal(t, "processed_base64_image", data["processed_image"])

	metadata := data["metadata"].(map[string]interface{})
	assert.Equal(t, 14.5, metadata["processing_time"])
	assert.Equal(t, "BiRefNet", metadata["model"])

	mockMLClient.AssertExpectations(t)
}

func TestRemoveBackground_WithDataURI(t *testing.T) {
	mockMLClient := new(MockMLClient)
	handler := NewMLHandler(mockMLClient)

	testImage := []byte("fake image data")
	base64Image := base64.StdEncoding.EncodeToString(testImage)
	dataURI := "data:image/png;base64," + base64Image

	// Mock should receive the clean base64 (without data URI prefix)
	mockMLClient.On("RemoveBackground", mock.Anything, base64Image, "png").Return(&models.RemoveBackgroundResponse{
		ProcessedImage: "processed_base64_image",
		Metadata: models.BackgroundRemovalMeta{
			ProcessingTime: 14.5,
			Model:          "BiRefNet",
			OriginalSize:   models.Size{Width: 1920, Height: 1080},
			ProcessedSize:  models.Size{Width: 1920, Height: 1080},
		},
	}, nil)

	router := setupTestRouter()
	router.POST("/remove-background", handler.RemoveBackground)

	reqBody := RemoveBackgroundRequest{
		ImageData: dataURI,
	}
	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/remove-background", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	data := response.Data.(map[string]interface{})
	assert.Equal(t, "png", data["original_format"])

	mockMLClient.AssertExpectations(t)
}

func TestRemoveBackground_MissingImageData(t *testing.T) {
	mockMLClient := new(MockMLClient)
	handler := NewMLHandler(mockMLClient)

	router := setupTestRouter()
	router.POST("/remove-background", handler.RemoveBackground)

	reqBody := RemoveBackgroundRequest{
		Format: "png",
		// ImageData is missing
	}
	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/remove-background", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)

	mockMLClient.AssertNotCalled(t, "RemoveBackground")
}

func TestRemoveBackground_InvalidBase64(t *testing.T) {
	mockMLClient := new(MockMLClient)
	handler := NewMLHandler(mockMLClient)

	router := setupTestRouter()
	router.POST("/remove-background", handler.RemoveBackground)

	reqBody := RemoveBackgroundRequest{
		ImageData: "not-valid-base64!!!",
		Format:    "png",
	}
	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/remove-background", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, utils.ErrCodeInvalidImage, response.Error.Code)

	mockMLClient.AssertNotCalled(t, "RemoveBackground")
}

func TestRemoveBackground_ImageTooLarge(t *testing.T) {
	mockMLClient := new(MockMLClient)
	handler := NewMLHandler(mockMLClient)

	// Create data larger than MaxImageSize
	largeData := make([]byte, utils.MaxImageSize+1)
	base64Image := base64.StdEncoding.EncodeToString(largeData)

	router := setupTestRouter()
	router.POST("/remove-background", handler.RemoveBackground)

	reqBody := RemoveBackgroundRequest{
		ImageData: base64Image,
		Format:    "png",
	}
	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/remove-background", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, utils.ErrCodeImageTooLarge, response.Error.Code)

	mockMLClient.AssertNotCalled(t, "RemoveBackground")
}

func TestRemoveBackground_MLServiceError(t *testing.T) {
	mockMLClient := new(MockMLClient)
	handler := NewMLHandler(mockMLClient)

	testImage := []byte("fake image data")
	base64Image := base64.StdEncoding.EncodeToString(testImage)

	mockMLClient.On("RemoveBackground", mock.Anything, base64Image, "png").Return(
		nil,
		utils.ErrMLService("Model loading failed", nil),
	)

	router := setupTestRouter()
	router.POST("/remove-background", handler.RemoveBackground)

	reqBody := RemoveBackgroundRequest{
		ImageData: base64Image,
		Format:    "png",
	}
	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/remove-background", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, utils.ErrCodeMLServiceError, response.Error.Code)

	mockMLClient.AssertExpectations(t)
}

func TestRemoveBackground_DefaultFormat(t *testing.T) {
	mockMLClient := new(MockMLClient)
	handler := NewMLHandler(mockMLClient)

	testImage := []byte("fake image data")
	base64Image := base64.StdEncoding.EncodeToString(testImage)

	// Should default to "png" format
	mockMLClient.On("RemoveBackground", mock.Anything, mock.Anything, "png").Return(&models.RemoveBackgroundResponse{
		ProcessedImage: base64.StdEncoding.EncodeToString([]byte("processed_image")),
		Metadata: models.BackgroundRemovalMeta{
			ProcessingTime: 14.5,
			Model:          "BiRefNet",
			OriginalSize:   models.Size{Width: 100, Height: 100},
			ProcessedSize:  models.Size{Width: 100, Height: 100},
		},
	}, nil)

	router := setupTestRouter()
	router.POST("/remove-background", handler.RemoveBackground)

	reqBody := RemoveBackgroundRequest{
		ImageData: base64Image,
		// Format not specified
	}
	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/remove-background", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockMLClient.AssertExpectations(t)
}

func TestRemoveBackgroundFromFile_Success(t *testing.T) {
	mockMLClient := new(MockMLClient)
	handler := NewMLHandler(mockMLClient)

	testImage := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}

	mockMLClient.On("RemoveBackground", mock.Anything, mock.Anything, "png").Return(&models.RemoveBackgroundResponse{
		ProcessedImage: base64.StdEncoding.EncodeToString([]byte("processed_image")),
		Metadata: models.BackgroundRemovalMeta{
			ProcessingTime: 14.5,
			Model:          "BiRefNet",
			OriginalSize:   models.Size{Width: 100, Height: 100},
			ProcessedSize:  models.Size{Width: 100, Height: 100},
		},
	}, nil)

	router := setupTestRouter()
	router.POST("/remove-background/upload", handler.RemoveBackgroundFromFile)

	// Create multipart form with proper Content-Type header
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file with proper headers
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="image"; filename="test.png"`)
	h.Set("Content-Type", "image/png")

	part, err := writer.CreatePart(h)
	require.NoError(t, err)
	_, err = part.Write(testImage)
	require.NoError(t, err)

	writer.WriteField("format", "png")
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/remove-background/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	assert.NotEmpty(t, w.Body.Bytes())
	mockMLClient.AssertExpectations(t)
}

func TestRemoveBackgroundFromFile_NoFile(t *testing.T) {
	mockMLClient := new(MockMLClient)
	handler := NewMLHandler(mockMLClient)

	router := setupTestRouter()
	router.POST("/remove-background/upload", handler.RemoveBackgroundFromFile)

	// Create empty multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/remove-background/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Contains(t, response.Error.Message, "No image file")

	mockMLClient.AssertNotCalled(t, "RemoveBackground")
}

func TestRemoveBackgroundFromFile_InvalidFileSize(t *testing.T) {
	mockMLClient := new(MockMLClient)
	handler := NewMLHandler(mockMLClient)

	router := setupTestRouter()
	router.POST("/remove-background/upload", handler.RemoveBackgroundFromFile)

	// Create a large file
	largeData := make([]byte, utils.MaxImageSize+1)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file with proper headers
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="image"; filename="large.png"`)
	h.Set("Content-Type", "image/png")

	part, err := writer.CreatePart(h)
	require.NoError(t, err)
	_, err = part.Write(largeData)
	require.NoError(t, err)

	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/remove-background/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response utils.Response
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, utils.ErrCodeImageTooLarge, response.Error.Code)

	mockMLClient.AssertNotCalled(t, "RemoveBackground")
}

package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/services"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

// MLHandler handles ML-related endpoints
type MLHandler struct {
	mlClient services.MLClient
}

// NewMLHandler creates a new ML handler
func NewMLHandler(mlClient services.MLClient) *MLHandler {
	return &MLHandler{
		mlClient: mlClient,
	}
}

// RemoveBackgroundRequest represents the request body for background removal
type RemoveBackgroundRequest struct {
	ImageData string `json:"image_data" binding:"required"`
	Format    string `json:"format,omitempty"`
}

// RemoveBackgroundResponse represents the response for background removal
type RemoveBackgroundResponse struct {
	ProcessedImage string          `json:"processed_image"`
	OriginalFormat string          `json:"original_format,omitempty"`
	Metadata       RemovalMetadata `json:"metadata"`
}

// RemovalMetadata contains metadata about the removal process
type RemovalMetadata struct {
	ProcessingTime float64   `json:"processing_time"`
	Model          string    `json:"model"`
	OriginalSize   ImageSize `json:"original_size"`
	ProcessedSize  ImageSize `json:"processed_size"`
}

// ImageSize represents image dimensions
type ImageSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// RemoveBackground handles background removal from uploaded images
// @Summary Remove background from image
// @Description Removes background from the provided image using ML
// @Tags ml
// @Accept json
// @Produce json
// @Param request body RemoveBackgroundRequest true "Image data in base64"
// @Success 200 {object} utils.Response{data=RemoveBackgroundResponse}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/remove-background [post]
func (h *MLHandler) RemoveBackground(c *gin.Context) {
	var req RemoveBackgroundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid request body", err))
		return
	}

	// Validate base64 image
	if err := utils.ValidateBase64Image(req.ImageData); err != nil {
		utils.RespondError(c, err)
		return
	}

	// Extract format from data URI if present
	originalFormat := utils.GetImageFormat(req.ImageData)

	// Strip data URI prefix
	cleanBase64 := utils.StripDataURIPrefix(req.ImageData)

	// Set default output format if not specified
	outputFormat := req.Format
	if outputFormat == "" {
		outputFormat = "png"
	}

	// Call ML service
	mlResponse, err := h.mlClient.RemoveBackground(c.Request.Context(), cleanBase64, outputFormat)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	// Build response
	response := RemoveBackgroundResponse{
		ProcessedImage: mlResponse.ProcessedImage,
		OriginalFormat: originalFormat,
		Metadata: RemovalMetadata{
			ProcessingTime: mlResponse.Metadata.ProcessingTime,
			Model:          mlResponse.Metadata.Model,
			OriginalSize: ImageSize{
				Width:  mlResponse.Metadata.OriginalSize.Width,
				Height: mlResponse.Metadata.OriginalSize.Height,
			},
			ProcessedSize: ImageSize{
				Width:  mlResponse.Metadata.ProcessedSize.Width,
				Height: mlResponse.Metadata.ProcessedSize.Height,
			},
		},
	}

	utils.RespondSuccess(c, http.StatusOK, response)
}

// RemoveBackgroundFromFile handles background removal from uploaded file
// @Summary Remove background from uploaded file
// @Description Removes background from an uploaded image file
// @Tags ml
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file"
// @Param format formData string false "Output format (png, webp)" default(png)
// @Success 200 {object} utils.Response{data=RemoveBackgroundResponse}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/remove-background/upload [post]
func (h *MLHandler) RemoveBackgroundFromFile(c *gin.Context) {
	// Get uploaded file
	fileHeader, err := c.FormFile("image")
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("No image file provided", err))
		return
	}

	// Validate file
	if err := utils.ValidateImageFile(fileHeader); err != nil {
		utils.RespondError(c, err)
		return
	}

	// Open file
	file, err := fileHeader.Open()
	if err != nil {
		utils.RespondError(c, utils.ErrInternalServer("Failed to open uploaded file", err))
		return
	}
	defer file.Close()

	// Convert to base64
	base64Data, err := utils.FileToBase64(file, fileHeader.Size)
	if err != nil {
		utils.RespondError(c, utils.ErrInternalServer("Failed to process file", err))
		return
	}

	// Get output format from form
	outputFormat := c.DefaultPostForm("format", "png")

	// Call ML service
	mlResponse, err := h.mlClient.RemoveBackground(c.Request.Context(), base64Data, outputFormat)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	// ❌ DON'T DO THIS (returns JSON):
	// utils.RespondSuccess(c, http.StatusOK, mlResponse)

	// ✅ DO THIS (returns PNG binary):
	// Decode base64 response back to bytes
	imageBytes, err := base64.StdEncoding.DecodeString(mlResponse.ProcessedImage)
	if err != nil {
		utils.RespondError(c, utils.ErrInternalServer("Failed to decode processed image", err))
		return
	}

	// Return the PNG image directly as binary
	c.Data(http.StatusOK, "image/png", imageBytes)
}

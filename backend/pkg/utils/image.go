package utils

import (
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"strings"
)

const (
	// MaxImageSize is the maximum allowed image size (10 MB)
	MaxImageSize = 10 * 1024 * 1024

	// Supported image formats
	FormatJPEG = "image/jpeg"
	FormatJPG  = "image/jpg"
	FormatPNG  = "image/png"
	FormatWEBP = "image/webp"
)

var (
	// SupportedImageFormats contains all supported MIME types
	SupportedImageFormats = map[string]bool{
		FormatJPEG: true,
		FormatJPG:  true,
		FormatPNG:  true,
		FormatWEBP: true,
	}

	// ImageExtensions maps MIME types to file extensions
	ImageExtensions = map[string]string{
		FormatJPEG: "jpg",
		FormatJPG:  "jpg",
		FormatPNG:  "png",
		FormatWEBP: "webp",
	}
)

// ValidateImageFile validates an uploaded image file
func ValidateImageFile(file *multipart.FileHeader) error {
	// Check file size
	if file.Size > MaxImageSize {
		return ErrImageTooLarge(MaxImageSize)
	}

	if file.Size == 0 {
		return ErrInvalidImage("File is empty", nil)
	}

	// Check content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		return ErrInvalidImage("Content-Type header is missing", nil)
	}

	if !SupportedImageFormats[contentType] {
		return ErrUnsupportedFormat(contentType)
	}

	return nil
}

// ValidateBase64Image validates a base64 encoded image
func ValidateBase64Image(base64Data string) error {
	if base64Data == "" {
		return ErrInvalidImage("Base64 data is empty", nil)
	}

	// Check if it has data URI prefix and extract it
	base64Data = StripDataURIPrefix(base64Data)

	// Try to decode
	decoded, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return ErrInvalidImage("Invalid base64 encoding", err)
	}

	// Check decoded size
	if len(decoded) > MaxImageSize {
		return ErrImageTooLarge(MaxImageSize)
	}

	if len(decoded) == 0 {
		return ErrInvalidImage("Decoded image is empty", nil)
	}

	return nil
}

// StripDataURIPrefix removes the data URI prefix from base64 string
// Example: "data:image/png;base64,iVBORw0KG..." -> "iVBORw0KG..."
func StripDataURIPrefix(base64Data string) string {
	if idx := strings.Index(base64Data, ","); idx != -1 {
		return base64Data[idx+1:]
	}
	return base64Data
}

// GetImageFormat extracts the image format from a data URI
// Example: "data:image/png;base64,..." -> "png"
func GetImageFormat(base64Data string) string {
	if !strings.HasPrefix(base64Data, "data:") {
		return ""
	}

	// Extract MIME type
	endIdx := strings.Index(base64Data, ";")
	if endIdx == -1 {
		return ""
	}

	mimeType := base64Data[5:endIdx] // Skip "data:"

	if ext, ok := ImageExtensions[mimeType]; ok {
		return ext
	}

	return ""
}

// FileToBase64 converts a multipart file to base64 string
func FileToBase64(file multipart.File, size int64) (string, error) {
	buffer := make([]byte, size)
	n, err := file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buffer[:n]), nil
}

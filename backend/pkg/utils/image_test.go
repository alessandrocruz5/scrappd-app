package utils

import (
	"bytes"
	"encoding/base64"
	"mime/multipart"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestFileHeader(filename string, contentType string, size int64) *multipart.FileHeader {
	header := make(textproto.MIMEHeader)
	header.Set("Content-Type", contentType)

	return &multipart.FileHeader{
		Filename: filename,
		Header:   header,
		Size:     size,
	}
}

func TestValidateImageFile_Success(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		size        int64
	}{
		{"JPEG image", FormatJPEG, 1024 * 1024},
		{"PNG image", FormatPNG, 2 * 1024 * 1024},
		{"WEBP image", FormatWEBP, 5 * 1024 * 1024},
		{"Small image", FormatPNG, 1024},
		{"Max size image", FormatPNG, MaxImageSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileHeader := createTestFileHeader("test.png", tt.contentType, tt.size)
			err := ValidateImageFile(fileHeader)
			assert.NoError(t, err)
		})
	}
}

func TestValidateImageFile_TooLarge(t *testing.T) {
	fileHeader := createTestFileHeader("large.png", FormatPNG, MaxImageSize+1)
	err := ValidateImageFile(fileHeader)

	require.Error(t, err)
	appErr, ok := err.(*AppError)
	require.True(t, ok)
	assert.Equal(t, ErrCodeImageTooLarge, appErr.Code)
}

func TestValidateImageFile_EmptyFile(t *testing.T) {
	fileHeader := createTestFileHeader("empty.png", FormatPNG, 0)
	err := ValidateImageFile(fileHeader)

	require.Error(t, err)
	appErr, ok := err.(*AppError)
	require.True(t, ok)
	assert.Equal(t, ErrCodeInvalidImage, appErr.Code)
	assert.Contains(t, appErr.Message, "empty")
}

func TestValidateImageFile_MissingContentType(t *testing.T) {
	header := make(textproto.MIMEHeader)
	// Don't set Content-Type

	fileHeader := &multipart.FileHeader{
		Filename: "test.png",
		Header:   header,
		Size:     1024,
	}

	err := ValidateImageFile(fileHeader)

	require.Error(t, err)
	appErr, ok := err.(*AppError)
	require.True(t, ok)
	assert.Equal(t, ErrCodeInvalidImage, appErr.Code)
	assert.Contains(t, appErr.Message, "Content-Type")
}

func TestValidateImageFile_UnsupportedFormat(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
	}{
		{"BMP format", "image/bmp"},
		{"GIF format", "image/gif"},
		{"TIFF format", "image/tiff"},
		{"SVG format", "image/svg+xml"},
		{"Generic format", "image/x-custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileHeader := createTestFileHeader("test.img", tt.contentType, 1024)
			err := ValidateImageFile(fileHeader)

			require.Error(t, err)
			appErr, ok := err.(*AppError)
			require.True(t, ok)
			assert.Equal(t, ErrCodeUnsupportedFormat, appErr.Code)
		})
	}
}

func TestValidateBase64Image_Success(t *testing.T) {
	// Create a valid base64 encoded string
	testData := []byte("fake image data")
	base64Data := base64.StdEncoding.EncodeToString(testData)

	err := ValidateBase64Image(base64Data)
	assert.NoError(t, err)
}

func TestValidateBase64Image_WithDataURI(t *testing.T) {
	testData := []byte("fake image data")
	base64Data := base64.StdEncoding.EncodeToString(testData)
	dataURI := "data:image/png;base64," + base64Data

	err := ValidateBase64Image(dataURI)
	assert.NoError(t, err)
}

func TestValidateBase64Image_Empty(t *testing.T) {
	err := ValidateBase64Image("")

	require.Error(t, err)
	appErr, ok := err.(*AppError)
	require.True(t, ok)
	assert.Equal(t, ErrCodeInvalidImage, appErr.Code)
	assert.Contains(t, appErr.Message, "empty")
}

func TestValidateBase64Image_InvalidEncoding(t *testing.T) {
	err := ValidateBase64Image("not-valid-base64!!!")

	require.Error(t, err)
	appErr, ok := err.(*AppError)
	require.True(t, ok)
	assert.Equal(t, ErrCodeInvalidImage, appErr.Code)
	assert.Contains(t, appErr.Message, "Invalid base64")
}

func TestValidateBase64Image_TooLarge(t *testing.T) {
	// Create data larger than MaxImageSize
	largeData := make([]byte, MaxImageSize+1)
	base64Data := base64.StdEncoding.EncodeToString(largeData)

	err := ValidateBase64Image(base64Data)

	require.Error(t, err)
	appErr, ok := err.(*AppError)
	require.True(t, ok)
	assert.Equal(t, ErrCodeImageTooLarge, appErr.Code)
}

func TestValidateBase64Image_EmptyDecoded(t *testing.T) {
	// Empty decoded data
	base64Data := base64.StdEncoding.EncodeToString([]byte{})

	err := ValidateBase64Image(base64Data)

	require.Error(t, err)
	appErr, ok := err.(*AppError)
	require.True(t, ok)
	assert.Equal(t, ErrCodeInvalidImage, appErr.Code)
}

func TestStripDataURIPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with PNG data URI",
			input:    "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
			expected: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
		},
		{
			name:     "with JPEG data URI",
			input:    "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCwAB//2Q==",
			expected: "/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCwAB//2Q==",
		},
		{
			name:     "without data URI prefix",
			input:    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
			expected: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripDataURIPrefix(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetImageFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "PNG format",
			input:    "data:image/png;base64,iVBORw0KG...",
			expected: "png",
		},
		{
			name:     "JPEG format",
			input:    "data:image/jpeg;base64,/9j/4AAQ...",
			expected: "jpg",
		},
		{
			name:     "WEBP format",
			input:    "data:image/webp;base64,UklGRiQA...",
			expected: "webp",
		},
		{
			name:     "no data URI prefix",
			input:    "iVBORw0KG...",
			expected: "",
		},
		{
			name:     "unsupported format",
			input:    "data:image/bmp;base64,Qk1eAA...",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetImageFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// mockFile implements multipart.File interface for testing
type mockFile struct {
	*bytes.Reader
}

func (m *mockFile) Close() error {
	return nil
}

func TestFileToBase64(t *testing.T) {
	testData := []byte("test image data")
	file := &mockFile{Reader: bytes.NewReader(testData)}

	result, err := FileToBase64(file, int64(len(testData)))
	require.NoError(t, err)

	// Decode and verify
	decoded, err := base64.StdEncoding.DecodeString(result)
	require.NoError(t, err)
	assert.Equal(t, testData, decoded)
}

package services

// This file contains example usage of the Storage interface and R2Storage implementation.
// These examples demonstrate how to integrate storage into your application.

/*
Example 1: Initialize R2 Storage in main.go or application setup

	import (
		"github.com/alessandrocruz5/scrappd-app/backend/internal/config"
		"github.com/alessandrocruz5/scrappd-app/backend/internal/services"
		"github.com/sirupsen/logrus"
	)

	func setupStorage(cfg *config.Config, logger *logrus.Logger) (services.Storage, error) {
		storage, err := services.NewR2Storage(&cfg.Storage, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize storage: %w", err)
		}
		return storage, nil
	}

Example 2: Upload a file from HTTP handler

	func (h *Handler) uploadImage(c *gin.Context) {
		// Parse multipart form
		file, header, err := c.Request.FormFile("image")
		if err != nil {
			c.JSON(400, gin.H{"error": "failed to get file"})
			return
		}
		defer file.Close()

		// Upload to storage
		key, err := h.storage.Upload(
			c.Request.Context(),
			file,
			header.Filename,
			header.Header.Get("Content-Type"),
		)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to upload file"})
			return
		}

		// Generate a presigned URL valid for 1 hour
		url, err := h.storage.GetURL(c.Request.Context(), key, 1*time.Hour)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to generate URL"})
			return
		}

		c.JSON(200, gin.H{
			"key": key,
			"url": url,
		})
	}

Example 3: Upload processed image from ML service

	func (h *Handler) processAndStoreImage(c *gin.Context) {
		// Get original image
		file, header, err := c.Request.FormFile("image")
		if err != nil {
			c.JSON(400, gin.H{"error": "failed to get file"})
			return
		}
		defer file.Close()

		// Process with ML service
		processedImage, err := h.mlClient.ProcessImage(c.Request.Context(), file)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to process image"})
			return
		}

		// Convert base64 to bytes
		imageData, err := base64.StdEncoding.DecodeString(processedImage)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to decode image"})
			return
		}

		// Upload processed image to storage
		key, err := h.storage.Upload(
			c.Request.Context(),
			bytes.NewReader(imageData),
			"processed_"+header.Filename,
			"image/png",
		)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to upload processed image"})
			return
		}

		c.JSON(200, gin.H{"key": key})
	}

Example 4: Download and serve a file

	func (h *Handler) downloadImage(c *gin.Context) {
		key := c.Param("key")

		// Download from storage
		data, err := h.storage.Download(c.Request.Context(), key)
		if err != nil {
			c.JSON(404, gin.H{"error": "file not found"})
			return
		}

		// Serve the file
		c.Data(200, "image/png", data)
	}

Example 5: Delete a file

	func (h *Handler) deleteImage(c *gin.Context) {
		key := c.Param("key")

		// Delete from storage
		err := h.storage.Delete(c.Request.Context(), key)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to delete file"})
			return
		}

		c.JSON(200, gin.H{"message": "file deleted successfully"})
	}

Example 6: List user's uploaded files

	func (h *Handler) listUserImages(c *gin.Context) {
		userID := c.GetString("user_id") // From JWT middleware

		// List files with user prefix
		prefix := fmt.Sprintf("uploads/%s/", userID)
		keys, err := h.storage.List(c.Request.Context(), prefix)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to list files"})
			return
		}

		c.JSON(200, gin.H{
			"files": keys,
			"count": len(keys),
		})
	}

Example 7: Check if file exists before processing

	func (h *Handler) checkAndProcess(c *gin.Context) {
		key := c.Param("key")

		// Check if file exists
		exists, err := h.storage.Exists(c.Request.Context(), key)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to check file"})
			return
		}

		if !exists {
			c.JSON(404, gin.H{"error": "file not found"})
			return
		}

		// Process the file...
		c.JSON(200, gin.H{"message": "file exists and ready for processing"})
	}

Environment Variables Required (.env):

	# Cloudflare R2 Configuration
	STORAGE_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com
	STORAGE_ACCESS_KEY_ID=<your-r2-access-key-id>
	STORAGE_SECRET_ACCESS_KEY=<your-r2-secret-access-key>
	STORAGE_BUCKET_NAME=scrappd-images
	STORAGE_REGION=auto

Notes:
- The endpoint URL format for R2 is: https://<account-id>.r2.cloudflarestorage.com
- You can get the account ID from your Cloudflare dashboard
- Create access keys from R2 dashboard under "Manage R2 API Tokens"
- The region should be "auto" for R2
- Files are automatically organized by date in the format: uploads/YYYY/MM/DD/UUID.ext
*/

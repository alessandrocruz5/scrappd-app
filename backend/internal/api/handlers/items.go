package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/services"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ItemsHandler struct {
	itemsService services.ItemsService
}

func NewItemsHandler(itemsService services.ItemsService) *ItemsHandler {
	return &ItemsHandler{
		itemsService: itemsService,
	}
}

// CreateItem handles uploading, processing, and storing an item.
func (h *ItemsHandler) CreateItem(c *gin.Context) {
	userID, tier, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("No image file provided", err))
		return
	}

	format := c.PostForm("format")
	itemName := c.PostForm("item_name")
	itemCategory := c.PostForm("item_category")
	tags := parseTags(c.PostForm("tags"))

	item, err := h.itemsService.CreateItem(c.Request.Context(), userID, tier, fileHeader, format, itemName, itemCategory, tags)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondSuccess(c, http.StatusAccepted, item)
}

// ListItems returns a paginated list of items for the current user.
func (h *ItemsHandler) ListItems(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	page := parseIntQuery(c, "page", 1)
	perPage := parseIntQuery(c, "per_page", 20)

	items, total, err := h.itemsService.ListItems(c.Request.Context(), userID, page, perPage)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	meta := &utils.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages(total, perPage),
	}

	utils.RespondSuccessWithMeta(c, http.StatusOK, items, meta)
}

// GetItem returns a single item with fresh signed URLs.
func (h *ItemsHandler) GetItem(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid item ID", err))
		return
	}

	item, err := h.itemsService.GetItem(c.Request.Context(), userID, itemID)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondSuccess(c, http.StatusOK, item)
}

// DeleteItem deletes an item and its storage objects.
func (h *ItemsHandler) DeleteItem(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid item ID", err))
		return
	}

	if err := h.itemsService.DeleteItem(c.Request.Context(), userID, itemID); err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondNoContent(c)
}

// CancelItemProcessing cancels background processing for an item.
func (h *ItemsHandler) CancelItemProcessing(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid item ID", err))
		return
	}

	if err := h.itemsService.CancelProcessing(c.Request.Context(), userID, itemID); err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondNoContent(c)
}

// GetUsage returns usage stats and rate limit headers.
func (h *ItemsHandler) GetUsage(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	stats, headers, err := h.itemsService.GetUsageStats(c.Request.Context(), userID)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	for key, value := range headers {
		c.Header(key, value)
	}

	utils.RespondSuccess(c, http.StatusOK, stats)
}

func getUserContext(c *gin.Context) (uuid.UUID, models.SubscriptionTier, error) {
	userIDRaw, ok := c.Get("user_id")
	if !ok {
		return uuid.Nil, models.TierFree, utils.ErrUnauthorized("User not authenticated", nil)
	}

	userID, err := uuid.Parse(userIDRaw.(string))
	if err != nil {
		return uuid.Nil, models.TierFree, utils.ErrBadRequest("Invalid user ID", err)
	}

	tier := models.TierFree
	if tierRaw, ok := c.Get("subscription_tier"); ok {
		if tierStr, ok := tierRaw.(string); ok {
			tier = models.SubscriptionTier(tierStr)
		}
	}

	return userID, tier, nil
}

func parseTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

func parseIntQuery(c *gin.Context, key string, defaultValue int) int {
	raw := c.Query(key)
	if raw == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(raw)
	if err != nil || val <= 0 {
		return defaultValue
	}
	return val
}

func parseFloatQuery(c *gin.Context, key string, defaultValue float64) float64 {
	raw := c.Query(key)
	if raw == "" {
		return defaultValue
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil || val <= 0 {
		return defaultValue
	}
	return val
}

func totalPages(total, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	if total == 0 {
		return 0
	}
	pages := total / perPage
	if total%perPage != 0 {
		pages++
	}
	if pages == 0 {
		pages = 1
	}
	return pages
}

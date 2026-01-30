package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/services"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PageItemsHandler struct {
	pageItemsService services.PageItemsService
}

func NewPageItemsHandler(pageItemsService services.PageItemsService) *PageItemsHandler {
	return &PageItemsHandler{pageItemsService: pageItemsService}
}

type createPageItemRequest struct {
	ItemID    string           `json:"item_id"`
	PositionX float64          `json:"position_x"`
	PositionY float64          `json:"position_y"`
	Width     float64          `json:"width"`
	Height    float64          `json:"height"`
	Rotation  float64          `json:"rotation"`
	ZIndex    *int             `json:"z_index,omitempty"`
	Opacity   *float64         `json:"opacity,omitempty"`
	Filters   *json.RawMessage `json:"filters,omitempty"`
}

type updatePageItemRequest struct {
	PositionX *float64         `json:"position_x,omitempty"`
	PositionY *float64         `json:"position_y,omitempty"`
	Width     *float64         `json:"width,omitempty"`
	Height    *float64         `json:"height,omitempty"`
	Rotation  *float64         `json:"rotation,omitempty"`
	ZIndex    *int             `json:"z_index,omitempty"`
	Opacity   *float64         `json:"opacity,omitempty"`
	Filters   *json.RawMessage `json:"filters,omitempty"`
}

// ListPageItems returns all items on a page.
func (h *PageItemsHandler) ListPageItems(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	pageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid page ID", err))
		return
	}

	items, err := h.pageItemsService.ListPageItems(c.Request.Context(), userID, pageID)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondSuccess(c, http.StatusOK, items)
}

// CreatePageItem adds an item to a page.
func (h *PageItemsHandler) CreatePageItem(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	pageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid page ID", err))
		return
	}

	var req createPageItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid request body", err))
		return
	}

	itemID, err := uuid.Parse(req.ItemID)
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid item ID", err))
		return
	}

	pageItem := &models.PageItem{
		PageID:    pageID,
		ItemID:    itemID,
		PositionX: req.PositionX,
		PositionY: req.PositionY,
		Width:     req.Width,
		Height:    req.Height,
		Rotation:  req.Rotation,
		ZIndex:    0,
		Opacity:   1.0,
		Filters:   rawMessageValue(req.Filters),
	}
	if req.ZIndex != nil {
		pageItem.ZIndex = *req.ZIndex
	}
	if req.Opacity != nil {
		pageItem.Opacity = *req.Opacity
	}

	created, err := h.pageItemsService.CreatePageItem(c.Request.Context(), userID, pageItem)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondCreated(c, created)
}

// UpdatePageItem updates a page item's position and styling.
func (h *PageItemsHandler) UpdatePageItem(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	pageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid page ID", err))
		return
	}

	pageItemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid page item ID", err))
		return
	}

	var req updatePageItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid request body", err))
		return
	}

	update := &models.PageItemUpdate{
		ID:        pageItemID,
		PageID:    pageID,
		PositionX: req.PositionX,
		PositionY: req.PositionY,
		Width:     req.Width,
		Height:    req.Height,
		Rotation:  req.Rotation,
		ZIndex:    req.ZIndex,
		Opacity:   req.Opacity,
		Filters:   rawMessageValue(req.Filters),
	}

	item, err := h.pageItemsService.UpdatePageItem(c.Request.Context(), userID, update)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondSuccess(c, http.StatusOK, item)
}

// DeletePageItem removes an item from a page.
func (h *PageItemsHandler) DeletePageItem(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	pageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid page ID", err))
		return
	}

	pageItemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid page item ID", err))
		return
	}

	if err := h.pageItemsService.DeletePageItem(c.Request.Context(), userID, pageID, pageItemID); err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondNoContent(c)
}

package handlers

import (
	"net/http"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/services"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InternalItemsHandler struct {
	itemsService services.ItemsService
}

func NewInternalItemsHandler(itemsService services.ItemsService) *InternalItemsHandler {
	return &InternalItemsHandler{itemsService: itemsService}
}

type processItemRequest struct {
	ItemID       string `json:"item_id"`
	UserID       string `json:"user_id"`
	OutputFormat string `json:"output_format"`
}

// ProcessItem handles async processing of an item.
func (h *InternalItemsHandler) ProcessItem(c *gin.Context) {
	var req processItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid request body", err))
		return
	}

	itemID, err := uuid.Parse(req.ItemID)
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid item ID", err))
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid user ID", err))
		return
	}

	if err := h.itemsService.ProcessItem(c.Request.Context(), userID, itemID, req.OutputFormat); err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"status": "ok"})
}

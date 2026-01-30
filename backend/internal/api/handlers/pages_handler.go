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

type PagesHandler struct {
	pagesService services.PagesService
}

func NewPagesHandler(pagesService services.PagesService) *PagesHandler {
	return &PagesHandler{
		pagesService: pagesService,
	}
}

type createPageRequest struct {
	ProjectID         string           `json:"project_id"`
	PageNumber        int              `json:"page_number"`
	Title             *string          `json:"title,omitempty"`
	CanvasWidth       *int             `json:"canvas_width,omitempty"`
	CanvasHeight      *int             `json:"canvas_height,omitempty"`
	BackgroundColor   *string          `json:"background_color,omitempty"`
	BackgroundImageURL *string          `json:"background_image_url,omitempty"`
	BackgroundPattern *string          `json:"background_pattern,omitempty"`
	LayoutTemplate    *json.RawMessage `json:"layout_template,omitempty"`
}

type updatePageRequest struct {
	PageNumber        *int             `json:"page_number,omitempty"`
	Title             *string          `json:"title,omitempty"`
	CanvasWidth       *int             `json:"canvas_width,omitempty"`
	CanvasHeight      *int             `json:"canvas_height,omitempty"`
	BackgroundColor   *string          `json:"background_color,omitempty"`
	BackgroundImageURL *string          `json:"background_image_url,omitempty"`
	BackgroundPattern *string          `json:"background_pattern,omitempty"`
	LayoutTemplate    *json.RawMessage `json:"layout_template,omitempty"`
}

// CreatePage creates a new page for a project.
func (h *PagesHandler) CreatePage(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	var req createPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid request body", err))
		return
	}

	if req.ProjectID == "" {
		utils.RespondError(c, utils.ErrBadRequest("Project ID is required", nil))
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid project ID", err))
		return
	}

	if req.PageNumber <= 0 {
		utils.RespondError(c, utils.ErrBadRequest("Page number must be greater than 0", nil))
		return
	}

	page := &models.Page{
		ProjectID:      projectID,
		PageNumber:     req.PageNumber,
		Title:          req.Title,
		LayoutTemplate: rawMessageValue(req.LayoutTemplate),
	}

	if req.CanvasWidth != nil {
		page.CanvasWidth = *req.CanvasWidth
	}
	if req.CanvasHeight != nil {
		page.CanvasHeight = *req.CanvasHeight
	}
	if req.BackgroundColor != nil {
		page.BackgroundColor = *req.BackgroundColor
	}
	if req.BackgroundImageURL != nil {
		page.BackgroundImageURL = req.BackgroundImageURL
	}
	if req.BackgroundPattern != nil {
		page.BackgroundPattern = req.BackgroundPattern
	}

	created, err := h.pagesService.CreatePage(c.Request.Context(), userID, page)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondCreated(c, created)
}

// ListPages returns pages for a project.
func (h *PagesHandler) ListPages(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	projectIDRaw := c.Query("project_id")
	if projectIDRaw == "" {
		utils.RespondError(c, utils.ErrBadRequest("project_id is required", nil))
		return
	}

	projectID, err := uuid.Parse(projectIDRaw)
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid project ID", err))
		return
	}

	pageNumber := parseIntQuery(c, "page", 1)
	perPage := parseIntQuery(c, "per_page", 20)

	pages, total, err := h.pagesService.ListPages(c.Request.Context(), userID, projectID, pageNumber, perPage)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	meta := &utils.Meta{
		Page:       pageNumber,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages(total, perPage),
	}

	utils.RespondSuccessWithMeta(c, http.StatusOK, pages, meta)
}

// GetPage returns a single page.
func (h *PagesHandler) GetPage(c *gin.Context) {
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

	page, err := h.pagesService.GetPage(c.Request.Context(), userID, pageID)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondSuccess(c, http.StatusOK, page)
}

// UpdatePage updates a page's metadata or layout.
func (h *PagesHandler) UpdatePage(c *gin.Context) {
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

	var req updatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid request body", err))
		return
	}

	update := &models.PageUpdate{
		ID:             pageID,
		PageNumber:     req.PageNumber,
		Title:          req.Title,
		CanvasWidth:    req.CanvasWidth,
		CanvasHeight:   req.CanvasHeight,
		BackgroundColor: req.BackgroundColor,
		BackgroundImageURL: req.BackgroundImageURL,
		BackgroundPattern: req.BackgroundPattern,
		LayoutTemplate: rawMessageValue(req.LayoutTemplate),
	}

	updated, err := h.pagesService.UpdatePage(c.Request.Context(), userID, update)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondSuccess(c, http.StatusOK, updated)
}

// DeletePage deletes a page.
func (h *PagesHandler) DeletePage(c *gin.Context) {
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

	if err := h.pagesService.DeletePage(c.Request.Context(), userID, pageID); err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondNoContent(c)
}

func rawMessageValue(raw *json.RawMessage) json.RawMessage {
	if raw == nil {
		return nil
	}
	return *raw
}

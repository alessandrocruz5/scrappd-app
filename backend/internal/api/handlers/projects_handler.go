package handlers

import (
	"net/http"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/services"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectsHandler struct {
	projectsService services.ProjectsService
}

func NewProjectsHandler(projectsService services.ProjectsService) *ProjectsHandler {
	return &ProjectsHandler{projectsService: projectsService}
}

type createProjectRequest struct {
	Title         string   `json:"title"`
	Description   *string  `json:"description,omitempty"`
	CoverImageURL *string  `json:"cover_image_url,omitempty"`
	Visibility    *string  `json:"visibility,omitempty"`
	IsTemplate    *bool    `json:"is_template,omitempty"`
	TemplatePrice *float64 `json:"template_price,omitempty"`
}

type updateProjectRequest struct {
	Title         *string  `json:"title,omitempty"`
	Description   *string  `json:"description,omitempty"`
	CoverImageURL *string  `json:"cover_image_url,omitempty"`
	Visibility    *string  `json:"visibility,omitempty"`
	IsTemplate    *bool    `json:"is_template,omitempty"`
	TemplatePrice *float64 `json:"template_price,omitempty"`
}

// CreateProject creates a new project.
func (h *ProjectsHandler) CreateProject(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid request body", err))
		return
	}

	project := &models.Project{
		Title:         req.Title,
		Description:   req.Description,
		CoverImageURL: req.CoverImageURL,
		TemplatePrice: req.TemplatePrice,
	}
	if req.Visibility != nil {
		project.Visibility = *req.Visibility
	}
	if req.IsTemplate != nil {
		project.IsTemplate = *req.IsTemplate
	}

	created, err := h.projectsService.CreateProject(c.Request.Context(), userID, project)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondCreated(c, created)
}

// ListProjects returns projects for the current user.
func (h *ProjectsHandler) ListProjects(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	page := parseIntQuery(c, "page", 1)
	perPage := parseIntQuery(c, "per_page", 20)

	projects, total, err := h.projectsService.ListProjects(c.Request.Context(), userID, page, perPage)
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

	utils.RespondSuccessWithMeta(c, http.StatusOK, projects, meta)
}

// GetProject returns a single project.
func (h *ProjectsHandler) GetProject(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid project ID", err))
		return
	}

	project, err := h.projectsService.GetProject(c.Request.Context(), userID, projectID)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondSuccess(c, http.StatusOK, project)
}

// UpdateProject updates a project's metadata.
func (h *ProjectsHandler) UpdateProject(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid project ID", err))
		return
	}

	var req updateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid request body", err))
		return
	}

	update := &models.ProjectUpdate{
		ID:            projectID,
		Title:         req.Title,
		Description:   req.Description,
		CoverImageURL: req.CoverImageURL,
		Visibility:    req.Visibility,
		IsTemplate:    req.IsTemplate,
		TemplatePrice: req.TemplatePrice,
	}

	project, err := h.projectsService.UpdateProject(c.Request.Context(), userID, update)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondSuccess(c, http.StatusOK, project)
}

// DeleteProject soft deletes a project.
func (h *ProjectsHandler) DeleteProject(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid project ID", err))
		return
	}

	if err := h.projectsService.DeleteProject(c.Request.Context(), userID, projectID); err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondNoContent(c)
}

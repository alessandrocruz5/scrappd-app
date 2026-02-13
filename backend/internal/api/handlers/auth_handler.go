package handlers

import (
	"net/http"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/models"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/services"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.CreateUserRequest true "User registration details"
// @Success 201 {object} utils.Response{data=models.User}
// @Failure 400 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid request body", err))
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	// Don't expose password hash
	user.PasswordHash = ""

	utils.RespondCreated(c, user)
}

// Login handles user login
// @Summary Login
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} utils.Response{data=models.LoginResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid request body", err))
		return
	}

	response, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	// Don't expose password hash
	response.User.PasswordHash = ""

	utils.RespondSuccess(c, http.StatusOK, response)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Get a new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Refresh token"
// @Success 200 {object} utils.Response{data=models.LoginResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Refresh token is required", err))
		return
	}

	response, err := h.authService.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	// Don't expose password hash
	response.User.PasswordHash = ""

	utils.RespondSuccess(c, http.StatusOK, response)
}

// Logout handles user logout
// @Summary Logout
// @Description Revoke refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Refresh token"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Refresh token is required", err))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// GetMe returns the current authenticated user
// @Summary Get current user
// @Description Get the currently authenticated user's information
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=models.User}
// @Failure 401 {object} utils.Response
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondError(c, utils.ErrUnauthorized("User not authenticated", nil))
		return
	}

	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid user ID", err))
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), parsedUserID)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	// Don't expose password hash
	user.PasswordHash = ""

	utils.RespondSuccess(c, http.StatusOK, user)
}

// ForgotPassword starts password reset flow.
// @Summary Request password reset
// @Description Generates a password reset token for the account email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.ForgotPasswordRequest true "Account email"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid request body", err))
		return
	}

	resetToken, err := h.authService.RequestPasswordReset(c.Request.Context(), req.Email)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	data := gin.H{
		"message": "If that email exists, a reset token has been generated.",
	}
	if gin.Mode() != gin.ReleaseMode && resetToken != "" {
		// Helpful for local/dev until outbound email delivery is wired.
		data["reset_token"] = resetToken
	}

	utils.RespondSuccess(c, http.StatusOK, data)
}

// ResetPassword completes password reset using a token.
// @Summary Reset password
// @Description Resets password with a valid reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid request body", err))
		return
	}

	if err := h.authService.ResetPassword(c.Request.Context(), req.Token, req.Password); err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"message": "Password has been reset successfully",
	})
}

// ResendVerification triggers email verification resend.
// @Summary Resend verification email
// @Description Resends account verification email if account exists
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.ResendVerificationRequest true "Account email"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /api/v1/auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req models.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid request body", err))
		return
	}

	if err := h.authService.ResendVerification(c.Request.Context(), req.Email); err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"message": "If that email exists, verification email has been sent.",
	})
}

// VerifyEmail verifies a user's email with a token.
// @Summary Verify email
// @Description Verifies account email using a token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Verify token"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.ErrBadRequest("Invalid request body", err))
		return
	}

	if err := h.authService.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}

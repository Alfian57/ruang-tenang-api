package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/features/auth/application"
	gamificationapp "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/application"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService        *application.AuthService
	levelConfigService *gamificationapp.LevelConfigService
}

func NewAuthHandler(authService *application.AuthService, levelConfigService *gamificationapp.LevelConfigService) *AuthHandler {
	return &AuthHandler{
		authService:        authService,
		levelConfigService: levelConfigService,
	}
}

// Helper to build UserDTO with level info
func (h *AuthHandler) buildUserDTO(ctx context.Context, user *model.User) dto.UserDTO {
	profileTheme := user.ProfileTheme
	if profileTheme == "" {
		profileTheme = "default"
	}

	userDTO := dto.UserDTO{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		Avatar:       user.Avatar,
		Role:         string(user.Role),
		Exp:          user.Exp,
		GoldCoins:    user.GoldCoins,
		IsPremium:    isUserPremiumActive(user),
		PremiumUntil: formatPremiumUntil(user.PremiumExpiresAt),
		Level:        1,
		BadgeName:    "Pemula",
		BadgeIcon:    "🌱",
		ProfileTheme: profileTheme,
		CreatedAt:    user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Get level info
	currentLevel, _, _ := h.levelConfigService.GetUserLevelInfo(ctx, user.Exp)
	if currentLevel != nil {
		userDTO.Level = currentLevel.Level
		userDTO.BadgeName = currentLevel.BadgeName
		userDTO.BadgeIcon = currentLevel.BadgeIcon
	}

	return userDTO
}

func isUserPremiumActive(user *model.User) bool {
	if user == nil {
		return false
	}

	if !user.IsPremium {
		return false
	}

	if user.PremiumExpiresAt == nil {
		return true
	}

	return user.PremiumExpiresAt.After(time.Now())
}

func formatPremiumUntil(premiumExpiresAt *time.Time) string {
	if premiumExpiresAt == nil {
		return ""
	}

	return premiumExpiresAt.Format("2006-01-02T15:04:05Z")
}

// Register godoc
// @Summary Register new user
// @Description Register a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register request"
// @Success 201 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	user, err := h.authService.Register(ctx, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(h.buildUserDTO(ctx, user), "Registration successful"))
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 200 {object} dto.LoginResponse
// @Failure 401 {object} dto.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	response, err := h.authService.Login(ctx, &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse(err.Error()))
		return
	}

	// Add level info to the user in response
	response.User.Level = 1
	response.User.BadgeName = "Pemula"
	response.User.BadgeIcon = "🌱"

	currentLevel, _, _ := h.levelConfigService.GetUserLevelInfo(ctx, response.User.Exp)
	if currentLevel != nil {
		response.User.Level = currentLevel.Level
		response.User.BadgeName = currentLevel.BadgeName
		response.User.BadgeIcon = currentLevel.BadgeIcon
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response, "Login successful"))
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Get authenticated user's profile
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserDTO
// @Failure 401 {object} dto.Response
// @Router /auth/me [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	user, err := h.authService.GetProfile(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("User not found"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(h.buildUserDTO(ctx, user), ""))
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update authenticated user's profile
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateProfileRequest true "Update profile request"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	user, err := h.authService.UpdateProfile(ctx, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(h.buildUserDTO(ctx, user), "Profile updated successfully"))
}

// UpdatePassword godoc
// @Summary Update password
// @Description Update authenticated user's password
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdatePasswordRequest true "Update password request"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /auth/password [put]
func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if err := h.authService.UpdatePassword(ctx, userID, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Password updated successfully"))
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Request a password reset token to be sent to email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "Forgot password request"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if err := h.authService.ForgotPassword(ctx, &req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "If the email is registered, a reset token has been sent."))
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset password using token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Reset password request"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if err := h.authService.ResetPassword(ctx, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Password has been reset successfully."))
}

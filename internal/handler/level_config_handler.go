package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LevelConfigHandler struct {
	levelConfigService *service.LevelConfigService
}

func NewLevelConfigHandler(levelConfigService *service.LevelConfigService) *LevelConfigHandler {
	return &LevelConfigHandler{levelConfigService: levelConfigService}
}

// GetAllConfigs godoc
// @Summary Get all level configurations
// @Description Get all level configurations (public)
// @Tags Levels
// @Produce json
// @Success 200 {object} dto.Response
// @Router /level-configs [get]
func (h *LevelConfigHandler) GetAllConfigs(c *gin.Context) {
	ctx := c.Request.Context()
	configs, err := h.levelConfigService.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get level configs"))
		return
	}

	// Convert to DTOs
	configDTOs := make([]dto.LevelConfigDTO, len(configs))
	for i, config := range configs {
		configDTOs[i] = dto.LevelConfigDTO{
			ID:        config.ID,
			Level:     config.Level,
			MinExp:    config.MinExp,
			BadgeName: config.BadgeName,
			BadgeIcon: config.BadgeIcon,
			CreatedAt: config.CreatedAt,
			UpdatedAt: config.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(configDTOs, ""))
}

// AdminGetAllConfigs godoc
// @Summary Get all level configurations (Admin)
// @Description Get all level configurations for admin management
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /admin/level-configs [get]
func (h *LevelConfigHandler) AdminGetAllConfigs(c *gin.Context) {

	h.GetAllConfigs(c)
}

// saveBadgeImage saves the uploaded badge image and returns the URL path
func saveBadgeImage(c *gin.Context) (string, error) {
	file, header, err := c.Request.FormFile("badge_image")
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Check file size (max 10MB)
	if header.Size > MaxUploadSize {
		return "", fmt.Errorf("file size exceeds 10MB limit")
	}

	// Check file type
	contentType := header.Header.Get("Content-Type")
	if !AllowedImageTypes[contentType] {
		return "", fmt.Errorf("invalid file type, allowed: jpg, png, gif, webp")
	}

	// Create upload directory
	uploadPath := filepath.Join(UploadDir, "images")
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory")
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = getExtensionFromMime(contentType)
	}
	filename := fmt.Sprintf("badge_%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	filePath := filepath.Join(uploadPath, filename)

	// Save file
	if err := c.SaveUploadedFile(header, filePath); err != nil {
		return "", fmt.Errorf("failed to save file")
	}

	return fmt.Sprintf("/uploads/images/%s", filename), nil
}

// CreateConfig godoc
// @Summary Create level configuration
// @Description Create a new level configuration with badge image upload (admin only)
// @Tags Admin
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param level formData int true "Level number"
// @Param min_exp formData int true "Minimum EXP"
// @Param badge_name formData string true "Badge name"
// @Param badge_image formData file true "Badge image file"
// @Success 201 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /admin/level-configs [post]
func (h *LevelConfigHandler) CreateConfig(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse form fields
	levelStr := c.PostForm("level")
	minExpStr := c.PostForm("min_exp")
	badgeName := c.PostForm("badge_name")

	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 1 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Level harus berupa angka >= 1"))
		return
	}

	minExp, err := strconv.Atoi(minExpStr)
	if err != nil || minExp < 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Min EXP harus berupa angka >= 0"))
		return
	}

	if badgeName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Badge name harus diisi"))
		return
	}

	// Handle badge image upload
	badgeIcon, err := saveBadgeImage(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Badge image diperlukan: "+err.Error()))
		return
	}

	config := &model.LevelConfig{
		Level:     level,
		MinExp:    minExp,
		BadgeName: badgeName,
		BadgeIcon: badgeIcon,
	}

	if err := h.levelConfigService.Create(ctx, config); err != nil {
		if err == service.ErrLevelExists {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("Level sudah ada"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to create level config"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(dto.LevelConfigDTO{
		ID:        config.ID,
		Level:     config.Level,
		MinExp:    config.MinExp,
		BadgeName: config.BadgeName,
		BadgeIcon: config.BadgeIcon,
		CreatedAt: config.CreatedAt,
		UpdatedAt: config.UpdatedAt,
	}, "Level config created successfully"))
}

// UpdateConfig godoc
// @Summary Update level configuration
// @Description Update an existing level configuration with optional badge image (admin only)
// @Tags Admin
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Level config ID"
// @Param level formData int true "Level number"
// @Param min_exp formData int true "Minimum EXP"
// @Param badge_name formData string true "Badge name"
// @Param badge_image formData file false "Badge image file (optional, keeps existing if not provided)"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /admin/level-configs/{id} [put]
func (h *LevelConfigHandler) UpdateConfig(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid ID"))
		return
	}

	// Parse form fields
	levelStr := c.PostForm("level")
	minExpStr := c.PostForm("min_exp")
	badgeName := c.PostForm("badge_name")

	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 1 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Level harus berupa angka >= 1"))
		return
	}

	minExp, err := strconv.Atoi(minExpStr)
	if err != nil || minExp < 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Min EXP harus berupa angka >= 0"))
		return
	}

	if badgeName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Badge name harus diisi"))
		return
	}

	// Handle optional badge image upload
	badgeIcon := ""
	if _, _, fileErr := c.Request.FormFile("badge_image"); fileErr == nil {
		// New image uploaded
		badgeIcon, err = saveBadgeImage(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("Gagal upload badge image: "+err.Error()))
			return
		}
	}

	config := &model.LevelConfig{
		Level:     level,
		MinExp:    minExp,
		BadgeName: badgeName,
		BadgeIcon: badgeIcon, // empty string means keep existing (handled in service)
	}

	if err := h.levelConfigService.Update(ctx, uint(id), config); err != nil {
		if err == service.ErrLevelExists {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("Level sudah ada"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to update level config"))
		return
	}

	updated, _ := h.levelConfigService.GetByID(ctx, uint(id))
	c.JSON(http.StatusOK, dto.SuccessResponse(dto.LevelConfigDTO{
		ID:        updated.ID,
		Level:     updated.Level,
		MinExp:    updated.MinExp,
		BadgeName: updated.BadgeName,
		BadgeIcon: updated.BadgeIcon,
		CreatedAt: updated.CreatedAt,
		UpdatedAt: updated.UpdatedAt,
	}, "Level config updated successfully"))
}

// DeleteConfig godoc
// @Summary Delete level configuration
// @Description Delete a level configuration (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Level config ID"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /admin/level-configs/{id} [delete]
func (h *LevelConfigHandler) DeleteConfig(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid ID"))
		return
	}

	if err := h.levelConfigService.Delete(ctx, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to delete level config"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Level config deleted successfully"))
}

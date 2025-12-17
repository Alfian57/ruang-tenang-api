package handlers

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/services"
	"github.com/gin-gonic/gin"
)

type LevelConfigHandler struct {
	levelConfigService *services.LevelConfigService
}

func NewLevelConfigHandler(levelConfigService *services.LevelConfigService) *LevelConfigHandler {
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
	configs, err := h.levelConfigService.GetAll()
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

// CreateConfig godoc
// @Summary Create level configuration
// @Description Create a new level configuration (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateLevelConfigRequest true "Level config data"
// @Success 201 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /admin/level-configs [post]
func (h *LevelConfigHandler) CreateConfig(c *gin.Context) {
	var req dto.CreateLevelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	config := &models.LevelConfig{
		Level:     req.Level,
		MinExp:    req.MinExp,
		BadgeName: req.BadgeName,
		BadgeIcon: req.BadgeIcon,
	}

	if err := h.levelConfigService.Create(config); err != nil {
		if err == services.ErrLevelExists {
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
// @Description Update an existing level configuration (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Level config ID"
// @Param request body dto.UpdateLevelConfigRequest true "Level config data"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /admin/level-configs/{id} [put]
func (h *LevelConfigHandler) UpdateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid ID"))
		return
	}

	var req dto.UpdateLevelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	config := &models.LevelConfig{
		Level:     req.Level,
		MinExp:    req.MinExp,
		BadgeName: req.BadgeName,
		BadgeIcon: req.BadgeIcon,
	}

	if err := h.levelConfigService.Update(uint(id), config); err != nil {
		if err == services.ErrLevelExists {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("Level sudah ada"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to update level config"))
		return
	}

	updated, _ := h.levelConfigService.GetByID(uint(id))
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
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid ID"))
		return
	}

	if err := h.levelConfigService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to delete level config"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Level config deleted successfully"))
}

package handler

import (
	gamificationapp "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/application"
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/gin-gonic/gin"

	"github.com/Alfian57/ruang-tenang-api/internal/features/auth/application")

type UserHandler struct {
	userService        *application.UserService
	levelConfigService *gamificationapp.LevelConfigService
}

func NewUserHandler(userService *application.UserService, levelConfigService *gamificationapp.LevelConfigService) *UserHandler {
	return &UserHandler{
		userService:        userService,
		levelConfigService: levelConfigService,
	}
}

// GetLeaderboard godoc
// @Summary Get user leaderboard
// @Description Get top users based on experience points
// @Tags leaderboard
// @Accept json
// @Produce json
// @Query limit query int false "Limit number of users (default 10)"
// @Success 200 {object} []model.User
// @Router /api/v1/leaderboard [get]
func (h *UserHandler) GetLeaderboard(c *gin.Context) {
	ctx := c.Request.Context()
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	if limit > 100 {
		limit = 100
	}

	users, err := h.userService.GetLeaderboard(ctx, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}

	// Build response with level info
	userDTOs := make([]dto.UserDTO, len(users))
	for i, user := range users {
		userDTO := dto.UserDTO{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Role:      string(user.Role),
			Exp:       user.Exp,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		// Get level info
		currentLevel, _, _ := h.levelConfigService.GetUserLevelInfo(ctx, user.Exp)
		if currentLevel != nil {
			userDTO.Level = currentLevel.Level
			userDTO.BadgeName = currentLevel.BadgeName
			userDTO.BadgeIcon = currentLevel.BadgeIcon
		} else {
			userDTO.Level = 1
			userDTO.BadgeName = "Pemula"
			userDTO.BadgeIcon = "🌱"
		}

		userDTOs[i] = userDTO
	}

	c.JSON(http.StatusOK, gin.H{"data": userDTOs})
}

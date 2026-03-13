package handler

import (
	dailytaskapp "github.com/Alfian57/ruang-tenang-api/internal/features/daily_task/application"
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/gin-gonic/gin"

	"github.com/Alfian57/ruang-tenang-api/internal/features/song/application")

type SongHandler struct {
	songService      *application.SongService
	dailyTaskService dailytaskapp.DailyTaskService
}

func NewSongHandler(songService *application.SongService) *SongHandler {
	return &SongHandler{songService: songService}
}

// SetDailyTaskService sets the daily task service for progress tracking
func (h *SongHandler) SetDailyTaskService(dailyTaskService dailytaskapp.DailyTaskService) {
	h.dailyTaskService = dailyTaskService
}

// GetCategories godoc
// @Summary Get song categories
// @Description Get all song categories with song count
// @Tags Songs
// @Produce json
// @Success 200 {object} dto.Response
// @Router /song-categories [get]
func (h *SongHandler) GetCategories(c *gin.Context) {
	ctx := c.Request.Context()
	categories, err := h.songService.GetCategories(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get categories"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(categories, ""))
}

// GetSongsByCategory godoc
// @Summary Get songs by category
// @Description Get all songs in a category
// @Tags Songs
// @Produce json
// @Param slug path string true "Category Slug"
// @Success 200 {object} dto.Response
// @Router /song-categories/{slug}/songs [get]
func (h *SongHandler) GetSongsByCategory(c *gin.Context) {
	ctx := c.Request.Context()
	categorySlug := c.Param("slug")

	songs, err := h.songService.GetSongsByCategoryBySlug(ctx, categorySlug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get songs"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(songs, ""))
}

// GetSong godoc
// @Summary Get song by ID
// @Description Get song details by ID
// @Tags Songs
// @Produce json
// @Param id path int true "Song ID"
// @Success 200 {object} dto.SongDTO
// @Failure 404 {object} dto.Response
// @Router /songs/{id} [get]
func (h *SongHandler) GetSong(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid song ID"))
		return
	}

	song, err := h.songService.GetSongByID(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Song not found"))
		return
	}

	// Update daily task progress for listening to songs (if user is authenticated)
	if h.dailyTaskService != nil {
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(uint); ok && uid > 0 {
				_ = h.dailyTaskService.UpdateTaskProgress(ctx, uid, model.TaskTypeListenSongs)
			}
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(song, ""))
}

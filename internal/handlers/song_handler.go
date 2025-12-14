package handlers

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/services"
	"github.com/gin-gonic/gin"
)

type SongHandler struct {
	songService *services.SongService
}

func NewSongHandler(songService *services.SongService) *SongHandler {
	return &SongHandler{songService: songService}
}

// GetCategories godoc
// @Summary Get song categories
// @Description Get all song categories with song count
// @Tags Songs
// @Produce json
// @Success 200 {object} dto.Response
// @Router /song-categories [get]
func (h *SongHandler) GetCategories(c *gin.Context) {
	categories, err := h.songService.GetCategories()
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
// @Param id path int true "Category ID"
// @Success 200 {object} dto.Response
// @Router /song-categories/{id}/songs [get]
func (h *SongHandler) GetSongsByCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid category ID"))
		return
	}

	songs, err := h.songService.GetSongsByCategory(uint(categoryID))
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
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid song ID"))
		return
	}

	song, err := h.songService.GetSongByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Song not found"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(song, ""))
}

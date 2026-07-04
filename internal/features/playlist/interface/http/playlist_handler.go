package handler

import (
	"github.com/Alfian57/ruang-tenang-api/internal/shared/serviceerror"
	"errors"
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/gin-gonic/gin"

	"github.com/Alfian57/ruang-tenang-api/internal/features/playlist/application")

// PlaylistHandler handles playlist-related HTTP requests
type PlaylistHandler struct {
	playlistService *application.PlaylistService
}

// NewPlaylistHandler creates a new playlist handler
func NewPlaylistHandler(playlistService *application.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{playlistService: playlistService}
}

// CreatePlaylist creates a new playlist
// @Summary Create a new playlist
// @Description Create a new playlist for the authenticated user
// @Tags Playlists
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreatePlaylistRequest true "Playlist data"
// @Success 201 {object} dto.Response{data=dto.PlaylistDTO}
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Router /playlists [post]
func (h *PlaylistHandler) CreatePlaylist(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")

	var req dto.CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	playlist, err := h.playlistService.CreatePlaylist(ctx, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   "Failed to create playlist: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.Response{
		Success: true,
		Message: "Playlist created successfully",
		Data:    playlist,
	})
}

// GetMyPlaylists gets all playlists for the authenticated user
// @Summary Get user's playlists
// @Description Get all playlists belonging to the authenticated user
// @Tags Playlists
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]dto.PlaylistListDTO}
// @Failure 401 {object} dto.Response
// @Router /playlists [get]
func (h *PlaylistHandler) GetMyPlaylists(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")

	playlists, err := h.playlistService.GetUserPlaylists(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   "Failed to get playlists: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    playlists,
	})
}

// GetPlaylist gets a playlist by UUID
// @Summary Get playlist details
// @Tags Playlists
// @Security BearerAuth
// @Param uuid path string true "Playlist UUID"
// @Router /playlists/{uuid} [get]
func (h *PlaylistHandler) GetPlaylist(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")
	playlistUUID := c.Param("uuid")

	playlist, err := h.playlistService.GetPlaylistByUUID(ctx, playlistUUID, userID)
	if err != nil {
		if errors.Is(err, serviceerror.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{Success: false, Error: "Playlist not found"})
			return
		}
		if errors.Is(err, serviceerror.ErrForbidden) {
			c.JSON(http.StatusForbidden, dto.Response{Success: false, Error: "You don't have access to this playlist"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: "Failed to get playlist: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: playlist})
}

// UpdatePlaylist updates a playlist
// @Summary Update a playlist
// @Tags Playlists
// @Security BearerAuth
// @Param uuid path string true "Playlist UUID"
// @Router /playlists/{uuid} [put]
func (h *PlaylistHandler) UpdatePlaylist(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")
	playlistUUID := c.Param("uuid")

	var req dto.UpdatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "Invalid request body: " + err.Error()})
		return
	}

	playlist, err := h.playlistService.UpdatePlaylistByUUID(ctx, playlistUUID, userID, &req)
	if err != nil {
		if errors.Is(err, serviceerror.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{Success: false, Error: "Playlist not found"})
			return
		}
		if errors.Is(err, serviceerror.ErrForbidden) {
			c.JSON(http.StatusForbidden, dto.Response{Success: false, Error: "You don't have permission to update this playlist"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: "Failed to update playlist: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "Playlist updated successfully", Data: playlist})
}

// DeletePlaylist deletes a playlist
// @Summary Delete a playlist
// @Tags Playlists
// @Security BearerAuth
// @Param uuid path string true "Playlist UUID"
// @Router /playlists/{uuid} [delete]
func (h *PlaylistHandler) DeletePlaylist(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")
	playlistUUID := c.Param("uuid")

	err := h.playlistService.DeletePlaylistByUUID(ctx, playlistUUID, userID)
	if err != nil {
		if errors.Is(err, serviceerror.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{Success: false, Error: "Playlist not found"})
			return
		}
		if errors.Is(err, serviceerror.ErrForbidden) {
			c.JSON(http.StatusForbidden, dto.Response{Success: false, Error: "You don't have permission to delete this playlist"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: "Failed to delete playlist: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "Playlist deleted successfully"})
}

// AddSongToPlaylist adds a song to a playlist
// @Summary Add song to playlist
// @Tags Playlists
// @Security BearerAuth
// @Param uuid path string true "Playlist UUID"
// @Router /playlists/{uuid}/songs [post]
func (h *PlaylistHandler) AddSongToPlaylist(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")
	playlistUUID := c.Param("uuid")

	var req dto.AddSongToPlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "Invalid request body: " + err.Error()})
		return
	}

	item, err := h.playlistService.AddSongToPlaylistByUUID(ctx, playlistUUID, userID, req.SongID)
	if err != nil {
		if errors.Is(err, serviceerror.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{Success: false, Error: "Song not found"})
			return
		}
		if errors.Is(err, serviceerror.ErrForbidden) {
			c.JSON(http.StatusForbidden, dto.Response{Success: false, Error: "You don't have permission to modify this playlist"})
			return
		}
		if err.Error() == "song already in playlist" {
			c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "Song is already in this playlist"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: "Failed to add song to playlist: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.Response{Success: true, Message: "Song added to playlist", Data: item})
}

// AddSongsToPlaylist adds multiple songs to a playlist
// @Summary Add multiple songs to playlist
// @Tags Playlists
// @Security BearerAuth
// @Param uuid path string true "Playlist UUID"
// @Router /playlists/{uuid}/songs/batch [post]
func (h *PlaylistHandler) AddSongsToPlaylist(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")
	playlistUUID := c.Param("uuid")

	var req dto.AddSongsToPlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "Invalid request body: " + err.Error()})
		return
	}

	items, err := h.playlistService.AddSongsToPlaylistByUUID(ctx, playlistUUID, userID, req.SongIDs)
	if err != nil {
		if errors.Is(err, serviceerror.ErrForbidden) {
			c.JSON(http.StatusForbidden, dto.Response{Success: false, Error: "You don't have permission to modify this playlist"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: "Failed to add songs to playlist: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.Response{Success: true, Message: "Songs added to playlist", Data: items})
}

// RemoveSongFromPlaylist removes a song from a playlist
// @Summary Remove song from playlist
// @Tags Playlists
// @Security BearerAuth
// @Param uuid path string true "Playlist UUID"
// @Param songId path int true "Song ID"
// @Router /playlists/{uuid}/songs/{songId} [delete]
func (h *PlaylistHandler) RemoveSongFromPlaylist(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("userID")
	playlistUUID := c.Param("uuid")

	songID, err := strconv.ParseUint(c.Param("songId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "Invalid song ID"})
		return
	}

	err = h.playlistService.RemoveSongFromPlaylistByUUID(ctx, playlistUUID, userID, uint(songID))
	if err != nil {
		if errors.Is(err, serviceerror.ErrForbidden) {
			c.JSON(http.StatusForbidden, dto.Response{Success: false, Error: "You don't have permission to modify this playlist"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: "Failed to remove song from playlist: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "Song removed from playlist"})
}

// RemoveItemFromPlaylist removes an item from a playlist
// @Summary Remove item from playlist
// @Tags Playlists
// @Security BearerAuth
// @Param uuid path string true "Playlist UUID"
// @Param itemId path int true "Item ID"
// @Router /playlists/{uuid}/items/{itemId} [delete]
func (h *PlaylistHandler) RemoveItemFromPlaylist(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("userID")
	playlistUUID := c.Param("uuid")

	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "Invalid item ID"})
		return
	}

	err = h.playlistService.RemoveItemFromPlaylistByUUID(ctx, playlistUUID, userID, uint(itemID))
	if err != nil {
		if errors.Is(err, serviceerror.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{Success: false, Error: "Item not found"})
			return
		}
		if errors.Is(err, serviceerror.ErrForbidden) {
			c.JSON(http.StatusForbidden, dto.Response{Success: false, Error: "You don't have permission to modify this playlist"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: "Failed to remove item from playlist: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "Item removed from playlist"})
}

// ReorderPlaylistItems reorders items in a playlist
// @Summary Reorder playlist items
// @Tags Playlists
// @Security BearerAuth
// @Param uuid path string true "Playlist UUID"
// @Router /playlists/{uuid}/reorder [put]
func (h *PlaylistHandler) ReorderPlaylistItems(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")
	playlistUUID := c.Param("uuid")

	var req dto.ReorderPlaylistItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Success: false, Error: "Invalid request body: " + err.Error()})
		return
	}

	err := h.playlistService.ReorderPlaylistItemsByUUID(ctx, playlistUUID, userID, req.ItemIDs)
	if err != nil {
		if errors.Is(err, serviceerror.ErrForbidden) {
			c.JSON(http.StatusForbidden, dto.Response{Success: false, Error: "You don't have permission to modify this playlist"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Success: false, Error: "Failed to reorder playlist: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "Playlist reordered successfully"})
}

// GetPublicPlaylists gets public playlists
// @Summary Get public playlists
// @Description Get all public playlists (paginated)
// @Tags Playlists
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.Response{data=[]dto.PlaylistDTO}
// @Router /playlists/public [get]
func (h *PlaylistHandler) GetPublicPlaylists(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	playlists, total, err := h.playlistService.GetPublicPlaylists(ctx, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Error:   "Failed to get public playlists: " + err.Error(),
		})
		return
	}

	totalPages := (int(total) + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"data":        playlists,
		"page":        page,
		"limit":       limit,
		"total_items": total,
		"total_pages": totalPages,
	})
}

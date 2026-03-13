package handler

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Alfian57/ruang-tenang-api/internal/features/guild/application")

type GuildHandler struct {
	guildService *application.GuildService
}

func NewGuildHandler(guildService *application.GuildService) *GuildHandler {
	return &GuildHandler{
		guildService: guildService,
	}
}

// CreateGuild godoc
// @Summary Create a new guild
// @Description Create a new guild with the current user as leader
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateGuildRequest true "Guild data"
// @Success 201 {object} dto.GuildResponse
// @Router /api/v1/guilds [post]
func (h *GuildHandler) CreateGuild(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	var req dto.CreateGuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Data guild tidak valid"))
		return
	}

	guild, err := h.guildService.CreateGuild(ctx, userID.(uint), req)
	if err != nil {
		switch err {
		case application.ErrAlreadyInGuild:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal membuat guild"))
		}
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(guild, "Guild berhasil dibuat"))
}

// GetGuild godoc
// @Summary Get guild details
// @Description Get detailed information about a guild
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Guild UUID"
// @Success 200 {object} dto.GuildDetailResponse
// @Router /api/v1/guilds/{id} [get]
func (h *GuildHandler) GetGuild(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := c.Get("user_id")

	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID guild tidak valid"))
		return
	}

	guild, err := h.guildService.GetGuild(ctx, guildID, userID.(uint))
	if err != nil {
		if err == application.ErrGuildNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil detail guild"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(guild, "Detail guild berhasil diambil"))
}

// UpdateGuild godoc
// @Summary Update guild
// @Description Update guild information (leader/admin only)
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Guild UUID"
// @Param request body dto.UpdateGuildRequest true "Updated guild data"
// @Success 200 {object} dto.GuildResponse
// @Router /api/v1/guilds/{id} [put]
func (h *GuildHandler) UpdateGuild(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := c.Get("user_id")

	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID guild tidak valid"))
		return
	}

	var req dto.UpdateGuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Data guild tidak valid"))
		return
	}

	guild, err := h.guildService.UpdateGuild(ctx, guildID, userID.(uint), req)
	if err != nil {
		switch err {
		case application.ErrGuildNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		case application.ErrNotGuildAdmin:
			c.JSON(http.StatusForbidden, dto.ErrorResponseWithCode(dto.ErrCodeForbidden, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengupdate guild"))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(guild, "Guild berhasil diupdate"))
}

// DeleteGuild godoc
// @Summary Delete guild
// @Description Delete a guild (leader only)
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Guild UUID"
// @Success 200 {object} dto.Response
// @Router /api/v1/guilds/{id} [delete]
func (h *GuildHandler) DeleteGuild(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := c.Get("user_id")

	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID guild tidak valid"))
		return
	}

	if err := h.guildService.DeleteGuild(ctx, guildID, userID.(uint)); err != nil {
		switch err {
		case application.ErrGuildNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		case application.ErrNotGuildLeader:
			c.JSON(http.StatusForbidden, dto.ErrorResponseWithCode(dto.ErrCodeForbidden, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal menghapus guild"))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Guild berhasil dihapus"))
}

// GetPublicGuilds godoc
// @Summary Get public guilds
// @Description Browse all public guilds
// @Tags guilds
// @Accept json
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} dto.PaginatedResponse
// @Router /api/v1/guilds [get]
func (h *GuildHandler) GetPublicGuilds(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	guilds, total, err := h.guildService.GetPublicGuilds(ctx, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil daftar guild"))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(guilds, page, limit, total))
}

// GetGuildLeaderboard godoc
// @Summary Get guild leaderboard
// @Description Get top guilds ranked by total XP
// @Tags guilds
// @Accept json
// @Produce json
// @Param limit query int false "Number of guilds (default 10)"
// @Success 200 {object} []dto.GuildLeaderboardEntry
// @Router /api/v1/guilds/leaderboard [get]
func (h *GuildHandler) GetGuildLeaderboard(c *gin.Context) {
	ctx := c.Request.Context()
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	leaderboard, err := h.guildService.GetGuildLeaderboard(ctx, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil leaderboard guild"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(leaderboard, "Leaderboard guild berhasil diambil"))
}

// GetMyGuild godoc
// @Summary Get my guild
// @Description Get the current user's guild info
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.MyGuildResponse
// @Router /api/v1/guilds/my-guild [get]
func (h *GuildHandler) GetMyGuild(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	myGuild, err := h.guildService.GetMyGuild(ctx, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil info guild"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(myGuild, "Info guild berhasil diambil"))
}

// JoinGuild godoc
// @Summary Join a guild
// @Description Join a public guild
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Guild UUID"
// @Success 200 {object} dto.Response
// @Router /api/v1/guilds/{id}/join [post]
func (h *GuildHandler) JoinGuild(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := c.Get("user_id")

	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID guild tidak valid"))
		return
	}

	if err := h.guildService.JoinGuild(ctx, guildID, userID.(uint)); err != nil {
		switch err {
		case application.ErrAlreadyInGuild:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		case application.ErrGuildFull:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		case application.ErrGuildNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal bergabung ke guild"))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Berhasil bergabung ke guild"))
}

// JoinByInviteCode godoc
// @Summary Join guild by invite code
// @Description Join a guild using an invite code
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Invite code"
// @Success 200 {object} dto.GuildResponse
// @Router /api/v1/guilds/join/{code} [post]
func (h *GuildHandler) JoinByInviteCode(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := c.Get("user_id")
	code := c.Param("code")

	guild, err := h.guildService.JoinByInviteCode(ctx, code, userID.(uint))
	if err != nil {
		switch err {
		case application.ErrAlreadyInGuild:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		case application.ErrGuildFull:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		case application.ErrInvalidInviteCode:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal bergabung ke guild"))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(guild, "Berhasil bergabung ke guild"))
}

// LeaveGuild godoc
// @Summary Leave guild
// @Description Leave the current guild
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Guild UUID"
// @Success 200 {object} dto.Response
// @Router /api/v1/guilds/{id}/leave [post]
func (h *GuildHandler) LeaveGuild(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := c.Get("user_id")

	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID guild tidak valid"))
		return
	}

	if err := h.guildService.LeaveGuild(ctx, guildID, userID.(uint)); err != nil {
		switch err {
		case application.ErrCannotLeaveAsLeader:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		case application.ErrGuildNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal meninggalkan guild"))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Berhasil meninggalkan guild"))
}

// KickMember godoc
// @Summary Kick member
// @Description Kick a member from the guild (leader/admin only)
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Guild UUID"
// @Param userId path int true "User ID to kick"
// @Success 200 {object} dto.Response
// @Router /api/v1/guilds/{id}/kick/{userId} [post]
func (h *GuildHandler) KickMember(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := c.Get("user_id")

	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID guild tidak valid"))
		return
	}

	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID user tidak valid"))
		return
	}

	if err := h.guildService.KickMember(ctx, guildID, uint(targetID), userID.(uint)); err != nil {
		switch err {
		case application.ErrCannotKickLeader:
			c.JSON(http.StatusForbidden, dto.ErrorResponseWithCode(dto.ErrCodeForbidden, err.Error()))
		case application.ErrNotGuildAdmin:
			c.JSON(http.StatusForbidden, dto.ErrorResponseWithCode(dto.ErrCodeForbidden, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengeluarkan anggota"))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Anggota berhasil dikeluarkan"))
}

// PromoteMember godoc
// @Summary Promote member
// @Description Promote a member to admin (leader only)
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Guild UUID"
// @Param userId path int true "User ID to promote"
// @Success 200 {object} dto.Response
// @Router /api/v1/guilds/{id}/promote/{userId} [post]
func (h *GuildHandler) PromoteMember(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := c.Get("user_id")

	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID guild tidak valid"))
		return
	}

	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID user tidak valid"))
		return
	}

	if err := h.guildService.PromoteMember(ctx, guildID, uint(targetID), userID.(uint)); err != nil {
		switch err {
		case application.ErrNotGuildLeader:
			c.JSON(http.StatusForbidden, dto.ErrorResponseWithCode(dto.ErrCodeForbidden, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mempromosikan anggota"))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Anggota berhasil dipromosikan menjadi admin"))
}

// TransferLeadership godoc
// @Summary Transfer leadership
// @Description Transfer guild leadership to another member (leader only)
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Guild UUID"
// @Param userId path int true "New leader User ID"
// @Success 200 {object} dto.Response
// @Router /api/v1/guilds/{id}/transfer/{userId} [post]
func (h *GuildHandler) TransferLeadership(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := c.Get("user_id")

	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID guild tidak valid"))
		return
	}

	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID user tidak valid"))
		return
	}

	if err := h.guildService.TransferLeadership(ctx, guildID, uint(targetID), userID.(uint)); err != nil {
		switch err {
		case application.ErrNotGuildLeader:
			c.JSON(http.StatusForbidden, dto.ErrorResponseWithCode(dto.ErrCodeForbidden, err.Error()))
		case application.ErrNotGuildMember:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mentransfer kepemimpinan"))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Kepemimpinan berhasil ditransfer"))
}

// CreateChallenge godoc
// @Summary Create guild challenge
// @Description Create a new challenge for the guild (leader/admin only)
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Guild UUID"
// @Param request body dto.CreateGuildChallengeRequest true "Challenge data"
// @Success 201 {object} dto.GuildChallengeResponse
// @Router /api/v1/guilds/{id}/challenges [post]
func (h *GuildHandler) CreateChallenge(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := c.Get("user_id")

	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID guild tidak valid"))
		return
	}

	var req dto.CreateGuildChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Data challenge tidak valid"))
		return
	}

	challenge, err := h.guildService.CreateChallenge(ctx, guildID, userID.(uint), req)
	if err != nil {
		switch err {
		case application.ErrNotGuildAdmin:
			c.JSON(http.StatusForbidden, dto.ErrorResponseWithCode(dto.ErrCodeForbidden, err.Error()))
		case application.ErrMaxActiveChallenges:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal membuat challenge"))
		}
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(challenge, "Challenge berhasil dibuat"))
}

// GetActiveChallenges godoc
// @Summary Get active challenges
// @Description Get all active challenges for a guild
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Guild UUID"
// @Success 200 {object} []dto.GuildChallengeResponse
// @Router /api/v1/guilds/{id}/challenges [get]
func (h *GuildHandler) GetActiveChallenges(c *gin.Context) {
	ctx := c.Request.Context()

	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID guild tidak valid"))
		return
	}

	challenges, err := h.guildService.GetActiveChallenges(ctx, guildID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil challenge"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(challenges, "Challenge aktif berhasil diambil"))
}

// GetChallengeHistory godoc
// @Summary Get challenge history
// @Description Get challenge history for a guild
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Guild UUID"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} dto.PaginatedResponse
// @Router /api/v1/guilds/{id}/challenges/history [get]
func (h *GuildHandler) GetChallengeHistory(c *gin.Context) {
	ctx := c.Request.Context()

	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID guild tidak valid"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	challenges, total, err := h.guildService.GetChallengeHistory(ctx, guildID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil riwayat challenge"))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(challenges, page, limit, total))
}

// GetRecentActivities godoc
// @Summary Get guild activities
// @Description Get recent activity log for a guild
// @Tags guilds
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Guild UUID"
// @Param limit query int false "Number of activities (default 20)"
// @Success 200 {object} []dto.GuildActivityResponse
// @Router /api/v1/guilds/{id}/activities [get]
func (h *GuildHandler) GetRecentActivities(c *gin.Context) {
	ctx := c.Request.Context()

	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID guild tidak valid"))
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	activities, err := h.guildService.GetRecentActivities(ctx, guildID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil aktivitas guild"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(activities, "Aktivitas guild berhasil diambil"))
}

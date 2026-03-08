package handler

import (
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FriendQuestHandler struct {
	questService *service.FriendQuestService
}

func NewFriendQuestHandler(questService *service.FriendQuestService) *FriendQuestHandler {
	return &FriendQuestHandler{questService: questService}
}

// @Summary Create a friend quest
// @Description Invite a friend for a collaborative quest
// @Tags friend-quests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateFriendQuestRequest true "Quest data"
// @Success 201 {object} dto.FriendQuestResponse
// @Router /api/v1/friend-quests [post]
func (h *FriendQuestHandler) CreateQuest(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req dto.CreateFriendQuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Data tidak valid"))
		return
	}

	quest, err := h.questService.CreateQuest(c.Request.Context(), userID.(uint), req)
	if err != nil {
		switch err {
		case service.ErrCannotQuestYourself:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		case service.ErrQuestAlreadyExists:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		case service.ErrMaxActiveQuests:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(quest, "Friend quest berhasil dibuat!"))
}

// @Summary Get my friend quests
// @Description Get paginated list of user's friend quests
// @Tags friend-quests
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (pending, active, completed, expired, declined)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} dto.PaginatedResponse
// @Router /api/v1/friend-quests [get]
func (h *FriendQuestHandler) GetMyQuests(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var filter dto.FriendQuestFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Parameter tidak valid"))
		return
	}

	quests, total, err := h.questService.GetMyQuests(c.Request.Context(), userID.(uint), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(quests, filter.Page, filter.Limit, total))
}

// @Summary Get friend quest detail
// @Description Get details of a specific friend quest
// @Tags friend-quests
// @Produce json
// @Security BearerAuth
// @Param id path string true "Quest UUID"
// @Success 200 {object} dto.FriendQuestResponse
// @Router /api/v1/friend-quests/{id} [get]
func (h *FriendQuestHandler) GetQuest(c *gin.Context) {
	userID, _ := c.Get("user_id")

	questID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID quest tidak valid"))
		return
	}

	quest, err := h.questService.GetQuest(c.Request.Context(), userID.(uint), questID)
	if err != nil {
		switch err {
		case service.ErrFriendQuestNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		case service.ErrNotQuestParticipant:
			c.JSON(http.StatusForbidden, dto.ErrorResponseWithCode(dto.ErrCodeForbidden, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(quest, "Detail friend quest"))
}

// @Summary Accept a friend quest
// @Description Accept a pending friend quest invitation
// @Tags friend-quests
// @Produce json
// @Security BearerAuth
// @Param id path string true "Quest UUID"
// @Success 200 {object} dto.FriendQuestResponse
// @Router /api/v1/friend-quests/{id}/accept [post]
func (h *FriendQuestHandler) AcceptQuest(c *gin.Context) {
	userID, _ := c.Get("user_id")

	questID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID quest tidak valid"))
		return
	}

	quest, err := h.questService.AcceptQuest(c.Request.Context(), userID.(uint), questID)
	if err != nil {
		switch err {
		case service.ErrFriendQuestNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		case service.ErrNotQuestParticipant:
			c.JSON(http.StatusForbidden, dto.ErrorResponseWithCode(dto.ErrCodeForbidden, err.Error()))
		case service.ErrQuestNotPending:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(quest, "Friend quest diterima!"))
}

// @Summary Decline a friend quest
// @Description Decline a pending friend quest invitation
// @Tags friend-quests
// @Produce json
// @Security BearerAuth
// @Param id path string true "Quest UUID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/friend-quests/{id}/decline [post]
func (h *FriendQuestHandler) DeclineQuest(c *gin.Context) {
	userID, _ := c.Get("user_id")

	questID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID quest tidak valid"))
		return
	}

	err = h.questService.DeclineQuest(c.Request.Context(), userID.(uint), questID)
	if err != nil {
		switch err {
		case service.ErrFriendQuestNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		case service.ErrNotQuestParticipant:
			c.JSON(http.StatusForbidden, dto.ErrorResponseWithCode(dto.ErrCodeForbidden, err.Error()))
		case service.ErrQuestNotPending:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Friend quest ditolak"))
}

package handler

import (
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TimedChallengeHandler struct {
	challengeService *service.TimedChallengeService
}

func NewTimedChallengeHandler(challengeService *service.TimedChallengeService) *TimedChallengeHandler {
	return &TimedChallengeHandler{challengeService: challengeService}
}

// @Summary Get challenge templates
// @Description Get available timed challenge templates
// @Tags timed-challenges
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.TimedChallengeTemplateResponse
// @Router /api/v1/challenges/templates [get]
func (h *TimedChallengeHandler) GetTemplates(c *gin.Context) {
	templates, err := h.challengeService.GetTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(templates, "Template challenge"))
}

// @Summary Start a timed challenge
// @Description Begin a new timed challenge from a template
// @Tags timed-challenges
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.StartTimedChallengeRequest true "Template ID"
// @Success 201 {object} dto.UserTimedChallengeResponse
// @Router /api/v1/challenges/start [post]
func (h *TimedChallengeHandler) StartChallenge(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req dto.StartTimedChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Data tidak valid"))
		return
	}

	challenge, err := h.challengeService.StartChallenge(c.Request.Context(), userID.(uint), req)
	if err != nil {
		switch err {
		case service.ErrAlreadyHasActiveChallenge:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		case service.ErrTimedChallengeTemplateNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(challenge, "Challenge dimulai! ⚡"))
}

// @Summary Get active challenge
// @Description Get user's current active timed challenge
// @Tags timed-challenges
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserTimedChallengeResponse
// @Router /api/v1/challenges/active [get]
func (h *TimedChallengeHandler) GetActiveChallenge(c *gin.Context) {
	userID, _ := c.Get("user_id")

	challenge, err := h.challengeService.GetActiveChallenge(c.Request.Context(), userID.(uint))
	if err != nil {
		switch err {
		case service.ErrTimedChallengeNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(challenge, "Challenge aktif"))
}

// @Summary Complete a challenge
// @Description Mark a timed challenge as completed
// @Tags timed-challenges
// @Produce json
// @Security BearerAuth
// @Param id path string true "Challenge UUID"
// @Success 200 {object} dto.UserTimedChallengeResponse
// @Router /api/v1/challenges/{id}/complete [post]
func (h *TimedChallengeHandler) CompleteChallenge(c *gin.Context) {
	userID, _ := c.Get("user_id")

	challengeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID challenge tidak valid"))
		return
	}

	challenge, err := h.challengeService.CompleteChallenge(c.Request.Context(), userID.(uint), challengeID)
	if err != nil {
		switch err {
		case service.ErrTimedChallengeNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		case service.ErrNotChallengeOwner:
			c.JSON(http.StatusForbidden, dto.ErrorResponseWithCode(dto.ErrCodeForbidden, err.Error()))
		case service.ErrChallengeNotActive:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(challenge, "Challenge selesai! 🎉"))
}

// @Summary Get challenge history
// @Description Get user's past timed challenges
// @Tags timed-challenges
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (active, completed, expired)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} dto.PaginatedResponse
// @Router /api/v1/challenges/history [get]
func (h *TimedChallengeHandler) GetHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var filter dto.TimedChallengeFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Parameter tidak valid"))
		return
	}

	challenges, total, err := h.challengeService.GetMyHistory(c.Request.Context(), userID.(uint), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(challenges, filter.Page, filter.Limit, total))
}

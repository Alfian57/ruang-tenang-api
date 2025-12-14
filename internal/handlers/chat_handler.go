package handlers

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/Alfian57/ruang-tenang-api/internal/services"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService *services.ChatService
}

func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// GetSessions godoc
// @Summary Get chat sessions
// @Description Get user's chat sessions with optional filtering
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param filter query string false "Filter: all, bookmarked, favorites"
// @Param search query string false "Search by title"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Router /chat-sessions [get]
func (h *ChatHandler) GetSessions(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var params dto.ChatSessionQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 50 {
		params.Limit = 20
	}

	sessions, total, err := h.chatService.GetSessions(userID, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get sessions"))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(sessions, params.Page, params.Limit, total))
}

// GetSession godoc
// @Summary Get chat session by ID
// @Description Get session with all messages
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} dto.ChatSessionDTO
// @Failure 404 {object} dto.Response
// @Router /chat-sessions/{id} [get]
func (h *ChatHandler) GetSession(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid session ID"))
		return
	}

	session, err := h.chatService.GetSessionByID(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(session, ""))
}

// CreateSession godoc
// @Summary Create new chat session
// @Description Create a new AI chat session
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateChatSessionRequest true "Create session request"
// @Success 201 {object} dto.Response
// @Router /chat-sessions [post]
func (h *ChatHandler) CreateSession(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req dto.CreateChatSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	session, err := h.chatService.CreateSession(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to create session"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(gin.H{
		"id":    session.ID,
		"title": session.Title,
	}, "Session created"))
}

// SendMessage godoc
// @Summary Send message to chat
// @Description Send a message and receive AI response
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Param request body dto.SendMessageRequest true "Message content"
// @Success 200 {object} dto.Response
// @Router /chat-sessions/{id}/messages [post]
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	sessionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid session ID"))
		return
	}

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	userMsg, aiMsg, err := h.chatService.SendMessage(uint(sessionID), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"user_message": userMsg,
		"ai_message":   aiMsg,
	}, ""))
}

// ToggleBookmark godoc
// @Summary Toggle session bookmark
// @Description Toggle bookmark status for a session
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} dto.Response
// @Router /chat-sessions/{id}/bookmark [put]
func (h *ChatHandler) ToggleBookmark(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	sessionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid session ID"))
		return
	}

	if err := h.chatService.ToggleBookmark(uint(sessionID), userID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Bookmark toggled"))
}

// ToggleFavorite godoc
// @Summary Toggle session favorite
// @Description Toggle favorite status for a session
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} dto.Response
// @Router /chat-sessions/{id}/favorite [put]
func (h *ChatHandler) ToggleFavorite(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	sessionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid session ID"))
		return
	}

	if err := h.chatService.ToggleFavorite(uint(sessionID), userID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Favorite toggled"))
}

// DeleteSession godoc
// @Summary Delete chat session
// @Description Soft delete a chat session
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param id path int true "Session ID"
// @Success 200 {object} dto.Response
// @Router /chat-sessions/{id} [delete]
func (h *ChatHandler) DeleteSession(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	sessionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid session ID"))
		return
	}

	if err := h.chatService.DeleteSession(uint(sessionID), userID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Session deleted"))
}

// ToggleMessageLike godoc
// @Summary Toggle message like
// @Description Toggle like status for a message
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param id path int true "Message ID"
// @Success 200 {object} dto.Response
// @Router /chat-messages/{id}/like [put]
func (h *ChatHandler) ToggleMessageLike(c *gin.Context) {
	messageID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid message ID"))
		return
	}

	if err := h.chatService.ToggleMessageLike(uint(messageID)); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Like toggled"))
}

// ToggleMessageDislike godoc
// @Summary Toggle message dislike
// @Description Toggle dislike status for a message
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param id path int true "Message ID"
// @Success 200 {object} dto.Response
// @Router /chat-messages/{id}/dislike [put]
func (h *ChatHandler) ToggleMessageDislike(c *gin.Context) {
	messageID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid message ID"))
		return
	}

	if err := h.chatService.ToggleMessageDislike(uint(messageID)); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Dislike toggled"))
}

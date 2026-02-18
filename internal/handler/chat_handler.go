package handler

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService      *service.ChatService
	dailyTaskService service.DailyTaskService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

func (h *ChatHandler) SetDailyTaskService(dailyTaskService service.DailyTaskService) {
	h.dailyTaskService = dailyTaskService
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
	ctx := c.Request.Context()
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

	sessions, total, err := h.chatService.GetSessions(ctx, userID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get sessions"))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(sessions, params.Page, params.Limit, total))
}

// GetSession godoc
// @Summary Get chat session by UUID
// @Description Get session with all messages
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Session UUID"
// @Success 200 {object} dto.ChatSessionDTO
// @Failure 404 {object} dto.Response
// @Router /chat-sessions/{uuid} [get]
func (h *ChatHandler) GetSession(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	sessionUUID := c.Param("uuid")

	session, err := h.chatService.GetSessionByUUID(ctx, sessionUUID, userID)
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
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	var req dto.CreateChatSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	session, err := h.chatService.CreateSession(ctx, userID, &req)
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
// @Param uuid path string true "Session UUID"
// @Param request body dto.SendMessageRequest true "Message content"
// @Success 200 {object} dto.Response
// @Router /chat-sessions/{uuid}/messages [post]
func (h *ChatHandler) SendMessage(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	sessionUUID := c.Param("uuid")

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	userMsg, aiMsg, err := h.chatService.SendMessageByUUID(ctx, sessionUUID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	// Update daily task progress for chatting with AI
	if h.dailyTaskService != nil {
		_ = h.dailyTaskService.UpdateTaskProgress(ctx, userID, model.TaskTypeChatAI)
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"user_message": userMsg,
		"ai_message":   aiMsg,
	}, ""))
}

// ToggleTrash godoc
// @Summary Toggle session trash status
// @Description Move session to/from trash
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Session UUID"
// @Success 200 {object} dto.Response
// @Router /chat-sessions/{uuid}/trash [put]
func (h *ChatHandler) ToggleTrash(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	sessionUUID := c.Param("uuid")

	if err := h.chatService.ToggleTrashByUUID(ctx, sessionUUID, userID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Trash status toggled"))
}

// ToggleFavorite godoc
// @Summary Toggle session favorite
// @Description Toggle favorite status for a session
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Session UUID"
// @Success 200 {object} dto.Response
// @Router /chat-sessions/{uuid}/favorite [put]
func (h *ChatHandler) ToggleFavorite(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	sessionUUID := c.Param("uuid")

	if err := h.chatService.ToggleFavoriteByUUID(ctx, sessionUUID, userID); err != nil {
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
// @Param uuid path string true "Session UUID"
// @Success 200 {object} dto.Response
// @Router /chat-sessions/{uuid} [delete]
func (h *ChatHandler) DeleteSession(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	sessionUUID := c.Param("uuid")

	if err := h.chatService.DeleteSessionByUUID(ctx, sessionUUID, userID); err != nil {
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
	ctx := c.Request.Context()
	messageID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid message ID"))
		return
	}

	userID := c.MustGet("user_id").(uint)

	if err := h.chatService.ToggleMessageLike(ctx, uint(messageID), userID); err != nil {
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
	ctx := c.Request.Context()
	messageID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid message ID"))
		return
	}

	userID := c.MustGet("user_id").(uint)

	if err := h.chatService.ToggleMessageDislike(ctx, uint(messageID), userID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Dislike toggled"))
}

// ================================
// Folder Management Handlers
// ================================

// GetFolders godoc
// @Summary Get chat folders
// @Description Get user's chat folders
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /chat-folders [get]
func (h *ChatHandler) GetFolders(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	folders, err := h.chatService.GetFolders(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(folders, ""))
}

// CreateFolder godoc
// @Summary Create chat folder
// @Description Create a new chat folder for organizing sessions
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateChatFolderRequest true "Create folder request"
// @Success 201 {object} dto.Response
// @Router /chat-folders [post]
func (h *ChatHandler) CreateFolder(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	var req dto.CreateChatFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	folder, err := h.chatService.CreateFolder(ctx, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(folder, "Folder created"))
}

// UpdateFolder godoc
// @Summary Update chat folder
// @Description Update folder name, color, or icon
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Folder ID"
// @Param request body dto.UpdateChatFolderRequest true "Update folder request"
// @Success 200 {object} dto.Response
// @Router /chat-folders/{id} [put]
func (h *ChatHandler) UpdateFolder(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	folderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid folder ID"))
		return
	}

	var req dto.UpdateChatFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	folder, err := h.chatService.UpdateFolder(ctx, uint(folderID), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(folder, "Folder updated"))
}

// DeleteFolder godoc
// @Summary Delete chat folder
// @Description Delete a chat folder (sessions will be moved out)
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param id path int true "Folder ID"
// @Success 200 {object} dto.Response
// @Router /chat-folders/{id} [delete]
func (h *ChatHandler) DeleteFolder(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	folderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid folder ID"))
		return
	}

	if err := h.chatService.DeleteFolder(ctx, uint(folderID), userID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Folder deleted"))
}

// ReorderFolders godoc
// @Summary Reorder chat folders
// @Description Reorder folders by providing folder IDs in desired order
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ReorderFoldersRequest true "Folder IDs in order"
// @Success 200 {object} dto.Response
// @Router /chat-folders/reorder [put]
func (h *ChatHandler) ReorderFolders(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	var req dto.ReorderFoldersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if err := h.chatService.ReorderFolders(ctx, userID, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Folders reordered"))
}

// MoveToFolder godoc
// @Summary Move session to folder
// @Tags Chat
// @Security BearerAuth
// @Param uuid path string true "Session UUID"
// @Router /chat-sessions/{uuid}/folder [put]
func (h *ChatHandler) MoveToFolder(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	sessionUUID := c.Param("uuid")

	var req dto.MoveSessionToFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if err := h.chatService.MoveSessionToFolderByUUID(ctx, sessionUUID, userID, req.FolderID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Session moved"))
}

// ================================
// Pin/Unpin Handlers
// ================================

// ToggleMessagePin godoc
// @Summary Toggle message pin
// @Description Toggle pin status for a message
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param id path int true "Message ID"
// @Success 200 {object} dto.Response
// @Router /chat-messages/{id}/pin [put]
func (h *ChatHandler) ToggleMessagePin(c *gin.Context) {
	ctx := c.Request.Context()
	messageID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid message ID"))
		return
	}

	userID := c.MustGet("user_id").(uint)

	if err := h.chatService.ToggleMessagePin(ctx, uint(messageID), userID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Pin toggled"))
}

// GetPinnedMessages godoc
// @Summary Get pinned messages
// @Tags Chat
// @Security BearerAuth
// @Param uuid path string true "Session UUID"
// @Router /chat-sessions/{uuid}/pinned [get]
func (h *ChatHandler) GetPinnedMessages(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	sessionUUID := c.Param("uuid")

	messages, err := h.chatService.GetPinnedMessagesByUUID(ctx, sessionUUID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(messages, ""))
}

// ================================
// Export Handlers
// ================================

// ExportChat godoc
// @Summary Export chat session
// @Tags Chat
// @Security BearerAuth
// @Param uuid path string true "Session UUID"
// @Router /chat-sessions/{uuid}/export [post]
func (h *ChatHandler) ExportChat(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	sessionUUID := c.Param("uuid")

	var req dto.ExportChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	export, err := h.chatService.ExportChatByUUID(ctx, sessionUUID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(export, ""))
}

// ================================
// Summary Handlers
// ================================

// GetSummary godoc
// @Summary Get session summary
// @Tags Chat
// @Security BearerAuth
// @Param uuid path string true "Session UUID"
// @Router /chat-sessions/{uuid}/summary [get]
func (h *ChatHandler) GetSummary(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	sessionUUID := c.Param("uuid")

	summary, err := h.chatService.GetSummaryByUUID(ctx, sessionUUID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(summary, ""))
}

// GenerateSummary godoc
// @Summary Generate session summary
// @Tags Chat
// @Security BearerAuth
// @Param uuid path string true "Session UUID"
// @Router /chat-sessions/{uuid}/summary [post]
func (h *ChatHandler) GenerateSummary(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	sessionUUID := c.Param("uuid")

	summary, err := h.chatService.GenerateSummaryByUUID(ctx, sessionUUID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(summary, "Summary generated"))
}

// ================================
// Suggested Prompts Handler
// ================================

// GetSuggestedPrompts godoc
// @Summary Get suggested prompts
// @Description Get context-aware suggested prompts based on mood, time, and session state
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param mood query string false "Current user mood"
// @Param time_of_day query string false "Time of day: morning, afternoon, evening, night"
// @Param has_messages query bool false "Whether session has existing messages"
// @Success 200 {object} dto.SuggestedPromptsResponse
// @Router /chat-prompts [get]
func (h *ChatHandler) GetSuggestedPrompts(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	var params dto.GetSuggestedPromptsRequest
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	prompts, err := h.chatService.GetSuggestedPrompts(ctx, userID, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(prompts.Prompts, ""))
}

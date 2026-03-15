package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	dailytaskapp "github.com/Alfian57/ruang-tenang-api/internal/features/daily_task/application"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Alfian57/ruang-tenang-api/internal/features/journal/application"
)

// JournalHandler handles HTTP requests for journals
type JournalHandler struct {
	service          *application.JournalService
	dailyTaskService dailytaskapp.DailyTaskService
}

// NewJournalHandler creates a new JournalHandler instance
func NewJournalHandler(service *application.JournalService) *JournalHandler {
	return &JournalHandler{service: service}
}

// SetDailyTaskService sets the daily task service for progress tracking
func (h *JournalHandler) SetDailyTaskService(dailyTaskService dailytaskapp.DailyTaskService) {
	h.dailyTaskService = dailyTaskService
}

// CreateJournal godoc
// @Summary Create a new journal entry
// @Description Create a new private journal entry
// @Tags journals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateJournalRequest true "Create journal request"
// @Success 201 {object} dto.JournalResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /journals [post]
func (h *JournalHandler) CreateJournal(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	var req dto.CreateJournalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	journal, err := h.service.CreateJournal(ctx, userID, req)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "blocked") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update daily task progress for writing journal
	if h.dailyTaskService != nil {
		if err := h.dailyTaskService.UpdateTaskProgress(ctx, userID, model.TaskTypeWriteJournal); err != nil {
			logger.Warn("failed to update daily task progress for writing journal",
				zap.Uint("user_id", userID), zap.Error(err))
		}
	}

	c.JSON(http.StatusCreated, gin.H{"data": journal})
}

// GetJournal godoc
// @Summary Get a journal entry
// @Description Get a specific journal entry by UUID
// @Tags journals
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Journal UUID"
// @Success 200 {object} dto.JournalResponse
// @Failure 404 {object} map[string]interface{}
// @Router /journals/{uuid} [get]
func (h *JournalHandler) GetJournal(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	journalUUID := c.Param("uuid")

	journal, err := h.service.GetJournalByUUID(ctx, userID, journalUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Journal not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": journal})
}

// UpdateJournal godoc
// @Summary Update a journal entry
// @Description Update an existing journal entry
// @Tags journals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Journal UUID"
// @Param request body dto.UpdateJournalRequest true "Update journal request"
// @Success 200 {object} dto.JournalResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /journals/{uuid} [put]
func (h *JournalHandler) UpdateJournal(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	journalUUID := c.Param("uuid")

	var req dto.UpdateJournalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	journal, err := h.service.UpdateJournalByUUID(ctx, userID, journalUUID, req)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "blocked") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Journal not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": journal})
}

// DeleteJournal godoc
// @Summary Delete a journal entry
// @Description Delete a journal entry permanently
// @Tags journals
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Journal UUID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /journals/{uuid} [delete]
func (h *JournalHandler) DeleteJournal(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	journalUUID := c.Param("uuid")

	if err := h.service.DeleteJournalByUUID(ctx, userID, journalUUID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Journal not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Journal deleted successfully"})
}

// ListJournals godoc
// @Summary List journal entries
// @Description Get paginated list of journal entries
// @Tags journals
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param tags query string false "Filter by tags (comma-separated)"
// @Param start_date query string false "Start date filter (YYYY-MM-DD)"
// @Param end_date query string false "End date filter (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Router /journals [get]
func (h *JournalHandler) ListJournals(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	var tags []string
	if tagsStr := c.Query("tags"); tagsStr != "" {
		tags = splitTags(tagsStr)
	}

	var startDate, endDate *time.Time
	if sd := c.Query("start_date"); sd != "" {
		if t, err := time.Parse("2006-01-02", sd); err == nil {
			startDate = &t
		}
	}
	if ed := c.Query("end_date"); ed != "" {
		if t, err := time.Parse("2006-01-02", ed); err == nil {
			t = t.Add(24*time.Hour - time.Second) // End of day
			endDate = &t
		}
	}

	// Parse mood filter
	var moodID *uint
	if m := c.Query("mood"); m != "" {
		if id, err := strconv.ParseUint(m, 10, 32); err == nil {
			mid := uint(id)
			moodID = &mid
		}
	}

	journals, total, err := h.service.ListJournals(ctx, userID, page, limit, tags, moodID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  journals,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// SearchJournals godoc
// @Summary Search journal entries
// @Description Search journals by content or title
// @Tags journals
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param limit query int false "Max results" default(10)
// @Success 200 {object} map[string]interface{}
// @Router /journals/search [get]
func (h *JournalHandler) SearchJournals(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	query := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	journals, err := h.service.SearchJournals(ctx, userID, query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": journals})
}

// ===== Settings Endpoints =====

// GetSettings godoc
// @Summary Get journal settings
// @Description Get user's journal privacy settings
// @Tags journals
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.JournalSettingsResponse
// @Router /journals/settings [get]
func (h *JournalHandler) GetSettings(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	settings, err := h.service.GetSettings(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

// UpdateSettings godoc
// @Summary Update journal settings
// @Description Update user's journal privacy settings
// @Tags journals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.JournalSettingsRequest true "Settings request"
// @Success 200 {object} dto.JournalSettingsResponse
// @Router /journals/settings [put]
func (h *JournalHandler) UpdateSettings(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	var req dto.JournalSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.service.UpdateSettings(ctx, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

// ===== AI Context Endpoints =====

// GetAIContext godoc
// @Summary Get journal context for AI
// @Description Get journal entries as context for AI chatbot
// @Tags journals
// @Produce json
// @Security BearerAuth
// @Param query query string false "Optional search query for relevant entries"
// @Param max_entries query int false "Maximum entries to return"
// @Param include_summary query bool false "Include AI-generated summary"
// @Success 200 {object} dto.JournalAIContext
// @Router /journals/ai-context [get]
func (h *JournalHandler) GetAIContext(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	req := dto.JournalAIContextRequest{
		Query:          c.Query("query"),
		IncludeSummary: c.Query("include_summary") == "true",
	}

	if me := c.Query("max_entries"); me != "" {
		if v, err := strconv.Atoi(me); err == nil {
			req.MaxEntries = v
		}
	}

	context, err := h.service.GetAIContext(ctx, userID, nil, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": context})
}

// GetAIAccessLogs godoc
// @Summary Get AI access logs
// @Description Get log of when AI accessed journal entries
// @Tags journals
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Max results" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /journals/ai-access-logs [get]
func (h *JournalHandler) GetAIAccessLogs(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	logs, err := h.service.GetAIAccessLogs(ctx, userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

// ===== Analytics Endpoints =====

// GetAnalytics godoc
// @Summary Get journal analytics
// @Description Get analytics data for user's journals
// @Tags journals
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.JournalAnalytics
// @Router /journals/analytics [get]
func (h *JournalHandler) GetAnalytics(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	analytics, err := h.service.GetAnalytics(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": analytics})
}

// ===== AI-Powered Features =====

// GetWritingPrompt godoc
// @Summary Get AI writing prompt
// @Description Get an AI-generated writing prompt based on mood and history
// @Tags journals
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.JournalPromptResponse
// @Router /journals/prompt [get]
func (h *JournalHandler) GetWritingPrompt(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	prompt, err := h.service.GetWritingPrompt(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": prompt})
}

// GetWeeklySummary godoc
// @Summary Get weekly summary
// @Description Get AI-generated weekly summary of journals
// @Tags journals
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.JournalWeeklySummary
// @Router /journals/weekly-summary [get]
func (h *JournalHandler) GetWeeklySummary(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	summary, err := h.service.GetWeeklySummary(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// ===== Export =====

// ExportJournals godoc
// @Summary Export journals
// @Description Export journals in PDF or TXT format
// @Tags journals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.JournalExportRequest true "Export request"
// @Success 200 {object} dto.JournalExportResponse
// @Router /journals/export [post]
func (h *JournalHandler) ExportJournals(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	var req dto.JournalExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	export, err := h.service.ExportJournals(ctx, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": export})
}

// ===== Toggle AI Share =====

// ToggleAIShare godoc
// @Summary Toggle AI sharing for a journal
// @Description Toggle whether a journal entry is shared with AI
// @Tags journals
// @Produce json
// @Security BearerAuth
// @Param uuid path string true "Journal UUID"
// @Success 200 {object} dto.JournalResponse
// @Router /journals/{uuid}/toggle-ai-share [post]
func (h *JournalHandler) ToggleAIShare(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)
	journalUUID := c.Param("uuid")

	// Get current state
	journal, err := h.service.GetJournalByUUID(ctx, userID, journalUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Journal not found"})
		return
	}

	// Toggle
	newValue := !journal.ShareWithAI
	updated, err := h.service.UpdateJournalByUUID(ctx, userID, journalUUID, dto.UpdateJournalRequest{
		ShareWithAI: &newValue,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updated})
}

// Helper function to split tags
func splitTags(tagsStr string) []string {
	var tags []string
	for _, t := range splitString(tagsStr, ",") {
		t = trimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

func splitString(s string, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

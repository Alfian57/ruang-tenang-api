package handlers

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BreathingHandler struct {
	service services.BreathingService
}

func NewBreathingHandler(service services.BreathingService) *BreathingHandler {
	return &BreathingHandler{service: service}
}

// GetTechniques godoc
// @Summary Get all breathing techniques
// @Description Get all available breathing techniques (system + user's custom)
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]dto.BreathingTechniqueResponse}
// @Failure 401 {object} dto.Response
// @Router /api/breathing/techniques [get]
func (h *BreathingHandler) GetTechniques(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	techniques, err := h.service.GetAllTechniques(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get techniques: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(techniques, "Techniques retrieved successfully"))
}

// GetTechniqueByID godoc
// @Summary Get a technique by ID
// @Description Get a specific breathing technique by ID
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Technique ID"
// @Success 200 {object} dto.Response{data=dto.BreathingTechniqueResponse}
// @Failure 404 {object} dto.Response
// @Router /api/breathing/techniques/{id} [get]
func (h *BreathingHandler) GetTechniqueByID(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	techniqueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid technique ID"))
		return
	}

	technique, err := h.service.GetTechniqueByID(userID, techniqueID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Technique not found"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(technique, "Technique retrieved successfully"))
}

// GetTechniqueBySlug godoc
// @Summary Get a technique by slug
// @Description Get a specific breathing technique by slug
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Technique slug"
// @Success 200 {object} dto.Response{data=dto.BreathingTechniqueResponse}
// @Failure 404 {object} dto.Response
// @Router /api/breathing/techniques/slug/{slug} [get]
func (h *BreathingHandler) GetTechniqueBySlug(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	slug := c.Param("slug")

	technique, err := h.service.GetTechniqueBySlug(userID, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Technique not found"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(technique, "Technique retrieved successfully"))
}

// CreateTechnique godoc
// @Summary Create a custom technique
// @Description Create a custom breathing technique
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateBreathingTechniqueRequest true "Technique data"
// @Success 201 {object} dto.Response{data=dto.BreathingTechniqueResponse}
// @Failure 400 {object} dto.Response
// @Router /api/breathing/techniques [post]
func (h *BreathingHandler) CreateTechnique(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req dto.CreateBreathingTechniqueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid request: "+err.Error()))
		return
	}

	technique, err := h.service.CreateCustomTechnique(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to create technique: "+err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(technique, "Technique created successfully"))
}

// UpdateTechnique godoc
// @Summary Update a custom technique
// @Description Update a custom breathing technique
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Technique ID"
// @Param request body dto.UpdateBreathingTechniqueRequest true "Technique data"
// @Success 200 {object} dto.Response{data=dto.BreathingTechniqueResponse}
// @Failure 400 {object} dto.Response
// @Router /api/breathing/techniques/{id} [put]
func (h *BreathingHandler) UpdateTechnique(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	techniqueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid technique ID"))
		return
	}

	var req dto.UpdateBreathingTechniqueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid request: "+err.Error()))
		return
	}

	technique, err := h.service.UpdateCustomTechnique(userID, techniqueID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to update technique: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(technique, "Technique updated successfully"))
}

// DeleteTechnique godoc
// @Summary Delete a custom technique
// @Description Delete a custom breathing technique
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Technique ID"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /api/breathing/techniques/{id} [delete]
func (h *BreathingHandler) DeleteTechnique(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	techniqueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid technique ID"))
		return
	}

	if err := h.service.DeleteCustomTechnique(userID, techniqueID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to delete technique: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Technique deleted successfully"))
}

// StartSession godoc
// @Summary Start a breathing session
// @Description Start a new breathing practice session
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.StartBreathingSessionRequest true "Session settings"
// @Success 201 {object} dto.Response{data=dto.BreathingSessionResponse}
// @Failure 400 {object} dto.Response
// @Router /api/breathing/sessions [post]
func (h *BreathingHandler) StartSession(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req dto.StartBreathingSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid request: "+err.Error()))
		return
	}

	session, err := h.service.StartSession(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to start session: "+err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(session, "Session started successfully"))
}

// CompleteSession godoc
// @Summary Complete a breathing session
// @Description Complete a breathing session and calculate XP
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param request body dto.CompleteBreathingSessionRequest true "Completion data"
// @Success 200 {object} dto.Response{data=dto.SessionCompletionResult}
// @Failure 400 {object} dto.Response
// @Router /api/breathing/sessions/{id}/complete [post]
func (h *BreathingHandler) CompleteSession(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid session ID"))
		return
	}

	var req dto.CompleteBreathingSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid request: "+err.Error()))
		return
	}

	result, err := h.service.CompleteSession(userID, sessionID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to complete session: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Session completed successfully"))
}

// GetSessionHistory godoc
// @Summary Get session history
// @Description Get paginated breathing session history
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param technique_id query string false "Filter by technique ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.Response{data=dto.SessionHistoryResponse}
// @Failure 400 {object} dto.Response
// @Router /api/breathing/sessions [get]
func (h *BreathingHandler) GetSessionHistory(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	req := dto.SessionHistoryRequest{
		StartDate:   c.Query("start_date"),
		EndDate:     c.Query("end_date"),
		TechniqueID: c.Query("technique_id"),
		Page:        page,
		Limit:       limit,
	}

	history, err := h.service.GetSessionHistory(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to get session history: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(history, "Session history retrieved successfully"))
}

// GetSessionByID godoc
// @Summary Get session by ID
// @Description Get a specific breathing session by ID
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} dto.Response{data=dto.BreathingSessionResponse}
// @Failure 404 {object} dto.Response
// @Router /api/breathing/sessions/{id} [get]
func (h *BreathingHandler) GetSessionByID(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid session ID"))
		return
	}

	session, err := h.service.GetSessionByID(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Session not found"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(session, "Session retrieved successfully"))
}

// GetPreferences godoc
// @Summary Get breathing preferences
// @Description Get user's breathing preferences
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=dto.BreathingPreferencesResponse}
// @Failure 400 {object} dto.Response
// @Router /api/breathing/preferences [get]
func (h *BreathingHandler) GetPreferences(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	prefs, err := h.service.GetPreferences(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to get preferences: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(prefs, "Preferences retrieved successfully"))
}

// UpdatePreferences godoc
// @Summary Update breathing preferences
// @Description Update user's breathing preferences
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateBreathingPreferencesRequest true "Preferences data"
// @Success 200 {object} dto.Response{data=dto.BreathingPreferencesResponse}
// @Failure 400 {object} dto.Response
// @Router /api/breathing/preferences [put]
func (h *BreathingHandler) UpdatePreferences(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req dto.UpdateBreathingPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid request: "+err.Error()))
		return
	}

	prefs, err := h.service.UpdatePreferences(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to update preferences: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(prefs, "Preferences updated successfully"))
}

// GetFavorites godoc
// @Summary Get favorite techniques
// @Description Get user's favorite breathing techniques
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]dto.BreathingTechniqueResponse}
// @Failure 400 {object} dto.Response
// @Router /api/breathing/favorites [get]
func (h *BreathingHandler) GetFavorites(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	favorites, err := h.service.GetFavorites(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to get favorites: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(favorites, "Favorites retrieved successfully"))
}

// AddFavorite godoc
// @Summary Add technique to favorites
// @Description Add a breathing technique to favorites
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Technique ID"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /api/breathing/favorites/{id} [post]
func (h *BreathingHandler) AddFavorite(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	techniqueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid technique ID"))
		return
	}

	if err := h.service.AddFavorite(userID, techniqueID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to add favorite: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Added to favorites"))
}

// RemoveFavorite godoc
// @Summary Remove technique from favorites
// @Description Remove a breathing technique from favorites
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Technique ID"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /api/breathing/favorites/{id} [delete]
func (h *BreathingHandler) RemoveFavorite(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	techniqueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid technique ID"))
		return
	}

	if err := h.service.RemoveFavorite(userID, techniqueID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to remove favorite: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Removed from favorites"))
}

// ReorderFavorites godoc
// @Summary Reorder favorite techniques
// @Description Reorder favorite breathing techniques
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body []string true "Technique IDs in new order"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /api/breathing/favorites/reorder [put]
func (h *BreathingHandler) ReorderFavorites(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var techniqueIDStrings []string
	if err := c.ShouldBindJSON(&techniqueIDStrings); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid request: "+err.Error()))
		return
	}

	techniqueIDs := make([]uuid.UUID, len(techniqueIDStrings))
	for i, idStr := range techniqueIDStrings {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid technique ID"))
			return
		}
		techniqueIDs[i] = id
	}

	if err := h.service.ReorderFavorites(userID, techniqueIDs); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to reorder favorites: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Favorites reordered successfully"))
}

// GetStats godoc
// @Summary Get breathing stats
// @Description Get user's breathing statistics
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=dto.BreathingStatsResponse}
// @Failure 400 {object} dto.Response
// @Router /api/breathing/stats [get]
func (h *BreathingHandler) GetStats(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	stats, err := h.service.GetStats(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to get stats: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(stats, "Stats retrieved successfully"))
}

// GetCalendar godoc
// @Summary Get breathing calendar
// @Description Get monthly calendar view of breathing practice
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param year query int true "Year"
// @Param month query int true "Month (1-12)"
// @Success 200 {object} dto.Response{data=dto.BreathingCalendarResponse}
// @Failure 400 {object} dto.Response
// @Router /api/breathing/calendar [get]
func (h *BreathingHandler) GetCalendar(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	year, err := strconv.Atoi(c.Query("year"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid year"))
		return
	}

	month, err := strconv.Atoi(c.Query("month"))
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid month"))
		return
	}

	calendar, err := h.service.GetCalendar(userID, year, month)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to get calendar: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(calendar, "Calendar retrieved successfully"))
}

// GetTechniqueUsage godoc
// @Summary Get technique usage stats
// @Description Get technique usage statistics for charts
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=[]dto.TechniqueUsageStats}
// @Failure 400 {object} dto.Response
// @Router /api/breathing/stats/usage [get]
func (h *BreathingHandler) GetTechniqueUsage(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	usage, err := h.service.GetTechniqueUsage(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to get technique usage: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(usage, "Technique usage retrieved successfully"))
}

// GetWidgetData godoc
// @Summary Get widget data
// @Description Get breathing data for dashboard widget
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response{data=dto.BreathingWidgetData}
// @Failure 400 {object} dto.Response
// @Router /api/breathing/widget [get]
func (h *BreathingHandler) GetWidgetData(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	widget, err := h.service.GetWidgetData(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to get widget data: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(widget, "Widget data retrieved successfully"))
}

// GetRecommendations godoc
// @Summary Get breathing recommendations
// @Description Get smart breathing technique recommendations
// @Tags Breathing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param mood query string false "Current mood (e.g., anxious, tired, stressed)"
// @Param time_of_day query string false "Time of day (morning, afternoon, evening, night)"
// @Success 200 {object} dto.Response{data=dto.RecommendationsResponse}
// @Failure 400 {object} dto.Response
// @Router /api/breathing/recommendations [get]
func (h *BreathingHandler) GetRecommendations(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	mood := c.Query("mood")
	timeOfDay := c.Query("time_of_day")

	recommendations, err := h.service.GetRecommendations(userID, mood, timeOfDay)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Failed to get recommendations: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(recommendations, "Recommendations retrieved successfully"))
}

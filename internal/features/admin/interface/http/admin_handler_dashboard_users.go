package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/gin-gonic/gin"
)

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Get platform statistics for admin dashboard
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /admin/stats [get]
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	ctx := c.Request.Context()
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonthStart := monthStart.AddDate(0, -1, 0)
	lastMonthEnd := monthStart.Add(-time.Second)

	var totalUsers int64
	var activeUsers int64
	var blockedUsers int64
	var usersThisMonth int64
	var userChartData []int64 = make([]int64, 7)
	var userGrowth float64
	var recentUsersDTO []gin.H = make([]gin.H, 0)

	// The first count doubles as a DB health check: if it fails, the database
	// is unreachable and returning all-zero stats would be misleading.
	if err := h.db.WithContext(ctx).Model(&model.User{}).Count(&totalUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to load dashboard statistics"))
		return
	}
	h.db.WithContext(ctx).Model(&model.User{}).Where("is_blocked = ?", false).Count(&activeUsers)
	h.db.WithContext(ctx).Model(&model.User{}).Where("is_blocked = ?", true).Count(&blockedUsers)
	h.db.WithContext(ctx).Model(&model.User{}).Where("created_at >= ?", monthStart).Count(&usersThisMonth)

	var usersLastMonth int64
	h.db.WithContext(ctx).Model(&model.User{}).Where("created_at >= ? AND created_at <= ?", lastMonthStart, lastMonthEnd).Count(&usersLastMonth)
	if usersLastMonth > 0 {
		userGrowth = float64(usersThisMonth-usersLastMonth) / float64(usersLastMonth) * 100
	}

	for i := 6; i >= 0; i-- {
		dayStart := todayStart.AddDate(0, 0, -i)
		dayEnd := dayStart.Add(24 * time.Hour)
		var count int64
		h.db.WithContext(ctx).Model(&model.User{}).Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).Count(&count)
		userChartData[6-i] = count
	}

	var recentUsers []model.User
	h.db.WithContext(ctx).Order("created_at DESC").Limit(5).Find(&recentUsers)
	for _, u := range recentUsers {
		recentUsersDTO = append(recentUsersDTO, gin.H{
			"id":         u.ID,
			"name":       u.Name,
			"email":      u.Email,
			"role":       u.Role,
			"is_blocked": u.IsBlocked,
			"created_at": u.CreatedAt,
		})
	}

	var totalArticles int64
	h.db.WithContext(ctx).Model(&model.Article{}).Count(&totalArticles)

	var articlesThisMonth int64
	h.db.WithContext(ctx).Model(&model.Article{}).Where("created_at >= ?", monthStart).Count(&articlesThisMonth)

	var totalChatSessions int64
	h.db.WithContext(ctx).Model(&model.ChatSession{}).Count(&totalChatSessions)

	var chatSessionsToday int64
	h.db.WithContext(ctx).Model(&model.ChatSession{}).Where("created_at >= ?", todayStart).Count(&chatSessionsToday)

	var totalMessages int64
	h.db.WithContext(ctx).Model(&model.ChatMessage{}).Count(&totalMessages)

	var messagesToday int64
	h.db.WithContext(ctx).Model(&model.ChatMessage{}).Where("created_at >= ?", todayStart).Count(&messagesToday)

	var totalSongs int64
	h.db.WithContext(ctx).Model(&model.Song{}).Count(&totalSongs)

	var totalSongCategories int64
	h.db.WithContext(ctx).Model(&model.SongCategory{}).Count(&totalSongCategories)

	var totalMoods int64
	h.db.WithContext(ctx).Model(&model.UserMood{}).Count(&totalMoods)

	var moodsToday int64
	h.db.WithContext(ctx).Model(&model.UserMood{}).Where("created_at >= ?", todayStart).Count(&moodsToday)

	chatChartData := make([]int64, 7)
	for i := 6; i >= 0; i-- {
		dayStart := todayStart.AddDate(0, 0, -i)
		dayEnd := dayStart.Add(24 * time.Hour)
		var count int64
		h.db.WithContext(ctx).Model(&model.ChatSession{}).Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).Count(&count)
		chatChartData[6-i] = count
	}

	var totalArticleCategories int64
	h.db.WithContext(ctx).Model(&model.ArticleCategory{}).Count(&totalArticleCategories)

	var blockedArticles int64
	h.db.WithContext(ctx).Model(&model.Article{}).Where("status = ?", "blocked").Count(&blockedArticles)

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"users": gin.H{
			"total":      totalUsers,
			"active":     activeUsers,
			"blocked":    blockedUsers,
			"this_month": usersThisMonth,
			"growth":     userGrowth,
			"chart_data": userChartData,
		},
		"articles": gin.H{
			"total":      totalArticles,
			"this_month": articlesThisMonth,
			"blocked":    blockedArticles,
			"categories": totalArticleCategories,
		},
		"chat_sessions": gin.H{
			"total":      totalChatSessions,
			"today":      chatSessionsToday,
			"chart_data": chatChartData,
		},
		"messages": gin.H{
			"total": totalMessages,
			"today": messagesToday,
		},
		"songs": gin.H{
			"total":      totalSongs,
			"categories": totalSongCategories,
		},
		"moods": gin.H{
			"total": totalMoods,
			"today": moodsToday,
		},
		"recent_users": recentUsersDTO,
	}, ""))
}

// GetUsers godoc
// @Summary Get all users
// @Description Get paginated list of users for admin
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param search query string false "Search by name or email"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.PaginatedResponse
// @Router /admin/users [get]
func (h *AdminHandler) GetUsers(c *gin.Context) {
	ctx := c.Request.Context()
	var params struct {
		Search string `form:"search"`
		Role   string `form:"role"`
		Page   int    `form:"page"`
		Limit  int    `form:"limit"`
	}
	c.ShouldBindQuery(&params)

	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 50 {
		params.Limit = 10
	}

	var users []model.User
	var total int64

	query := h.db.WithContext(ctx).Model(&model.User{})
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", searchTerm, searchTerm)
	}
	if role := strings.TrimSpace(strings.ToLower(params.Role)); role != "" {
		if role != string(model.RoleUser) && role != string(model.RoleMitra) && role != string(model.RoleAdmin) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid role filter"))
			return
		}
		query = query.Where("role = ?", role)
	}

	query.Count(&total)
	query.Offset((params.Page - 1) * params.Limit).Limit(params.Limit).Order("created_at DESC").Find(&users)

	journalBlockedByUserID := make(map[uint]bool, len(users))
	if len(users) > 0 {
		userIDs := make([]uint, 0, len(users))
		for _, u := range users {
			userIDs = append(userIDs, u.ID)
		}

		var journalSettings []model.JournalSettings
		h.db.WithContext(ctx).
			Where("user_id IN ?", userIDs).
			Find(&journalSettings)

		for _, setting := range journalSettings {
			journalBlockedByUserID[setting.UserID] = setting.IsBlocked
		}
	}

	result := make([]gin.H, len(users))
	for i, u := range users {
		result[i] = gin.H{
			"id":               u.ID,
			"name":             u.Name,
			"email":            u.Email,
			"avatar":           u.Avatar,
			"role":             u.Role,
			"is_blocked":       u.IsBlocked,
			"is_banned":        u.IsBanned,
			"is_suspended":     u.IsSuspended(),
			"suspension_end":   u.SuspensionEnd,
			"journal_blocked":  journalBlockedByUserID[u.ID],
			"is_forum_blocked": u.IsForumBlocked,
			"created_at":       u.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(result, params.Page, params.Limit, total))
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req struct {
		Role string `json:"role" binding:"required,oneof=user mitra"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	var user model.User
	if err := h.db.WithContext(ctx).First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("User not found"))
		return
	}

	if user.Role == model.RoleAdmin {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Admin role cannot be changed here"))
		return
	}

	nextRole := model.UserRole(strings.TrimSpace(strings.ToLower(req.Role)))
	if nextRole != model.RoleUser && nextRole != model.RoleMitra {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid role"))
		return
	}

	user.Role = nextRole
	if err := h.db.WithContext(ctx).Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to update user role"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"avatar":     user.Avatar,
		"role":       user.Role,
		"is_blocked": user.IsBlocked,
		"created_at": user.CreatedAt,
	}, "User role updated"))
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user by ID
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.Response
// @Router /admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var user model.User
	if err := h.db.WithContext(ctx).First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("User not found"))
		return
	}

	// Never allow deleting admin accounts.
	if user.Role == model.RoleAdmin {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Cannot delete admin users"))
		return
	}

	// Never allow an admin to delete their own account.
	if requesterID, ok := middleware.GetUserID(c); ok && requesterID == user.ID {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("You cannot delete your own account"))
		return
	}

	if err := h.db.WithContext(ctx).Delete(&model.User{}, user.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to delete user"))
		return
	}

	h.invalidateAccountStatus(user.ID)
	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "User deleted"))
}

// BlockUser godoc
// @Summary Block a user
// @Description Block a user by ID (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.Response
// @Router /admin/users/{id}/block [put]
func (h *AdminHandler) BlockUser(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var user model.User
	if err := h.db.WithContext(ctx).First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("User not found"))
		return
	}

	if user.Role == model.RoleAdmin {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Cannot block admin users"))
		return
	}

	user.IsBlocked = true
	if err := h.db.WithContext(ctx).Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to block user"))
		return
	}

	h.invalidateAccountStatus(user.ID)
	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "User blocked"))
}

// UnblockUser godoc
// @Summary Unblock a user
// @Description Unblock a user by ID (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.Response
// @Router /admin/users/{id}/unblock [put]
func (h *AdminHandler) UnblockUser(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var user model.User
	if err := h.db.WithContext(ctx).First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("User not found"))
		return
	}

	user.IsBlocked = false
	if err := h.db.WithContext(ctx).Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to unblock user"))
		return
	}

	h.invalidateAccountStatus(user.ID)
	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "User unblocked"))
}

// invalidateAccountStatus clears the cached ban/block/suspend status so changes
// take effect on the user's very next request instead of after the cache TTL.
func (h *AdminHandler) invalidateAccountStatus(userID uint) {
	middleware.InvalidateAccountStatus(userID)
}

// BanUser godoc
// @Summary Ban a user
// @Description Permanently ban a user by ID (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.Response
// @Router /admin/users/{id}/ban [put]
func (h *AdminHandler) BanUser(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	var user model.User
	if err := h.db.WithContext(ctx).First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("User not found"))
		return
	}

	if user.Role == model.RoleAdmin {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Cannot ban admin users"))
		return
	}

	if err := h.db.WithContext(ctx).Model(&user).Updates(map[string]interface{}{
		"is_banned":  true,
		"ban_reason": strings.TrimSpace(req.Reason),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to ban user"))
		return
	}

	h.invalidateAccountStatus(user.ID)
	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "User banned"))
}

// UnbanUser godoc
// @Summary Unban a user
// @Description Lift a permanent ban from a user by ID (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.Response
// @Router /admin/users/{id}/unban [put]
func (h *AdminHandler) UnbanUser(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var user model.User
	if err := h.db.WithContext(ctx).First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("User not found"))
		return
	}

	if err := h.db.WithContext(ctx).Model(&user).Updates(map[string]interface{}{
		"is_banned":  false,
		"ban_reason": "",
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to unban user"))
		return
	}

	h.invalidateAccountStatus(user.ID)
	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "User unbanned"))
}

// UnsuspendUser godoc
// @Summary Lift a user's suspension
// @Description Remove an active suspension from a user by ID (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.Response
// @Router /admin/users/{id}/unsuspend [put]
func (h *AdminHandler) UnsuspendUser(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var user model.User
	if err := h.db.WithContext(ctx).First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("User not found"))
		return
	}

	if err := h.db.WithContext(ctx).Model(&user).Updates(map[string]interface{}{
		"suspension_end":    nil,
		"suspension_reason": "",
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to lift suspension"))
		return
	}

	h.invalidateAccountStatus(user.ID)
	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "User suspension lifted"))
}

// ToggleJournalBlock godoc
// @Summary Toggle journal block for a user
// @Description Block or unblock a user from accessing journal features (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.Response
// @Router /admin/users/{id}/block-journal [put]
func (h *AdminHandler) ToggleJournalBlock(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")

	var user model.User
	if err := h.db.WithContext(ctx).First(&user, idStr).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("User not found"))
		return
	}

	if user.Role == model.RoleAdmin {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Cannot block admin users"))
		return
	}

	settings, err := h.journalService.ToggleJournalBlock(ctx, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to toggle journal block: "+err.Error()))
		return
	}

	status := "unblocked"
	if settings.IsBlocked {
		status = "blocked"
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(settings, "User journal access "+status))
}

// ToggleForumBlock godoc
// @Summary Toggle forum block for a user
// @Description Block or unblock a user from accessing forum features (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.Response
// @Router /admin/users/{id}/block-forum [put]
func (h *AdminHandler) ToggleForumBlock(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")

	var user model.User
	if err := h.db.WithContext(ctx).First(&user, idStr).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("User not found"))
		return
	}

	if user.Role == model.RoleAdmin {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Cannot block admin users from forum"))
		return
	}

	user.IsForumBlocked = !user.IsForumBlocked
	if err := h.db.WithContext(ctx).Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to toggle forum block"))
		return
	}

	status := "unblocked"
	if user.IsForumBlocked {
		status = "blocked"
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"is_forum_blocked": user.IsForumBlocked,
	}, "User forum access "+status))
}

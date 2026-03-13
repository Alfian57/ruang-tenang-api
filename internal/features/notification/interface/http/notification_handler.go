package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Alfian57/ruang-tenang-api/internal/features/notification/application")

type NotificationHandler struct {
	notifService *application.NotificationService
}

func NewNotificationHandler(notifService *application.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifService: notifService}
}

// GetNotifications returns paginated notifications for the authenticated user.
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID := c.GetUint("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	unreadOnly := c.Query("unread_only") == "true"

	result, err := h.notifService.GetNotifications(c.Request.Context(), userID, page, limit, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GetUnreadCount returns the count of unread notifications.
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := c.GetUint("user_id")

	count, err := h.notifService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"unread_count": count}})
}

// MarkAsRead marks a specific notification as read.
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.GetUint("user_id")
	notifIDStr := c.Param("id")

	notifID, err := uuid.Parse(notifIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	if err := h.notifService.MarkAsRead(c.Request.Context(), notifID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllAsRead marks all notifications as read for the authenticated user.
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetUint("user_id")

	if err := h.notifService.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

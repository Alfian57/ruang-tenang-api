package handler

import (
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
)

type PushHandler struct {
	pushService *service.PushService
}

func NewPushHandler(pushService *service.PushService) *PushHandler {
	return &PushHandler{pushService: pushService}
}

// GetVAPIDKey returns the VAPID public key for the frontend.
func (h *PushHandler) GetVAPIDKey(c *gin.Context) {
	key := h.pushService.GetVAPIDPublicKey()
	c.JSON(http.StatusOK, gin.H{"data": dto.PushVAPIDKeyResponse{PublicKey: key}})
}

// Subscribe registers a push subscription for the authenticated user.
func (h *PushHandler) Subscribe(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req dto.PushSubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.pushService.Subscribe(c.Request.Context(), userID, req.Endpoint, req.P256dh, req.Auth); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to subscribe"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscribed to push notifications"})
}

// Unsubscribe removes a push subscription.
func (h *PushHandler) Unsubscribe(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req dto.PushUnsubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.pushService.Unsubscribe(c.Request.Context(), userID, req.Endpoint); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unsubscribe"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed from push notifications"})
}

package handler

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/gin-gonic/gin"

	"github.com/Alfian57/ruang-tenang-api/internal/features/broadcast/application")

type BroadcastHandler struct {
	broadcastService *application.BroadcastService
}

func NewBroadcastHandler(broadcastService *application.BroadcastService) *BroadcastHandler {
	return &BroadcastHandler{broadcastService: broadcastService}
}

func (h *BroadcastHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req dto.CreateBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	var scheduledAt *time.Time
	if req.ScheduledAt != nil && *req.ScheduledAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format waktu jadwal tidak valid (gunakan RFC3339)"})
			return
		}
		if t.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Waktu jadwal harus di masa depan"})
			return
		}
		scheduledAt = &t
	}

	b, err := h.broadcastService.Create(c.Request.Context(), userID, req.Title, req.Body, req.Icon, req.URL, scheduledAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat broadcast"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": toBroadcastResponse(b)})
}

func (h *BroadcastHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	broadcasts, total, err := h.broadcastService.GetAll(c.Request.Context(), page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat broadcast"})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	var items []dto.BroadcastResponse
	for i := range broadcasts {
		items = append(items, toBroadcastResponse(&broadcasts[i]))
	}
	if items == nil {
		items = []dto.BroadcastResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total_items": total,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	})
}

func (h *BroadcastHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	b, err := h.broadcastService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Broadcast tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toBroadcastResponse(b)})
}

func (h *BroadcastHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	var scheduledAt *time.Time
	if req.ScheduledAt != nil && *req.ScheduledAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format waktu jadwal tidak valid"})
			return
		}
		if t.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Waktu jadwal harus di masa depan"})
			return
		}
		scheduledAt = &t
	}

	b, err := h.broadcastService.Update(c.Request.Context(), id, req.Title, req.Body, req.Icon, req.URL, scheduledAt)
	if err != nil {
		if _, ok := err.(*application.BroadcastError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui broadcast"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toBroadcastResponse(b)})
}

func (h *BroadcastHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.broadcastService.Delete(c.Request.Context(), id)
	if err != nil {
		if _, ok := err.(*application.BroadcastError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus broadcast"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Broadcast berhasil dihapus"})
}

func (h *BroadcastHandler) SendNow(c *gin.Context) {
	id := c.Param("id")

	b, err := h.broadcastService.SendNow(c.Request.Context(), id)
	if err != nil {
		if _, ok := err.(*application.BroadcastError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim broadcast"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    toBroadcastResponse(b),
		"message": "Broadcast sedang dikirim",
	})
}

func (h *BroadcastHandler) Cancel(c *gin.Context) {
	id := c.Param("id")

	b, err := h.broadcastService.Cancel(c.Request.Context(), id)
	if err != nil {
		if _, ok := err.(*application.BroadcastError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membatalkan broadcast"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toBroadcastResponse(b)})
}

func toBroadcastResponse(b *model.BroadcastNotification) dto.BroadcastResponse {
	creatorName := ""
	if b.Creator.Name != "" {
		creatorName = b.Creator.Name
	}
	return dto.BroadcastResponse{
		ID:          b.ID.String(),
		Title:       b.Title,
		Body:        b.Body,
		Icon:        b.Icon,
		URL:         b.URL,
		Status:      string(b.Status),
		ScheduledAt: b.ScheduledAt,
		SentAt:      b.SentAt,
		SentCount:   b.SentCount,
		FailedCount: b.FailedCount,
		CreatedBy:   b.CreatedBy,
		CreatorName: creatorName,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

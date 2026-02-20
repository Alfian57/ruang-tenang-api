package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/uuid"
)

type NotificationService struct {
	notifRepo *repository.NotificationRepository
}

func NewNotificationService(notifRepo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifRepo: notifRepo}
}

// CreateHeartNotification creates a notification when a story receives a heart
func (s *NotificationService) CreateHeartNotification(ctx context.Context, authorID uint, heartGiverName string, storyTitle string, storyID string) {
	data, _ := json.Marshal(map[string]string{"story_id": storyID})

	notification := &model.Notification{
		ID:      uuid.New(),
		UserID:  authorID,
		Type:    model.NotificationTypeHeart,
		Title:   "Heart Baru! ❤️",
		Message: fmt.Sprintf("%s menyukai ceritamu \"%s\"", heartGiverName, truncateTitle(storyTitle, 50)),
		Data:    string(data),
	}

	// Best-effort - don't fail the heart action if notification fails
	s.notifRepo.Create(ctx, notification)
}

// CreateStoryApprovedNotification creates a notification when a story is approved
func (s *NotificationService) CreateStoryApprovedNotification(ctx context.Context, authorID uint, storyTitle string, storyID string) {
	data, _ := json.Marshal(map[string]string{"story_id": storyID})

	notification := &model.Notification{
		ID:      uuid.New(),
		UserID:  authorID,
		Type:    model.NotificationTypeStoryApproved,
		Title:   "Cerita Disetujui! 🎉",
		Message: fmt.Sprintf("Ceritamu \"%s\" telah disetujui dan dipublikasikan", truncateTitle(storyTitle, 50)),
		Data:    string(data),
	}

	s.notifRepo.Create(ctx, notification)
}

// CreateStoryRejectedNotification creates a notification when a story is rejected
func (s *NotificationService) CreateStoryRejectedNotification(ctx context.Context, authorID uint, storyTitle string, storyID string, feedback string) {
	dataMap := map[string]string{"story_id": storyID}
	if feedback != "" {
		dataMap["feedback"] = feedback
	}
	data, _ := json.Marshal(dataMap)

	notification := &model.Notification{
		ID:      uuid.New(),
		UserID:  authorID,
		Type:    model.NotificationTypeStoryRejected,
		Title:   "Cerita Perlu Revisi",
		Message: fmt.Sprintf("Ceritamu \"%s\" memerlukan revisi", truncateTitle(storyTitle, 50)),
		Data:    string(data),
	}

	s.notifRepo.Create(ctx, notification)
}

// GetNotifications returns paginated user notifications
func (s *NotificationService) GetNotifications(ctx context.Context, userID uint, page, limit int, unreadOnly bool) (*dto.NotificationListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	notifications, total, err := s.notifRepo.GetByUserID(ctx, userID, page, limit, unreadOnly)
	if err != nil {
		return nil, err
	}

	items := make([]dto.NotificationResponse, len(notifications))
	for i, n := range notifications {
		items[i] = dto.NotificationResponse{
			ID:        n.ID,
			Type:      string(n.Type),
			Title:     n.Title,
			Message:   n.Message,
			IsRead:    n.IsRead,
			Data:      n.Data,
			CreatedAt: n.CreatedAt,
		}
	}

	unreadCount, _ := s.notifRepo.GetUnreadCount(ctx, userID)

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &dto.NotificationListResponse{
		Notifications: items,
		Total:         total,
		Page:          page,
		Limit:         limit,
		TotalPages:    totalPages,
		UnreadCount:   unreadCount,
	}, nil
}

// GetUnreadCount returns count of unread notifications
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	return s.notifRepo.GetUnreadCount(ctx, userID)
}

// MarkAsRead marks a single notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uint) error {
	return s.notifRepo.MarkAsRead(ctx, notificationID, userID)
}

// MarkAllAsRead marks all notifications as read
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uint) error {
	return s.notifRepo.MarkAllAsRead(ctx, userID)
}

func truncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen-3] + "..."
}

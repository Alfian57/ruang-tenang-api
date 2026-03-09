package service

import (
	"context"
	"encoding/json"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PushService struct {
	pushRepo       *repository.PushSubscriptionRepository
	vapidPublicKey string
	vapidPrivKey   string
	vapidContact   string
	logger         *zap.Logger
}

func NewPushService(
	pushRepo *repository.PushSubscriptionRepository,
	vapidPublicKey, vapidPrivKey, vapidContact string,
) *PushService {
	logger, _ := zap.NewProduction()
	return &PushService{
		pushRepo:       pushRepo,
		vapidPublicKey: vapidPublicKey,
		vapidPrivKey:   vapidPrivKey,
		vapidContact:   vapidContact,
		logger:         logger,
	}
}

func (s *PushService) GetVAPIDPublicKey() string {
	return s.vapidPublicKey
}

// Subscribe registers or updates a push subscription for a user.
func (s *PushService) Subscribe(ctx context.Context, userID uint, endpoint, p256dh, auth string) error {
	sub := &model.PushSubscription{
		ID:       uuid.New(),
		UserID:   userID,
		Endpoint: endpoint,
		P256dh:   p256dh,
		Auth:     auth,
	}
	return s.pushRepo.Upsert(ctx, sub)
}

// Unsubscribe removes a push subscription by endpoint.
func (s *PushService) Unsubscribe(ctx context.Context, userID uint, endpoint string) error {
	return s.pushRepo.DeleteByEndpoint(ctx, endpoint, userID)
}

// PushPayload is the JSON structure sent to the browser push service.
type PushPayload struct {
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Icon  string            `json:"icon,omitempty"`
	Badge string            `json:"badge,omitempty"`
	Tag   string            `json:"tag,omitempty"`
	Data  map[string]string `json:"data,omitempty"`
}

// SendToUser sends a push notification to all subscriptions of a user.
// Best-effort: logs errors but does not fail the caller.
func (s *PushService) SendToUser(ctx context.Context, userID uint, payload PushPayload) {
	if s.vapidPublicKey == "" || s.vapidPrivKey == "" {
		return // VAPID not configured — skip silently
	}

	subs, err := s.pushRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Warn("push: failed to get subscriptions", zap.Uint("user_id", userID), zap.Error(err))
		return
	}

	if len(subs) == 0 {
		return
	}

	if payload.Icon == "" {
		payload.Icon = "/favicon/android-chrome-192x192.png"
	}
	if payload.Badge == "" {
		payload.Badge = "/favicon/favicon-32x32.png"
	}

	body, err := json.Marshal(payload)
	if err != nil {
		s.logger.Warn("push: failed to marshal payload", zap.Error(err))
		return
	}

	for _, sub := range subs {
		wpSub := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				P256dh: sub.P256dh,
				Auth:   sub.Auth,
			},
		}
		resp, err := webpush.SendNotification(body, wpSub, &webpush.Options{
			VAPIDPublicKey:  s.vapidPublicKey,
			VAPIDPrivateKey: s.vapidPrivKey,
			Subscriber:      s.vapidContact,
			TTL:             60,
		})
		if err != nil {
			s.logger.Warn("push: send failed", zap.String("endpoint", sub.Endpoint), zap.Error(err))
			// Remove stale/invalid subscription
			if resp != nil && (resp.StatusCode == 404 || resp.StatusCode == 410) {
				_ = s.pushRepo.DeleteByID(ctx, sub.ID.String())
			}
			continue
		}
		resp.Body.Close()
	}
}

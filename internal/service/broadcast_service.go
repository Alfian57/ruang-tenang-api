package service

import (
	"context"
	"encoding/json"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BroadcastService struct {
	broadcastRepo  *repository.BroadcastNotificationRepository
	pushRepo       *repository.PushSubscriptionRepository
	vapidPublicKey string
	vapidPrivKey   string
	vapidContact   string
	logger         *zap.Logger
	stopCh         chan struct{}
}

func NewBroadcastService(
	broadcastRepo *repository.BroadcastNotificationRepository,
	pushRepo *repository.PushSubscriptionRepository,
	vapidPublicKey, vapidPrivKey, vapidContact string,
) *BroadcastService {
	logger, _ := zap.NewProduction()
	return &BroadcastService{
		broadcastRepo:  broadcastRepo,
		pushRepo:       pushRepo,
		vapidPublicKey: vapidPublicKey,
		vapidPrivKey:   vapidPrivKey,
		vapidContact:   vapidContact,
		logger:         logger,
		stopCh:         make(chan struct{}),
	}
}

func (s *BroadcastService) Create(ctx context.Context, userID uint, title, body, icon, url string, scheduledAt *time.Time) (*model.BroadcastNotification, error) {
	status := model.BroadcastStatusDraft
	if scheduledAt != nil {
		status = model.BroadcastStatusScheduled
	}

	b := &model.BroadcastNotification{
		ID:          uuid.New(),
		Title:       title,
		Body:        body,
		Icon:        icon,
		URL:         url,
		Status:      status,
		ScheduledAt: scheduledAt,
		CreatedBy:   userID,
	}

	if err := s.broadcastRepo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *BroadcastService) GetByID(ctx context.Context, id string) (*model.BroadcastNotification, error) {
	return s.broadcastRepo.GetByID(ctx, id)
}

func (s *BroadcastService) GetAll(ctx context.Context, page, limit int, search string) ([]model.BroadcastNotification, int64, error) {
	return s.broadcastRepo.GetAll(ctx, page, limit, search)
}

func (s *BroadcastService) Update(ctx context.Context, id string, title, body, icon, url string, scheduledAt *time.Time) (*model.BroadcastNotification, error) {
	b, err := s.broadcastRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if b.Status == model.BroadcastStatusSent || b.Status == model.BroadcastStatusSending {
		return nil, ErrBroadcastAlreadySent
	}

	b.Title = title
	b.Body = body
	b.Icon = icon
	b.URL = url
	if scheduledAt != nil {
		b.ScheduledAt = scheduledAt
		b.Status = model.BroadcastStatusScheduled
	} else {
		b.ScheduledAt = nil
		b.Status = model.BroadcastStatusDraft
	}

	if err := s.broadcastRepo.Update(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *BroadcastService) Delete(ctx context.Context, id string) error {
	b, err := s.broadcastRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if b.Status == model.BroadcastStatusSending {
		return ErrBroadcastSending
	}
	return s.broadcastRepo.Delete(ctx, id)
}

func (s *BroadcastService) Cancel(ctx context.Context, id string) (*model.BroadcastNotification, error) {
	b, err := s.broadcastRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b.Status != model.BroadcastStatusScheduled && b.Status != model.BroadcastStatusDraft {
		return nil, ErrBroadcastCannotCancel
	}
	b.Status = model.BroadcastStatusCancelled
	if err := s.broadcastRepo.Update(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *BroadcastService) SendNow(ctx context.Context, id string) (*model.BroadcastNotification, error) {
	b, err := s.broadcastRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b.Status == model.BroadcastStatusSent || b.Status == model.BroadcastStatusSending {
		return nil, ErrBroadcastAlreadySent
	}

	go s.executeBroadcast(b)

	return b, nil
}

func (s *BroadcastService) executeBroadcast(b *model.BroadcastNotification) {
	ctx := context.Background()

	b.Status = model.BroadcastStatusSending
	_ = s.broadcastRepo.Update(ctx, b)

	subs, err := s.pushRepo.GetAll(ctx)
	if err != nil {
		s.logger.Error("broadcast: failed to get subscriptions", zap.Error(err))
		b.Status = model.BroadcastStatusDraft
		_ = s.broadcastRepo.Update(ctx, b)
		return
	}

	payload := PushPayload{
		Title: b.Title,
		Body:  b.Body,
		Icon:  b.Icon,
		Tag:   "broadcast-" + b.ID.String(),
	}
	if b.URL != "" {
		payload.Data = map[string]string{"url": b.URL}
	}
	if payload.Icon == "" {
		payload.Icon = "/favicon/android-chrome-192x192.png"
	}

	body, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("broadcast: failed to marshal payload", zap.Error(err))
		return
	}

	sentCount := 0
	failedCount := 0

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
			TTL:             86400,
		})
		if err != nil {
			failedCount++
			if resp != nil && (resp.StatusCode == 404 || resp.StatusCode == 410) {
				_ = s.pushRepo.DeleteByID(ctx, sub.ID.String())
			}
			continue
		}
		resp.Body.Close()
		sentCount++
	}

	now := time.Now()
	b.Status = model.BroadcastStatusSent
	b.SentAt = &now
	b.SentCount = sentCount
	b.FailedCount = failedCount
	_ = s.broadcastRepo.Update(ctx, b)

	s.logger.Info("broadcast: sent", zap.String("id", b.ID.String()), zap.Int("sent", sentCount), zap.Int("failed", failedCount))
}

// StartScheduler starts a background goroutine that checks for scheduled broadcasts.
func (s *BroadcastService) StartScheduler() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.processScheduledBroadcasts()
			case <-s.stopCh:
				return
			}
		}
	}()
	s.logger.Info("broadcast: scheduler started")
}

// StopScheduler stops the scheduler goroutine.
func (s *BroadcastService) StopScheduler() {
	close(s.stopCh)
}

func (s *BroadcastService) processScheduledBroadcasts() {
	ctx := context.Background()
	broadcasts, err := s.broadcastRepo.GetScheduledDue(ctx)
	if err != nil {
		s.logger.Warn("broadcast: failed to get scheduled broadcasts", zap.Error(err))
		return
	}

	for i := range broadcasts {
		s.executeBroadcast(&broadcasts[i])
	}
}

// Sentinel errors
var (
	ErrBroadcastAlreadySent  = &BroadcastError{"broadcast sudah terkirim"}
	ErrBroadcastSending      = &BroadcastError{"broadcast sedang dikirim"}
	ErrBroadcastCannotCancel = &BroadcastError{"broadcast tidak dapat dibatalkan"}
)

type BroadcastError struct {
	Message string
}

func (e *BroadcastError) Error() string {
	return e.Message
}

package application

import (
	"context"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

// BadgeRepository defines the badge data access needed by the gamification feature.
// This port interface breaks the circular dependency between gamification and badge.
type BadgeRepository interface {
	GetRecentlyEarnedBadges(ctx context.Context, userID uint, since time.Time) ([]model.UserBadge, error)
	GetUserBadgeCount(ctx context.Context, userID uint) (int, error)
}

// FeatureUnlockRepository defines the feature unlock data access needed by the gamification feature.
// This port interface breaks the circular dependency between gamification and feature_unlock.
type FeatureUnlockRepository interface {
	GetFeaturesByLevel(ctx context.Context, level int) ([]model.FeatureDefinition, error)
}

// UserRepository defines the user data access needed by the gamification feature.
// This port interface breaks the circular dependency between gamification and auth.
type UserRepository interface {
	FindByID(ctx context.Context, id uint) (*model.User, error)
}

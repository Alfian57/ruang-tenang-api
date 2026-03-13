package application

import (
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

// LevelConfigRepository defines the level config data access needed by the feature_unlock feature.
// This port interface breaks the circular dependency between feature_unlock and gamification.
type LevelConfigRepository interface {
	GetAll(ctx context.Context) ([]model.LevelConfig, error)
	GetLevelByExp(ctx context.Context, exp int64) (*model.LevelConfig, error)
}

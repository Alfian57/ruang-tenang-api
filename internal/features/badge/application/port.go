package application

import (
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

// LevelConfigRepository defines the level config data access needed by the badge feature.
// This port interface breaks the circular dependency between badge and gamification.
type LevelConfigRepository interface {
	GetLevelByExp(ctx context.Context, exp int64) (*model.LevelConfig, error)
}

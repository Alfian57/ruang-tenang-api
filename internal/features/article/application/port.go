package application

import (
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

// ModerationService defines the moderation operations needed by the article feature.
// This port interface breaks the circular dependency between article and moderation.
type ModerationService interface {
	ModerateNewArticle(ctx context.Context, article *model.Article) (*dto.AIModerationResult, error)
}

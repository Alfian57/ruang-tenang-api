package application

import (
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

// ArticleRepository defines the article data access needed by the moderation feature.
// This port interface breaks the circular dependency between moderation and article.
type ArticleRepository interface {
	FindByID(ctx context.Context, id uint) (*model.Article, error)
	UpdateStatus(ctx context.Context, id uint, status model.ArticleStatus) error
}

// ForumRepository defines the forum data access needed by the moderation feature.
type ForumRepository interface {
	// Add methods as needed by moderation
}

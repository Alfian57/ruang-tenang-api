package handler

import (
	articleinfra "github.com/Alfian57/ruang-tenang-api/internal/features/article/infrastructure"
	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	foruminfra "github.com/Alfian57/ruang-tenang-api/internal/features/forum/infrastructure"
	journalapp "github.com/Alfian57/ruang-tenang-api/internal/features/journal/application"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/cache"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db             *gorm.DB
	userRepo       *authinfra.UserRepository
	articleRepo    *articleinfra.ArticleRepository
	forumRepo      foruminfra.ForumRepository
	cacheService   *cache.CacheService
	journalService *journalapp.JournalService
}

func NewAdminHandler(
	db *gorm.DB,
	userRepo *authinfra.UserRepository,
	articleRepo *articleinfra.ArticleRepository,
	forumRepo foruminfra.ForumRepository,
	cacheService *cache.CacheService,
	journalService *journalapp.JournalService,
) *AdminHandler {
	return &AdminHandler{
		db:             db,
		userRepo:       userRepo,
		articleRepo:    articleRepo,
		forumRepo:      forumRepo,
		cacheService:   cacheService,
		journalService: journalService,
	}
}

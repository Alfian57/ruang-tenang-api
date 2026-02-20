package handler

import (
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db             *gorm.DB
	userRepo       *repository.UserRepository
	articleRepo    *repository.ArticleRepository
	forumRepo      repository.ForumRepository
	cacheService   *service.CacheService
	journalService *service.JournalService
}

func NewAdminHandler(
	db *gorm.DB,
	userRepo *repository.UserRepository,
	articleRepo *repository.ArticleRepository,
	forumRepo repository.ForumRepository,
	cacheService *service.CacheService,
	journalService *service.JournalService,
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

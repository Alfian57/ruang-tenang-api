package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
)

type ArticleService struct {
	articleRepo           *repository.ArticleRepository
	categoryRepo          *repository.ArticleCategoryRepository
	gamificationService   *GamificationService
	contentContextService *ContentContextService
	cacheService          *CacheService
}

func NewArticleService(articleRepo *repository.ArticleRepository, categoryRepo *repository.ArticleCategoryRepository, gamificationService *GamificationService, contentContextService *ContentContextService, cacheService *CacheService) *ArticleService {
	return &ArticleService{
		articleRepo:           articleRepo,
		categoryRepo:          categoryRepo,
		gamificationService:   gamificationService,
		contentContextService: contentContextService,
		cacheService:          cacheService,
	}
}

// GetPublishedArticles returns only published articles for public view
func (s *ArticleService) GetPublishedArticles(ctx context.Context, params *dto.ArticleQueryParams) ([]dto.ArticleListDTO, int64, error) {
	articles, total, err := s.articleRepo.FindPublished(ctx, params.CategoryID, params.Search, params.Page, params.Limit)
	if err != nil {
		return nil, 0, err
	}

	return s.articlesToListDTO(ctx, articles), total, nil
}

// GetArticles returns articles with optional filters (for admin)
func (s *ArticleService) GetArticles(ctx context.Context, params *dto.ArticleQueryParams) ([]dto.ArticleListDTO, int64, error) {
	articles, total, err := s.articleRepo.FindAll(ctx, params.CategoryID, params.Search, params.Page, params.Limit, params.Status, params.UserID)
	if err != nil {
		return nil, 0, err
	}

	return s.articlesToListDTO(ctx, articles), total, nil
}

// GetUserArticles returns articles owned by a specific user
func (s *ArticleService) GetUserArticles(ctx context.Context, userID uint, page, limit int) ([]dto.ArticleListDTO, int64, error) {
	articles, total, err := s.articleRepo.FindByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	return s.articlesToListDTO(ctx, articles), total, nil
}

func (s *ArticleService) articlesToListDTO(ctx context.Context, articles []model.Article) []dto.ArticleListDTO {
	var result []dto.ArticleListDTO
	for _, article := range articles {
		excerpt := article.Content
		if len(excerpt) > 150 {
			excerpt = excerpt[:150] + "..."
		}
		// Remove HTML tags from excerpt
		excerpt = strings.ReplaceAll(excerpt, "<p>", "")
		excerpt = strings.ReplaceAll(excerpt, "</p>", " ")

		item := dto.ArticleListDTO{
			ID:         article.ID,
			Title:      article.Title,
			Thumbnail:  article.Thumbnail,
			Excerpt:    excerpt,
			CategoryID: article.ArticleCategoryID,
			Category: dto.ArticleCategoryDTO{
				ID:        article.Category.ID,
				Name:      article.Category.Name,
				CreatedAt: article.Category.CreatedAt,
			},
			UserID:           article.UserID,
			Status:           string(article.Status),
			ModerationStatus: string(article.ModerationStatus),
			CreatedAt:        article.CreatedAt,
		}

		if article.Author != nil {
			item.Author = &dto.ArticleAuthorDTO{
				ID:   article.Author.ID,
				Name: article.Author.Name,
			}
		}

		result = append(result, item)
	}

	return result
}

func (s *ArticleService) GetArticleByID(ctx context.Context, id uint) (*dto.ArticleDTO, error) {
	article, err := s.articleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &dto.ArticleDTO{
		ID:         article.ID,
		Title:      article.Title,
		Thumbnail:  article.Thumbnail,
		Content:    article.Content,
		CategoryID: article.ArticleCategoryID,
		Category: dto.ArticleCategoryDTO{
			ID:        article.Category.ID,
			Name:      article.Category.Name,
			CreatedAt: article.Category.CreatedAt,
		},
		UserID:           article.UserID,
		Status:           string(article.Status),
		ModerationStatus: string(article.ModerationStatus),
		CreatedAt:        article.CreatedAt,
		UpdatedAt:        article.UpdatedAt,
	}

	if article.Author != nil {
		result.Author = &dto.ArticleAuthorDTO{
			ID:   article.Author.ID,
			Name: article.Author.Name,
		}
	}

	return result, nil
}

// GetPublishedArticleByID returns an article only if it's published
func (s *ArticleService) GetPublishedArticleByID(ctx context.Context, id uint) (*dto.ArticleDTO, error) {
	article, err := s.GetArticleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if article.Status != string(model.ArticleStatusPublished) {
		return nil, errors.New("article not found")
	}

	return article, nil
}

func (s *ArticleService) GetCategories(ctx context.Context) ([]dto.ArticleCategoryDTO, error) {
	// Check cache first
	if s.cacheService != nil {
		if cached := s.cacheService.Get(CacheKeyArticleCategories); cached != nil {
			return cached.([]dto.ArticleCategoryDTO), nil
		}
	}

	categories, err := s.categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var result []dto.ArticleCategoryDTO
	for _, category := range categories {
		result = append(result, dto.ArticleCategoryDTO{
			ID:          category.ID,
			Name:        category.Name,
			Description: category.Description,
			CreatedAt:   category.CreatedAt,
		})
	}

	// Store in cache
	if s.cacheService != nil {
		s.cacheService.SetWithTTL(CacheKeyArticleCategories, result, s.cacheService.CategoryTTL)
	}
	return result, nil
}

// CreateArticle creates a new article (admin)
func (s *ArticleService) CreateArticle(ctx context.Context, article *model.Article) error {
	return s.articleRepo.Create(ctx, article)
}

// CreateUserArticle creates a new article for a user (pending moderation)
func (s *ArticleService) CreateUserArticle(ctx context.Context, userID uint, req *dto.CreateUserArticleRequest) (*model.Article, error) {
	article := &model.Article{
		Title:             req.Title,
		Thumbnail:         req.Thumbnail,
		Content:           req.Content,
		ArticleCategoryID: req.CategoryID,
		UserID:            userID,
		Status:            model.ArticleStatusDraft,
		IsUserGenerated:   true,
	}

	if err := s.articleRepo.Create(ctx, article); err != nil {
		return nil, err
	}

	// Notify content context cache
	if s.contentContextService != nil {
		article, _ = s.articleRepo.FindByID(ctx, article.ID) // Reload with category
		s.contentContextService.NotifyArticleChange(ctx, article)
	}

	// Note: EXP is awarded when article is approved by moderator, not on creation

	return article, nil
}

// UpdateUserArticle updates an article owned by the user
func (s *ArticleService) UpdateUserArticle(ctx context.Context, userID uint, articleID uint, req *dto.UpdateUserArticleRequest) (*model.Article, error) {
	article, err := s.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if article.UserID != userID {
		return nil, errors.New("not authorized to update this article")
	}

	// Check if blocked
	if article.Status == model.ArticleStatusBlocked {
		return nil, errors.New("cannot update blocked article")
	}

	article.Title = req.Title
	article.Thumbnail = req.Thumbnail
	article.Content = req.Content
	article.ArticleCategoryID = req.CategoryID

	if err := s.articleRepo.Update(ctx, article); err != nil {
		return nil, err
	}

	// Notify content context cache
	if s.contentContextService != nil {
		s.contentContextService.NotifyArticleChange(ctx, article)
	}

	return article, nil
}

// DeleteUserArticle deletes an article owned by the user
func (s *ArticleService) DeleteUserArticle(ctx context.Context, userID uint, articleID uint) error {
	article, err := s.articleRepo.FindByID(ctx, articleID)
	if err != nil {
		return err
	}

	// Check ownership
	if article.UserID != userID {
		return errors.New("not authorized to delete this article")
	}

	err = s.articleRepo.Delete(ctx, articleID)
	if err == nil && s.contentContextService != nil {
		s.contentContextService.NotifyArticleDelete(ctx, articleID)
	}
	return err
}

// BlockArticle blocks an article (admin only)
func (s *ArticleService) BlockArticle(ctx context.Context, articleID uint) error {
	return s.articleRepo.UpdateStatus(ctx, articleID, model.ArticleStatusBlocked)
}

// UnblockArticle unblocks an article (admin only)
func (s *ArticleService) UnblockArticle(ctx context.Context, articleID uint) error {
	return s.articleRepo.UpdateStatus(ctx, articleID, model.ArticleStatusPublished)
}

func (s *ArticleService) CreateCategory(ctx context.Context, category *model.ArticleCategory) error {
	return s.categoryRepo.Create(ctx, category)
}

func (s *ArticleService) UpdateArticle(ctx context.Context, article *model.Article) error {
	return s.articleRepo.Update(ctx, article)
}

func (s *ArticleService) DeleteArticle(ctx context.Context, articleID uint) error {
	return s.articleRepo.Delete(ctx, articleID)
}
